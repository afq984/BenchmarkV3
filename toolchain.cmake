set(CMAKE_SYSTEM_NAME Linux)
set(CMAKE_SYSTEM_PROCESSOR aarch64)

set(triple aarch64-linux)

set(CMAKE_C_COMPILER_TARGET ${triple})
set(CMAKE_CXX_COMPILER_TARGET ${triple})

set(clang_bin ${CMAKE_CURRENT_LIST_DIR}/clang-bin)

set(CMAKE_C_COMPILER ${clang_bin}/clang)
set(CMAKE_CXX_COMPILER ${clang_bin}/clang++)
set(CMAKE_RANLIB ${clang_bin}/llvm-ranlib)
set(CMAKE_AR ${clang_bin}/llvm-ar)

set(CMAKE_SYSROOT ${CMAKE_CURRENT_LIST_DIR}/sysroot)
add_link_options("-fuse-ld=lld")

set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)
set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE ONLY)
