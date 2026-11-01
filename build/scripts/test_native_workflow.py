import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
WORKFLOW = ROOT / ".github/workflows/native-platform.yml"
POWERSHELL = shutil.which("powershell")
BASH = shutil.which("bash")


def packaging_step(test, platform, shell):
    steps = re.split(r"(?m)^      - ", WORKFLOW.read_text(encoding="utf-8"))[1:]
    selected = [step for step in steps if "build/scripts/build-platform.py" in step
                and f"if: runner.os == '{platform}'" in step]
    test.assertEqual(len(selected), 1, f"Expected one native {platform} packaging step")
    step = selected[0]
    test.assertIn(f"shell: {shell}", step)
    test.assertIn("BUILD_PLATFORM: ${{ inputs.platform }}", step)
    test.assertIn("BUILD_ARCH: ${{ inputs.arch }}", step)
    return re.search(r"(?m)^        run: (.+)$", step).group(1)


class NativeWorkflowRoutingTests(unittest.TestCase):
    def test_packaging_routes_use_platform_native_shells(self):
        packaging_step(self, "Windows", "powershell")
        packaging_step(self, "Linux", "bash")


@unittest.skipUnless(os.name == "nt" and POWERSHELL, "requires Windows PowerShell")
class WindowsNativeStepTests(unittest.TestCase):
    def run_step(self, arch, status):
        body = packaging_step(self, "Windows", "powershell")
        with tempfile.TemporaryDirectory(prefix="musicalc native workflow ") as temporary:
            script = Path(temporary) / "step.ps1"
            script.write_text(
                'function python { $global:LASTEXITCODE = [int]$env:MUSICALC_TEST_EXIT; '
                'foreach ($value in $args) { [Console]::WriteLine("argument:" + $value) } }\n'
                + body + '\nexit $LASTEXITCODE\n', encoding="utf-8",
            )
            return subprocess.run(
                [POWERSHELL, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", str(script)],
                env={**os.environ, "PATH": "", "BUILD_PLATFORM": "windows",
                     "BUILD_ARCH": arch, "MUSICALC_TEST_EXIT": str(status)},
                capture_output=True, text=True, timeout=30,
            )

    def test_windows_arguments_work_without_bash_on_path(self):
        for arch in ("amd64", "arm64", "an argument with spaces; $literal"):
            with self.subTest(arch=arch):
                result = self.run_step(arch, 0)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(result.stdout.splitlines(),
                                 ["argument:build/scripts/build-platform.py", "argument:windows", f"argument:{arch}"])

    def test_windows_builder_failure_is_propagated(self):
        result = self.run_step("amd64", 7)
        self.assertEqual(result.returncode, 7, result.stderr)


@unittest.skipUnless(BASH, "requires GNU Bash")
class LinuxNativeStepTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        result = subprocess.run([BASH, "--noprofile", "--norc", "-c", 'printf "%s" "${BASH_VERSION:-}"'],
                                capture_output=True, text=True, timeout=30)
        if result.returncode or not result.stdout:
            raise unittest.SkipTest("requires working GNU Bash")

    def test_linux_arguments_and_failure_status_are_preserved(self):
        body = packaging_step(self, "Linux", "bash")
        for arch, status in (("amd64", 0), ("arm64", 0), ("amd64", 7)):
            with self.subTest(arch=arch, status=status):
                result = subprocess.run(
                    [BASH, "--noprofile", "--norc", "-e", "-o", "pipefail", "-c",
                     'python() { printf "argument:%s\\n" "$@"; return "$MUSICALC_TEST_EXIT"; }\n' + body],
                    env={**os.environ, "BUILD_PLATFORM": "linux", "BUILD_ARCH": arch,
                         "MUSICALC_TEST_EXIT": str(status)},
                    capture_output=True, text=True, timeout=30,
                )
                self.assertEqual(result.returncode, status, result.stderr)
                self.assertEqual(result.stdout.splitlines(),
                                 ["argument:build/scripts/build-platform.py", "argument:linux", f"argument:{arch}"])


if __name__ == "__main__":
    unittest.main()
