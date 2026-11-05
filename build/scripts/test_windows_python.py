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


@unittest.skipUnless(os.name == "nt" and shutil.which("powershell"), "requires Windows PowerShell")
class SystemPythonTests(unittest.TestCase):
    shell = "powershell"

    def run_helper(self, root, executable, version=VERSION):
        exports = root / "github-path.txt"
        exports.write_text("existing\n", encoding="utf-8")
        result = subprocess.run(
            [self.shell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", str(HELPER),
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
                [self.shell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command",
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
        self.assertIn(str(root / "Python313-arm64"), result.stderr)
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


@unittest.skipUnless(os.name == "nt" and shutil.which("pwsh"), "requires Windows and PowerShell 7")
class CoreSystemPythonTests(SystemPythonTests):
    shell = "pwsh"


@unittest.skipUnless(os.name == "nt" and shutil.which("powershell"), "requires Windows PowerShell")
class SystemPythonLocationTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.fixture = tempfile.TemporaryDirectory(prefix="musicalc python probe ")
        cls.addClassCleanup(cls.fixture.cleanup)
        cls.probe = Path(cls.fixture.name) / "python.exe"
        source = ('public class Probe { public static void Main() { '
                  'System.Console.WriteLine(@"{""version"":""' + VERSION +
                  '"",""bits"":64}"); } }')
        result = subprocess.run(["powershell", "-NoProfile", "-NonInteractive", "-Command",
            "Add-Type -TypeDefinition $env:MUSICALC_TEST_PROBE_SOURCE "
            "-OutputAssembly $env:MUSICALC_TEST_PROBE_PATH -OutputType ConsoleApplication"],
            env={**os.environ, "MUSICALC_TEST_PROBE_SOURCE": source, "MUSICALC_TEST_PROBE_PATH": str(cls.probe)},
            capture_output=True, text=True, timeout=30)
        if result.returncode:
            raise RuntimeError(result.stderr)

    def run_locations(self, folders, runner_arch, expected, processor_arch="AMD64", explicit=False, broken=False):
        for shell in ("powershell", "pwsh"):
            if not shutil.which(shell):
                continue
            with self.subTest(shell=shell), tempfile.TemporaryDirectory(prefix="musicalc Python locations ") as temporary:
                root = Path(temporary)
                for folder in folders:
                    directory = root / folder
                    directory.mkdir()
                    (directory / "Scripts").mkdir()
                    shutil.copy2(self.probe, directory / "python.exe")
                if broken:
                    (root / "Python313-arm64/python.exe").write_bytes(b"invalid executable")
                exports = root / "github-path.txt"
                exports.write_text("existing\n", encoding="utf-8")
                command = ('$env:ProgramFiles = $env:MUSICALC_TEST_PROGRAMFILES; '
                    '& $env:MUSICALC_TEST_SELECTOR -Version $env:MUSICALC_TEST_VERSION')
                if explicit:
                    command += ' -PythonPath $env:MUSICALC_TEST_EXPLICIT'
                result = subprocess.run([shell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", command],
                    env={**os.environ, "MUSICALC_TEST_PROGRAMFILES": str(root),
                        "MUSICALC_TEST_SELECTOR": str(HELPER), "MUSICALC_TEST_VERSION": VERSION,
                        "MUSICALC_TEST_EXPLICIT": str(root / "Python313/python.exe"),
                        "RUNNER_ARCH": runner_arch, "PROCESSOR_ARCHITECTURE": processor_arch,
                        "GITHUB_PATH": str(exports)}, capture_output=True, text=True, timeout=30)
                exported = exports.read_text(encoding="utf-8-sig").splitlines()
                if expected:
                    self.assertEqual(result.returncode, 0, result.stderr)
                    self.assertEqual(exported, ["existing", str(root / expected), str(root / expected / "Scripts")])
                    self.assertIn(str(root / expected / "python.exe"), result.stdout)
                else:
                    self.assertNotEqual(result.returncode, 0)
                    self.assertIn("could not run", result.stderr)
                    self.assertEqual(exported, ["existing"])

    def test_arm64_installation_is_found_at_its_standard_suffixed_path(self):
        self.run_locations(["Python313-arm64"], "ARM64", "Python313-arm64")

    def test_preference_matches_runner_or_native_powershell_architecture(self):
        for runner, processor, expected in (("ARM64", "AMD64", "Python313-arm64"),
                                            ("X64", "AMD64", "Python313"),
                                            ("", "ARM64", "Python313-arm64")):
            with self.subTest(runner=runner, processor=processor):
                self.run_locations(["Python313", "Python313-arm64"], runner, expected, processor)

    def test_legacy_unsuffixed_installation_remains_usable_on_arm64(self):
        self.run_locations(["Python313"], "ARM64", "Python313")

    def test_explicit_path_takes_precedence_over_native_default(self):
        self.run_locations(["Python313", "Python313-arm64"], "ARM64", "Python313", explicit=True)

    def test_broken_preferred_interpreter_stops_without_falling_back(self):
        self.run_locations(["Python313", "Python313-arm64"], "ARM64", None, broken=True)


@unittest.skipUnless(os.name == "nt" and shutil.which("powershell"), "requires Windows PowerShell")
class WindowsPowerShellSyntaxTests(unittest.TestCase):
    def test_selected_scripts_parse_in_windows_powershell(self):
        scripts = [ROOT / "build/scripts" / name for name in (
            "use-system-python.ps1", "build-win.ps1", "build-android.ps1", "build-installer.ps1")]
        result = subprocess.run(
            ["powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command",
             '$ErrorActionPreference = "Stop"; '
             'foreach ($file in ($env:MUSICALC_TEST_SCRIPTS -split "[|]")) { '
             '$tokens = $null; $errors = $null; '
             '$null = [Management.Automation.Language.Parser]::ParseFile($file, [ref]$tokens, [ref]$errors); '
             'if ($errors.Count) { throw ($errors | Out-String) } }'],
            env={**os.environ, "MUSICALC_TEST_SCRIPTS": "|".join(map(str, scripts))},
            capture_output=True, text=True, timeout=30,
        )
        self.assertEqual(result.returncode, 0, result.stderr)


class WindowsPythonWorkflowTests(unittest.TestCase):
    def test_windows_routes_use_system_python_and_linux_keeps_setup_python(self):
        for name in ("native-platform.yml", "native-android-platform.yml", "native-windows-installer.yml"):
            with self.subTest(workflow=name):
                text = (ROOT / ".github/workflows" / name).read_text(encoding="utf-8")
                steps = re.split(r"(?m)^      - ", text)[1:]
                selection = [step for step in steps if "use-system-python.ps1" in step]
                self.assertEqual(len(selection), 1, "Windows job must select system Python")
                self.assertIn("shell: powershell", selection[0])
                self.assertNotIn("shell: pwsh", text)
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
