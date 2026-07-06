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
	// unix names (clang++ -> clang -> clang-22 and ld.lld -> lld are symlinks)
	"bin/clang", "bin/clang++", "bin/clang-22",
	"bin/lld", "bin/ld.lld",
	"bin/llvm-tblgen", "bin/llvm-ar", "bin/llvm-ranlib",
	// windows names (separate .exe copies; no clang-22)
	"bin/clang.exe", "bin/clang++.exe",
	"bin/lld.exe", "bin/ld.lld.exe",
	"bin/llvm-tblgen.exe", "bin/llvm-ar.exe", "bin/llvm-ranlib.exe",
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
	// Interpreter path inside an extracted python-build-standalone archive; its
	// tarballs contain a top-level python/ directory. (Windows instead uses the
	// PSF embeddable, whose python.exe sits at the archive root.)
	standalonePython = "python/bin/python3"
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

	// Bundled Python 3 for LLVM's cmake, so no system python3 is required.
	PythonPkg: &Archive{
		URL:    "https://github.com/astral-sh/python-build-standalone/releases/download/20260623/cpython-3.13.14+20260623-x86_64-unknown-linux-gnu-install_only_stripped.tar.gz",
		Sha256: "459ed79967acc207bef2ff5124dac35d74d5108528e37b15395d14e2922f2c92",
	},
	Python: standalonePython,

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

	// Bundled Python 3 for LLVM's cmake, so no system python3 is required.
	PythonPkg: &Archive{
		URL:    "https://github.com/astral-sh/python-build-standalone/releases/download/20260623/cpython-3.13.14+20260623-aarch64-unknown-linux-gnu-install_only_stripped.tar.gz",
		Sha256: "e931d7a393f54902503f8745ceb35420e7dd50a067e78e5f45c71404f7a15b30",
	},
	Python: standalonePython,

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

	// Bundled Python 3 for LLVM's cmake, so no system python3 is required. macOS
	// ships no usable python3 (only the Command Line Tools provide one), so this
	// lets the benchmark run on a stock machine.
	PythonPkg: &Archive{
		URL:    "https://github.com/astral-sh/python-build-standalone/releases/download/20260623/cpython-3.13.14+20260623-aarch64-apple-darwin-install_only_stripped.tar.gz",
		Sha256: "795a5aeeb050f00aa8a2214d779bad9f1b9113edb6923317a80c042a11a087d7",
	},
	Python: standalonePython,

	LLVMSrc:        defaultLLVMSrc,
	LLVMSrcArchive: defaultLLVMSrcArchive,

	DebianSysrootArchive: defaultDebianSysrootArchive,
}

var WindowsAmd64Config = &Config{
	ClangBin: "clang+llvm-22.1.8-x86_64-pc-windows-msvc/bin",
	ClangPkg: &Archive{
		URL:    "https://github.com/llvm/llvm-project/releases/download/llvmorg-22.1.8/clang+llvm-22.1.8-x86_64-pc-windows-msvc.tar.xz",
		Sha256: "d96c2cc1736f4eb7fa43cb9bbdf56d93551a9ae0a9aadb9c99c3c3b2b712a234",
		Keep:   toolchainKeep,
	},

	CmakeBin: "cmake-3.22.1-windows-x86_64/bin",
	CmakePkg: &Archive{
		URL:    "https://github.com/Kitware/CMake/releases/download/v3.22.1/cmake-3.22.1-windows-x86_64.zip",
		Sha256: "35fbbb7d9ffa491834bbc79cdfefc6c360088a3c9bf55c29d111a5afa04cdca3",
	},

	NinjaBin: defaultNinjaBin,
	NinjaPkg: &Archive{
		URL:    "https://github.com/ninja-build/ninja/releases/download/v1.12.1/ninja-win.zip",
		Sha256: "f550fec705b6d6ff58f2db3c374c2277a37691678d6aba463adcbb129108467a",
	},

	// LLVM's cmake requires Python 3; provide the embeddable interpreter since
	// Windows has none by default.
	PythonPkg: &Archive{
		URL:       "https://www.python.org/ftp/python/3.13.1/python-3.13.1-embed-amd64.zip",
		Sha256:    "7b7923ff0183a8b8fca90f6047184b419b108cb437f75fc1c002f9d2f8bcec16",
		ExtractTo: "python",
	},
	Python: "python/python.exe",

	// A Windows host running a clang that targets aarch64-linux is neither MSVC
	// nor MinGW, so LLVM's cmake falls back to config.guess (a shell script it
	// cannot run) and fails to detect the host arch. Set the host triple directly.
	CmakeArgs: []string{"-DLLVM_HOST_TRIPLE=x86_64-pc-windows-msvc"},

	LLVMSrc:        defaultLLVMSrc,
	LLVMSrcArchive: defaultLLVMSrcArchive,

	DebianSysrootArchive: defaultDebianSysrootArchive,
}

var WindowsArm64Config = &Config{
	ClangBin: "clang+llvm-22.1.8-aarch64-pc-windows-msvc/bin",
	ClangPkg: &Archive{
		URL:    "https://github.com/llvm/llvm-project/releases/download/llvmorg-22.1.8/clang+llvm-22.1.8-aarch64-pc-windows-msvc.tar.xz",
		Sha256: "de718c58ebbc5f61d58c17b90457fcf42983bc2c4a4aba3e010d108713bfd7f1",
		Keep:   toolchainKeep,
	},

	// cmake only ships Windows ARM64 binaries from 3.24.0 on. cmake is not part
	// of the timed build, so the version needn't match the other configs' 3.22.1.
	CmakeBin: "cmake-3.24.0-windows-arm64/bin",
	CmakePkg: &Archive{
		URL:    "https://github.com/Kitware/CMake/releases/download/v3.24.0/cmake-3.24.0-windows-arm64.zip",
		Sha256: "552c3c922460a05b1ee14b560750d2deb7a16cf55ad780a0b81bce81fe38e93d",
	},

	NinjaBin: defaultNinjaBin,
	NinjaPkg: &Archive{
		URL:    "https://github.com/ninja-build/ninja/releases/download/v1.12.1/ninja-winarm64.zip",
		Sha256: "79c96a50e0deafec212cfa85aa57c6b74003f52d9d1673ddcd1eab1c958c5900",
	},

	PythonPkg: &Archive{
		URL:       "https://www.python.org/ftp/python/3.13.1/python-3.13.1-embed-arm64.zip",
		Sha256:    "ae8561bf958f77c68cb6c44ced983e5267fe965a7e4168f41ec2291350b81d55",
		ExtractTo: "python",
	},
	Python: "python/python.exe",

	CmakeArgs: []string{"-DLLVM_HOST_TRIPLE=aarch64-pc-windows-msvc"},

	LLVMSrc:        defaultLLVMSrc,
	LLVMSrcArchive: defaultLLVMSrcArchive,

	DebianSysrootArchive: defaultDebianSysrootArchive,
}

var configs = map[string]*Config{
	"linux-amd64":   LinuxAmd64Config,
	"linux-arm64":   LinuxArm64Config,
	"macos-arm64":   MacOSArm64Config,
	"windows-amd64": WindowsAmd64Config,
	"windows-arm64": WindowsArm64Config,
}
