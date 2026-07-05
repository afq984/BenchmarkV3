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

# Windows sets PROCESSOR_ARCHITECTURE to the process architecture, and (only under
# emulation) PROCESSOR_ARCHITEW6432 to the native one. Prefer the native value so an
# emulated x64 shell on ARM64 still benchmarks the native binary. This is far more
# reliable across PowerShell versions than [RuntimeInformation]::OSArchitecture.
if ($env:PROCESSOR_ARCHITEW6432) { $arch = $env:PROCESSOR_ARCHITEW6432 } else { $arch = $env:PROCESSOR_ARCHITECTURE }
switch ($arch) {
    "AMD64" { $variant = "windows-amd64" }
    "ARM64" { $variant = "windows-arm64" }
    default {
        Write-Error "BenchmarkV3: unsupported architecture: '$arch' (supported: AMD64, ARM64)"
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
