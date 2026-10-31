import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile
import textwrap
import unittest


WORKFLOW = Path(__file__).resolve().parents[2] / ".github/workflows/native-ios-platform.yml"


class IPhoneWorkflowTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.bash = shutil.which("bash")
        if cls.bash is None:
            raise unittest.SkipTest("Bash is required to test the iPhone workflow")
        version = subprocess.run(
            [cls.bash, "-c", 'printf "%s" "${BASH_VERSION:-}"'],
            capture_output=True, text=True, check=True,
        ).stdout
        if not version:
            raise unittest.SkipTest("The bash command must be GNU Bash")
        # Exercise the workflow body, including strict-mode and unlock ordering.
        blocks = re.findall(
            r"(?m)^        run: \|\n((?:          .*\n|\n)+)",
            WORKFLOW.read_text(encoding="utf-8"),
        )
        if len(blocks) != 1:
            raise AssertionError("Expected one multiline iPhone packaging step")
        cls.body = textwrap.dedent(blocks[0])

    def run_step(self, upload, unlock_exit=0):
        with tempfile.TemporaryDirectory(prefix="musicalc iphone ") as temporary:
            scripts = Path(temporary) / "build/scripts"
            scripts.mkdir(parents=True)
            scripts.joinpath("apple-signing.sh").write_text(
                'unlock_signing_keychain() { return "$TEST_UNLOCK_EXIT"; }\n'
                'python3() { printf "argument:%s\\n" "$@"; }\n',
                encoding="utf-8",
            )
            env = dict(os.environ, UPLOAD_TESTFLIGHT=upload, TEST_UNLOCK_EXIT=str(unlock_exit))
            return subprocess.run(
                [self.bash, "--noprofile", "--norc", "-e", "-o", "pipefail", "-c", self.body],
                cwd=temporary, env=env, capture_output=True, text=True,
            )

    def test_build_without_upload_has_no_optional_arguments(self):
        result = self.run_step("false")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.splitlines(), ["argument:build/scripts/build-ios.py"])

    def test_upload_passes_exactly_one_flag(self):
        result = self.run_step("true")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.splitlines(), ["argument:build/scripts/build-ios.py", "argument:--upload"])

    def test_unlock_failure_stops_before_packaging(self):
        result = self.run_step("false", unlock_exit=7)
        self.assertEqual(result.returncode, 7, result.stderr)
        self.assertEqual(result.stdout, "")


if __name__ == "__main__":
    unittest.main()
