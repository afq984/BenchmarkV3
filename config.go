package main

import "path/filepath"

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
}

func (c *Config) Packages() []Package {
	return []Package{
		c.ClangPkg,
		c.CmakePkg,
		c.NinjaPkg,
		c.LLVMSrcArchive,
		c.DebianSysrootArchive,
	}
}

// llvm-tblgen path relative to buildDir
func (c *Config) LLVMTblgen() string {
	return filepath.Join(c.ClangBin, "llvm-tblgen")
}

// cmake path relative to buildDir
func (c *Config) Cmake() string {
	return filepath.Join(c.CmakeBin, "cmake")
}

// ninja path relative to buildDir
func (c *Config) Ninja() string {
	return filepath.Join(c.NinjaBin, "ninja")
}
