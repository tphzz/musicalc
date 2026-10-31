$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Resolve-Path (Join-Path $scriptDir "..\..")
$distDir = Join-Path $projectRoot "build\dist"
New-Item -ItemType Directory -Force $distDir | Out-Null
$env:CGO_ENABLED = "1"
$env:CC = "clang --target=x86_64-w64-windows-gnu"
$env:CXX = "clang++ --target=x86_64-w64-windows-gnu"
$env:CGO_CFLAGS = "-O3 -march=x86-64-v3"
$env:CGO_LDFLAGS = "-O3"
Push-Location $projectRoot
try {
    $metadata = python (Join-Path $scriptDir 'version.py') | ConvertFrom-Json
    if ($LASTEXITCODE -ne 0) { throw "Version resolution failed" }
    go build -ldflags="-s -w -H=windowsgui -extld=x86_64-w64-mingw32-gcc -X main.version=$($metadata.version)" -o (Join-Path $distDir "musicalc.exe")
    if ($LASTEXITCODE -ne 0) { throw "Windows build failed with exit code $LASTEXITCODE" }
} finally {
    Pop-Location
}
