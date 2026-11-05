from datetime import datetime, timezone
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

import ios_build_number as numbers


class BuildNumberTests(unittest.TestCase):
    def test_utc_calendar_boundaries_produce_ordered_packager_integers(self):
        moments = [datetime(2026, 11, 2, 23, 59, 59, tzinfo=timezone.utc),
                   datetime(2026, 11, 3, 0, 0, 0, tzinfo=timezone.utc),
                   datetime(2026, 12, 31, 23, 59, 59, tzinfo=timezone.utc),
                   datetime(2027, 1, 1, 0, 0, 0, tzinfo=timezone.utc)]
        actual = []
        with patch.object(numbers, "datetime") as clock, patch.object(numbers.time, "sleep"):
            for moment in moments:
                clock.now.return_value = moment
                actual.append(numbers.current_build())
            clock.now.assert_called_with(timezone.utc)
        self.assertEqual(actual, [20261102235959, 20261103000000, 20261231235959, 20270101000000])
        self.assertEqual(actual, sorted(actual))
        self.assertTrue(all(0 < value < 2**63 for value in actual))

    def test_pause_finishes_before_the_fresh_utc_sample(self):
        events = []
        moment = datetime(2026, 11, 2, 12, 0, 0, tzinfo=timezone.utc)
        def read_clock(zone):
            events.append("sample")
            self.assertEqual(zone, timezone.utc)
            return moment
        with patch.object(numbers.time, "sleep", side_effect=lambda delay: events.append(("pause", delay))), \
             patch.object(numbers, "datetime") as clock:
            clock.now.side_effect = read_clock
            self.assertEqual(numbers.current_build(), 20261102120000)
        self.assertEqual(events, [("pause", 1), "sample"])

    def test_real_sequential_attempts_need_no_state_and_leave_legacy_files_untouched(self):
        with tempfile.TemporaryDirectory(prefix="ios build numbers ") as temporary:
            folder = Path(temporary)
            legacy = folder / "build-numbers.json"
            original = b"not JSON; old bundle and reservations are irrelevant"
            legacy.write_bytes(original)
            command = [sys.executable, "-B", str(Path(numbers.__file__))]
            results = [subprocess.run(command, cwd=folder, capture_output=True, text=True) for _ in range(2)]
            for result in results:
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertRegex(result.stdout.strip(), r"^[0-9]{14}$")
            self.assertLess(int(results[0].stdout), int(results[1].stdout))
            self.assertEqual(legacy.read_bytes(), original)
            self.assertEqual(list(folder.iterdir()), [legacy])

    def test_obsolete_seed_and_migration_commands_fail_without_writing_files(self):
        with tempfile.TemporaryDirectory() as temporary:
            journal = Path(temporary) / "build-numbers.json"
            for operation in (["--version", "0.9.0", "--seed", "0"],
                              ["--migrate-bundle-from", "com.musicalc"]):
                command = [sys.executable, "-B", str(Path(numbers.__file__)), "--file", str(journal),
                           "--bundle", "io.example.musicalc", *operation]
                result = subprocess.run(command, capture_output=True, text=True)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(list(journal.parent.iterdir()), [])


if __name__ == "__main__":
    unittest.main()
