from contextlib import ExitStack, redirect_stderr
import importlib.util
import io
import json
import subprocess
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import ios_app_store as asc

spec = importlib.util.spec_from_file_location("build_ios_flow", Path(__file__).with_name("build-ios.py"))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class BuildFlowTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        self.home = Path(temporary.name)
        self.directory = self.home / ".config/musicalc-ios"
        self.directory.mkdir(parents=True)
        self.config = {"bundle_id": "io.example.musicalc", "team_id": "ABCDEFGHIJ",
                       "certificate": "Apple Distribution: Example (ABCDEFGHIJ)", "profile": "Example",
                       "app_id": "1234567890", "api_key_id": "0123456789",
                       "api_issuer_id": "00000000-0000-0000-0000-000000000000",
                       "key_directory": str(self.directory / "private_keys")}
        self.key = Path(self.config["key_directory"]) / "AuthKey_0123456789.p8"
        self.key.parent.mkdir()
        self.key.write_text("test fixture", encoding="utf-8")
        self.save_config()
        owner = (self.directory / "config.json").stat().st_uid
        stack = ExitStack()
        self.addCleanup(stack.close)
        stack.enter_context(patch.object(module.sys, "argv", ["build-ios.py"]))
        stack.enter_context(patch.object(module.sys, "platform", "darwin"))
        stack.enter_context(patch.object(module.Path, "home", return_value=self.home))
        stack.enter_context(patch.object(module.os, "getuid", return_value=owner, create=True))
        stack.enter_context(patch.object(module.stat, "S_IMODE", return_value=0o600))
        stack.enter_context(patch.object(module, "ROOT", self.home / "checkout"))
        stack.enter_context(patch.object(module, "current", return_value={
            "marketing_version": "0.10.0", "commit": "a"*40}))
        self.number = stack.enter_context(patch.object(module, "current_build", return_value=20261102120000))
        self.remote = stack.enter_context(patch.object(module, "latest_build", return_value=(8, 0, 0)))
        self.run = stack.enter_context(patch.object(module, "run", side_effect=self.run_command))
        self.icon_fails = False
        self.signing_fails = False
        stack.enter_context(patch.object(module.shutil, "copyfile", side_effect=lambda src, dst: Path(dst).write_bytes(b"signing probe")))

    def save_config(self):
        (self.directory / "config.json").write_text(json.dumps(self.config), encoding="utf-8")

    def run_command(self, *args, **kwargs):
        if args[:2] == ("security", "find-identity"):
            return (("A"*40) + ' "' + self.config["certificate"] + '"').encode()
        if args[0] == "codesign" and self.signing_fails:
            raise subprocess.CalledProcessError(1, list(args))
        if args[0] == "xcodebuild":
            return b"Xcode 26.0"
        if args[:2] == ("xcrun", "--sdk"):
            return b"26.0"
        if args[:2] == ("xcrun", "swift") and self.icon_fails:
            raise RuntimeError("icon failed")
        if len(args) > 1 and args[1] == "release":
            raise RuntimeError("packaging boundary")
        return b""

    def package(self):
        with self.assertRaisesRegex(RuntimeError, "packaging boundary"):
            module.main()
        self.assertFalse((self.directory / "build-numbers.lock").exists())
        self.assertFalse(list(self.directory.glob("*.bak")))
        return next(call.args for call in reversed(self.run.call_args_list)
                    if len(call.args) > 1 and call.args[1] == "release")

    def test_build_only_needs_no_api_config_key_or_network(self):
        for key in ("app_id", "api_key_id", "api_issuer_id", "key_directory"):
            self.config.pop(key)
        self.save_config()
        self.key.unlink()
        self.remote.side_effect = AssertionError("Build-only must not contact Apple")
        arguments = self.package()
        self.assertEqual(arguments[arguments.index("--app-id") + 1], "io.example.musicalc")
        self.assertEqual(arguments[arguments.index("--app-version") + 1], "0.10.0")
        self.assertEqual(arguments[arguments.index("--app-build") + 1], 20261102120000)
        self.remote.assert_not_called()
        self.assertFalse((self.directory / "build-numbers.json").exists())

    def test_corrupt_or_mismatched_legacy_journals_are_ignored_and_untouched(self):
        journal = self.directory / "build-numbers.json"
        for old in (b"invalid JSON", b'{"bundle":"com.musicalc","versions":{"0.10.0":9999}}'):
            with self.subTest(old=old):
                journal.write_bytes(old)
                self.package()
                self.assertEqual(journal.read_bytes(), old)
        self.remote.assert_not_called()

    def test_retry_of_same_commit_uses_fresh_time_without_state(self):
        self.number.side_effect = [20261102235959, 20261103000000]
        for expected in (20261102235959, 20261103000000):
            arguments = self.package()
            self.assertEqual(arguments[arguments.index("--app-build") + 1], expected)
        self.assertFalse((self.directory / "build-numbers.json").exists())

    def test_upload_verifies_configured_app_and_accepts_timestamp_above_old_builds(self):
        with patch.object(module.sys, "argv", ["build-ios.py", "--upload"]):
            self.package()
        self.remote.assert_called_once_with(self.config, "io.example.musicalc", "0.10.0")
        self.assertFalse((self.directory / "build-numbers.json").exists())

    def test_upload_requires_api_settings_before_icon_or_packaging(self):
        self.config.pop("app_id")
        self.save_config()
        error = io.StringIO()
        with patch.object(module.sys, "argv", ["build-ios.py", "--upload"]), \
             redirect_stderr(error), self.assertRaises(SystemExit):
            module.main()
        self.assertIn("app_id", error.getvalue())
        self.number.assert_not_called()
        self.remote.assert_not_called()
        self.assertFalse(any(call.args[:2] == ("xcrun", "swift") for call in self.run.call_args_list))

    def test_upload_missing_private_key_stops_before_icon_and_lookup(self):
        self.key.unlink()
        error = io.StringIO()
        with patch.object(module.sys, "argv", ["build-ios.py", "--upload"]), \
             redirect_stderr(error), self.assertRaises(SystemExit):
            module.main()
        self.assertIn(str(self.key), error.getvalue())
        self.assertIn("not found", error.getvalue())
        self.remote.assert_not_called()
        self.number.assert_not_called()
        self.assertFalse(any(call.args[:2] == ("xcrun", "swift") for call in self.run.call_args_list))

    def test_upload_rejects_clock_behind_existing_build_before_packaging(self):
        for latest in ((20261102120000, 0, 0), (20270101000000, 1, 0)):
            with self.subTest(latest=latest):
                self.remote.return_value = latest
                error = io.StringIO()
                with patch.object(module.sys, "argv", ["build-ios.py", "--upload"]), \
                     redirect_stderr(error), self.assertRaises(SystemExit):
                    module.main()
                self.assertIn("clock", error.getvalue())
        self.assertFalse(any(len(call.args) > 1 and call.args[1] == "release" for call in self.run.call_args_list))
        self.assertFalse((self.directory / "build-numbers.json").exists())

    def test_apple_failure_blocks_upload_but_not_build_only(self):
        self.remote.side_effect = ValueError("App Store Connect unavailable")
        with patch.object(module.sys, "argv", ["build-ios.py", "--upload"]), \
             self.assertRaisesRegex(ValueError, "unavailable"):
            module.main()
        self.assertFalse(any(len(call.args) > 1 and call.args[1] == "release" for call in self.run.call_args_list))
        self.package()
        self.assertFalse((self.directory / "build-numbers.json").exists())

    def test_malformed_apple_pagination_blocks_upload_before_packaging(self):
        self.remote.side_effect = asc.latest_build
        app = {"data": {"id": self.config["app_id"], "attributes": {"bundleId": self.config["bundle_id"]}}}
        with patch.object(module.sys, "argv", ["build-ios.py", "--upload"]), \
             patch.object(asc, "api_token", return_value="fixture-token"), \
             patch.object(asc, "request_json", side_effect=[app, {"data": []}]), \
             self.assertRaises(ValueError):
            module.main()
        self.assertFalse(any(len(call.args) > 1 and call.args[1] == "release" for call in self.run.call_args_list))

    def test_signing_failure_stops_before_icon_build_number_apple_and_tool(self):
        self.signing_fails = True
        with patch.object(module.sys, "argv", ["build-ios.py", "--upload"]), \
             self.assertRaisesRegex(ValueError, "signing preflight failed.*runner session"):
            module.main()
        self.number.assert_not_called()
        self.remote.assert_not_called()
        self.assertFalse(any(call.args[:2] == ("xcrun", "swift") for call in self.run.call_args_list))
        self.assertFalse(any(str(call.args[0]).endswith("prepare-ios-tool.py") or
                             any(str(arg).endswith("prepare-ios-tool.py") for arg in call.args)
                             for call in self.run.call_args_list))

    def test_packaging_uses_one_explicit_login_keychain(self):
        self.package()
        keychain = self.home / "Library/Keychains/login.keychain-db"
        lookup = next(call for call in self.run.call_args_list if call.args[:2] == ("security", "find-identity"))
        self.assertEqual(lookup.args[-1], keychain)
        probe = next(call for call in self.run.call_args_list if call.args[0] == "codesign")
        self.assertEqual(probe.args[probe.args.index("--keychain") + 1], keychain)
        release = next(call for call in self.run.call_args_list if len(call.args) > 1 and call.args[1] == "release")
        self.assertEqual(release.kwargs["env"]["MUSICALC_IOS_SIGNING_KEYCHAIN"], str(keychain))

    def test_packaging_temp_stays_on_checkout_volume_and_cleans_failed_retries(self):
        cache = module.ROOT / "build/.cache"
        cache.mkdir(parents=True)
        sentinel = cache / "unrelated"
        sentinel.write_text("keep", encoding="utf-8")
        workspaces = []
        inherited_temp = self.home / "system temporary"
        inherited_temp.mkdir()

        def observe_release(*args, **kwargs):
            if len(args) > 1 and args[1] == "release":
                workspace = Path(kwargs["env"]["TMPDIR"])
                self.assertEqual(workspace.parent, cache)
                self.assertTrue(workspace.is_dir())
                self.assertEqual(workspace.stat().st_dev, module.ROOT.stat().st_dev)
                (workspace / "partial app").write_bytes(b"partial package")
                self.assertEqual(module.os.environ["TMPDIR"], str(inherited_temp))
                workspaces.append(workspace)
            return self.run_command(*args, **kwargs)

        self.run.side_effect = observe_release
        with patch.dict(module.os.environ, {"TMPDIR": str(inherited_temp)}):
            self.package()
            self.package()
            self.assertEqual(module.os.environ["TMPDIR"], str(inherited_temp))
        self.assertEqual(len(set(workspaces)), 2)
        self.assertTrue(all(not directory.exists() for directory in workspaces))
        self.assertEqual(sentinel.read_text(encoding="utf-8"), "keep")
        self.assertTrue(inherited_temp.is_dir())

    def test_successful_release_cleans_temp_before_ipa_verification_and_keeps_upload_env(self):
        workspaces = []

        def release_then_stop_at_verification(*args, **kwargs):
            if len(args) > 1 and args[1] == "release":
                environment = kwargs["env"]
                workspace = Path(environment["TMPDIR"])
                self.assertTrue(workspace.is_dir())
                self.assertEqual(environment["API_PRIVATE_KEYS_DIR"], self.config["key_directory"])
                self.assertEqual(environment["MUSICALC_IOS_SIGNING_KEYCHAIN"],
                                 str(self.home / "Library/Keychains/login.keychain-db"))
                self.assertEqual(environment["UNCHANGED_PARENT_SETTING"], "preserved")
                workspaces.append(workspace)
                return b""
            if args[:3] == ("ditto", "-x", "-k"):
                self.assertEqual(len(workspaces), 1)
                self.assertFalse(workspaces[0].exists())
                self.assertEqual(kwargs.get("env"), None)
                raise RuntimeError("IPA verification boundary")
            return self.run_command(*args, **kwargs)

        self.run.side_effect = release_then_stop_at_verification
        with patch.object(module.sys, "argv", ["build-ios.py", "--upload"]), \
             patch.dict(module.os.environ, {"UNCHANGED_PARENT_SETTING": "preserved"}), \
             self.assertRaisesRegex(RuntimeError, "IPA verification boundary"):
            module.main()
        self.assertEqual(len(workspaces), 1)
        self.assertFalse(workspaces[0].exists())

    def test_icon_preflight_failure_does_not_start_packaging(self):
        self.icon_fails = True
        with self.assertRaisesRegex(RuntimeError, "icon failed"):
            module.main()
        self.number.assert_not_called()
        self.remote.assert_not_called()
        self.assertFalse((self.directory / "build-numbers.json").exists())


class SigningPreflightTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory(prefix="musicalc signing ")
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name)
        self.fixture = self.root / "source"
        self.fixture.write_bytes(b"probe binary fixture")
        self.real_copyfile = module.shutil.copyfile
        self.failure = ""
        stack = ExitStack()
        self.addCleanup(stack.close)
        stack.enter_context(patch.object(module, "ROOT", self.root))
        stack.enter_context(patch.object(module, "run", side_effect=self.run_command))
        stack.enter_context(patch.object(module.shutil, "copyfile", side_effect=self.copy_probe))
        # A protected system binary's metadata is unnecessary for this test copy.
        stack.enter_context(patch.object(module.shutil, "copystat", side_effect=PermissionError("system metadata denied")))

    def copy_probe(self, source, destination):
        if self.failure == "copy":
            raise PermissionError("probe copy denied")
        self.assertEqual(source, "/usr/bin/true")
        return self.real_copyfile(self.fixture, destination)

    def run_command(self, *args, **kwargs):
        if args[0] == "codesign":
            self.assertEqual(Path(args[-1]).read_bytes(), b"probe binary fixture")
            if self.failure == "sign" and args[1] == "--force":
                raise subprocess.CalledProcessError(78, list(args))
            if self.failure == "verify" and args[1] == "--verify":
                raise subprocess.CalledProcessError(79, list(args))
        return b""

    def test_probe_needs_no_system_metadata_and_cleans_up(self):
        module.verify_signing_identity("A" * 40, self.root / "login keychain.db")
        self.assertEqual(list((self.root / "build/.cache").iterdir()), [])

    def test_copy_sign_and_verify_errors_retain_cause_and_clean_up(self):
        for operation, message in [("copy", "probe copy denied"),
                                   ("sign", "exit status 78"), ("verify", "exit status 79")]:
            with self.subTest(operation=operation):
                self.failure = operation
                with self.assertRaisesRegex(ValueError, message) as stopped:
                    module.verify_signing_identity("A" * 40, self.root / "login keychain.db")
                self.assertIsNotNone(stopped.exception.__cause__)
                self.assertEqual(list((self.root / "build/.cache").iterdir()), [])


if __name__ == "__main__":
    unittest.main()
