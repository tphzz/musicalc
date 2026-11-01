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

3. **Extract Zig 0.16.0**

   Download the **0.16.0** Windows archive matching the build host (x86_64 or
   aarch64) from the [official Zig downloads](https://ziglang.org/download/).
   Extract it, for example to `C:\tools\zig-0.16.0`. Zig supplies the C/C++
   compiler, linker, headers and Windows target libraries for **both x64 and ARM64**.
   See [Zig's cross-compilation and libc support](https://ziglang.org/learn/overview/#cross-compiling-is-a-first-class-use-case).
   Separate LLVM, GCC, MSYS2, w64devkit and Visual Studio installations are not
   required. Zig uses the GNU Windows target ABI; no external MinGW installation
   is used. The existing Go/Fyne Cgo build does not use MSVC's compiler interface.

4. **Expose Zig to the build account**

   Add the directory containing `zig.exe` to the runner account's PATH. Restart
   runner services when idle after updating their environment. For the current
   PowerShell session:
   ```powershell
   $env:Path = "C:\tools\zig-0.16.0;$env:Path"
   zig version
   ```
   The command must report `0.16.0`; build entry points reject other versions.
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
   sudo apt-get install gcc libgl1-mesa-dev xorg-dev libasound2-dev pkg-config
   ```
3. **Linux ARM64 Cross-compilation support (requires a more recent Ubuntu Server version for building)**
   ```bash
   sudo dpkg --add-architecture arm64
   sudo apt-get update
   sudo apt install gcc-aarch64-linux-gnu libasound2-dev:arm64 libgl1-mesa-dev:arm64 libx11-dev:arm64 libxrandr-dev:arm64 libxxf86vm-dev:arm64 libxi-dev:arm64 libxcursor-dev:arm64 libxinerama-dev:arm64
   ```

4. **Windows x64 and ARM64 cross-compilation support**

   Extract **Zig 0.16.0** for the Linux build host from the
   [official downloads](https://ziglang.org/download/) and put its directory on
   PATH. `zig version` must report `0.16.0`. The same archive supplies both Windows
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

1. **Android SDK and NDK**

   Install Android Studio, then open the SDK Manager → "SDK Tools" and install:
   - **Android SDK Build-Tools** (provides `zipalign`)
   - **NDK (Side by side)**

   The build script uses `%LOCALAPPDATA%\Android\Sdk` and the newest NDK inside it by default. Set `ANDROID_HOME` and/or `ANDROID_NDK_HOME` to override.

2. **Fyne CLI**
   ```powershell
   go install fyne.io/tools/cmd/fyne@latest
   ```

### Build

```powershell
.\build\scripts\build-android.ps1
```

The script builds a release APK with the `no_emoji`, `no_animations` and `migrated_fynedo` tags for a smaller, snappier app, aligns it to 16 KB boundaries (required for Android 15+), and writes it to `build\dist\musicalc.apk`. App metadata (name, ID, icon) comes from `FyneApp.toml`; the version comes from the latest Git tag and the build number from the commit count.

Install it on a connected device with:
```powershell
adb install -r build\dist\musicalc.apk
```

## Requirements

- Go 1.26.2 or later
- Zig 0.16.0 for Windows x64/ARM64 Cgo builds; native GCC/development libraries for Linux
- Fyne v2.8.2
