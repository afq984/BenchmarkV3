package main

import (
	"path/filepath"
	"runtime"
)

// exe appends the host executable suffix (".exe" on Windows) to a tool path.
func exe(path string) string {
	if runtime.GOOS == "windows" {
		return path + ".exe"
	}
	return path
}

type Config struct {
	ClangBin             string
	ClangPkg             Package
	CmakeBin             string
	CmakePkg             Package
	NinjaBin             string
	NinjaPkg             Package
	LLVMSrc              string
	LLVMSrcArchive       *Archive
	DebianSysrootArchive *Archive

	// PythonPkg provides the Python 3 interpreter that LLVM's cmake requires;
	// Python is its path relative to buildDir, passed as -DPython3_EXECUTABLE.
	PythonPkg Package
	Python    string

	// CmakeArgs are extra -D flags appended to the cmake configure command,
	// for host-specific quirks (e.g. LLVM_HOST_TRIPLE on Windows).
	CmakeArgs []string
}

func (c *Config) Packages() []Package {
	pkgs := []Package{
		c.ClangPkg,
		c.CmakePkg,
		c.NinjaPkg,
		c.LLVMSrcArchive,
		c.DebianSysrootArchive,
	}
	if c.PythonPkg != nil {
		pkgs = append(pkgs, c.PythonPkg)
	}
	return pkgs
}

// llvm-tblgen path relative to buildDir
func (c *Config) LLVMTblgen() string {
	return exe(filepath.Join(c.ClangBin, "llvm-tblgen"))
}

// cmake path relative to buildDir
func (c *Config) Cmake() string {
	return exe(filepath.Join(c.CmakeBin, "cmake"))
}

// ninja path relative to buildDir
func (c *Config) Ninja() string {
	return exe(filepath.Join(c.NinjaBin, "ninja"))
}
