from concurrent.futures import ThreadPoolExecutor
import json
from pathlib import Path
import tempfile
import unittest
from ios_build_number import reserve


class CounterTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.path = Path(self.temp.name) / "counter.json"

    def test_missing_and_new_versions_need_explicit_seed(self):
        with self.assertRaises(ValueError):
            reserve(self.path, "com.musicalc", "0.9.0")
        self.assertEqual(reserve(self.path, "com.musicalc", "0.9.0", 0), 0)
        self.assertEqual(reserve(self.path, "com.musicalc", "0.9.0"), 1)
        with self.assertRaises(ValueError):
            reserve(self.path, "com.musicalc", "0.9.1")

    def test_concurrent_reservations_are_unique_and_durable(self):
        reserve(self.path, "com.musicalc", "0.9.0", 0)
        with ThreadPoolExecutor(max_workers=4) as pool:
            values = list(pool.map(lambda _: reserve(self.path, "com.musicalc", "0.9.0"), range(12)))
        self.assertEqual(sorted(values), list(range(1, 13)))
        self.assertEqual(reserve(self.path, "com.musicalc", "0.9.0"), 13)

    def test_recovery_cannot_reuse_reserved_numbers(self):
        reserve(self.path, "com.musicalc", "0.9.0", 20)
        with self.assertRaises(ValueError):
            reserve(self.path, "com.musicalc", "0.9.0", 19)
        self.assertEqual(reserve(self.path, "com.musicalc", "0.9.0"), 21)
        with self.assertRaises(ValueError):
            reserve(self.path, "other.bundle", "0.9.0")

    def test_corrupt_or_exhausted_state_fails(self):
        self.path.write_text('{"bundle":"com.musicalc","versions":{"0.9.0":9999}}')
        with self.assertRaises(ValueError):
            reserve(self.path, "com.musicalc", "0.9.0")
        self.path.write_text("bad json")
        with self.assertRaises(ValueError):
            reserve(self.path, "com.musicalc", "0.9.0")


if __name__ == "__main__":
    unittest.main()
