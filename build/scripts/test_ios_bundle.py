import datetime
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


if __name__ == "__main__":
    unittest.main()
