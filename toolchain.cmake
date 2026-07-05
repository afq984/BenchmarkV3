set(CMAKE_SYSTEM_NAME Linux)
set(CMAKE_SYSTEM_PROCESSOR aarch64)

set(triple aarch64-linux-gnu)

set(CMAKE_C_COMPILER_TARGET ${triple})
set(CMAKE_CXX_COMPILER_TARGET ${triple})

# Host tools are .exe on Windows; the Linux/macOS toolchains have no suffix.
if(CMAKE_HOST_WIN32)
  set(host_exe ".exe")
else()
  set(host_exe "")
endif()

set(clang_bin ${CMAKE_CURRENT_LIST_DIR}/clang-bin)

set(CMAKE_C_COMPILER ${clang_bin}/clang${host_exe})
set(CMAKE_CXX_COMPILER ${clang_bin}/clang++${host_exe})
set(CMAKE_RANLIB ${clang_bin}/llvm-ranlib${host_exe})
set(CMAKE_AR ${clang_bin}/llvm-ar${host_exe})

set(CMAKE_SYSROOT ${CMAKE_CURRENT_LIST_DIR}/sysroot)
add_link_options("-fuse-ld=lld")

# LLVM's lib/Support/CMakeLists.txt links the target's Unix system libraries
# (dl, rt, m, ...) only when the BUILD host is Unix (elseif(CMAKE_HOST_UNIX)), so
# a Windows host cross-compiling to Linux never links them and fails on symbols
# like dladdr. Add them back for that case.
if(CMAKE_HOST_WIN32)
  add_link_options(-ldl -lrt -lm)
endif()

set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)
set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE ONLY)
