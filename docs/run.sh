#!/bin/sh
# Download and run the latest BenchmarkV3 prebuilt for this platform.
#
#   curl -fsSL https://afq984.github.io/BenchmarkV3/run.sh | sh
#
# Pass flags after "sh -s --", e.g. a quick build:
#
#   curl -fsSL https://afq984.github.io/BenchmarkV3/run.sh | sh -s -- --quick
set -eu

os=$(uname -s)
arch=$(uname -m)
case "$os $arch" in
	"Linux x86_64")  variant=linux-amd64 ;;
	"Linux aarch64") variant=linux-arm64 ;;
	"Darwin arm64")  variant=macos-arm64 ;;
	*)
		echo "BenchmarkV3: unsupported platform: $os $arch" >&2
		echo "supported: Linux x86_64, Linux aarch64, macOS Apple Silicon" >&2
		exit 1
		;;
esac

url="https://github.com/afq984/BenchmarkV3/releases/download/latest/BenchmarkV3-$variant"
bin=$(mktemp)
trap 'rm -f "$bin"' EXIT

echo "BenchmarkV3: downloading $variant ..." >&2
curl -fL --progress-bar -o "$bin" "$url"
chmod +x "$bin"

set +e
"$bin" "$@"
status=$?
rm -f "$bin"
exit "$status"
