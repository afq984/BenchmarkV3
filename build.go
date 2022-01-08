package main

import (
	"context"
	_ "embed"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/pflag"
)

//go:embed toolchain.cmake
var toolchainContents []byte

const toolchainFileName = "toolchain.cmake"

var quick bool

func init() {
	pflag.BoolVar(&quick, "quick", false, "do a quick build to check configuration")
}

func run(name string, args ...string) error {
	cmd := exec.Command(
		name,
		args...,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Println("running:", cmd)

	err := cmd.Run()
	if err != nil {
		log.Println("command failed:", cmd)
		log.Println(err)
	}
	return err
}

func Build(c *Config) (time.Duration, error) {
	var buildDir string
	var err error

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	buildDir, err = ioutil.TempDir(".", "build.*")
	if err != nil {
		log.Println("failed to create build directory")
		return 0, err
	}
	defer func() {
		log.Println("cleaning up", buildDir)
		os.RemoveAll(buildDir)
	}()
	log.Println("using build directory:", buildDir)

	buildAbsPath := func(rel string) string {
		p, err := filepath.Abs(filepath.Join(buildDir, rel))
		if err != nil {
			panic(err)
		}
		return p
	}

	// parallel download and extract
	{
		errMux := sync.Mutex{}
		wg := sync.WaitGroup{}
		wg.Add(len(c.Packages()))

		for _, p := range c.Packages() {
			go func(p Package) {
				defer wg.Done()

				lerr := p.SetUp(ctx, buildDir)
				if lerr != nil {
					errMux.Lock()
					defer errMux.Unlock()

					err = lerr
					cancel()
				}
			}(p)
		}

		wg.Wait()
		if err != nil {
			return 0, err
		}
	}

	log.Println("writing", toolchainFileName)
	err = os.WriteFile(filepath.Join(buildDir, toolchainFileName), toolchainContents, 0644)
	if err != nil {
		log.Println("failed to write toolchain.cmake:", err)
		return 0, err
	}

	// symlink so we can have a static cmake toolchain file
	err = os.Symlink(c.ClangBin, filepath.Join(buildDir, "clang-bin"))
	if err != nil {
		log.Println("cannot create symlink for clang-bin")
		return 0, err
	}

	err = run(
		buildAbsPath(c.Cmake()),
		"-B", buildAbsPath("out"),
		"-S", buildAbsPath(c.LLVMSrc),
		"-G", "Ninja",
		"-DCMAKE_MAKE_PROGRAM="+buildAbsPath(c.Ninja()),
		"-DCMAKE_TOOLCHAIN_FILE="+buildAbsPath(toolchainFileName),
		"-DCMAKE_BUILD_TYPE=Release", // debug builds sadly take too much disk space
		"-DLLVM_ENABLE_PROJECTS=",
		"-DLLVM_TABLEGEN="+buildAbsPath(c.LLVMTblgen()),
		"-DLLVM_TARGETS_TO_BUILD=X86",
	)
	if err != nil {
		return 0, err
	}

	t0 := time.Now()
	buildTarget := "llc"
	if quick {
		buildTarget = "llvm-cxxfilt"
	}
	err = run(
		buildAbsPath(c.Ninja()),
		"-C", buildAbsPath("out"),
		buildTarget,
	)
	t1 := time.Now()
	if err != nil {
		return 0, err
	}

	dt := t1.Sub(t0)

	return dt, nil
}
