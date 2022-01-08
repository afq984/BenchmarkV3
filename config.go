package main

import "path/filepath"

type Config struct {
	ClangBin             string
	ClangArchive         *Archive
	CmakeBin             string
	CmakeArchive         *Archive
	NinjaBin             string
	NinjaArchive         *Archive
	LLVMSrc              string
	LLVMSrcArchive       *Archive
	DebianSysrootArchive *Archive
}

func (c *Config) Archives() []*Archive {
	return []*Archive{
		c.ClangArchive,
		c.CmakeArchive,
		c.NinjaArchive,
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
