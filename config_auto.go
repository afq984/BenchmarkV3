package main

import (
	"log"
	"runtime"
)

func autoselectConfig() string {
	if runtime.GOOS == "linux" {
		switch runtime.GOARCH {
		case "amd64":
			return "linux-amd64"
		case "arm64":
			return "linux-arm64"
		}
	}
	if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "arm64":
			return "macos-arm64"
		case "amd64":
			// Intel Macs are no longer supported: LLVM 22 ships no native
			// x86_64 macOS toolchain (only arm64). Use an Apple Silicon Mac.
			log.Fatalf("Intel Macs are not supported by this benchmark version (no native x86_64 macOS LLVM build); run on Apple Silicon")
		}
	}
	if runtime.GOOS == "windows" {
		switch runtime.GOARCH {
		case "amd64":
			return "windows-amd64"
		case "arm64":
			return "windows-arm64"
		}
	}
	log.Fatalf("Unknown GOOS and GOARCH combination: %s, %s", runtime.GOOS, runtime.GOARCH)
	panic("unreachable")
}
