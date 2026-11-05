import importlib.util
import os
from pathlib import Path
import sys
import unittest
from unittest.mock import patch


def load(name, filename):
    spec = importlib.util.spec_from_file_location(name, Path(__file__).with_name(filename))
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


platform = load("build_platform", "build-platform.py")
DATA = {"version":"0.9.0","marketing_version":"0.9.0","release":True,"commit":"a"*40,"build":1}


class CommandTests(unittest.TestCase):
    def test_selected_tag_is_forwarded_to_goreleaser(self):
        # The selected tag remains authoritative even if another tag aliases HEAD.
        with patch.object(sys,"argv",["build-platform.py","linux","amd64"]), \
             patch.object(platform,"current",return_value=DATA), \
             patch.object(platform.subprocess,"check_output",return_value="linux\namd64\n"), \
             patch.object(platform,"run") as run:
            platform.main()
        calls = [call for call in run.call_args_list if call.args[0] == "goreleaser"]
        self.assertTrue(calls)
        for call in calls:
            self.assertEqual(call.kwargs["env"]["GORELEASER_CURRENT_TAG"],"0.9.0")

    def test_linux_missing_wayland_stops_before_tests_and_packaging(self):
        for arch in ("amd64", "arm64"):
            for missing in ("wayland-client", "wayland-cursor", "wayland-egl", "xkbcommon"):
                commands = []

                def run(*args, **kwargs):
                    commands.append(args)
                    if args[0] == "pkg-config" and missing in args:
                        raise platform.subprocess.CalledProcessError(1, args)

                with self.subTest(arch=arch, missing=missing), \
                     patch.object(sys, "argv", ["build-platform.py", "linux", arch]), \
                     patch.object(platform, "current", return_value=DATA), \
                     patch.object(platform.subprocess, "check_output", return_value=f"linux\n{arch}\n"), \
                     patch.object(platform, "run", side_effect=run), \
                     self.assertRaises(platform.subprocess.CalledProcessError):
                    platform.main()
                self.assertEqual(len(commands), 1, "Do not test or package with missing native libraries")

    def test_windows_targets_build_with_supported_zig(self):
        for zig_version in ("0.16.0", "0.17.0"):
            for arch, target in (("amd64", "x86_64"), ("arm64", "aarch64")):
                with self.subTest(arch=arch, zig_version=zig_version), \
                     patch.object(sys, "argv", ["build-platform.py", "windows", arch]), \
                     patch.object(platform, "current", return_value=DATA), \
                     patch.object(platform.subprocess, "check_output", side_effect=[f"windows\n{arch}\n", zig_version + "\n"]), \
                     patch.object(platform.shutil, "which", side_effect=lambda name: "zig" if name == "zig" else None), \
                     patch.object(platform.shutil, "copy2"), \
                     patch.object(platform, "verify_windows_binary"), \
                     patch.object(platform, "git", side_effect=[str(platform.ROOT / ".git"), str(platform.ROOT)]), \
                     patch.dict(os.environ, {"GOFLAGS": ""}), \
                     patch.object(platform, "run") as run:
                    platform.main()
                    self.assertEqual(run.call_args_list[0].args, ("go", "test", "./..."))
                    for call in run.call_args_list:
                        env = call.kwargs["env"]
                        self.assertEqual(env["CC"], f"zig cc -target {target}-windows-gnu")
                        self.assertEqual(env["CXX"], f"zig c++ -target {target}-windows-gnu")
                        self.assertEqual(env["CGO_ENABLED"], "1")
                        self.assertEqual(env["GIT_DIR"], str(platform.ROOT / ".git"))
                        self.assertEqual(env["GIT_WORK_TREE"], str(platform.ROOT))
                        self.assertNotIn("mingw32-gcc", env.get("GOFLAGS", ""))
                        self.assertTrue(Path(env["ZIG_GLOBAL_CACHE_DIR"]).is_absolute())
                        self.assertTrue(Path(env["ZIG_LOCAL_CACHE_DIR"]).is_absolute())
                    build = run.call_args_list[-1].args
                    self.assertEqual(build[:2], ("goreleaser", "build"))
                    self.assertEqual(build[-2:], ("--id", "win-" + arch))

    def test_missing_or_wrong_zig_stops_windows_before_building(self):
        for arch in ("amd64", "arm64"):
            for installed, version, message in ((False, "", "Missing Zig"),
                                                (True, "0.15.2\n", "Expected Zig 0.16.0 or 0.17.0"),
                                                (True, "0.18.0\n", "Expected Zig 0.16.0 or 0.17.0"),
                                                (True, "0.17.0-dev.1\n", "Expected Zig 0.16.0 or 0.17.0")):
                with self.subTest(arch=arch, installed=installed), \
                     patch.object(sys, "argv", ["build-platform.py", "windows", arch]), \
                     patch.object(platform, "current", return_value=DATA), \
                     patch.object(platform.subprocess, "check_output", side_effect=[f"windows\n{arch}\n", version]), \
                     patch.object(platform.shutil, "which", return_value="zig" if installed else None), \
                     patch.object(platform, "run") as run, \
                     patch("sys.stderr") as stderr, self.assertRaises(SystemExit) as failure:
                    platform.main()
                self.assertEqual(failure.exception.code, 2)
                self.assertIn(message, "".join(call.args[0] for call in stderr.write.call_args_list))
                run.assert_not_called()

if __name__ == "__main__":
    unittest.main()
