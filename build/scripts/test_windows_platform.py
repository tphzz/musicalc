from contextlib import ExitStack
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location("windows_platform", Path(__file__).with_name("build-platform.py"))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class WindowsPlatformTests(unittest.TestCase):
    def test_unverified_binary_is_never_copied_to_versioned_output(self):
        with tempfile.TemporaryDirectory(prefix="musicalc platform ") as temporary:
            root = Path(temporary)
            for arch, name in (("amd64", "musicalc_x64.exe"), ("arm64", "musicalc_arm64.exe")):
                for valid in (True, False):
                    with self.subTest(arch=arch, valid=valid), ExitStack() as stack:
                        stack.enter_context(patch.object(module, "ROOT", root))
                        stack.enter_context(patch.object(module.sys, "argv", ["build-platform.py", "windows", arch]))
                        stack.enter_context(patch.object(module, "current", return_value={
                            "version": "0.9.0-snapshot.g1234567", "release": False}))
                        stack.enter_context(patch.object(module, "git", side_effect=lambda *args: str(root)))
                        stack.enter_context(patch.object(module.shutil, "which", return_value="zig"))
                        stack.enter_context(patch.object(module.subprocess, "check_output", side_effect=[
                            f"windows {arch}", "0.17.0"]))
                        stack.enter_context(patch.object(module, "run"))
                        events = []
                        def check(path, selected_arch):
                            events.append(("check", path, selected_arch))
                            if not valid:
                                raise ValueError("wrong GUI subsystem")
                        checker = stack.enter_context(patch.object(module, "verify_windows_binary",
                                                                    side_effect=check, create=True))
                        copier = stack.enter_context(patch.object(module.shutil, "copy2",
                                                                   side_effect=lambda *args: events.append(("copy", *args))))
                        if valid:
                            module.main()
                            self.assertEqual([event[0] for event in events], ["check", "copy"])
                            copier.assert_called_once_with(root / "build/dist" / name,
                                root / f"build/dist/musicalc_0.9.0-snapshot.g1234567_windows_{arch}.exe")
                        else:
                            with self.assertRaisesRegex(ValueError, "GUI subsystem"):
                                module.main()
                            copier.assert_not_called()
                        checker.assert_called_once_with(root / "build/dist" / name, arch)


if __name__ == "__main__":
    unittest.main()
