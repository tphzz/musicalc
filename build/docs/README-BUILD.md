# Build Instructions

## Windows

### Prerequisites

1. **Clone the repository**
   ```powershell
   git clone https://github.com/bzeiss/musicalc.git
   cd musicalc
   ```

2. **Install LLVM and MSYS2** (provides clang plus the MinGW linker required for Fyne)
   ```powershell
   winget install -e --id MSYS2.MSYS2
   winget install -e --id LLVM.LLVM
   ```

3. **Install GCC via MSYS2**
   
   Open "MSYS2 MSYS" from Start Menu and run:
   ```bash
   pacman -S mingw-w64-x86_64-gcc
   ```

4. **Add MinGW to PATH**
   
   Add `C:\msys64\mingw64\bin` to your system PATH environment variable
   
   - Press `Win+R` and type `sysdm.cpl`, press Enter
   - Go to "Advanced" tab → "Environment Variables"
   - Edit "Path" under System variables
   - Add new entry: `C:\msys64\mingw64\bin`
   - Click OK and restart your terminal

5. **Verify compiler installation**
   ```powershell
   clang --version
   x86_64-w64-mingw32-gcc --version
   ```
   Both commands should output version information.

6. **Install Go dependencies**
   ```powershell
   go mod download
   ```

7. **Build the application**
   ```powershell
   .\build\scripts\build-win.ps1
   ```
   The script injects the application version from the current Git tag. If the current commit is not exactly tagged, it injects a local dev version derived from the latest tag and commit.

8. **Run the application**
   ```powershell
   .\build\dist\musicalc.exe
   ```

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

3. **Windows AMD64 Cross-compilation support (requires a more recent Ubuntu Server version for building)**
   ```bash
   sudo apt-get update
   sudo apt install gcc-mingw-w64-x86-64 g++-mingw-w64-x86-64
   ```

4. **Windows ARM64 Cross-compilation support (requires a more recent Ubuntu Server version for building)**
   - Download Zig from: https://ziglang.org/download/
   - Put `zig` somewhere into your environment `PATH`
   - Test by calling `zig` and `zig c++`
   - The Windows GoReleaser config stores Zig cache data under `build/.cache/zig-global` and `build/.cache/zig-local` so build-generated files stay inside the project tree.

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
   export CGO_CFLAGS="-O3 -flto=auto -march=x86-64-v3 -m64"
   export CGO_LDFLAGS="-O3 -flto=auto"
   export CGO_ENABLED=1
   export CC=x86_64-w64-mingw32-gcc
   export CXX=x86_64-w64-mingw32-g++
   export GOOS=windows
   export GOARCH=amd64
   VERSION="$(git describe --tags --exact-match 2>/dev/null || echo dev-$(git rev-parse --short HEAD))"
   go build -ldflags="-s -w -X main.version=${VERSION}" -o build/dist/musicalc_win_amd64.exe
   ```

9. **Build the application for Windows ARM64**
   ```bash
   mkdir -p build/dist
   export CGO_CFLAGS="-O3 -mcpu=oryon_1 -fomit-frame-pointer"
   export CGO_LDFLAGS="-O3"
   export CGO_ENABLED=1
   export CC="zig cc -target aarch64-windows-gnu"
   export CXX="zig c++ -target aarch64-windows-gnu"
   export ZIG_GLOBAL_CACHE_DIR=build/.cache/zig-global
   export ZIG_LOCAL_CACHE_DIR=build/.cache/zig-local
   export GOOS=windows
   export GOARCH=arm64
   VERSION="$(git describe --tags --exact-match 2>/dev/null || echo dev-$(git rev-parse --short HEAD))"
   go build -ldflags="-s -w -X main.version=${VERSION}" -o build/dist/musicalc_win_arm64.exe
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
- GCC/MinGW (for CGO support on Windows)
- Fyne v2.8.2
