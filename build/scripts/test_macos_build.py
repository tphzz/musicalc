import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest


SCRIPT = Path(__file__).with_name("build-macos.sh")


@unittest.skipUnless(sys.platform == "darwin", "Native macOS tools are required")
class MacOSArchitectureTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        temporary = tempfile.TemporaryDirectory(prefix="musicalc macOS architecture ")
        cls.addClassCleanup(temporary.cleanup)
        cls.root = Path(temporary.name)
        source = cls.root / "slice.c"
        source.write_text("int musicalc_fixture(void) { return 0; }\n", encoding="utf-8")
        for arch in ("arm64", "x86_64"):
            subprocess.run(
                ["xcrun", "--sdk", "macosx", "clang", "-arch", arch, "-c", str(source),
                 "-o", str(cls.root / arch)],
                check=True, timeout=60,
            )
        subprocess.run(
            ["lipo", "-create", str(cls.root / "arm64"), str(cls.root / "x86_64"),
             "-output", str(cls.root / "universal")],
            check=True, timeout=60,
        )
        # Execute the packaging script's verification commands with real Mach-O files.
        cls.body = "\n".join(
            line for line in SCRIPT.read_text(encoding="utf-8").splitlines()
            if line.startswith("lipo ") and "-verify_arch" in line
        )
        if not cls.body:
            raise AssertionError("Packaging must verify the executable architectures")

    def verify(self, fixture):
        app = self.root / self._testMethodName / "MusiCalc.app"
        executable = app / "Contents/MacOS/MusiCalc"
        executable.parent.mkdir(parents=True)
        shutil.copyfile(self.root / fixture, executable)
        return subprocess.run(
            ["/bin/bash", "--noprofile", "--norc", "-e", "-o", "pipefail", "-c", self.body],
            env=dict(os.environ, app=str(app)), capture_output=True, text=True, timeout=30,
        )

    def test_universal_binary_passes(self):
        result = self.verify("universal")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_missing_arm64_slice_fails(self):
        result = self.verify("x86_64")
        self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_missing_intel_slice_fails(self):
        result = self.verify("arm64")
        self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)


if __name__ == "__main__":
    unittest.main()
