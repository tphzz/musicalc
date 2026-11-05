import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import textwrap
import unittest


HELPER = Path(__file__).with_name("apple-signing.sh")


class SigningKeychainTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.bash = shutil.which("bash")
        if cls.bash is None:
            raise unittest.SkipTest("Bash is required to test Apple signing")
        version = subprocess.run(
            [cls.bash, "-c", 'printf "%s" "${BASH_VERSION:-}"'],
            capture_output=True, text=True, check=True,
        ).stdout
        if not version:
            raise unittest.SkipTest("The bash command must be GNU Bash")

    def run_helper(self, elapsed=0, failure="", sleep=False, mode="600"):
        with tempfile.TemporaryDirectory(prefix="musicalc signing ") as temporary:
            directory = Path(temporary)
            directory.joinpath("apple-signing.sh").write_text(
                HELPER.read_text(encoding="utf-8"), encoding="utf-8", newline="\n",
            )
            runner_home = directory / "runner home"
            password_file = runner_home / ".config/developer-id-signing/login-keychain-password"
            password_file.parent.mkdir(parents=True)
            password_file.write_text("fixture-password\n", encoding="utf-8")
            # Model the macOS security boundary; no real keychain or credentials.
            script = textwrap.dedent(r"""
                now=0
                locked=1
                timeout_enabled=1
                timeout_seconds=300
                lock_on_sleep=1
                expires=0
                signing_access=0
                stat() {
                  case "$2" in
                    %u) id -u ;;
                    %Lp) printf '%s\n' "$TEST_MODE" ;;
                    *) return 90 ;;
                  esac
                }
                security() {
                  local operation="$1"
                  shift
                  if [[ "$operation" == "$TEST_FAILURE" ]]; then
                    case "$operation" in
                      unlock-keychain) return 7 ;;
                      set-keychain-settings) return 8 ;;
                      set-key-partition-list) return 9 ;;
                    esac
                  fi
                  case "$operation" in
                    unlock-keychain)
                      [[ "$1" == -p && "$2" == fixture-password ]] || return 10
                      locked=0
                      expires=$((now + timeout_seconds))
                      ;;
                    set-keychain-settings)
                      [[ "$locked" == 0 ]] || return 11
                      local OPTIND=1 option
                      timeout_enabled=0
                      lock_on_sleep=0
                      while getopts 'lut:' option; do
                        case "$option" in
                          l) lock_on_sleep=1 ;;
                          u) timeout_enabled=1 ;;
                          t) timeout_seconds="$OPTARG" ;;
                          *) return 12 ;;
                        esac
                      done
                      expires=$((now + timeout_seconds))
                      ;;
                    set-key-partition-list)
                      [[ "$locked" == 0 ]] || return 13
                      signing_access=1
                      ;;
                    *) return 91 ;;
                  esac
                }
                source ./apple-signing.sh
                # Also require errors to propagate when a caller checks status.
                if unlock_signing_keychain; then
                  now="$TEST_ELAPSED"
                else
                  exit "$?"
                fi
                if [[ "$TEST_SLEEP" == true && "$lock_on_sleep" == 1 ]]; then
                  locked=1
                fi
                if (( locked || (timeout_enabled && now >= expires) )); then
                  exit 31
                fi
                [[ "$signing_access" == 1 ]] || exit 32
                printf 'Signed probe\n'
            """)
            env = dict(
                os.environ, HOME=str(runner_home), GITHUB_ACTIONS="",
                TEST_ELAPSED=str(elapsed), TEST_FAILURE=failure,
                TEST_SLEEP=str(sleep).lower(), TEST_MODE=mode,
            )
            return subprocess.run(
                [self.bash, "--noprofile", "--norc", "-e", "-o", "pipefail", "-c", script],
                cwd=directory, env=env, capture_output=True, text=True,
            )

    def test_signing_survives_ninety_minute_build(self):
        result = self.run_helper(elapsed=90 * 60)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout, "Signed probe\n")

    def test_keychain_still_locks_after_two_hours(self):
        result = self.run_helper(elapsed=2 * 60 * 60 + 1)
        self.assertEqual(result.returncode, 31, result.stderr)
        self.assertEqual(result.stdout, "")

    def test_sleep_still_locks_keychain(self):
        result = self.run_helper(sleep=True)
        self.assertEqual(result.returncode, 31, result.stderr)

    def test_security_failures_stop_before_packaging(self):
        for operation, status in [
            ("unlock-keychain", 7),
            ("set-keychain-settings", 8),
            ("set-key-partition-list", 9),
        ]:
            with self.subTest(operation=operation):
                result = self.run_helper(failure=operation)
                self.assertEqual(result.returncode, status, result.stderr)
                self.assertEqual(result.stdout, "")

    def test_invalid_password_permissions_stop_packaging(self):
        result = self.run_helper(mode="644")
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("owned by the runner and mode 600", result.stderr)
        self.assertEqual(result.stdout, "")

    def test_password_is_not_printed(self):
        result = self.run_helper()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotIn("fixture-password", result.stdout + result.stderr)


if __name__ == "__main__":
    unittest.main()
