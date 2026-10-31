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


    def test_automatic_new_version_and_state_use_apple_high_water_mark(self):
        self.assertEqual(reserve(self.path, "io.example.app", "0.9.0", minimum=0), 1)
        self.assertEqual(reserve(self.path, "io.example.app", "0.9.0", minimum=0), 2)
        self.assertEqual(reserve(self.path, "io.example.app", "0.10.0", minimum=0), 1)
        self.assertEqual(reserve(self.path, "io.example.app", "0.10.0", minimum=7), 8)
        state = json.loads(self.path.read_text(encoding="utf-8"))
        self.assertEqual(state, {"bundle": "io.example.app", "versions": {"0.9.0": 2, "0.10.0": 8}})

    def test_automatic_existing_version_preserves_local_reservations(self):
        reserve(self.path, "io.example.app", "0.9.0", 12)
        self.assertEqual(reserve(self.path, "io.example.app", "0.9.0", minimum=3), 13)
        self.assertEqual(reserve(self.path, "io.example.app", "0.9.0", minimum=18), 19)

    def test_automatic_recovery_and_exhaustion_do_not_replace_good_state(self):
        reserve(self.path, "io.example.app", "0.9.0", 12)
        before = self.path.read_bytes()
        for value in (True, -1, 10000, "12", 9999):
            with self.subTest(value=value), self.assertRaises(ValueError):
                reserve(self.path, "io.example.app", "0.9.0", minimum=value)
            self.assertEqual(self.path.read_bytes(), before)
        with self.assertRaises(ValueError):
            reserve(self.path, "wrong.bundle", "0.9.0", minimum=0)
        self.assertEqual(self.path.read_bytes(), before)

    def test_automatic_allocation_rejects_corrupt_local_version(self):
        self.path.write_text('{"bundle":"io.example.app","versions":{"0.9.0":"12"}}', encoding="utf-8")
        before = self.path.read_bytes()
        with self.assertRaises(ValueError):
            reserve(self.path, "io.example.app", "0.9.0", minimum=0)
        self.assertEqual(self.path.read_bytes(), before)

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

    def test_bundle_mismatch_reports_both_ids_without_changing_state(self):
        reserve(self.path, "com.musicalc", "0.9.0", 0)
        before = self.path.read_bytes()
        with self.assertRaises(ValueError) as stopped:
            reserve(self.path, "io.example.musicalc", "0.9.0")
        message = str(stopped.exception)
        self.assertIn(str(self.path), message)
        self.assertIn("com.musicalc", message)
        self.assertIn("io.example.musicalc", message)
        self.assertEqual(self.path.read_bytes(), before)

    def test_invalid_counter_shape_reports_recovery_without_changing_state(self):
        for state in ([], {"bundle": "com.musicalc", "versions": []}):
            with self.subTest(state=state):
                self.path.write_text(json.dumps(state), encoding="utf-8")
                before = self.path.read_bytes()
                with self.assertRaisesRegex(ValueError, "restore|reconcile"):
                    reserve(self.path, "com.musicalc", "0.9.0")
                self.assertEqual(self.path.read_bytes(), before)

    def test_corrupt_or_exhausted_state_fails(self):
        self.path.write_text('{"bundle":"com.musicalc","versions":{"0.9.0":9999}}')
        with self.assertRaises(ValueError):
            reserve(self.path, "com.musicalc", "0.9.0")
        self.path.write_text("bad json")
        with self.assertRaises(ValueError):
            reserve(self.path, "com.musicalc", "0.9.0")


if __name__ == "__main__":
    unittest.main()
