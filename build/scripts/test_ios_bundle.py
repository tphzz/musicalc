import datetime
from contextlib import redirect_stderr
import io
import os
import stat
import tempfile
from unittest.mock import patch
import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location("build_ios", Path(__file__).with_name("build-ios.py"))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class BundleTests(unittest.TestCase):
    def setUp(self):
        self.info = {"CFBundleIdentifier":"com.musicalc", "CFBundleShortVersionString":"0.9.0",
                     "CFBundleVersion":"1", "UIDeviceFamily":[1], "MinimumOSVersion":"15.0",
                     "DTSDKName":"iphoneos26.0"}
        self.ent = {"application-identifier":"ABCDEFGHIJ.com.musicalc",
                    "get-task-allow":False, "beta-reports-active":True,
                    "com.apple.developer.team-identifier":"ABCDEFGHIJ"}
        self.profile = {"TeamIdentifier":["ABCDEFGHIJ"], "Entitlements":dict(self.ent),
                        "ExpirationDate":datetime.datetime(2099,1,1)}

    def verify(self):
        module.verify_bundle(self.info,self.profile,self.ent,"com.musicalc","ABCDEFGHIJ","0.9.0",1)

    def test_missing_runner_config_reports_path_before_building(self):
        with tempfile.TemporaryDirectory() as temporary:
            home = Path(temporary)
            expected = home / ".config/musicalc-ios/config.json"
            error = io.StringIO()
            with patch.object(module.sys, "argv", ["build-ios.py"]), \
                 patch.object(module.sys, "platform", "darwin"), \
                 patch.object(module.Path, "home", return_value=home), \
                 patch.object(module, "run") as run, \
                 patch.object(module, "reserve") as reserve, \
                 redirect_stderr(error), self.assertRaises(SystemExit) as stopped:
                module.main()
            self.assertEqual(stopped.exception.code, 2)
            self.assertIn(str(expected), error.getvalue())
            self.assertIn("runner account", error.getvalue())
            run.assert_not_called()
            reserve.assert_not_called()

    def test_valid_distribution(self):
        self.verify()

    def test_wrong_version_or_platform_is_rejected(self):
        for key,value in (("CFBundleVersion","2"),("UIDeviceFamily",[1,2]),("DTSDKName","iphoneos18.0")):
            original=self.info[key]
            self.info[key]=value
            with self.subTest(key=key),self.assertRaises(ValueError):
                self.verify()
            self.info[key]=original

    def test_development_wildcard_and_expired_profiles_are_rejected(self):
        for key,value in (("get-task-allow",True),("beta-reports-active",False),
                          ("application-identifier","ABCDEFGHIJ.*")):
            original=self.ent[key]
            self.ent[key]=value
            with self.subTest(key=key),self.assertRaises(ValueError):
                self.verify()
            self.ent[key]=original
        self.profile["ExpirationDate"]=datetime.datetime(2000,1,1)
        with self.assertRaises(ValueError):
            self.verify()



class PrivateKeyTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory(prefix="musicalc private key ")
        self.addCleanup(temporary.cleanup)
        self.key = Path(temporary.name) / "AuthKey_0123456789.p8"
        self.secret = "private key fixture must not be printed"
        self.key.write_text(self.secret, encoding="utf-8")
        self.metadata = self.key.stat()

    def verify(self, uid=None, mode=0o600, owner=None):
        values = list(self.metadata)
        values[0] = stat.S_IFREG | mode
        values[4] = self.metadata.st_uid if owner is None else owner
        with patch.object(module.Path, "stat", return_value=os.stat_result(values)), \
             patch.object(module.Path, "read_text", side_effect=AssertionError("Do not read the key")):
            module.verify_private_key(self.key, self.metadata.st_uid if uid is None else uid)

    def test_valid_key_metadata_passes_without_reading_contents(self):
        self.verify()

    def test_missing_key_reports_the_resolved_path(self):
        self.key.unlink()
        with self.assertRaises(ValueError) as error:
            module.verify_private_key(self.key, self.metadata.st_uid)
        self.assertIn(str(self.key), str(error.exception))
        self.assertIn("not found", str(error.exception))

    def test_directory_is_not_a_key(self):
        self.key.unlink()
        self.key.mkdir()
        with self.assertRaisesRegex(ValueError, "regular file"):
            module.verify_private_key(self.key, self.metadata.st_uid)

    def test_wrong_owner_and_mode_are_reported_together(self):
        with self.assertRaises(ValueError) as error:
            self.verify(uid=42, owner=43, mode=0o644)
        message = str(error.exception)
        for expected in (str(self.key), "43", "42", "0644", "0600"):
            self.assertIn(expected, message)
        self.assertNotIn(self.secret, message)

    def test_ownership_is_still_required_with_correct_mode(self):
        with self.assertRaisesRegex(ValueError, "owner"):
            self.verify(uid=42, owner=43)

    def test_exact_mode_600_is_still_required(self):
        for mode in (0o400, 0o640, 0o644, 0o700):
            with self.subTest(mode=mode), self.assertRaises(ValueError) as error:
                self.verify(mode=mode)
            self.assertIn(f"{mode:04o}", str(error.exception))
            self.assertIn("0600", str(error.exception))

    def test_inspection_failure_is_actionable_and_does_not_print_contents(self):
        with patch.object(module.Path, "stat", side_effect=PermissionError(13, "Permission denied")), \
             self.assertRaises(ValueError) as error:
            module.verify_private_key(self.key, self.metadata.st_uid)
        self.assertIn(str(self.key), str(error.exception))
        self.assertIn("Permission denied", str(error.exception))
        self.assertNotIn(self.secret, str(error.exception))


if __name__ == "__main__":
    unittest.main()
