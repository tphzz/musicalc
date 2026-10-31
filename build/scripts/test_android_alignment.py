import unittest
from android_alignment import verify


class AlignmentTests(unittest.TestCase):
    def test_valid_multiple_segments(self):
        self.assertEqual(verify(" LOAD 0 0 0 0 R 0x4000\n LOAD 0 0 0 0 RW 0x10000"), [16384,65536])

    def test_empty_failed_or_malformed_output_cannot_pass(self):
        for output in ("", "readelf: invalid file", " LOAD 0 0 R bogus", " LOAD 0 R 0x1000",
                       " LOAD 0 R 0x4001"):
            with self.subTest(output=output), self.assertRaises(ValueError):
                verify(output)


if __name__ == "__main__":
    unittest.main()
