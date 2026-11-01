param(
    [ValidateSet("amd64", "arm64")]
    [string]$Arch = "amd64"
)
$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Resolve-Path (Join-Path $scriptDir "..\..")
if (-not (Get-Command zig -ErrorAction SilentlyContinue)) {
    throw "Missing Zig; install 0.17.0 or 0.16.0 and add it to PATH as described in README-BUILD.md."
}
$zigVersion = & zig version
if ($LASTEXITCODE -ne 0 -or $zigVersion -notin @("0.16.0", "0.17.0")) {
    throw "Expected Zig 0.16.0 or 0.17.0; found '$zigVersion'."
}
$distDir = Join-Path $projectRoot "build\dist"
New-Item -ItemType Directory -Force $distDir | Out-Null
$environment = @{
    CGO_ENABLED = "1"
    GOOS = "windows"
    GOARCH = $Arch
    GOAMD64 = "v3"
    ZIG_GLOBAL_CACHE_DIR = Join-Path $projectRoot "build\.cache\zig-global"
    ZIG_LOCAL_CACHE_DIR = Join-Path $projectRoot "build\.cache\zig-local"
    CGO_LDFLAGS = "-O3"
    GIT_DIR = $null
    GIT_WORK_TREE = $null
}
if ($Arch -eq "amd64") {
    $target = "x86_64-windows-gnu"
    $environment.CGO_CFLAGS = "-O3 -mcpu=x86_64_v3"
    $filename = "musicalc.exe"
} else {
    $target = "aarch64-windows-gnu"
    $environment.CGO_CFLAGS = "-O3 -mcpu=oryon_1 -fomit-frame-pointer"
    $filename = "musicalc_arm64.exe"
}
$environment.CC = "zig cc -target $target"
$environment.CXX = "zig c++ -target $target"
$previousEnvironment = @{}
foreach ($name in $environment.Keys) {
    $previousEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, "Process")
}
Push-Location $projectRoot
try {
    foreach ($name in $environment.Keys) {
        [Environment]::SetEnvironmentVariable($name, $environment[$name], "Process")
    }
    # Keep Go 1.26 VCS stamps on this checkout for submodules and worktrees.
    $env:GIT_DIR = & git rev-parse --absolute-git-dir
    if ($LASTEXITCODE -ne 0) { throw "Cannot resolve Git directory" }
    $env:GIT_WORK_TREE = & git rev-parse --show-toplevel
    if ($LASTEXITCODE -ne 0) { throw "Cannot resolve Git worktree" }
    $metadata = python (Join-Path $scriptDir 'version.py') | ConvertFrom-Json
    if ($LASTEXITCODE -ne 0) { throw "Version resolution failed" }
    go build -trimpath -ldflags="-s -w -H=windowsgui -X main.version=$($metadata.version)" -o (Join-Path $distDir $filename)
    if ($LASTEXITCODE -ne 0) { throw "Windows build failed with exit code $LASTEXITCODE" }
} finally {
    foreach ($name in $previousEnvironment.Keys) {
        [Environment]::SetEnvironmentVariable($name, $previousEnvironment[$name], "Process")
    }
    Pop-Location
}
