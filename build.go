package main

import (
	_ "embed"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

//go:embed toolchain.cmake
var toolchainContents []byte

const toolchainFileName = "toolchain.cmake"

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

func buildAbsPath(rel string) string {
	p, err := filepath.Abs(filepath.Join(buildDir, rel))
	if err != nil {
		panic(err)
	}
	return p
}

func Build(c *Config) (err error) {
	log.Println("cleaning build directory")
	err = os.RemoveAll(buildDir)
	if err != nil {
		log.Println("faild to cleanup build directory")
		return err
	}

	// parallel download and extract
	{
		errMux := sync.Mutex{}
		wg := sync.WaitGroup{}
		wg.Add(len(c.Archives()))

		for _, a := range c.Archives() {
			go func(a *Archive) {
				defer wg.Done()

				lerr := a.DownloadAndExtract()
				if lerr != nil {
					errMux.Lock()
					defer errMux.Unlock()

					err = lerr
				}
			}(a)
		}

		wg.Wait()
		if err != nil {
			return err
		}
	}

	log.Println("writing", toolchainFileName)
	err = os.WriteFile(filepath.Join(buildDir, toolchainFileName), toolchainContents, 0644)
	if err != nil {
		log.Println("failed to write toolchain.cmake:", err)
		return err
	}

	// symlink executables so we can have a static cmake toolchain file
	err = os.Symlink(c.Clang(), filepath.Join(buildDir, "clang"))
	if err != nil {
		log.Println("cannot create symlink for clang")
		return err
	}
	err = os.Symlink(c.Clangpp(), filepath.Join(buildDir, "clang++"))
	if err != nil {
		log.Println("cannot create symlink for clang++")
		return err
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
		return err
	}

	t0 := time.Now()
	err = run(
		buildAbsPath(c.Ninja()),
		"-C", buildAbsPath("out"),
		"llc",
	)
	t1 := time.Now()
	if err != nil {
		return err
	}

	dt := t1.Sub(t0)
	log.Println("build completed in", dt)

	return nil
}
