# Download and run the latest BenchmarkV3 prebuilt for this Windows machine.
#
#   irm https://afq984.github.io/BenchmarkV3/run.ps1 | iex
#
# To pass flags (e.g. a quick build), invoke it as a scriptblock so the
# arguments reach the script:
#
#   & ([scriptblock]::Create((irm https://afq984.github.io/BenchmarkV3/run.ps1))) --quick
$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"  # Invoke-WebRequest is far faster without it

# OSArchitecture reports the native architecture even from an emulated process.
$arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
switch ($arch) {
    "X64"   { $variant = "windows-amd64" }
    "Arm64" { $variant = "windows-arm64" }
    default {
        Write-Error "BenchmarkV3: unsupported architecture: $arch (supported: x64, arm64)"
        return
    }
}

$url = "https://github.com/afq984/BenchmarkV3/releases/download/latest/BenchmarkV3-$variant.exe"
$exe = Join-Path ([System.IO.Path]::GetTempPath()) "BenchmarkV3-$variant.exe"

Write-Host "BenchmarkV3: downloading $variant ..."
Invoke-WebRequest -Uri $url -OutFile $exe

# No `exit`: this may be dot-sourced via `iex`, and exit would close the session.
try {
    & $exe @args
} finally {
    Remove-Item -Force -ErrorAction SilentlyContinue -Path $exe
}
