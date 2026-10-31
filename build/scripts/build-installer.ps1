param(
    [string]$Version,
    [switch]$Release
)
$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Resolve-Path (Join-Path $scriptDir "..\..")
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
    iscc "/DMyAppVersion=$($metadata.version)" "/DMyAppNumericVersion=$($metadata.marketing_version)" (Join-Path $projectRoot "build\installer\musicalc.iss")
    if ($LASTEXITCODE -ne 0) { throw "Inno Setup failed with exit code $LASTEXITCODE" }
} finally {
    Pop-Location
}
