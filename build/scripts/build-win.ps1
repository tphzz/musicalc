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
# Removes the variable for $null. PowerShell passes $null to .NET string parameters as "", and
# only Windows PowerShell's .NET Framework deletes a variable set to ""; PowerShell 7 keeps it
# empty, and Git rejects an empty GIT_DIR.
function Set-ProcessVariable([string]$Name, $Value) {
    if ($null -eq $Value) {
        Remove-Item "Env:$Name" -ErrorAction SilentlyContinue
    } else {
        Set-Item "Env:$Name" $Value
    }
}
Push-Location $projectRoot
try {
    foreach ($name in $environment.Keys) {
        Set-ProcessVariable $name $environment[$name]
    }
    # Keep Go 1.26 VCS stamps on this checkout for submodules and worktrees.
    $env:GIT_DIR = & git rev-parse --absolute-git-dir
    if ($LASTEXITCODE -ne 0) { throw "Cannot resolve Git directory" }
    $env:GIT_WORK_TREE = & git rev-parse --show-toplevel
    if ($LASTEXITCODE -ne 0) { throw "Cannot resolve Git worktree" }
    $metadata = python (Join-Path $scriptDir 'version.py') | ConvertFrom-Json
    if ($LASTEXITCODE -ne 0) { throw "Version resolution failed" }
    # Zig requires the GUI subsystem explicitly at the external link step.
    go build -trimpath -ldflags="-s -w -H=windowsgui -extldflags=-Wl,--subsystem,windows -X main.version=$($metadata.version)" -o (Join-Path $distDir $filename)
    if ($LASTEXITCODE -ne 0) { throw "Windows build failed with exit code $LASTEXITCODE" }
    python (Join-Path $scriptDir 'windows_binary.py') (Join-Path $distDir $filename) $Arch
    if ($LASTEXITCODE -ne 0) { throw "Windows binary verification failed with exit code $LASTEXITCODE" }
} finally {
    foreach ($name in $previousEnvironment.Keys) {
        Set-ProcessVariable $name $previousEnvironment[$name]
    }
    Pop-Location
}
