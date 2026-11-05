$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Resolve-Path (Join-Path $scriptDir "..\..")
$distDir = Join-Path $projectRoot "build\dist"

# Build tags for a snappier mobile app:
#   no_emoji        - drop the 4 MB bundled emoji font (UI uses theme icons instead)
#   no_animations   - instant transitions, like a hardware calculator
#   migrated_fynedo - lock-free Fyne internals; all UI work runs on the main thread
$buildTags = "no_emoji,no_animations,migrated_fynedo"

# --- Locate Android SDK and NDK ---
function Resolve-AndroidSdkPath {
    param([string]$ProcessPath, [string]$MachinePath, [string]$ProfilePath)
    $sdkPath = $ProcessPath
    if ([string]::IsNullOrWhiteSpace($sdkPath)) { $sdkPath = $MachinePath }
    if ([string]::IsNullOrWhiteSpace($sdkPath)) { $sdkPath = $ProfilePath }
    if (-not (Test-Path -LiteralPath $sdkPath -PathType Container)) {
        throw "Android SDK not found at '$sdkPath'. Set ANDROID_HOME to an installed SDK directory accessible to the runner account (process setting, then machine setting, then profile default)."
    }
    return (Resolve-Path -LiteralPath $sdkPath).Path
}

$env:ANDROID_HOME = Resolve-AndroidSdkPath $env:ANDROID_HOME `
    ([Environment]::GetEnvironmentVariable("ANDROID_HOME", "Machine")) `
    (Join-Path $env:LOCALAPPDATA "Android\Sdk")

if ([string]::IsNullOrWhiteSpace($env:ANDROID_NDK_HOME) -or -not (Test-Path $env:ANDROID_NDK_HOME)) {
    $ndkDir = Join-Path $env:ANDROID_HOME "ndk"
    $ndk = Get-ChildItem $ndkDir -Directory -ErrorAction SilentlyContinue |
        Sort-Object { [version]($_.Name -replace '[^0-9.].*$', '') } |
        Select-Object -Last 1
    if ($null -eq $ndk) {
        throw "Android NDK not found in $ndkDir. Install 'NDK (Side by side)' via Android Studio's SDK Manager or set ANDROID_NDK_HOME."
    }
    $env:ANDROID_NDK_HOME = $ndk.FullName
}

# --- Locate zipalign (newest build-tools) ---
$zipalign = Get-ChildItem (Join-Path $env:ANDROID_HOME "build-tools") -Directory -ErrorAction SilentlyContinue |
    Sort-Object { [version]($_.Name -replace '[^0-9.].*$', '') } |
    ForEach-Object { Join-Path $_.FullName "zipalign.exe" } |
    Where-Object { Test-Path $_ } |
    Select-Object -Last 1
if ($null -eq $zipalign) {
    throw "zipalign not found in $env:ANDROID_HOME\build-tools."
}

$readelf = Join-Path $env:ANDROID_NDK_HOME "toolchains\llvm\prebuilt\windows-x86_64\bin\llvm-readelf.exe"
if (-not (Test-Path $readelf)) {
    $readelf = (Get-Command llvm-readelf -ErrorAction SilentlyContinue).Source
}

Write-Host "Android SDK: $env:ANDROID_HOME"
Write-Host "Android NDK: $env:ANDROID_NDK_HOME"
Write-Host "zipalign:    $zipalign"

# --- Build configuration ---
# In release mode fyne adds -Wl,-z,max-page-size=16384 to the link step itself.
$env:CGO_ENABLED = "1"
$env:CGO_CFLAGS = "-O3"

New-Item -ItemType Directory -Force $distDir | Out-Null

Push-Location $projectRoot
try {
    # Android needs a plain x.y.z version plus a build number that increases with every build
    $metadata = python (Join-Path $scriptDir 'version.py') | ConvertFrom-Json
    if ($LASTEXITCODE -ne 0) { throw 'Version resolution failed' }
    $appVersion = $metadata.marketing_version
    $appBuild = $metadata.build

    Write-Host "Building APK $appVersion (build $appBuild) with tags $buildTags..."
    Get-ChildItem -Filter *.apk | Remove-Item -Force
    fyne package --os android/arm64 --release --tags $buildTags --app-version $appVersion --app-build $appBuild
    if ($LASTEXITCODE -ne 0) {
        throw "fyne package failed with exit code $LASTEXITCODE."
    }

    $apk = Get-ChildItem -Filter *.apk | Select-Object -First 1
    if ($null -eq $apk) {
        throw "fyne package did not produce an APK."
    }

    # --- Align to 16 KB boundaries (Android 15+ / Google Play requirement) ---
    $unaligned = Join-Path $distDir "musicalc-unaligned.apk"
    $aligned = Join-Path $distDir ("musicalc_{0}_android_arm64.apk" -f $metadata.version)
    Move-Item -Force $apk.FullName $unaligned

    Write-Host "Aligning APK to 16 KB boundary..."
    & $zipalign -f -P 16 4 $unaligned $aligned
    if ($LASTEXITCODE -ne 0) { throw "zipalign failed." }
    & $zipalign -c -P 16 4 $aligned
    if ($LASTEXITCODE -ne 0) { throw "zipalign verification failed." }
    Remove-Item -Force $unaligned

    # --- Verify 16 KB alignment of the native library's LOAD segments ---
    if ($null -ne $readelf) {
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        $zip = [System.IO.Compression.ZipFile]::OpenRead($aligned)
        try {
            $entry = $zip.Entries | Where-Object { $_.FullName -like "lib/arm64-v8a/*.so" } | Select-Object -First 1
            if ($null -eq $entry) { throw "No arm64-v8a native library found in APK." }
            $tmpLib = Join-Path $distDir "tmp_lib.so"
            [System.IO.Compression.ZipFileExtensions]::ExtractToFile($entry, $tmpLib, $true)
        } finally {
            $zip.Dispose()
        }

        try {
            python (Join-Path $scriptDir 'android_alignment.py') $readelf $tmpLib
            if ($LASTEXITCODE -ne 0) { throw "Native library alignment verification failed" }
        } finally {
            Remove-Item -LiteralPath $tmpLib -Force -ErrorAction SilentlyContinue
        }
    } else {
        throw "llvm-readelf is required to verify native library alignment."
    }

    Write-Host "Build complete: $aligned ($([math]::Round((Get-Item $aligned).Length / 1MB, 1)) MB)"
} finally {
    Pop-Location
}
