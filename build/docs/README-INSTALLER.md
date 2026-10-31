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
| Windows x64 | self-hosted, Windows, X64 | LLVM clang, MinGW x64 linker, Inno Setup 6; Android SDK/NDK/JDK for APKs |
| Windows ARM64 | self-hosted, Windows, ARM64 | Zig 0.16.0 on PATH |
| Apple platforms | self-hosted, macOS, ARM64 | Selected Xcode, Rosetta for Intel test execution; Apple signing setup |

Actions select Go **1.26.2**, Python **3.13.14**, and GoReleaser **2.15.4**.
Fyne packaging uses **1.7.3**, with the checked-in iPhone compatibility patch.
Persistent runners fail with actionable missing-tool errors rather than installing
system packages during each build. See [Apple setup](README-APPLE.md).

Host concurrency groups serialize the x64 lane and the shared Apple Silicon /
ARM64-VM lane across this repository's platform wrappers and aggregate runs.
GitHub concurrency groups do not coordinate separate repositories.

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

An existing Inno Setup installation must be uninstalled before this NSIS package.
Later NSIS installs reuse their recorded installation directory. Current aggregate
CI still assembles the existing dual-architecture Inno installer; this NSIS
builder is the local x64 route. Installation/upgrade testing is separate from
successful compilation and payload verification.

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
