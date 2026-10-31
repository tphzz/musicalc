import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


@unittest.skipUnless(os.name == "nt" and shutil.which("pwsh"), "requires Windows and PowerShell 7")
class ExecutableVersionTests(unittest.TestCase):
    def check_version(self, actual, expected):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            reporter = root / "report-version.cmd"
            reporter.write_text("@echo off\necho %MUSICALC_TEST_ACTUAL_VERSION%\nexit /b 0\n", encoding="utf-8")
            verifier = root / "verify.ps1"
            verifier.write_text("""
$ErrorActionPreference = "Stop"
$tokens = $null
$errors = $null
$ast = [Management.Automation.Language.Parser]::ParseFile(
    $env:MUSICALC_TEST_BUILDER, [ref]$tokens, [ref]$errors)
if ($errors.Count) { throw "Builder has PowerShell syntax errors" }
$function = $ast.Find({
    param($node)
    $node -is [Management.Automation.Language.FunctionDefinitionAst] -and
        $node.Name -eq "Assert-ExecutableVersion"
}, $true)
if ($null -eq $function) { throw "Missing executable-version assertion" }
& $function.Body.GetScriptBlock() -Path $env:MUSICALC_TEST_REPORTER -ExpectedVersion $env:MUSICALC_TEST_EXPECTED_VERSION
""", encoding="utf-8")
            result = subprocess.run(
                ["pwsh", "-NoProfile", "-NonInteractive", "-File", str(verifier)],
                env={**os.environ,
                     "MUSICALC_TEST_BUILDER": str(Path(__file__).with_name("build-nsis-installer.ps1")),
                     "MUSICALC_TEST_REPORTER": str(reporter),
                     "MUSICALC_TEST_ACTUAL_VERSION": actual,
                     "MUSICALC_TEST_EXPECTED_VERSION": expected},
                capture_output=True, text=True, timeout=30,
            )
            return result

    def test_exact_version_is_accepted(self):
        result = self.check_version("0.9.0-snapshot.g123abcd", "0.9.0-snapshot.g123abcd")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_prefix_only_version_is_rejected(self):
        result = self.check_version("0.9.0-snapshot.g123abcd", "0.9.0-snapshot")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Executable version mismatch", result.stderr)

    def test_snapshot_cannot_be_packaged_as_release(self):
        result = self.check_version("0.9.0-snapshot.g123abcd", "0.9.0")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Executable version mismatch", result.stderr)


if __name__ == "__main__":
    unittest.main()
