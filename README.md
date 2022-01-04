# BenchmarkV3

Benchmark how fast your system compiles C++ projects.
Created to answer the question of whether it is worth investing a new dev machine.

This benchmark measures the time to build LLVM's [llc] using clang, cmake and ninja.
It uses prebuilt toolchains and build for the same target sysroot
so hopefully it benchmarks the same thing on different systems.

## Run from source

```
git clone https://github.com/afq984/BenchmarkV3
cd BenchmarkV3
go build .
./BenchmarkV3 -c <config>
```

Substitute `<config>` to one of the below:
*   `linux-amd64` - Works on ubuntu 20.04 or up-to-date linux systems
*   `linux-amd64-ubuntu1604` - Try this one on older linux systems
*   `macos-amd64` - Use this on macOS. Should also work on M1 through [Rosetta].

## Run prebuilt

TODO

## Other BenchmarkV3 flags

*   `-quick`: Do a quick build instead. This measures the build time of `llvm-cxxfilt`
    instead of `llc` which has only about 1/8 build targets.

[llc]: https://llvm.org/docs/CommandGuide/llc.html
[Rosetta]: https://developer.apple.com/documentation/apple-silicon/about-the-rosetta-translation-environment
