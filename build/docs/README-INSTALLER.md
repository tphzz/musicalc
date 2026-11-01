# MusiCalc packages and releases

All GitHub Actions workflows are manual. Select **00 Full - All Platforms** to
build every platform, or one numbered platform workflow for an individual build.
The workflow definitions must exist on the default branch before GitHub exposes
their manual buttons.

## Versions

Release tags are explicit `MAJOR.MINOR.PATCH` values, for example `0.9.0`,
without a leading `v` or leading-zero components. No workflow increments a
version or creates/moves a tag.

A numeric branch produces `0.9.0-snapshot.g<sha>`. Other branches use the latest
reachable numeric tag, or `0.0.0`, as the snapshot base. The public
`build/scripts/version.py` helper is shared by CI and local packaging.
An optional `MUSICALC_SNAPSHOT_VERSION` must be valid prerelease SemVer.
Unset/empty overrides use the normal fallback. A release tag rejects snapshot
overrides.

Android/iOS require a numeric marketing version. They use the snapshot's numeric
core and a separate build number; the artifact name retains the full build version.

## Runner prerequisites

The workflows use these existing self-hosted labels:

| Platform | Labels | Extra prerequisites |
|---|---|---|
| Linux x64 | self-hosted, Linux, X64, jammy, docker | GCC/G++, pkg-config, Xvfb, GL/X11/ALSA development packages |
| Linux ARM64 | self-hosted, Linux, ARM64, jammy, docker | Native AArch64 GCC/G++, the same GUI/audio packages |
| Windows x64 | self-hosted, Windows, X64 | Zig 0.16.0 on PATH, Inno Setup 6; Android SDK/NDK/JDK for APKs |
| Windows ARM64 | self-hosted, Windows, ARM64 | Zig 0.16.0 on PATH |
| Apple platforms | self-hosted, macOS, ARM64 | Selected Xcode, Rosetta for Intel test execution; Apple signing setup |

Actions select Go **1.26.2** and GoReleaser **2.15.4**. All platforms use Python
**3.13.14**. Windows jobs select a preinstalled, all-users 64-bit Python from
`C:\Program Files\Python313\python.exe`; they never invoke `setup-python` or
install/repair Python during a build. Linux and Apple retain `setup-python`.
Fyne packaging uses **1.7.3**, with the checked-in iPhone compatibility patch.
System compiler/mobile prerequisites are not installed by builds. See
[Windows Python installation](#windows-python-installation) and [Apple setup](README-APPLE.md).

Manual desktop and Android platform workflows are independent, including repeated
dispatches on the same ref. Each performs its own checks and can run concurrently
with other dispatches when matching runners are available. Checks/build helpers do
not impose shared host queues.

The iPhone build job remains serialized and non-cancelling through its Apple host
concurrency group across branches, aggregate runs, standalone runs and retries.
Other platforms do not join that group. The aggregate workflow retains its explicit
job dependency order and same-ref aggregate-run queue, without blocking independent
platform dispatches. GitHub concurrency groups are repository-scoped.

Both Windows architectures use the same pinned Zig C/C++ toolchain, including its
Windows headers, libraries and linker. No MSYS2, w64devkit, standalone LLVM/GCC or
Visual Studio build installation is required. See the [Windows toolchain setup](README-BUILD.md#windows)
and [standalone EXE requirements](README-BUILD.md#standalone-windows-executable).
The native CI helper validates Zig before tests and packaging and passes absolute
cache paths under build/.cache to GoReleaser. The release config inherits these
paths and the selected Git checkout from the helper. Use the build entry points
for Go 1.26 submodule/worktree builds so VCS stamping follows that checkout.

## Windows Python installation

Provision Python once per Windows host, then run every runner service under its
normal service account. Windows workflow steps use the built-in Windows
PowerShell 5.1 (`shell: powershell`), including native tests and packaging;
PowerShell 7 (`pwsh`) and Bash are not required by the Windows steps. Linux native
packaging uses Bash and its quoted environment-variable syntax.
The Windows desktop, Android and installer jobs call
`build/scripts/use-system-python.ps1 -Version "3.13.14"`. It validates the exact
runtime version and 64-bit interpreter before adding its directory and Scripts to
job PATH. A missing, broken, 32-bit or mismatched interpreter fails with a setup
message; nothing is installed or changed in the registry by that helper. It selects
Program Files directly, independently of user PATH and Actions tool caches.

1. Wait for the Windows runners to be idle and stop their services. If Python 3.13
   is registered to an old runner tool-cache directory, remove that installation
   through Windows Apps & Features or its original installer as Administrator
   before installing it in the new location. Deleting cache folders or registry
   keys manually does not perform an uninstall.
2. Download the [official Python 3.13.14 full installer](https://www.python.org/downloads/release/python-31314/)
   matching the host architecture. Use Customize installation, install for all
   users, and choose `C:\Program Files\Python313`. The full installer supports
   [all-users installation](https://docs.python.org/3.13/using/windows.html#installation-steps).
   For an **x64 host**, these Administrator PowerShell one-liners download, verify
   and install the published installer:

~~~powershell
Invoke-WebRequest 'https://www.python.org/ftp/python/3.13.14/python-3.13.14-amd64.exe' -OutFile "$env:TEMP\python-3.13.14-amd64.exe"
if ((Get-FileHash -Algorithm SHA256 -LiteralPath "$env:TEMP\python-3.13.14-amd64.exe").Hash -ine 'c54d9b9bbb8a36e6489363ddd01139707fd781d72f1f9e90c7ec65d0061368e0') { throw 'Python installer checksum mismatch' }
Start-Process -FilePath "$env:TEMP\python-3.13.14-amd64.exe" -ArgumentList '/quiet InstallAllUsers=1 TargetDir="C:\Program Files\Python313" PrependPath=1 Include_test=0' -Wait -PassThru -WindowStyle Hidden | Select-Object ExitCode
~~~

Installer exit 0 means success; 3010 means success with a reboot required. The
above file/hash are x64-specific; ARM64 hosts use the ARM64 full installer from
the same official release page.

3. Verify the independent installation, then restart the normal runner services.
   Run the selection helper from the updated public checkout as the service user
   to check the same selection used by jobs:

~~~powershell
& 'C:\Program Files\Python313\python.exe' --version
./build/scripts/use-system-python.ps1 -Version "3.13.14"
~~~

Git for Windows (`git`), the native compiler and other listed tools
must also be available to the service account. Its workspaces must be owned by
that account. Elevated tool provisioning and ordinary checkout/build execution
are separate operations; keep builds under the normal service account.

Retry a commit containing the updated workflows. Older Windows workflows on this
host that still call `setup-python` for 3.13 can modify the same machine-wide Python
registration; update those workflows to the same system-Python policy too. Future
Python pin changes require provisioning the new version once per host before use.

## Local packaging

Run from the public repository root. Ensure Python, Go and the relevant native
prerequisites are on PATH. Install GoReleaser 2.15.4 explicitly if needed.

~~~bash
python build/scripts/version.py
python -m unittest discover -s build/scripts -p 'test_*.py' -v

# On the matching native architecture:
python build/scripts/build-platform.py linux amd64
python build/scripts/build-platform.py linux arm64
python build/scripts/build-platform.py windows amd64
python build/scripts/build-platform.py windows arm64
~~~

GoReleaser configuration files also support direct snapshot builds with a safe
default when `MUSICALC_SNAPSHOT_VERSION` is absent. To use the shared resolver's
version when invoking GoReleaser directly from Bash:

~~~bash
version_exports="$(python build/scripts/version.py --shell)" || exit
eval "$version_exports"
goreleaser release --snapshot --clean --skip=publish --config build/release/goreleaser-linux-amd64.yaml
~~~

Individual Windows workflows produce their architecture's executable. The full
workflow downloads both and assembles the existing installer, which selects the
native executable at installation time. To assemble locally after both binaries
are present as `build/dist/musicalc_x64.exe` and `musicalc_arm64.exe`:

~~~powershell
./build/scripts/build-installer.ps1
~~~

### Local NSIS installer for Windows x64

For a local x64 NSIS package, use the Windows build prerequisites above plus
NSIS 3.11 and run:

~~~powershell
./build/scripts/build-nsis-installer.ps1
# If makensis.exe is not on PATH, supply its location:
./build/scripts/build-nsis-installer.ps1 -MakeNSIS "C:\Program Files (x86)\NSIS\makensis.exe"
~~~

The builder runs `build/scripts/build-win.ps1`, then checks the rebuilt
`build/dist/musicalc.exe` for AMD64 PE architecture, Git revision and an exact
version match before compiling `build/installer/musicalc.nsi`. It writes
`build/dist/installer/musicalc_<version>_windows_amd64_nsis_setup.exe`.
A missing compiler, failed build or identity mismatch stops before compilation.

The executable's `--version` argument prints its full desktop version and exits
before opening a window. The builder captures this output through a pipe,
which is required for the Windows GUI executable in PowerShell.

This Unicode installer supports Windows 10+ on native x64 and retains the
executable's CPU requirements. It requests administrator privileges, installs
for all users under Program Files, adds a Start menu shortcut and an optional
desktop shortcut, includes the MIT licence/icons, and registers an uninstaller.
The standard wizard supports English and German. Uninstall removes only known
installed files and its own shortcuts/registry entries; user preferences remain.

The installer detects an older Inno Setup installation and reuses its directory.
After you click Install, it runs that installation's registered uninstaller
silently without reboot, then installs/registers the new version. This removes
the old package's registration while retaining application preferences and
unrelated files. It checks that MusiCalc is closed and that the destination
accepts file creation/writes before removing the old package. Cancellation or failure of the old uninstaller stops the
upgrade before the new payload is copied.

Existing NSIS installations are updated in their recorded directory. Canceling
the wizard before Install leaves the old installation untouched. Current
aggregate CI still assembles the existing dual-architecture Inno installer;
this NSIS builder remains the local x64 route.

#### Native installer upgrade tests

On x64 Windows with NSIS 3 and Inno Setup 6, run:

~~~powershell
$env:MUSICALC_TEST_MAKENSIS = "C:\Program Files (x86)\NSIS\makensis.exe"
$env:MUSICALC_TEST_ISCC = "C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
python -m unittest discover -s build/scripts -p test_nsis_upgrade.py -v
~~~

These tests compile the production control flow with inert payloads, temporary
file paths and unique HKCU test keys. They exercise a real Inno uninstaller,
NSIS updates/uninstall, locked files, an ACL-denied destination and failed/incomplete migration. They do
not change an installed MusiCalc or its preferences. The tests skip when their
native compilers/platform are unavailable.

Linux packages retain the existing CPU optimization requirements. Tar archives
contain a `musicalc/` directory and executable `install.sh`. Windows builds retain
their existing CPU requirements too; an architecture label is not a promise of
compatibility with every processor in that architecture.

## Review and tag an exact release revision

After validating the release branch, integrate it into main deliberately.
These commands are an example for an as-yet-uncreated `0.9.0` tag; never replace
an existing tag.

~~~bash
git fetch origin
git switch main
git pull --ff-only origin main
git merge --no-ff refs/remotes/origin/0.9.0
git merge-base --is-ancestor refs/remotes/origin/0.9.0 HEAD
# Run the appropriate verification on the integrated revision.
git push origin HEAD:refs/heads/main
git tag -a 0.9.0 -m "Release 0.9.0"
git push origin refs/tags/0.9.0:refs/tags/0.9.0
~~~

Run **00 Full - All Platforms** manually against that tag. Tagged macOS packages
must be signed and notarized. After all platforms succeed, the final job creates
or updates a **draft** GitHub release containing Windows, Linux, macOS and Android
assets plus SHA-256 checksums. It refuses to replace a published release.
Publishing the reviewed draft is a separate user action.

The iPhone IPA is an Actions artifact and uses App Store Connect/TestFlight,
rather than a GitHub release asset. Upload is an explicit manual-workflow input;
it defaults off until Apple onboarding is complete. Upload acceptance does not
mean Apple has finished processing or made the build available to testers.

The existing `utils/release.py` remains a local exact-tag GoReleaser helper.
It skips publishing by default and does not create tags or push commits.
