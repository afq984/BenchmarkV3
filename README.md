# BenchmarkV3

Benchmark how fast your system compiles C++ projects.
Created to answer the question of whether it is worth investing a new dev machine.

This benchmark measures the time to build LLVM's [llc] using [clang], [cmake] and [ninja].
It uses prebuilt toolchains and build for the same target sysroot
so hopefully it benchmarks the same thing on different systems.

## Run prebuilt

Run the latest build for your platform:

```
curl -fsSL https://afq984.github.io/BenchmarkV3/run.sh | sh
```

Append `-s -- --quick` for a quick build:

```
curl -fsSL https://afq984.github.io/BenchmarkV3/run.sh | sh -s -- --quick
```

On Windows, in PowerShell:

```
irm https://afq984.github.io/BenchmarkV3/run.ps1 | iex
```

For a quick build, invoke it as a scriptblock so the flag passes through:

```
& ([scriptblock]::Create((irm https://afq984.github.io/BenchmarkV3/run.ps1))) --quick
```

Or download the binary yourself from the [release page](https://github.com/afq984/BenchmarkV3/releases/tag/latest):

```
# V is one of: linux-amd64, linux-arm64, macos-arm64
V=linux-amd64
curl -o BenchmarkV3 -L https://github.com/afq984/BenchmarkV3/releases/download/latest/BenchmarkV3-$V
chmod +x BenchmarkV3
./BenchmarkV3
```

## Run from source

```
git clone https://github.com/afq984/BenchmarkV3
cd BenchmarkV3
go build .
./BenchmarkV3
```

## BenchmarkV3 command line flags

*   `--quick`: Do a quick build instead. This measures the build time of `llvm-cxxfilt`
    instead of `llc` which has only about 1/8 build targets.

*   `--detect`: Detect the system only. Does not actually run the benchmark.

*   `-c <config>`: Use a specific config:
    *   `auto` - Auto detect the config to use (default)
    *   `linux-amd64` - For x86-64 Linux systems (requires glibc 2.34+, e.g. Ubuntu 22.04 / Debian 12 / RHEL 9 or newer)
    *   `linux-arm64` - For ARM64 Linux systems (same glibc 2.34+ requirement)
    *   `macos-arm64` - For Apple Silicon Macs (requires macOS 14 Sonoma or newer)

## FAQ

*   Why is there no `macos-amd64` (Intel Mac) config?

    Since [LLVM 20](https://github.com/llvm/llvm-project/releases/tag/llvmorg-20.1.0) the official releases no longer ship a native x86-64 macOS toolchain — only Apple Silicon (`arm64`).
    This benchmark now uses the native `arm64` build, so Intel Macs are no longer supported.

*   What are the OS requirements?

    The prebuilt toolchain is produced by LLVM's release CI on Ubuntu 22.04 / macOS 14, so:

    *   Linux (`linux-amd64`, `linux-arm64`): glibc 2.34 and a GCC 12-era `libstdc++`
        — Ubuntu 22.04, Debian 12, RHEL/Rocky/Alma 9, Fedora 35 or newer.
        Older distros (RHEL 8, CentOS 7, Ubuntu 20.04) are not supported.
    *   macOS (`macos-arm64`): macOS 14 (Sonoma) or newer.

    Unlike older versions, the LLVM toolchain no longer links `libtinfo` / `libncurses`,
    so no ncurses package is required.

*   Do I need `libxml2` installed on linux?

    No. The toolchain's `lld` linker lists `libxml2.so.2` as a dependency, but
    only uses it for Windows output — it never calls it when linking for Linux.
    On hosts that lack `libxml2.so.2` (minimal installs, or rolling-release
    distros that renamed it to `libxml2.so.16`), the benchmark automatically
    generates a tiny stub `libxml2.so.2` in its build directory so linking works.
    No package or symlink is required.

[clang]: https://clang.llvm.org/
[cmake]: https://cmake.org/
[ninja]: https://ninja-build.org/
[llc]: https://llvm.org/docs/CommandGuide/llc.html
