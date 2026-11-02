import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


SCRIPTS = Path(__file__).resolve().parent
POWERSHELL = shutil.which("powershell")


@unittest.skipUnless(os.name == "nt" and POWERSHELL, "Windows PowerShell is required")
class RunnerDiscoveryTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory(prefix="musicalc runner discovery ")
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name)

    def resolve(self, script, function, arguments, path=None):
        harness = self.root / "resolve.ps1"
        harness.write_text(r"""
$ErrorActionPreference = "Stop"
try {
$parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($env:MUSICALC_TEST_SCRIPT, [ref]$null, [ref]$parseErrors)
if ($parseErrors.Count) { throw $parseErrors[0] }
$definition = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq $env:MUSICALC_TEST_FUNCTION }, $true)
if (-not $definition) { throw "Resolver function not found: $env:MUSICALC_TEST_FUNCTION" }
. ([scriptblock]::Create($definition.Extent.Text))
$parameters = ConvertFrom-Json $env:MUSICALC_TEST_ARGUMENTS
& $env:MUSICALC_TEST_FUNCTION @parameters
} catch { [Console]::Error.WriteLine($_.Exception.Message); exit 1 }
""", encoding="utf-8")
        env = dict(os.environ)
        env.update({
            "MUSICALC_TEST_SCRIPT": str(SCRIPTS / script),
            "MUSICALC_TEST_FUNCTION": function,
            "MUSICALC_TEST_ARGUMENTS": json.dumps(arguments),
            "PATH": path or str(Path(os.environ["SystemRoot"]) / "System32"),
            "ProgramFiles": str(self.root / "Program Files"),
            "ProgramFiles(x86)": str(self.root / "Program Files (x86)"),
            "ProgramW6432": str(self.root / "Program Files"),
            "LOCALAPPDATA": str(self.root / "Local AppData"),
        })
        return subprocess.run(
            [POWERSHELL, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", str(harness)],
            env=env, capture_output=True, text=True, timeout=30,
        )

    def sdk(self, process="", machine="", profile=""):
        return self.resolve("build-android.ps1", "Resolve-AndroidSdkPath", [process, machine, profile])

    def test_android_machine_sdk_works_without_an_inherited_setting(self):
        sdk = self.root / "Shared Android SDK"
        sdk.mkdir()
        result = self.sdk(machine=str(sdk), profile=str(self.root / "Missing profile SDK"))
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(Path(result.stdout.strip()), sdk)

    def test_android_process_override_precedes_machine_sdk(self):
        process, machine = self.root / "Selected SDK", self.root / "Machine SDK"
        process.mkdir()
        machine.mkdir()
        result = self.sdk(str(process), str(machine))
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(Path(result.stdout.strip()), process)

    def test_android_profile_sdk_is_the_last_default(self):
        sdk = self.root / "Profile SDK"
        sdk.mkdir()
        result = self.sdk(profile=str(sdk))
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(Path(result.stdout.strip()), sdk)

    def test_invalid_configured_sdk_fails_instead_of_switching_installations(self):
        valid, missing = self.root / "Other SDK", self.root / "Missing SDK"
        valid.mkdir()
        for arguments in [(str(missing), str(valid), str(valid)), ("", str(missing), str(valid))]:
            with self.subTest(arguments=arguments):
                result = self.sdk(*arguments)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(str(missing), result.stderr)
                self.assertIn("ANDROID_HOME", result.stderr)

    def compiler(self, override="", path=None):
        return self.resolve("build-installer.ps1", "Resolve-InnoCompiler", [override], path)

    def test_inno_explicit_compiler_path_supports_spaces(self):
        compiler = self.root / "Custom Inno" / "ISCC.exe"
        compiler.parent.mkdir()
        compiler.touch()
        result = self.compiler(str(compiler))
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(Path(result.stdout.strip()), compiler)

    def test_inno_finds_standard_install_without_path_changes(self):
        compiler = self.root / "Program Files (x86)" / "Inno Setup 6" / "ISCC.exe"
        compiler.parent.mkdir(parents=True)
        compiler.touch()
        result = self.compiler()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(Path(result.stdout.strip()), compiler)

    def test_inno_path_selection_precedes_standard_install(self):
        standard = self.root / "Program Files (x86)" / "Inno Setup 6" / "ISCC.exe"
        standard.parent.mkdir(parents=True)
        standard.touch()
        compiler = self.root / "Path Compiler" / "ISCC.exe"
        compiler.parent.mkdir()
        compiler.touch()
        result = self.compiler(path=str(compiler.parent))
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(Path(result.stdout.strip()), compiler)

    def test_bad_inno_override_does_not_fall_back(self):
        missing = self.root / "Missing ISCC.exe"
        result = self.compiler(str(missing))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn(str(missing), result.stderr)
        self.assertIn("ISCC_PATH", result.stderr)

    def test_missing_inno_names_the_installation_and_override(self):
        result = self.compiler()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Inno Setup 6", result.stderr)
        self.assertIn("ISCC_PATH", result.stderr)


if __name__ == "__main__":
    unittest.main()
