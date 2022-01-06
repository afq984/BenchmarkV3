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
		case "amd64", "arm64":
			return "macos-amd64"
		}
	}
	log.Fatalf("Unknown GOOS and GOARCH combination: %s, %s", runtime.GOOS, runtime.GOARCH)
	panic("unreachable")
}
