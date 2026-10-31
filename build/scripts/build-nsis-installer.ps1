param(
    [string]$MakeNSIS = "makensis"
)
$ErrorActionPreference = "Stop"

function Assert-ExecutableVersion {
    param([string]$Path, [string]$ExpectedVersion)
    # An explicit pipe captures stdout from the Windows GUI executable.
    $lines = @(& $Path --version | Out-String -Stream)
    if ($LASTEXITCODE -ne 0 -or $lines.Count -ne 1 -or $lines[0] -cne $ExpectedVersion) {
        throw "Executable version mismatch: expected '$ExpectedVersion', received '$($lines -join ' ')'."
    }
}

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "../..")).Path
$executable = Join-Path $projectRoot "build/dist/musicalc.exe"
$compiler = Get-Command $MakeNSIS -CommandType Application -ErrorAction SilentlyContinue
if ($null -eq $compiler) {
    throw "NSIS 3 makensis.exe not found. Install NSIS 3 or provide -MakeNSIS with its full path."
}

Push-Location $projectRoot
try {
    $metadata = python (Join-Path $PSScriptRoot "version.py") | ConvertFrom-Json
    if ($LASTEXITCODE -ne 0) { throw "Version resolution failed" }
    & (Join-Path $PSScriptRoot "build-win.ps1")
    $buildInfo = (& go version -m $executable) -join "`n"
    if ($LASTEXITCODE -ne 0) { throw "Cannot read Go build information from the Windows executable" }
    if (-not $buildInfo.Contains("vcs.revision=$($metadata.commit)")) {
        throw "Executable does not match the selected Git revision. Rebuild with build/scripts/build-win.ps1."
    }
    $binary = [IO.File]::ReadAllBytes($executable)
    if ($binary.Length -lt 64) { throw "Invalid Windows executable" }
    $peOffset = [BitConverter]::ToInt32($binary, 60)
    if ($peOffset -lt 64 -or $peOffset -gt $binary.Length - 26 -or
        [BitConverter]::ToUInt32($binary, $peOffset) -ne 0x4550 -or
        [BitConverter]::ToUInt16($binary, $peOffset + 4) -ne 0x8664) {
        throw "NSIS packaging requires an AMD64 Windows executable."
    }
    Assert-ExecutableVersion -Path $executable -ExpectedVersion $metadata.version
    $outputDirectory = Join-Path $projectRoot "build/dist/installer"
    New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null
    $output = Join-Path $outputDirectory "musicalc_$($metadata.version)_windows_amd64_nsis_setup.exe"
    $arguments = @(
        "/NOCONFIG", "/WX", "/V3", "/INPUTCHARSET", "UTF8",
        "/DAppVersion=$($metadata.version)",
        "/DNumericVersion=$($metadata.marketing_version).0",
        "/DSourceRoot=$projectRoot",
        "/DAppExe=$executable",
        "/DInstallerOutput=$output",
        (Join-Path $projectRoot "build/installer/musicalc.nsi")
    )
    & $compiler.Source @arguments
    if ($LASTEXITCODE -ne 0) { throw "NSIS failed with exit code $LASTEXITCODE" }
    if (-not (Test-Path -LiteralPath $output -PathType Leaf)) { throw "NSIS did not produce the installer" }
    Write-Host "Built NSIS installer: $output"
} finally {
    Pop-Location
}
