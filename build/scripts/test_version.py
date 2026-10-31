import unittest
from version import resolve

SHA = "a" * 40


class VersionTests(unittest.TestCase):
    def test_release_and_numeric_branch_are_distinct(self):
        self.assertEqual(resolve("tag", "0.9.0", SHA)["version"], "0.9.0")
        data = resolve("branch", "0.9.0", SHA)
        self.assertEqual(data["version"], "0.9.0-snapshot.gaaaaaaa")
        self.assertFalse(data["release"])

    def test_branch_fallback_and_empty_override(self):
        self.assertEqual(resolve("branch", "main", SHA, "0.8.8", "")["version"],
                         "0.8.8-snapshot.gaaaaaaa")
        self.assertEqual(resolve("branch", "feature/a", SHA)["version"],
                         "0.0.0-snapshot.gaaaaaaa")

    def test_invalid_tags_and_release_override_fail(self):
        for tag in ("v0.9.0", "0.9", "0.9.0-beta", "01.2.3", "1.2.03", "1.2.3\n", "１.2.3"):
            with self.subTest(tag=tag), self.assertRaises(ValueError):
                resolve("tag", tag, SHA)
        with self.assertRaises(ValueError):
            resolve("tag", "0.9.0", SHA, override="0.9.0-test")

    def test_snapshot_override_is_validated(self):
        self.assertEqual(resolve("branch", "main", SHA, override="1.0.0-rc.2")["marketing_version"], "1.0.0")
        for value in ("dev", "1.2.3", "1.2.3-a.01", "1.2.3-a/$()", "1.2.3-a\n"):
            with self.subTest(value=value), self.assertRaises(ValueError):
                resolve("branch", "main", SHA, override=value)


if __name__ == "__main__":
    unittest.main()
