"""Run the production installer flow with inert payloads and isolated temp/HKCU paths."""
import ctypes
import os
from pathlib import Path
import platform
import re
import shutil
import subprocess
import tempfile
import unittest
import uuid

if os.name == "nt":
    import winreg

MAKENSIS = os.environ.get("MUSICALC_TEST_MAKENSIS") or shutil.which("makensis")
ISCC = os.environ.get("MUSICALC_TEST_ISCC") or shutil.which("iscc")


@unittest.skipUnless(os.name == "nt" and platform.machine().lower() in ("amd64", "x86_64")
                     and MAKENSIS and ISCC, "requires x64 Windows, NSIS 3 and Inno Setup 6")
class InstallerUpgradeTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory(prefix="musicalc-installer-test-")
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name).resolve()
        self.assertTrue(self.root.is_relative_to(Path(tempfile.gettempdir()).resolve()))
        self.product = Path(__file__).resolve().parents[2]
        self.registry = r"Software\MusiCalcInstallerTests" + "\\" + str(uuid.uuid4())
        self.legacy_key = self.registry + r"\inno"
        self.app_key = self.registry + r"\nsis"
        self.uninstall_key = self.registry + r"\uninstall"
        self.addCleanup(self.remove_test_registry)
        self.old = self.root / "old app with spaces"
        self.old.mkdir()
        (self.old / "musicalc.exe").write_bytes(b"old")
        self.default = self.root / "fresh app"
        self.preferences = self.root / "user-data/preferences.json"
        self.preferences.parent.mkdir()
        self.preferences.write_bytes(b'{"retained":true}')
        self.payload = self.root / "new.exe"
        self.payload.write_bytes(b"new")
        self.setup = self.compile_setup()

    def run_checked(self, arguments):
        result = subprocess.run(list(map(str, arguments)), capture_output=True, text=True, timeout=60)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        return result

    def compile_nsis(self, script, output, extra=()):
        self.run_checked([MAKENSIS, "/NOCONFIG", "/WX", "/V2", "/INPUTCHARSET", "UTF8",
                          "/DInstallerOutput=" + str(output), *extra, script])
        return output

    def compile_setup(self, destination=None):
        text = (self.product / "build/installer/musicalc.nsi").read_text(encoding="utf-8")
        # Scope-only substitutions keep the production upgrade/installation control flow.
        for name, value in (("AppKey", self.app_key), ("UninstallKey", self.uninstall_key),
                            ("InnoKey", self.legacy_key)):
            text, count = re.subn(r'^!define ' + name + r' "[^"]+"$',
                                 lambda _: f'!define {name} "{value}"', text, flags=re.M)
            self.assertEqual(count, 1)
        self.assertIn('RequestExecutionLevel admin', text)
        text = text.replace("RequestExecutionLevel admin", "RequestExecutionLevel user")
        text = text.replace("HKLM", "HKCU").replace("SetShellVarContext all", "SetShellVarContext current")
        self.assertIn('InstallDir "$PROGRAMFILES64\\${AppName}"', text)
        text = text.replace('InstallDir "$PROGRAMFILES64\\${AppName}"', f'InstallDir "{self.default}"')
        (self.root / "links").mkdir(exist_ok=True)
        for variable in ("$SMPROGRAMS", "$DESKTOP"):
            text = text.replace(variable + r"\MusiCalc.lnk", str(self.root / "links" / (variable[1:] + ".lnk")))
        self.assertNotIn("HKLM", text)
        self.assertNotIn("8F9A3D2E", text)
        self.assertNotIn("$PROGRAMFILES", text)
        if destination is not None:
            self.assertTrue(destination.resolve().is_relative_to(self.root))
            # Simulate choosing a destination on the Directory page after .onInit.
            text = text.replace("  Call PrepareUpgrade",
                                f'  StrCpy $INSTDIR "{destination}"\n  Call PrepareUpgrade', 1)
        script = self.root / "fixture.nsi"
        script.write_text(text, encoding="utf-8")
        return self.compile_nsis(script, self.root / "musicalc-fixture.exe", [
            "/DAppVersion=1.0.0-test", "/DNumericVersion=1.0.0.0",
            "/DSourceRoot=" + str(self.product), "/DAppExe=" + str(self.payload)])

    def write_registry(self, path, values):
        with winreg.CreateKeyEx(winreg.HKEY_CURRENT_USER, path, 0,
                                winreg.KEY_WRITE | winreg.KEY_WOW64_64KEY) as key:
            for name, value in values.items():
                winreg.SetValueEx(key, name, 0, winreg.REG_SZ, str(value))

    def read_registry(self, path, name):
        try:
            with winreg.OpenKey(winreg.HKEY_CURRENT_USER, path, 0,
                                winreg.KEY_READ | winreg.KEY_WOW64_64KEY) as key:
                return winreg.QueryValueEx(key, name)[0]
        except FileNotFoundError:
            return None

    def remove_test_registry(self):
        def remove(path):
            self.assertTrue(path == self.registry or path.startswith(self.registry + "\\"))
            try:
                with winreg.OpenKey(winreg.HKEY_CURRENT_USER, path, 0,
                                    winreg.KEY_READ | winreg.KEY_WOW64_64KEY) as key:
                    children = [winreg.EnumKey(key, i) for i in range(winreg.QueryInfoKey(key)[0])]
                for name in children:
                    remove(path + "\\" + name)
                winreg.DeleteKeyEx(winreg.HKEY_CURRENT_USER, path, winreg.KEY_WOW64_64KEY)
            except FileNotFoundError:
                pass
        remove(self.registry)

    def install_legacy_inno(self):
        script = self.root / "legacy.iss"
        script.write_text(f'''[Setup]
AppId=Fixture-{uuid.uuid4()}
AppName=MusiCalc upgrade test fixture
AppVersion=0.1.0
DefaultDirName={self.old}
PrivilegesRequired=lowest
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
CreateUninstallRegKey=no
DisableStartupPrompt=yes
OutputDir={self.root}
OutputBaseFilename=legacy-setup
[Files]
Source: "{self.old / 'musicalc.exe'}"; DestDir: "{{app}}"; Flags: ignoreversion
[Registry]
Root: HKCU; Subkey: "{self.legacy_key}"; ValueType: string; ValueName: "InstallLocation"; ValueData: "{{app}}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "{self.legacy_key}"; ValueType: string; ValueName: "UninstallString"; ValueData: """{{uninstallexe}}"""; Flags: uninsdeletekey
''', encoding="utf-8")
        self.run_checked([ISCC, "/Q", script])
        self.run_checked([self.root / "legacy-setup.exe", "/VERYSILENT", "/SUPPRESSMSGBOXES", "/NORESTART"])
        self.assertIsNotNone(self.read_registry(self.legacy_key, "UninstallString"))

    def register_fake_uninstaller(self, exit_code):
        marker = self.root / "uninstaller-called.txt"
        script = self.root / "fake-uninstall.nsi"
        script.write_text(f'''Unicode true
Name "Uninstaller test fixture"
OutFile "${{InstallerOutput}}"
RequestExecutionLevel user
SilentInstall silent
Page instfiles
Section
  FileOpen $0 "{marker}" w
  FileWrite $0 "called"
  FileClose $0
  SetErrorLevel {exit_code}
SectionEnd
''', encoding="utf-8")
        executable = self.compile_nsis(script, self.root / "fake-uninstall.exe")
        self.write_registry(self.legacy_key, {
            "UninstallString": '"' + str(executable) + '"', "InstallLocation": self.old})
        return marker

    def run_setup(self):
        return subprocess.run([str(self.setup), "/S"], capture_output=True, text=True, timeout=60)

    def assert_not_replaced(self):
        self.assertEqual((self.old / "musicalc.exe").read_bytes(), b"old")
        self.assertIsNone(self.read_registry(self.uninstall_key, "DisplayVersion"))
        self.assertEqual(self.preferences.read_bytes(), b'{"retained":true}')
        self.assertEqual(list(self.old.glob("*.tmp")), [])

    def test_fresh_install(self):
        result = self.run_setup()
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertEqual((self.default / "musicalc.exe").read_bytes(), b"new")
        self.assertEqual(list(self.default.glob("*.tmp")), [])
        self.assertEqual(self.read_registry(self.uninstall_key, "DisplayVersion"), "1.0.0-test")

    def test_existing_inno_install_is_migrated_in_place(self):
        self.install_legacy_inno()
        (self.old / "unrelated.txt").write_bytes(b"keep")
        result = self.run_setup()
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertEqual((self.old / "musicalc.exe").read_bytes(), b"new")
        self.assertIsNone(self.read_registry(self.legacy_key, "UninstallString"))
        self.assertEqual(self.read_registry(self.uninstall_key, "InstallLocation"), str(self.old))
        self.assertEqual((self.old / "unrelated.txt").read_bytes(), b"keep")
        self.assertEqual(self.preferences.read_bytes(), b'{"retained":true}')

    def test_existing_nsis_install_is_updated_and_uninstall_preserves_other_files(self):
        self.write_registry(self.app_key, {"InstallDir": self.old})
        (self.old / "unrelated.txt").write_bytes(b"keep")
        self.assertEqual(self.run_setup().returncode, 0)
        self.assertEqual((self.old / "musicalc.exe").read_bytes(), b"new")
        # NSIS requires _?= to be the final unquoted command-line suffix.
        command = '"' + str(self.old / "Uninstall.exe") + '" /S _?=' + str(self.old)
        removed = subprocess.run(command, capture_output=True, text=True, timeout=60)
        self.assertEqual(removed.returncode, 0, removed.stdout + removed.stderr)
        self.assertFalse((self.old / "musicalc.exe").exists())
        self.assertEqual((self.old / "unrelated.txt").read_bytes(), b"keep")
        self.assertEqual(self.preferences.read_bytes(), b'{"retained":true}')
        self.assertIsNone(self.read_registry(self.uninstall_key, "DisplayVersion"))

    def test_failed_or_incomplete_inno_uninstall_stops_before_new_payload(self):
        for exit_code in (7, 0):
            with self.subTest(exit_code=exit_code):
                marker = self.register_fake_uninstaller(exit_code)
                result = self.run_setup()
                self.assertNotEqual(result.returncode, 0)
                self.assertTrue(marker.exists())
                self.assert_not_replaced()

    def test_unwritable_destination_does_not_run_old_uninstaller(self):
        marker = self.register_fake_uninstaller(0)
        destination = self.root / "selected but unwritable"
        destination.mkdir()
        self.setup = self.compile_setup(destination)
        identity = self.run_checked(["whoami", "/user", "/fo", "csv", "/nh"]).stdout
        sid = re.search(r"S-1-[0-9]+(?:-[0-9]+)+", identity).group()
        self.assertTrue(destination.resolve().is_relative_to(self.root))
        self.run_checked(["icacls", destination, "/deny", "*" + sid + ":(WD)"])
        try:
            result = self.run_setup()
        finally:
            self.run_checked(["icacls", destination, "/remove:d", "*" + sid])
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse(marker.exists())
        self.assert_not_replaced()
        self.assertIsNotNone(self.read_registry(self.legacy_key, "UninstallString"))

    def test_locked_app_stops_before_old_uninstaller(self):
        marker = self.register_fake_uninstaller(0)
        create = ctypes.windll.kernel32.CreateFileW
        create.argtypes = [ctypes.c_wchar_p, ctypes.c_uint32, ctypes.c_uint32, ctypes.c_void_p,
                           ctypes.c_uint32, ctypes.c_uint32, ctypes.c_void_p]
        create.restype = ctypes.c_void_p
        close = ctypes.windll.kernel32.CloseHandle
        close.argtypes, close.restype = [ctypes.c_void_p], ctypes.c_int
        handle = create(str(self.old / "musicalc.exe"), 0x80000000, 0, None, 3, 0x80, None)
        self.assertNotEqual(handle, ctypes.c_void_p(-1).value)
        try:
            self.assertNotEqual(self.run_setup().returncode, 0)
        finally:
            close(handle)
        self.assertFalse(marker.exists())
        self.assert_not_replaced()


if __name__ == "__main__":
    unittest.main()
