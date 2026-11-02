param(
    [string]$Version,
    [switch]$Release
)
$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Resolve-Path (Join-Path $scriptDir "..\..")

function Resolve-InnoCompiler {
    param([string]$CompilerPath)
    if (-not [string]::IsNullOrWhiteSpace($CompilerPath)) {
        if (-not (Test-Path -LiteralPath $CompilerPath -PathType Leaf)) {
            throw "ISCC_PATH points to a missing compiler: $CompilerPath"
        }
        return (Resolve-Path -LiteralPath $CompilerPath).Path
    }
    $command = Get-Command iscc -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($command) { return $command.Source }
    foreach ($directory in @(${env:ProgramFiles(x86)}, $env:ProgramFiles, $env:ProgramW6432, (Join-Path $env:LOCALAPPDATA "Programs"))) {
        if ([string]::IsNullOrWhiteSpace($directory)) { continue }
        $candidate = Join-Path $directory "Inno Setup 6\ISCC.exe"
        if (Test-Path -LiteralPath $candidate -PathType Leaf) { return $candidate }
    }
    throw "Inno Setup 6 compiler (ISCC.exe) not found. Install Inno Setup 6 or set ISCC_PATH to its full executable path for the runner account."
}

$compilerPath = $env:ISCC_PATH
if ([string]::IsNullOrWhiteSpace($compilerPath)) {
    $compilerPath = [Environment]::GetEnvironmentVariable("ISCC_PATH", "Machine")
}
$iscc = Resolve-InnoCompiler $compilerPath
Write-Host "Inno Setup compiler: $iscc"

Push-Location $projectRoot
try {
    $metadata = python (Join-Path $scriptDir 'version.py') | ConvertFrom-Json
    if ($LASTEXITCODE -ne 0) { throw "Version resolution failed" }
    if ($Version -and $Version -cne $metadata.version) {
        throw "Installer version must match the selected ref: $($metadata.version)"
    }
    if ($Release -and -not $metadata.release) {
        throw "Release installer builds require a numeric tag at HEAD"
    }
    & $iscc "/DMyAppVersion=$($metadata.version)" "/DMyAppNumericVersion=$($metadata.marketing_version)" (Join-Path $projectRoot "build\installer\musicalc.iss")
    if ($LASTEXITCODE -ne 0) { throw "Inno Setup failed with exit code $LASTEXITCODE" }
} finally {
    Pop-Location
}
