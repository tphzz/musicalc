import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
HELPER = ROOT / "build/scripts/use-system-python.ps1"
VERSION = ".".join(map(str, sys.version_info[:3]))


@unittest.skipUnless(os.name == "nt" and shutil.which("pwsh"), "requires Windows and PowerShell 7")
class SystemPythonTests(unittest.TestCase):
    def run_helper(self, root, executable, version=VERSION):
        exports = root / "github-path.txt"
        exports.write_text("existing\n", encoding="utf-8")
        result = subprocess.run(
            ["pwsh", "-NoProfile", "-NonInteractive", "-File", str(HELPER),
             "-Version", version, "-PythonPath", str(executable)],
            env={**os.environ, "GITHUB_PATH": str(exports),
                 "RUNNER_TOOL_CACHE": str(root / "unusable-cache"),
                 "AGENT_TOOLSDIRECTORY": str(root / "unusable-cache")},
            capture_output=True, text=True, timeout=30,
        )
        return result, exports.read_text(encoding="utf-8-sig").splitlines()

    def test_real_python_is_validated_and_exported_without_using_toolcache(self):
        with tempfile.TemporaryDirectory(prefix="musicalc system python ") as temporary:
            result, exported = self.run_helper(Path(temporary), sys.executable)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn(f"Using system Python {VERSION}", result.stdout)
        self.assertEqual(exported[0], "existing")
        self.assertEqual(exported[1], str(Path(sys.executable).parent))
        self.assertFalse(any("unusable-cache" in line for line in exported))

    def test_missing_python_reports_installation_before_exporting(self):
        with tempfile.TemporaryDirectory(prefix="musicalc system python ") as temporary:
            missing = Path(temporary) / "Program Files/Python313/python.exe"
            result, exported = self.run_helper(Path(temporary), missing)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Install Python", result.stderr)
        self.assertIn("for all users", result.stderr)
        self.assertEqual(exported, ["existing"])

    def test_wrong_version_is_rejected_without_exporting(self):
        with tempfile.TemporaryDirectory() as temporary:
            result, exported = self.run_helper(Path(temporary), sys.executable, "0.0.0")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn(f"found {VERSION}", result.stderr)
        self.assertIn("expected 0.0.0", result.stderr)
        self.assertEqual(exported, ["existing"])

    def test_default_location_does_not_fall_back_to_path_or_a_runner_cache(self):
        with tempfile.TemporaryDirectory(prefix="musicalc Program Files ") as temporary:
            root = Path(temporary)
            result = subprocess.run(
                ["pwsh", "-NoProfile", "-NonInteractive", "-Command",
                 '$env:ProgramFiles = $env:MUSICALC_TEST_PROGRAMFILES; '
                 'try { & $env:MUSICALC_TEST_SELECTOR -Version "3.13.14" } '
                 'catch { [Console]::Error.WriteLine($_.Exception.Message); exit 1 }'],
                env={**os.environ, "MUSICALC_TEST_PROGRAMFILES": str(root),
                     "MUSICALC_TEST_SELECTOR": str(HELPER),
                     "PATH": str(Path(sys.executable).parent) + os.pathsep + os.environ["PATH"]},
                capture_output=True, text=True, timeout=30,
            )
        self.assertNotEqual(result.returncode, 0)
        self.assertIn(str(root / "Python313"), result.stderr)
        self.assertIn("for all users", result.stderr)

    def test_failed_interpreter_and_32_bit_python_are_rejected(self):
        for body, message in [
            ("@exit /b 7\n", "could not run"),
            ("@echo invalid-json\n@exit /b 0\n", "invalid version information"),
            (f'@echo {{"version":"{VERSION}","bits":32}}\n@exit /b 0\n', "64-bit"),
        ]:
            with self.subTest(message=message), tempfile.TemporaryDirectory() as temporary:
                root = Path(temporary)
                reporter = root / "python.cmd"
                reporter.write_text(body, encoding="utf-8")
                result, exported = self.run_helper(root, reporter)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(message, result.stderr)
                self.assertEqual(exported, ["existing"])


class WindowsPythonWorkflowTests(unittest.TestCase):
    def test_windows_routes_use_system_python_and_linux_keeps_setup_python(self):
        for name in ("native-platform.yml", "native-android-platform.yml", "native-windows-installer.yml"):
            with self.subTest(workflow=name):
                text = (ROOT / ".github/workflows" / name).read_text(encoding="utf-8")
                steps = re.split(r"(?m)^      - ", text)[1:]
                selection = [step for step in steps if "use-system-python.ps1" in step]
                self.assertEqual(len(selection), 1, "Windows job must select system Python")
                self.assertIn("shell: pwsh", selection[0])
                self.assertIn('-Version "3.13.14"', selection[0])
                setup = [step for step in steps if "actions/setup-python@" in step]
                if name == "native-platform.yml":
                    self.assertIn("if: runner.os == 'Windows'", selection[0])
                    self.assertEqual(len(setup), 1)
                    self.assertIn("if: runner.os != 'Windows'", setup[0])
                else:
                    self.assertEqual(setup, [], "Windows job must not install Python")


if __name__ == "__main__":
    unittest.main()
