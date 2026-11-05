param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern('^[0-9]+\.[0-9]+\.[0-9]+$')]
    [string]$Version,
    [string]$PythonPath
)
$ErrorActionPreference = "Stop"

# Select the all-users installation independently of PATH and Actions tool caches.
$pythonCandidates = @($PythonPath)
if (-not $PythonPath) {
    $directoryName = "Python" + (($Version.Split('.'))[0..1] -join '')
    $directoryNames = @($directoryName, "$directoryName-arm64")
    if ($env:RUNNER_ARCH -eq 'ARM64' -or $env:PROCESSOR_ARCHITECTURE -eq 'ARM64') {
        $directoryNames = @("$directoryName-arm64", $directoryName)
    }
    $pythonCandidates = @(foreach ($name in $directoryNames) {
        Join-Path $env:ProgramFiles "$name/python.exe"
    })
}
$PythonPath = $pythonCandidates | Where-Object { Test-Path -LiteralPath $_ -PathType Leaf } | Select-Object -First 1
if (-not $PythonPath) {
    $locations = "'" + ($pythonCandidates -join "' or '") + "'"
    throw "Install Python $Version (64-bit) for all users at $locations. See build/docs/README-INSTALLER.md#windows-python-installation."
}
$PythonPath = (Resolve-Path -LiteralPath $PythonPath).Path
# Single-quoted Python literals survive Windows PowerShell's native argument quoting.
$probe = "import json, struct, sys; print(json.dumps({'version': '.'.join(map(str, sys.version_info[:3])), 'bits': struct.calcsize('P') * 8}))"
try {
    $output = & $PythonPath -I -c $probe
} catch {
    throw "System Python could not run at '$PythonPath': $_"
}
if ($LASTEXITCODE -ne 0) {
    throw "System Python could not run at '$PythonPath' (exit $LASTEXITCODE). Repair the all-users installation."
}
try {
    $details = $output | ConvertFrom-Json
} catch {
    throw "System Python returned invalid version information at '$PythonPath'. Repair the all-users installation."
}
if ($details.version -cne $Version) {
    throw "System Python version mismatch: expected $Version, found $($details.version) at '$PythonPath'."
}
if ($details.bits -ne 64) {
    throw "System Python must be 64-bit; found $($details.bits)-bit at '$PythonPath'."
}
if ([IO.Path]::GetFileName($PythonPath) -ine 'python.exe') {
    throw "The selected system interpreter must be named python.exe."
}

$paths = @((Split-Path -Parent $PythonPath))
$scripts = Join-Path $paths[0] 'Scripts'
if (Test-Path -LiteralPath $scripts -PathType Container) { $paths += $scripts }
$env:PATH = ($paths -join [IO.Path]::PathSeparator) + [IO.Path]::PathSeparator + $env:PATH
if ($env:GITHUB_PATH) {
    Add-Content -LiteralPath $env:GITHUB_PATH -Value $paths -Encoding utf8
}
Write-Host "Using system Python $Version (64-bit): $PythonPath"
