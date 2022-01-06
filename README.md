# BenchmarkV3

Benchmark how fast your system compiles C++ projects.
Created to answer the question of whether it is worth investing a new dev machine.

This benchmark measures the time to build LLVM's [llc] using clang, cmake and ninja.
It uses prebuilt toolchains and build for the same target sysroot
so hopefully it benchmarks the same thing on different systems.

## Run prebuilt

Go to the [release page](https://github.com/afq984/BenchmarkV3/releases/tag/latest)
to download the latest build and run directly.

```
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

*   `-quick`: Do a quick build instead. This measures the build time of `llvm-cxxfilt`
    instead of `llc` which has only about 1/8 build targets.

*   `-detect`: Detect the system only. Does not actually run the benchmark.

*   `-c <config>`: Use a specific config:
    *   `auto` - Auto detect the config to use (default)
    *   `linux-amd64` - Works on ubuntu 20.04 or up-to-date linux systems
    *   `linux-amd64-ubuntu1604` - Try this one on older linux systems
    *   `macos-amd64` - Use this on macOS. Should also work on M1 through [Rosetta].

[llc]: https://llvm.org/docs/CommandGuide/llc.html
[Rosetta]: https://developer.apple.com/documentation/apple-silicon/about-the-rosetta-translation-environment
