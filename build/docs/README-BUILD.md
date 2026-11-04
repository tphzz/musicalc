# Build Instructions

## Windows

### Prerequisites

1. **Clone the repository**
   ```powershell
   git clone https://github.com/bzeiss/musicalc.git
   cd musicalc
   ```

2. **Install the build runtimes**

   Use Go **1.26.2** and Python **3.13.14**. Windows Actions use a preinstalled
   all-users Python; see [runner setup](README-INSTALLER.md#windows-python-installation).

3. **Extract Zig 0.17.0**

   Download the **0.17.0** Windows archive matching the build host (x86_64 or
   aarch64) from the [official Zig downloads](https://ziglang.org/download/).
   Extract it, for example to `C:\tools\zig-0.17.0`. Zig supplies the C/C++
   compiler, linker, headers and Windows target libraries for **both x64 and ARM64**.
   **0.16.0 remains supported**; an existing installation does not need replacing.
   See [Zig's cross-compilation and libc support](https://ziglang.org/learn/overview/#cross-compiling-is-a-first-class-use-case).
   Separate LLVM, GCC, MSYS2, w64devkit and Visual Studio installations are not
   required. Zig uses the GNU Windows target ABI; no external MinGW installation
   is used. The existing Go/Fyne Cgo build does not use MSVC's compiler interface.

4. **Expose Zig to the build account**

   Add the directory containing `zig.exe` to the runner account's PATH. Restart
   runner services when idle after updating their environment. For the current
   PowerShell session:
   ```powershell
   $env:Path = "C:\tools\zig-0.17.0;$env:Path"
   zig version
   ```
   The command must report `0.17.0` or `0.16.0`. Both stable releases are verified;
   build entry points reject unverified versions and development snapshots.
   Zig caches are placed under `build/.cache` by the local/CI build scripts.

5. **Install Go dependencies**
   ```powershell
   go mod download
   ```

6. **Build the application**
   ```powershell
   .\build\scripts\build-win.ps1 -Arch amd64
   .\build\scripts\build-win.ps1 -Arch arm64
   ```
   Either target can be compiled from either Windows host architecture. Outputs
   are `build/dist/musicalc.exe` (x64) and `build/dist/musicalc_arm64.exe` (ARM64).
   Omitting `-Arch` builds x64. Versions come from `build/scripts/version.py`:
   exact numeric tags produce release versions; other refs produce snapshots.
   Existing x86-64-v3 and Oryon-1 C compiler CPU settings are retained.
   The script restores the caller's environment on success or failure and binds
   Git metadata to the selected checkout, including submodules/worktrees on Go 1.26.

   CI runs tests on the matching native architecture before packaging through
   GoReleaser. Cross-compiling an EXE does not execute that target's tests.

7. **Run the executable matching the machine**
   ```powershell
   .\build\dist\musicalc.exe
   # On Windows ARM64:
   .\build\dist\musicalc_arm64.exe
   ```

### Windows GUI startup

Both Windows targets are linked as GUI applications, so opening the EXE or its
installer shortcut starts the calculator without a console window. Go's
`-H=windowsgui` flag alone is insufficient with the current Zig external linker;
the builders also pass `-extldflags=-Wl,--subsystem,windows`.

The local builder and native CI helper verify the output's PE32+ header, target
architecture and Windows GUI subsystem before accepting it. To inspect an EXE
without running it:

~~~powershell
python build/scripts/windows_binary.py build/dist/musicalc.exe amd64
python build/scripts/windows_binary.py build/dist/musicalc_arm64.exe arm64
~~~

A missing, malformed, wrong-architecture or console executable fails the build.
The checker uses only Python's standard library. `--version` remains available
to scripts that capture or pipe output.

### Standalone Windows executable

The EXE embeds application assets/help. End users need the supported Windows OS
and graphics driver; build tools are not runtime prerequisites. The Windows
Universal CRT is an [OS component on Windows 10 / Server 2016 and later](https://learn.microsoft.com/en-us/cpp/windows/universal-crt-deployment?view=msvc-170).
Its `api-ms-win-crt-*.dll` imports do not require shipping compiler runtime DLLs
beside the EXE or installing a separate VC++ runtime package on supported Windows.

Inspect each artifact with a PE inspection tool. For example, from a Visual Studio
Developer PowerShell when those optional tools are available:
```powershell
dumpbin /DEPENDENTS build/dist/musicalc.exe
dumpbin /DEPENDENTS build/dist/musicalc_arm64.exe
```
A standalone deliverable must not require adjacent third-party compiler runtime
DLLs such as libgcc, libstdc++, libwinpthread, msys-2.0, vcruntime or msvcp.
Import inspection covers declared imports; also smoke-test the GUI on a supported
Windows machine without development tools and test ARM64 on a native ARM64 host.

## Linux

### Ubuntu

1. **Clone the repository**
   ```bash
   git clone https://github.com/bzeiss/musicalc.git
   cd musicalc
   ```

2. **Install dependencies**
   ```bash
   sudo apt-get update
   sudo apt-get install -y build-essential pkg-config libgl1-mesa-dev xorg-dev libasound2-dev libwayland-dev libxkbcommon-dev xvfb xauth libgl1-mesa-dri
   ```
   The pinned GLFW dependency builds X11 and Wayland support by default, so both
   development stacks are required. Xvfb/xauth and Mesa software rendering support
   headless native CI tests. For Docker runners, install these in each runner image;
   packages on the Docker host are not available inside the container. See
   [Linux runner images](README-INSTALLER.md#linux-runner-images) for verification.

3. **Linux ARM64 Cross-compilation support (requires a more recent Ubuntu Server version for building)**
   ```bash
   sudo dpkg --add-architecture arm64
   sudo apt-get update
   sudo apt install gcc-aarch64-linux-gnu libasound2-dev:arm64 libgl1-mesa-dev:arm64 libx11-dev:arm64 libxrandr-dev:arm64 libxxf86vm-dev:arm64 libxi-dev:arm64 libxcursor-dev:arm64 libxinerama-dev:arm64 libwayland-dev:arm64 libxkbcommon-dev:arm64
   ```

4. **Windows x64 and ARM64 cross-compilation support**

   Extract **Zig 0.17.0** (or supported **0.16.0**) for the Linux build host from the
   [official downloads](https://ziglang.org/download/) and put its directory on
   PATH. `zig version` must report `0.17.0` or `0.16.0`. The same archive supplies both Windows
   target toolchains; no separate MinGW package is needed.

5. **Install Go dependencies**
   ```bash
   go mod download
   ```

6. **Build the application for Linux AMD64**
   ```bash
   mkdir -p build/dist
   export CGO_CFLAGS="-O3 -flto=auto -march=x86-64-v3"
   export CGO_LDFLAGS="-O3 -flto=auto"
   export CGO_ENABLED=1
   export CC=gcc
   export CXX=g++
   export GOOS=linux
   export GOARCH=amd64
   VERSION="$(git describe --tags --exact-match 2>/dev/null || echo dev-$(git rev-parse --short HEAD))"
   go build -ldflags="-s -w -X main.version=${VERSION}" -o build/dist/musicalc_linux_amd64
   ```
7. **Build the application for Linux ARM64**
   ```bash
   mkdir -p build/dist
   export CGO_CFLAGS="-O3 -flto=auto -march=armv8.4-a+crc+crypto -fomit-frame-pointer"
   export CGO_LDFLAGS="-O3 -flto=auto -Wl,--gc-sections"
   export CGO_ENABLED=1
   export CC=aarch64-linux-gnu-gcc
   export PKG_CONFIG_PATH=/usr/lib/aarch64-linux-gnu/pkgconfig:/usr/share/pkgconfig
   export GOOS=linux
   export GOARCH=arm64
   VERSION="$(git describe --tags --exact-match 2>/dev/null || echo dev-$(git rev-parse --short HEAD))"
   go build -ldflags="-s -w -X main.version=${VERSION}" -o build/dist/musicalc_linux_arm64
   ```

8. **Build the application for Windows AMD64**
   ```bash
   mkdir -p build/dist
   export CGO_CFLAGS="-O3 -mcpu=x86_64_v3"
   export CGO_LDFLAGS="-O3"
   export CGO_ENABLED=1
   export CC="zig cc -target x86_64-windows-gnu"
   export CXX="zig c++ -target x86_64-windows-gnu"
   export ZIG_GLOBAL_CACHE_DIR="$PWD/build/.cache/zig-global"
   export ZIG_LOCAL_CACHE_DIR="$PWD/build/.cache/zig-local"
   export GOOS=windows
   export GOARCH=amd64
   export GOAMD64=v3
   eval "$(python3 build/scripts/version.py --shell)"
   go build -trimpath -ldflags="-s -w -H=windowsgui -X main.version=${MUSICALC_VERSION}" -o build/dist/musicalc_win_amd64.exe
   ```

9. **Build the application for Windows ARM64**
   ```bash
   mkdir -p build/dist
   export CGO_CFLAGS="-O3 -mcpu=oryon_1 -fomit-frame-pointer"
   export CGO_LDFLAGS="-O3"
   export CGO_ENABLED=1
   export CC="zig cc -target aarch64-windows-gnu"
   export CXX="zig c++ -target aarch64-windows-gnu"
   export ZIG_GLOBAL_CACHE_DIR="$PWD/build/.cache/zig-global"
   export ZIG_LOCAL_CACHE_DIR="$PWD/build/.cache/zig-local"
   export GOOS=windows
   export GOARCH=arm64
   eval "$(python3 build/scripts/version.py --shell)"
   go build -trimpath -ldflags="-s -w -H=windowsgui -X main.version=${MUSICALC_VERSION}" -o build/dist/musicalc_win_arm64.exe
   ```

10. **Run the application**
   ```bash
   ./build/dist/musicalc_xxx
   ```

## Android

The Android build runs on Windows with PowerShell and produces an ARM64 APK.

### Prerequisites

Provision the Android SDK for the **runner service account**, using an accessible
shared directory if several runner accounts use it. Android Studio's Tools → SDK
Manager shows the SDK Location. Install an SDK Platform, Android SDK Build-Tools
(provides `zipalign`), and NDK (Side by side); a JDK is also required by packaging.
The workflow does not install these prerequisites.

SDK selection is: process `ANDROID_HOME`, machine-level `ANDROID_HOME`, then
`%LOCALAPPDATA%\Android\Sdk` for the runner account. The machine setting is read
from Windows directly when the process setting is empty. A nonempty configured
path that is missing fails explicitly instead of switching installations. Inspect
both settings with:

~~~powershell
[pscustomobject]@{Process=$env:ANDROID_HOME; Machine=[Environment]::GetEnvironmentVariable('ANDROID_HOME','Machine')}
~~~

For an existing SDK, set its location from Administrator PowerShell:

~~~powershell
$androidSdkPath = (Read-Host 'Android SDK location').Trim().Trim('"'); if (-not (Test-Path -LiteralPath $androidSdkPath -PathType Container)) { throw 'SDK directory does not exist' }; [Environment]::SetEnvironmentVariable('ANDROID_HOME', $androidSdkPath, 'Machine')
~~~

An explicit process setting takes precedence. If a running runner retains an old
nonempty value, restart it while idle after correcting that setting. The script
uses `ANDROID_NDK_HOME` when it points to an installed NDK; otherwise it selects
the newest side-by-side NDK under the selected SDK. It requires `llvm-readelf` to
verify native library alignment. These settings do not install an SDK or grant
access to another user's private profile.

The workflow installs the pinned Fyne packaging CLI. For manual packaging use:

~~~powershell
go install fyne.io/tools/cmd/fyne@v1.7.3
~~~

### NDK compiler fails inside another user's profile

If the runner can hash/read clang.exe and its file ACL grants RX, but even
`clang.exe --version` prints `Permission denied`, inspect the SDK's parent path.
The NDK's [LLVM Windows startup](https://android.googlesource.com/toolchain/llvm-project/+/9f872551d3c681d06fd303b36f16ed5c274735eb/llvm/lib/Support/Windows/Process.inc#205)
calls GetLongPathNameW before processing compiler options. That API also needs
[List Folder/Read Data/Read Attributes on parent directories](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getlongpathnamew).
Access to the executable alone does not supply access to another user's private
profile. This can fail after the executable has already started.

Provision the SDK in a shared location outside private profiles, for example
`C:\Android\Sdk`. For an existing SDK on C:, close Android Studio and wait for
all jobs using it to be idle. Run the following in Administrator PowerShell.
It validates the source SDK, rejects an existing destination, and moves within
the same volume rather than making another SDK copy. Change the destination drive
if the current SDK is on a different volume.

~~~powershell
$ErrorActionPreference = 'Stop'
$oldSdk = [Environment]::GetEnvironmentVariable('ANDROID_HOME','Machine')
if ([string]::IsNullOrWhiteSpace($oldSdk)) { throw 'Machine ANDROID_HOME is not set' }
$oldSdk = (Resolve-Path -LiteralPath $oldSdk).Path
$newSdk = 'C:\Android\Sdk'
if ($oldSdk -eq [IO.Path]::GetPathRoot($oldSdk)) { throw 'SDK source must not be a drive root' }
foreach ($component in @('build-tools','platforms','ndk')) {
    if (-not (Test-Path -LiteralPath (Join-Path $oldSdk $component) -PathType Container)) { throw "SDK component missing: $component" }
}
if ([IO.Path]::GetPathRoot($oldSdk) -ne [IO.Path]::GetPathRoot($newSdk)) { throw 'Choose a destination on the same volume' }
if (Test-Path -LiteralPath $newSdk) { throw "Destination already exists: $newSdk" }
New-Item -ItemType Directory -Path (Split-Path -Parent $newSdk) -Force | Out-Null
Move-Item -LiteralPath $oldSdk -Destination $newSdk
[Environment]::SetEnvironmentVariable('ANDROID_HOME',$newSdk,'Machine')
$env:ANDROID_HOME = $newSdk
~~~

The existing file permissions move with the SDK. Ensure the runner account has
read/execute access to the shared parent directory and SDK tree. Update Android
Studio's SDK Location to the new directory, and update any separately configured
SDK/NDK paths that still name the old location. Restart the runner with a refreshed
environment; rebooting the idle host ensures services inherit the new machine
setting. A nonempty old process ANDROID_HOME takes precedence over the new machine
setting until refreshed.

In a fresh runner-account PowerShell, verify before retrying CI:

~~~powershell
$androidSdkPath = [Environment]::GetEnvironmentVariable('ANDROID_HOME','Machine')
& "$androidSdkPath\ndk\30.0.16248370\toolchains\llvm\prebuilt\windows-x86_64\bin\clang.exe" --version
~~~

That example uses NDK30.0.16248370; substitute the installed NDK version. A passing
version command confirms startup, not a completed APK or native device test.

### Build

~~~powershell
.\build\scripts\build-android.ps1
~~~

The release APK uses `no_emoji`, `no_animations` and `migrated_fynedo`, and mandatory
16 KB ZIP/native-library alignment checks. Output is
`build/dist/musicalc_<version>_android_arm64.apk`; the successful build prints the
exact filename. `FyneApp.toml` supplies app identity/icon metadata; the shared
version resolver supplies marketing version and build number. Use that exact
reported APK filename with `adb install -r` for device installation.

## Requirements

- Go 1.26.2 or later
- Zig 0.17.0 (recommended) or 0.16.0 for Windows x64/ARM64 Cgo builds; native GCC/development libraries for Linux
- Fyne v2.8.2
