import hashlib
from pathlib import Path
import struct
import subprocess
import sys
import tempfile
import unittest


@unittest.skipUnless(sys.platform == "darwin", "requires macOS AppKit and Xcode")
class IPhoneIconTests(unittest.TestCase):
    def test_real_icon_helper_creates_opaque_1024_png(self):
        root = Path(__file__).resolve().parents[2]
        source = root / "icons/appicon.png"
        original = hashlib.sha256(source.read_bytes()).digest()
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "icon.png"
            result = subprocess.run(
                ["xcrun", "swift", str(root / "build/scripts/ios-icon.swift"), str(source), str(output)],
                capture_output=True, text=True, timeout=120,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            data = output.read_bytes()
            self.assertEqual(data[:8], b"\x89PNG\r\n\x1a\n")
            self.assertEqual(data[12:16], b"IHDR")
            width, height, depth, color = struct.unpack(">IIBB", data[16:26])
            self.assertEqual((width, height, depth, color), (1024, 1024, 8, 2))
        self.assertEqual(hashlib.sha256(source.read_bytes()).digest(), original)


if __name__ == "__main__":
    unittest.main()
