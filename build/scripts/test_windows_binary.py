from pathlib import Path
import struct
import subprocess
import sys
import tempfile
import unittest

from windows_binary import verify_windows_binary


CHECKER = Path(__file__).with_name("windows_binary.py")


def pe_header(machine=0x8664, subsystem=2):
    data = bytearray(512)
    data[:2] = b"MZ"
    struct.pack_into("<I", data, 0x3c, 128)
    data[128:132] = b"PE\0\0"
    struct.pack_into("<HHIIIHH", data, 132, machine, 0, 0, 0, 0, 240, 0x22)
    struct.pack_into("<H", data, 152, 0x20b)
    struct.pack_into("<H", data, 152 + 68, subsystem)
    return data


class WindowsBinaryTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory(prefix="musicalc PE ")
        self.addCleanup(temporary.cleanup)
        self.executable = Path(temporary.name) / "app with spaces.exe"

    def test_gui_headers_for_both_architectures(self):
        for arch, machine in (("amd64", 0x8664), ("arm64", 0xaa64)):
            with self.subTest(arch=arch):
                self.executable.write_bytes(pe_header(machine))
                verify_windows_binary(self.executable, arch)

    def test_console_and_other_subsystems_are_rejected(self):
        for arch, machine in (("amd64", 0x8664), ("arm64", 0xaa64)):
            for subsystem in (0, 1, 3):
                with self.subTest(arch=arch, subsystem=subsystem):
                    self.executable.write_bytes(pe_header(machine, subsystem))
                    with self.assertRaisesRegex(ValueError, "GUI subsystem.*" + str(subsystem)):
                        verify_windows_binary(self.executable, arch)

    def test_wrong_architecture_is_rejected(self):
        for arch, machine in (("amd64", 0xaa64), ("arm64", 0x8664), ("amd64", 0x14c)):
            with self.subTest(arch=arch, machine=machine):
                self.executable.write_bytes(pe_header(machine))
                with self.assertRaisesRegex(ValueError, "machine.*" + arch):
                    verify_windows_binary(self.executable, arch)

    def test_malformed_or_truncated_headers_are_rejected(self):
        bad_dos = pe_header()
        bad_dos[:2] = b"NO"
        bad_offset = pe_header()
        struct.pack_into("<I", bad_offset, 0x3c, 0xffffffff)
        bad_signature = pe_header()
        bad_signature[128:132] = b"NONE"
        bad_magic = pe_header()
        struct.pack_into("<H", bad_magic, 152, 0x10b)
        short_optional = pe_header()
        struct.pack_into("<H", short_optional, 128 + 20, 60)
        cases = [b"", b"MZ", bad_dos, bad_offset, bad_signature, bad_magic,
                 short_optional, pe_header()[:140], pe_header()[:300]]
        for index, data in enumerate(cases):
            with self.subTest(case=index):
                self.executable.write_bytes(data)
                with self.assertRaises(ValueError):
                    verify_windows_binary(self.executable, "amd64")

    def test_cli_reports_success_and_failure_without_tracebacks(self):
        for subsystem, expected in ((2, 0), (3, 1)):
            with self.subTest(subsystem=subsystem):
                self.executable.write_bytes(pe_header(subsystem=subsystem))
                result = subprocess.run([sys.executable, "-B", str(CHECKER), str(self.executable), "amd64"],
                                        capture_output=True, text=True, timeout=15)
                self.assertEqual(result.returncode, expected, result.stderr)
                self.assertNotIn("Traceback", result.stderr)
                if expected:
                    self.assertIn(str(self.executable), result.stderr)
                    self.assertIn("GUI subsystem", result.stderr)

    def test_missing_file_is_reported(self):
        result = subprocess.run([sys.executable, "-B", str(CHECKER), str(self.executable), "amd64"],
                                capture_output=True, text=True, timeout=15)
        self.assertEqual(result.returncode, 1)
        self.assertIn(str(self.executable), result.stderr)
        self.assertNotIn("Traceback", result.stderr)


if __name__ == "__main__":
    unittest.main()
