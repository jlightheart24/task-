# Build Go core as a Windows DLL (C-ABI) for WinUI 3.

$Root = Split-Path -Parent $MyInvocation.MyCommand.Path | Split-Path -Parent
$OutDir = Join-Path $Root "build\windows"

New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

# Requires cgo and a C toolchain (e.g. MinGW) installed on Windows.
# Produces: taskppcore.dll and taskppcore.h

$Env:CGO_ENABLED = "1"
$Env:GOOS = "windows"
$Env:GOARCH = "amd64"

$Target = Join-Path $OutDir "taskppcore.dll"

Write-Host "Building $Target"

go build -buildmode=c-shared -o $Target ./cmd/corelib
