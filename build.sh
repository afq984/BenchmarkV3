cmake-3.22.1-linux-x86_64/bin/cmake -B build \
    -S llvm-13.0.0.src \
    -DCMAKE_TOOLCHAIN_FILE=$PWD/cmake.toolchain \
    -G Ninja \
    -DCMAKE_MAKE_PROGRAM=$PWD/ninja \
    -DLLVM_ENABLE_PROJECTS= \
    -DLLVM_TABLEGEN=$PWD/clang+llvm-13.0.0-x86_64-linux-gnu-ubuntu-20.04/bin/llvm-tblgen \
    -DLLVM_TARGETS_TO_BUILD=X86 \
    -DCMAKE_BUILD_TYPE=Release
