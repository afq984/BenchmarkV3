package main

import "strings"

const downloadDir = "dl"

// keepPaths builds an Archive.Keep filter. A pattern ending in "/" keeps that
// directory and everything under it; any other pattern keeps an exact path.
func keepPaths(patterns ...string) func(string) bool {
	return func(p string) bool {
		for _, pat := range patterns {
			if strings.HasSuffix(pat, "/") {
				if strings.HasPrefix(p, pat) {
					return true
				}
			} else if p == pat {
				return true
			}
		}
		return false
	}
}

// The build runs only a few of the toolchain's (statically linked) tools, so we
// extract just those plus clang's resource headers instead of the full ~12 GB.
var toolchainKeep = keepPaths(
	"bin/clang", "bin/clang++", "bin/clang-22",
	"bin/lld", "bin/ld.lld",
	"bin/llvm-tblgen", "bin/llvm-ar", "bin/llvm-ranlib",
	"lib/clang/",
)

// Building llc needs only the llvm project and the shared cmake / third-party
// modules it references; the monorepo's other projects are skipped.
var llvmSrcKeep = keepPaths("llvm/", "cmake/", "third-party/")

const (
	defaultNinjaBin = "."
	// The LLVM monorepo source tree; cmake is pointed at the llvm/ subdir.
	// Since LLVM 20 the per-subproject source tarballs (llvm-X.src.tar.xz) are
	// no longer published, so we use the full llvm-project source tarball.
	defaultLLVMSrc = "llvm-project-22.1.8.src/llvm"
)

var defaultLLVMSrcArchive = &Archive{
	URL:    "https://github.com/llvm/llvm-project/releases/download/llvmorg-22.1.8/llvm-project-22.1.8.src.tar.xz",
	Sha256: "922f1817a0df7b1489272d18134ee0087a8b068828f87ac63b9861b1a9965888",
	Keep:   llvmSrcKeep,
}

var defaultDebianSysrootArchive = &Archive{
	URL:       "https://commondatastorage.googleapis.com/chrome-linux-sysroot/toolchain/2befe8ce3e88be6080e4fb7e6d412278ea6a7625/debian_sid_arm64_sysroot.tar.xz",
	Sha256:    "e4389eab2fe363f3fbdfa4d3ce9d94457d78fd2c0e62171a7534867623eadc90",
	ExtractTo: "sysroot",
}

var LinuxAmd64Config = &Config{
	ClangBin: "LLVM-22.1.8-Linux-X64/bin",
	ClangPkg: &Archive{
		URL:    "https://github.com/llvm/llvm-project/releases/download/llvmorg-22.1.8/LLVM-22.1.8-Linux-X64.tar.xz",
		Sha256: "df0e1ecf16caf3489a272a5eea4eec9b0d82878f6477fa309504f918a0006384",
		Keep:   toolchainKeep,
	},

	CmakeBin: "cmake-3.22.1-linux-x86_64/bin",
	CmakePkg: &Archive{
		URL:    "https://github.com/Kitware/CMake/releases/download/v3.22.1/cmake-3.22.1-linux-x86_64.tar.gz",
		Sha256: "73565c72355c6652e9db149249af36bcab44d9d478c5546fd926e69ad6b43640",
	},

	NinjaBin: defaultNinjaBin,
	NinjaPkg: &Archive{
		URL:    "https://github.com/ninja-build/ninja/releases/download/v1.12.1/ninja-linux.zip",
		Sha256: "6f98805688d19672bd699fbbfa2c2cf0fc054ac3df1f0e6a47664d963d530255",
	},

	LLVMSrc:        defaultLLVMSrc,
	LLVMSrcArchive: defaultLLVMSrcArchive,

	DebianSysrootArchive: defaultDebianSysrootArchive,
}

var LinuxArm64Config = &Config{
	ClangBin: "LLVM-22.1.8-Linux-ARM64/bin",
	ClangPkg: &Archive{
		URL:    "https://github.com/llvm/llvm-project/releases/download/llvmorg-22.1.8/LLVM-22.1.8-Linux-ARM64.tar.xz",
		Sha256: "805efad2bb91cb4967fa569e0881d10c0f69c04461cf671cccbae19f547acc34",
		Keep:   toolchainKeep,
	},

	CmakeBin: "cmake-3.22.1-linux-aarch64/bin",
	CmakePkg: &Archive{
		URL:    "https://github.com/Kitware/CMake/releases/download/v3.22.1/cmake-3.22.1-linux-aarch64.tar.gz",
		Sha256: "601443375aa1a48a1a076bda7e3cca73af88400463e166fffc3e1da3ce03540b",
	},

	NinjaBin: defaultNinjaBin,
	NinjaPkg: &Archive{
		URL:    "https://github.com/ninja-build/ninja/releases/download/v1.12.1/ninja-linux-aarch64.zip",
		Sha256: "5c25c6570b0155e95fce5918cb95f1ad9870df5768653afe128db822301a05a1",
	},

	LLVMSrc:        defaultLLVMSrc,
	LLVMSrcArchive: defaultLLVMSrcArchive,

	DebianSysrootArchive: defaultDebianSysrootArchive,
}

var MacOSArm64Config = &Config{
	ClangBin: "LLVM-22.1.8-macOS-ARM64/bin",
	ClangPkg: &Archive{
		URL:    "https://github.com/llvm/llvm-project/releases/download/llvmorg-22.1.8/LLVM-22.1.8-macOS-ARM64.tar.xz",
		Sha256: "f260f4f7c0d430828a81ae8a3826a1d63fc0963ec2459489308cc23b1f7eab4f",
		Keep:   toolchainKeep,
	},

	CmakeBin: "cmake-3.22.1-macos-universal/CMake.app/Contents/bin",
	CmakePkg: &Archive{
		URL:    "https://github.com/Kitware/CMake/releases/download/v3.22.1/cmake-3.22.1-macos-universal.tar.gz",
		Sha256: "9ba46ce69d524f5bcdf98076a6b01f727604fb31cf9005ec03dea1cf16da9514",
	},

	NinjaBin: defaultNinjaBin,
	NinjaPkg: &Archive{
		URL:    "https://github.com/ninja-build/ninja/releases/download/v1.12.1/ninja-mac.zip",
		Sha256: "89a287444b5b3e98f88a945afa50ce937b8ffd1dcc59c555ad9b1baf855298c9",
	},

	LLVMSrc:        defaultLLVMSrc,
	LLVMSrcArchive: defaultLLVMSrcArchive,

	DebianSysrootArchive: defaultDebianSysrootArchive,
}

var configs = map[string]*Config{
	"linux-amd64": LinuxAmd64Config,
	"linux-arm64": LinuxArm64Config,
	"macos-arm64": MacOSArm64Config,
}
