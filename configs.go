package main

const downloadDir = "dl"

const (
	defaultNinjaBin = "."
	defaultLLVMSrc  = "llvm-13.0.0.src"
)

var defaultLLVMSrcArchive = &Archive{
	URL:    "https://github.com/llvm/llvm-project/releases/download/llvmorg-13.0.0/llvm-13.0.0.src.tar.xz",
	Sha256: "408d11708643ea826f519ff79761fcdfc12d641a2510229eec459e72f8163020",
}

var defaultDebianSysrootArchive = &Archive{
	URL:       "https://commondatastorage.googleapis.com/chrome-linux-sysroot/toolchain/2befe8ce3e88be6080e4fb7e6d412278ea6a7625/debian_sid_arm64_sysroot.tar.xz",
	Sha256:    "e4389eab2fe363f3fbdfa4d3ce9d94457d78fd2c0e62171a7534867623eadc90",
	ExtractTo: "sysroot",
}

var LinuxAmd64Config = &Config{
	ClangBin: "clang+llvm-13.0.0-x86_64-linux-gnu-ubuntu-20.04/bin",
	ClangPkg: &Archive{
		URL:    "https://github.com/llvm/llvm-project/releases/download/llvmorg-13.0.0/clang+llvm-13.0.0-x86_64-linux-gnu-ubuntu-20.04.tar.xz",
		Sha256: "2c2fb857af97f41a5032e9ecadf7f78d3eff389a5cd3c9ec620d24f134ceb3c8",
	},

	CmakeBin: "cmake-3.22.1-linux-x86_64/bin",
	CmakePkg: &Archive{
		URL:    "https://github.com/Kitware/CMake/releases/download/v3.22.1/cmake-3.22.1-linux-x86_64.tar.gz",
		Sha256: "73565c72355c6652e9db149249af36bcab44d9d478c5546fd926e69ad6b43640",
	},

	NinjaBin: defaultNinjaBin,
	NinjaPkg: &Archive{
		URL:    "https://github.com/ninja-build/ninja/releases/download/v1.10.2/ninja-linux.zip",
		Sha256: "763464859c7ef2ea3a0a10f4df40d2025d3bb9438fcb1228404640410c0ec22d",
	},

	LLVMSrc:        defaultLLVMSrc,
	LLVMSrcArchive: defaultLLVMSrcArchive,

	DebianSysrootArchive: defaultDebianSysrootArchive,
}

var LinuxAmd64Ubuntu1604Config = &Config{
	ClangBin: "clang+llvm-13.0.0-x86_64-linux-gnu-ubuntu-16.04/bin",
	ClangPkg: &Archive{
		URL:    "https://github.com/llvm/llvm-project/releases/download/llvmorg-13.0.0/clang+llvm-13.0.0-x86_64-linux-gnu-ubuntu-16.04.tar.xz",
		Sha256: "76d0bf002ede7a893f69d9ad2c4e101d15a8f4186fbfe24e74856c8449acd7c1",
	},

	CmakeBin: LinuxAmd64Config.CmakeBin,
	CmakePkg: LinuxAmd64Config.CmakePkg,

	NinjaBin: defaultNinjaBin,
	NinjaPkg: LinuxAmd64Config.NinjaPkg,

	LLVMSrc:        defaultLLVMSrc,
	LLVMSrcArchive: defaultLLVMSrcArchive,

	DebianSysrootArchive: defaultDebianSysrootArchive,
}

var LinuxArm64Config = &Config{
	ClangBin: "clang+llvm-13.0.0-aarch64-linux-gnu/bin",
	ClangPkg: &Archive{
		URL:    "https://github.com/llvm/llvm-project/releases/download/llvmorg-13.0.0/clang+llvm-13.0.0-aarch64-linux-gnu.tar.xz",
		Sha256: "968d65d2593850ee9b37fcda074fb7641529bd45d2f976af6c8197de3c22612f",
	},

	CmakeBin: "cmake-3.22.1-linux-aarch64/bin",
	CmakePkg: &Archive{
		URL:    "https://github.com/Kitware/CMake/releases/download/v3.22.1/cmake-3.22.1-linux-aarch64.tar.gz",
		Sha256: "601443375aa1a48a1a076bda7e3cca73af88400463e166fffc3e1da3ce03540b",
	},

	NinjaBin: defaultNinjaBin,
	NinjaPkg: &System{
		Name: "ninja",
	},

	LLVMSrc:        defaultLLVMSrc,
	LLVMSrcArchive: defaultLLVMSrcArchive,

	DebianSysrootArchive: defaultDebianSysrootArchive,
}

var MacOSAmd64Config = &Config{
	ClangBin: "clang+llvm-13.0.0-x86_64-apple-darwin/bin",
	ClangPkg: &Archive{
		URL:    "https://github.com/llvm/llvm-project/releases/download/llvmorg-13.0.0/clang+llvm-13.0.0-x86_64-apple-darwin.tar.xz",
		Sha256: "d051234eca1db1f5e4bc08c64937c879c7098900f7a0370f3ceb7544816a8b09",
	},

	CmakeBin: "cmake-3.22.1-macos-universal/CMake.app/Contents/bin",
	CmakePkg: &Archive{
		URL:    "https://github.com/Kitware/CMake/releases/download/v3.22.1/cmake-3.22.1-macos-universal.tar.gz",
		Sha256: "9ba46ce69d524f5bcdf98076a6b01f727604fb31cf9005ec03dea1cf16da9514",
	},

	NinjaBin: defaultNinjaBin,
	NinjaPkg: &Archive{
		URL:    "https://github.com/ninja-build/ninja/releases/download/v1.10.2/ninja-mac.zip",
		Sha256: "6fa359f491fac7e5185273c6421a000eea6a2f0febf0ac03ac900bd4d80ed2a5",
	},

	LLVMSrc:        defaultLLVMSrc,
	LLVMSrcArchive: defaultLLVMSrcArchive,

	DebianSysrootArchive: defaultDebianSysrootArchive,
}

var configs = map[string]*Config{
	"linux-amd64":            LinuxAmd64Config,
	"linux-amd64-ubuntu1604": LinuxAmd64Ubuntu1604Config,
	"linux-arm64":            LinuxArm64Config,
	"macos-amd64":            MacOSAmd64Config,
}
