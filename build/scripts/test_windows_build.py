import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


BUILDER = Path(__file__).with_name("build-win.ps1")
POWERSHELL = shutil.which("powershell")


@unittest.skipUnless(os.name == "nt" and POWERSHELL, "requires Windows PowerShell")
class WindowsBuildTests(unittest.TestCase):
    def run_builder(self, arch, zig_version="0.16.0", build_status=0):
        with tempfile.TemporaryDirectory(prefix="musicalc windows build ") as temporary:
            root = Path(temporary)
            builder = root / "build/scripts/build-win.ps1"
            builder.parent.mkdir(parents=True)
            shutil.copy2(BUILDER, builder)
            harness = root / "run.ps1"
            harness.write_text("""
$ErrorActionPreference = "Stop"
function zig { $global:LASTEXITCODE = 0; $env:MUSICALC_TEST_ZIG_VERSION }
function python { $global:LASTEXITCODE = 0; '{"version":"0.9.0"}' }
function git {
    $global:LASTEXITCODE = 0
    if ($args[-1] -eq "--absolute-git-dir") { Join-Path $env:MUSICALC_TEST_ROOT ".git" }
    else { $env:MUSICALC_TEST_ROOT }
}
function go {
    @{ arguments = @($args); CC = $env:CC; CXX = $env:CXX;
       GOOS = $env:GOOS; GOARCH = $env:GOARCH;
       GIT_DIR = $env:GIT_DIR; GIT_WORK_TREE = $env:GIT_WORK_TREE;
       CGO_CFLAGS = $env:CGO_CFLAGS;
       ZIG_GLOBAL_CACHE_DIR = $env:ZIG_GLOBAL_CACHE_DIR;
       ZIG_LOCAL_CACHE_DIR = $env:ZIG_LOCAL_CACHE_DIR } | ConvertTo-Json -Compress
    $global:LASTEXITCODE = [int]$env:MUSICALC_TEST_BUILD_STATUS
}
$original = @{}
foreach ($name in @("CGO_ENABLED", "CC", "CXX", "GOOS", "GOARCH", "GOAMD64", "CGO_CFLAGS",
                    "CGO_LDFLAGS", "ZIG_GLOBAL_CACHE_DIR", "ZIG_LOCAL_CACHE_DIR", "GIT_DIR", "GIT_WORK_TREE")) {
    $original[$name] = "caller-$name"
    [Environment]::SetEnvironmentVariable($name, $original[$name], "Process")
}
$status = 0
try { & $env:MUSICALC_TEST_BUILDER -Arch $env:MUSICALC_TEST_ARCH }
catch { [Console]::Error.WriteLine($_.Exception.Message); $status = 1 }
foreach ($name in $original.Keys) {
    if ([Environment]::GetEnvironmentVariable($name, "Process") -cne $original[$name]) {
        [Console]::Error.WriteLine("Environment was not restored: $name"); exit 3
    }
}
exit $status
""", encoding="utf-8")
            return subprocess.run(
                [POWERSHELL, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", str(harness)],
                env={**os.environ, "MUSICALC_TEST_BUILDER": str(builder), "MUSICALC_TEST_ARCH": arch,
                     "MUSICALC_TEST_ROOT": str(root),
                     "MUSICALC_TEST_ZIG_VERSION": zig_version, "MUSICALC_TEST_BUILD_STATUS": str(build_status)},
                capture_output=True, text=True, timeout=30,
            )

    def test_both_targets_select_zig_and_distinct_executables(self):
        for zig_version in ("0.16.0", "0.17.0"):
            for arch, target, filename, cpu in (
                ("amd64", "x86_64", "musicalc.exe", "x86_64_v3"),
                ("arm64", "aarch64", "musicalc_arm64.exe", "oryon_1"),
            ):
                with self.subTest(arch=arch, zig_version=zig_version):
                    result = self.run_builder(arch, zig_version=zig_version)
                    self.assertEqual(result.returncode, 0, result.stderr)
                    data = json.loads(result.stdout)
                    self.assertEqual(data["CC"], f"zig cc -target {target}-windows-gnu")
                    self.assertEqual(data["CXX"], f"zig c++ -target {target}-windows-gnu")
                    self.assertEqual((data["GOOS"], data["GOARCH"]), ("windows", arch))
                    self.assertEqual(Path(data["GIT_DIR"]).parent, Path(data["GIT_WORK_TREE"]))
                    self.assertEqual(Path(data["GIT_DIR"]).name, ".git")
                    self.assertIn(f"-mcpu={cpu}", data["CGO_CFLAGS"])
                    self.assertEqual(Path(data["arguments"][-1]).name, filename)
                    self.assertEqual(Path(data["arguments"][-1]).parent.name, "dist")
                    self.assertNotIn("mingw32-gcc", " ".join(data["arguments"]))
                    self.assertTrue(Path(data["ZIG_GLOBAL_CACHE_DIR"]).is_absolute())
                    self.assertTrue(Path(data["ZIG_LOCAL_CACHE_DIR"]).is_absolute())

    def test_wrong_zig_version_stops_before_compilation(self):
        for version in ("0.15.2", "0.18.0", "0.17.0-dev.1"):
            with self.subTest(version=version):
                result = self.run_builder("amd64", zig_version=version)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("Expected Zig 0.16.0 or 0.17.0", result.stderr)
                self.assertEqual(result.stdout.strip(), "")

    def test_failed_compilation_is_reported(self):
        result = self.run_builder("arm64", build_status=7)
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("Windows build failed with exit code 7", result.stderr)


if __name__ == "__main__":
    unittest.main()
