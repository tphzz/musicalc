from contextlib import ExitStack
import importlib.util
import json
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
        key = Path(self.config["key_directory"]) / "AuthKey_0123456789.p8"
        key.parent.mkdir()
        key.write_text("test fixture", encoding="utf-8")
        config_path = self.directory / "config.json"
        config_path.write_text(json.dumps(self.config), encoding="utf-8")
        owner = config_path.stat().st_uid
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
        self.remote = stack.enter_context(patch.object(module, "latest_build", return_value=8))
        self.run = stack.enter_context(patch.object(module, "run", side_effect=self.run_command))
        self.icon_fails = False

    def run_command(self, *args, **kwargs):
        if args[:2] == ("security", "find-identity"):
            return (("A"*40) + ' "' + self.config["certificate"] + '"').encode()
        if args[0] == "xcodebuild":
            return b"Xcode 26.0"
        if args[:2] == ("xcrun", "--sdk"):
            return b"26.0"
        if args[:2] == ("xcrun", "swift") and self.icon_fails:
            raise RuntimeError("icon failed")
        if len(args) > 1 and str(args[1]).endswith("prepare-ios-tool.py"):
            raise RuntimeError("packaging boundary")
        return b""

    def test_checkout_version_and_config_bundle_are_reserved_automatically(self):
        with self.assertRaisesRegex(RuntimeError, "packaging boundary"):
            module.main()
        self.remote.assert_called_once_with(self.config, "io.example.musicalc", "0.10.0")
        state = json.loads((self.directory / "build-numbers.json").read_text(encoding="utf-8"))
        self.assertEqual(state, {"bundle": "io.example.musicalc", "versions": {"0.10.0": 9}})
        # A packaging failure has consumed its reservation, including on retry.
        with self.assertRaisesRegex(RuntimeError, "packaging boundary"):
            module.main()
        state = json.loads((self.directory / "build-numbers.json").read_text(encoding="utf-8"))
        self.assertEqual(state["versions"]["0.10.0"], 10)

    def test_apple_failure_does_not_reserve_or_start_packaging(self):
        self.remote.side_effect = ValueError("App Store Connect unavailable")
        with self.assertRaisesRegex(ValueError, "unavailable"):
            module.main()
        self.assertFalse((self.directory / "build-numbers.json").exists())
        self.assertFalse(any(str(call.args[1]).endswith("prepare-ios-tool.py")
                             for call in self.run.call_args_list if len(call.args) > 1))

    def test_malformed_apple_pagination_does_not_reserve_or_package(self):
        self.remote.side_effect = asc.latest_build
        app = {"data": {"id": self.config["app_id"], "attributes": {"bundleId": self.config["bundle_id"]}}}
        with patch.object(asc, "api_token", return_value="fixture-token"), \
             patch.object(asc, "request_json", side_effect=[app, {"data": []}]), \
             self.assertRaises(ValueError):
            module.main()
        self.assertFalse((self.directory / "build-numbers.json").exists())
        self.assertFalse(any(str(call.args[1]).endswith("prepare-ios-tool.py")
                             for call in self.run.call_args_list if len(call.args) > 1))

    def test_icon_preflight_failure_does_not_reserve(self):
        self.icon_fails = True
        with self.assertRaisesRegex(RuntimeError, "icon failed"):
            module.main()
        self.remote.assert_not_called()
        self.assertFalse((self.directory / "build-numbers.json").exists())


if __name__ == "__main__":
    unittest.main()
