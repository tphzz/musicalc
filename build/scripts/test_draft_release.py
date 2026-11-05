import contextlib
import copy
import hashlib
import importlib.util
import io
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch
from urllib.parse import parse_qs, unquote, urlsplit


spec = importlib.util.spec_from_file_location("draft_release", Path(__file__).with_name("draft-release.py"))
draft = importlib.util.module_from_spec(spec)
spec.loader.exec_module(draft)
REPOSITORY = "example/musicalc"
COMMIT = "a" * 40


def package_names(version):
    return {
        f"musicalc_{version}_windows_amd64.exe", f"musicalc_{version}_windows_arm64.exe",
        f"musicalc-setup-{version}.exe", f"musicalc_{version}_linux_amd64v3.tar.gz",
        f"musicalc_{version}_linux_arm64.tar.gz", f"MusiCalc_{version}_universal.app.zip",
        f"musicalc_{version}_universal.dmg", f"musicalc_{version}_android_arm64.apk",
        *(f"musicalc_{version}_{arch}.{extension}" for arch in ("amd64", "arm64")
          for extension in ("deb", "rpm")),
    }


def release_package_names(version):
    return {
        f"musicalc-setup-{version}.exe", f"musicalc_{version}_universal.dmg",
        f"musicalc_{version}_android_arm64.apk", f"musicalc_{version}_linux_amd64v3.tar.gz",
        f"musicalc_{version}_linux_arm64.tar.gz",
        *(f"musicalc_{version}_{arch}.{extension}" for arch in ("amd64", "arm64")
          for extension in ("deb", "rpm")),
    }


class GitHub:
    """In-memory release API; file bytes exercise uploads and asset digest verification."""

    def __init__(self):
        self.releases = []
        self.tags = {"0.9.0": {"type": "commit", "sha": COMMIT}}
        self.tag_objects = {}
        self.mutations = []
        self.uploads = []
        self.fail_upload = False
        self.fail_delete = False
        self.corrupt_upload = False
        self.publish_on_upload = False
        self.next_asset = 1

    def add_release(self, tag, **fields):
        release = dict(id=len(self.releases) + 1, tag_name=tag, draft=True, body="Original notes",
                       target_commitish=COMMIT, assets=[], name="Original title",
                       html_url=f"https://github.com/{REPOSITORY}/releases/untagged-{len(self.releases) + 1}")
        release.update(fields)
        self.releases.append(release)
        return release

    def api(self, endpoint, method="GET", payload=None):
        parsed = urlsplit(endpoint)
        path = parsed.path.removeprefix(f"repos/{REPOSITORY}/")
        page = int(parse_qs(parsed.query).get("page", ["1"])[0])
        if method != "GET":
            self.mutations.append((method, path, copy.deepcopy(payload)))
        if path == "releases":
            if method == "GET":
                return copy.deepcopy(self.releases[(page - 1) * 100:page * 100])
            if method == "POST":
                return copy.deepcopy(self.add_release(**dict(payload, tag=payload["tag_name"])))
        if path == "releases/generate-notes":
            return {"body": "Generated release notes"}
        if path.startswith("git/matching-refs/tags/"):
            prefix = unquote(path.removeprefix("git/matching-refs/tags/"))
            return [{"ref": "refs/tags/" + tag, "object": obj}
                    for tag, obj in self.tags.items() if tag.startswith(prefix)]
        if path.startswith("git/tags/"):
            return {"object": self.tag_objects[path.rsplit("/", 1)[1]]}
        parts = path.split("/")
        if parts[:2] == ["releases", "assets"]:
            if self.fail_delete:
                raise RuntimeError("asset deletion failed")
            for release in self.releases:
                release["assets"][:] = [asset for asset in release["assets"] if asset["id"] != int(parts[2])]
            return None
        if len(parts) >= 2 and parts[0] == "releases":
            release = next(release for release in self.releases if release["id"] == int(parts[1]))
            if len(parts) == 3 and parts[2] == "assets":
                return copy.deepcopy(release["assets"][(page - 1) * 100:page * 100])
            if method == "PATCH":
                release.update(payload)
            elif method == "DELETE":
                self.releases.remove(release)
                return None
            return copy.deepcopy(release)
        raise AssertionError((method, endpoint, payload))

    def gh(self, *args, **kwargs):
        if args[:2] != ("release", "upload"):
            raise AssertionError(args)
        release = next(release for release in self.releases if release["tag_name"] == args[2])
        for value in args[5:]:
            if value == "--clobber":
                continue
            path = Path(value)
            content = path.read_bytes()
            self.uploads.append(path.name)
            if self.fail_upload and len(self.uploads) > 1:
                self.fail_upload = False
                raise RuntimeError("upload interrupted")
            release["assets"][:] = [asset for asset in release["assets"] if asset["name"] != path.name]
            release["assets"].append(dict(id=self.next_asset, name=path.name, size=len(content),
                state="uploaded", digest="sha256:" + hashlib.sha256(content).hexdigest()))
            self.next_asset += 1
        if self.corrupt_upload:
            release["assets"][-1]["digest"] = "sha256:" + "0" * 64
        if self.publish_on_upload:
            release["draft"] = False
        return ""


class DraftReleaseTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix="musicalc releases ")
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name) / "artifacts"
        self.root.mkdir()
        self.github = GitHub()
        self.ref = "main"
        self.data = dict(version="0.9.0-snapshot.gaaaaaaa", marketing_version="0.9.0",
                         release=False, commit=COMMIT, build=1)
        self.write_packages()

    def write_packages(self):
        for path in self.root.iterdir():
            path.unlink()
        for name in package_names(self.data["version"]):
            (self.root / name).write_bytes((self.data["commit"] + name).encode())

    def run_draft(self):
        with patch.object(sys, "argv", ["draft-release.py", str(self.root)]), \
             patch.dict(os.environ, {"GITHUB_REPOSITORY": REPOSITORY, "GITHUB_REF_NAME": self.ref,
                 "GITHUB_REF_TYPE": "tag" if self.data["release"] else "branch",
                 "GITHUB_RUN_ID": "123", "GITHUB_SERVER_URL": "https://github.com",
                 "GITHUB_STEP_SUMMARY": str(self.root.parent / "summary.md")}), \
             patch.object(draft, "current", return_value=self.data), \
             patch.object(draft, "api", side_effect=self.github.api, create=True), \
             patch.object(draft, "gh", side_effect=self.github.gh), \
             patch.object(draft.subprocess, "run", side_effect=AssertionError("Unexpected process")), \
             contextlib.redirect_stdout(io.StringIO()) as output:
            draft.main()
        return output.getvalue()

    def select_tag(self):
        self.ref = "0.9.0"
        self.data.update(version="0.9.0", release=True)
        self.write_packages()

    def test_new_branch_uploads_nine_packages_and_keeps_all_build_inputs(self):
        output = self.run_draft()
        release = self.github.releases[0]
        self.assertEqual(release["tag_name"], "snapshot/main")
        self.assertEqual(release["target_commitish"], COMMIT)
        self.assertTrue(release["draft"])
        self.assertTrue(release["prerelease"])
        self.assertIn(self.data["version"], release["body"])
        self.assertIn(COMMIT, release["body"])
        self.assertIn("actions/runs/123", release["body"])
        self.assertNotIn("incomplete", release["body"].lower())
        self.assertIn(release["html_url"], output)
        checksum = self.root / f"musicalc_{self.data['version']}_SHA256SUMS.txt"
        self.assertFalse(checksum.exists())
        self.assertEqual({asset["name"] for asset in release["assets"]}, release_package_names(self.data["version"]))
        self.assertEqual(len(release["assets"]), 9)
        for asset in release["assets"]:
            self.assertEqual(asset["digest"], "sha256:" + hashlib.sha256((self.root / asset["name"]).read_bytes()).hexdigest())
        self.assertTrue(all((self.root / name).is_file() for name in package_names(self.data["version"])))

    def test_branch_refresh_removes_old_assets_and_retains_draft_identity(self):
        self.run_draft()
        original = self.github.releases[0]["id"]
        self.data.update(commit="b" * 40, version="0.9.0-snapshot.gbbbbbbb")
        self.write_packages()
        self.run_draft()
        self.assertEqual(len(self.github.releases), 1)
        release = self.github.releases[0]
        self.assertEqual(release["id"], original)
        self.assertEqual(release["target_commitish"], "b" * 40)
        self.assertEqual({asset["name"] for asset in release["assets"]}, release_package_names(self.data["version"]))
        self.assertTrue(all("gaaaaaaa" not in asset["name"] for asset in release["assets"]))

    def test_same_commit_branch_rerun_replaces_bytes_without_a_checksum_manifest(self):
        self.run_draft()
        path = self.root / next(iter(release_package_names(self.data["version"])))
        path.write_bytes(b"rebuilt output")
        self.run_draft()
        release = self.github.releases[0]
        self.assertEqual(len(release["assets"]), 9)
        self.assertEqual(next(asset for asset in release["assets"] if asset["name"] == path.name)["size"], 14)
        self.assertFalse((self.root / f"musicalc_{self.data['version']}_SHA256SUMS.txt").exists())

    def test_branch_names_and_numeric_tag_have_separate_identities(self):
        for ref in ("main", "0.9.0", "feature/äudio", "feature-audio"):
            self.ref = ref
            self.write_packages()
            self.run_draft()
        self.select_tag()
        self.run_draft()
        self.assertEqual({release["tag_name"] for release in self.github.releases},
                         {"snapshot/main", "snapshot/0.9.0", "snapshot/feature/äudio", "snapshot/feature-audio", "0.9.0"})

    def test_new_tag_is_created_but_existing_draft_or_published_release_is_preserved(self):
        self.select_tag()
        self.run_draft()
        release = self.github.releases[0]
        self.assertEqual(release["tag_name"], "0.9.0")
        self.assertFalse(release["prerelease"])
        self.assertIn("Generated release notes", release["body"])
        self.assertEqual({asset["name"] for asset in release["assets"]}, release_package_names(self.data["version"]))
        for published in (False, True):
            release["draft"] = not published
            before = copy.deepcopy(self.github.releases)
            self.github.mutations.clear()
            self.github.uploads.clear()
            output = self.run_draft()
            self.assertIn("Preserved", output)
            self.assertEqual(before, self.github.releases)
            self.assertFalse(self.github.mutations)
            self.assertFalse(self.github.uploads)

    def test_paginated_lookup_finds_existing_release_after_first_page(self):
        for number in range(105):
            self.github.add_release(f"old-{number}")
        self.select_tag()
        self.github.add_release("0.9.0")
        self.assertIn("Preserved", self.run_draft())
        self.assertFalse(self.github.mutations)

    def test_ambiguous_drafts_do_not_get_modified(self):
        self.github.add_release("snapshot/main")
        self.github.add_release("snapshot/main")
        with self.assertRaisesRegex(ValueError, "Multiple"):
            self.run_draft()
        self.assertFalse(self.github.mutations)

    def test_published_branch_or_existing_snapshot_tag_is_rejected(self):
        for actual_tag in (False, True):
            with self.subTest(actual_tag=actual_tag):
                self.github = GitHub()
                if actual_tag:
                    self.github.tags["snapshot/main"] = {"type": "commit", "sha": COMMIT}
                else:
                    self.github.add_release("snapshot/main", draft=False)
                with self.assertRaisesRegex(ValueError, "Published|tag already exists"):
                    self.run_draft()
                self.assertFalse(self.github.mutations)

    def test_wrong_or_missing_remote_tag_stops_before_creation(self):
        self.select_tag()
        for tags in ({}, {"0.9.0": {"type": "commit", "sha": "b" * 40}}):
            self.github.tags = tags
            with self.assertRaisesRegex(ValueError, "tag.*built commit"):
                self.run_draft()
            self.assertFalse(self.github.mutations)

    def test_annotated_release_tag_is_peeled(self):
        self.select_tag()
        self.github.tags["0.9.0"] = {"type": "tag", "sha": "c" * 40}
        self.github.tag_objects["c" * 40] = {"type": "commit", "sha": COMMIT}
        self.run_draft()
        self.assertEqual(self.github.releases[0]["target_commitish"], COMMIT)

    def test_every_build_package_is_required_before_any_mutation(self):
        for name in package_names(self.data["version"]):
            self.write_packages()
            (self.root / name).unlink()
            with self.subTest(missing=name), self.assertRaisesRegex(ValueError, "artifact"):
                self.run_draft()
            self.assertFalse(self.github.mutations)

    def test_unexpected_versioned_files_and_duplicate_names_are_rejected(self):
        for name in ("secret.p8", "musicalc_0.9.0_old.exe", "readme.txt"):
            self.write_packages()
            (self.root / name).write_bytes(b"unexpected")
            with self.subTest(name=name), self.assertRaisesRegex(ValueError, "artifact"):
                self.run_draft()
        self.write_packages()
        nested = self.root / "nested"
        nested.mkdir()
        duplicate_name = next(iter(package_names(self.data["version"])))
        (nested / duplicate_name).write_bytes(b"duplicate")
        with self.assertRaisesRegex(ValueError, "Duplicate"):
            self.run_draft()
        self.assertFalse(self.github.mutations)

    def test_failed_branch_upload_keeps_old_assets_and_retry_repairs_draft(self):
        self.run_draft()
        old_names = {asset["name"] for asset in self.github.releases[0]["assets"]}
        self.data.update(commit="b" * 40, version="0.9.0-snapshot.gbbbbbbb")
        self.write_packages()
        self.github.uploads.clear()
        self.github.fail_upload = True
        with self.assertRaisesRegex(RuntimeError, "upload interrupted"):
            self.run_draft()
        self.assertIn("incomplete", self.github.releases[0]["body"].lower())
        self.assertTrue(old_names <= {asset["name"] for asset in self.github.releases[0]["assets"]})
        self.run_draft()
        self.assertEqual(len(self.github.releases[0]["assets"]), 9)
        self.assertNotIn("incomplete", self.github.releases[0]["body"].lower())

    def test_failed_new_tag_upload_removes_only_new_draft_and_preserves_git_tag(self):
        self.select_tag()
        untouched = self.github.add_release("unrelated")
        self.github.fail_upload = True
        with self.assertRaisesRegex(RuntimeError, "upload interrupted"):
            self.run_draft()
        self.assertEqual(self.github.releases, [untouched])
        self.assertIn("0.9.0", self.github.tags)
        self.run_draft()
        self.assertEqual(len(self.github.releases), 2)

    def test_corrupt_upload_or_cleanup_failure_never_marks_branch_ready(self):
        for cleanup in (False, True):
            self.github = GitHub()
            self.data.update(version="0.9.0-snapshot.gaaaaaaa", commit=COMMIT)
            self.write_packages()
            self.run_draft()
            self.data.update(version="0.9.0-snapshot.gbbbbbbb", commit="b" * 40)
            self.write_packages()
            self.github.corrupt_upload = not cleanup
            self.github.fail_delete = cleanup
            with self.subTest(cleanup=cleanup), self.assertRaisesRegex(RuntimeError, "asset|digest"):
                self.run_draft()
            self.assertIn("incomplete", self.github.releases[0]["body"].lower())

    def test_release_published_during_refresh_is_never_finalized_or_deleted(self):
        self.github.publish_on_upload = True
        with self.assertRaisesRegex(ValueError, "Published"):
            self.run_draft()
        self.assertFalse(any(method == "DELETE" for method, _, _ in self.github.mutations))

    def test_lookup_auth_or_network_error_is_not_treated_as_missing(self):
        with patch.object(self.github, "api", side_effect=RuntimeError("HTTP 403")), \
             self.assertRaisesRegex(RuntimeError, "HTTP 403"):
            self.run_draft()
        self.assertFalse(self.github.uploads)

    def test_zero_length_package_is_rejected_before_mutation(self):
        (self.root / next(iter(package_names(self.data["version"])))).write_bytes(b"")
        with self.assertRaisesRegex(ValueError, "nonempty"):
            self.run_draft()
        self.assertFalse(self.github.mutations)

    def test_branch_asset_lookup_is_paginated_before_obsolete_files_are_removed(self):
        release = self.github.add_release("snapshot/main")
        release["assets"] = [dict(id=number + 100, name=f"old-{number}", size=1,
                                  state="uploaded") for number in range(105)]
        self.run_draft()
        self.assertEqual(len(release["assets"]), 9)
        self.assertFalse(any(asset["name"].startswith("old-") for asset in release["assets"]))

    def legacy_release(self, tag):
        version = self.data["version"]
        names = package_names(version) | {f"musicalc_{version}_SHA256SUMS.txt"}
        return self.github.add_release(tag, assets=[dict(id=100 + number, name=name,
            size=1, state="uploaded") for number, name in enumerate(sorted(names))])

    def test_branch_refresh_trims_legacy_thirteen_assets_to_nine(self):
        release = self.legacy_release("snapshot/main")
        original_id = release["id"]
        self.run_draft()
        self.assertEqual(release["id"], original_id)
        self.assertEqual({asset["name"] for asset in release["assets"]}, release_package_names(self.data["version"]))
        self.assertEqual(len([method for method, _, _ in self.github.mutations if method == "DELETE"]), 4)

    def test_failed_upload_does_not_prune_legacy_extras_and_retry_finishes(self):
        release = self.legacy_release("snapshot/main")
        self.github.fail_upload = True
        with self.assertRaisesRegex(RuntimeError, "upload interrupted"):
            self.run_draft()
        self.assertEqual(len(release["assets"]), 13)
        self.assertFalse(any(method == "DELETE" for method, _, _ in self.github.mutations))
        self.assertIn("incomplete", release["body"].lower())
        self.run_draft()
        self.assertEqual({asset["name"] for asset in release["assets"]}, release_package_names(self.data["version"]))
        self.assertNotIn("incomplete", release["body"].lower())

    def test_old_local_manifest_is_neither_rewritten_nor_uploaded(self):
        checksum = self.root / f"musicalc_{self.data['version']}_SHA256SUMS.txt"
        checksum.write_bytes(b"old manifest")
        self.run_draft()
        self.assertEqual(checksum.read_bytes(), b"old manifest")
        self.assertNotIn(checksum.name, self.github.uploads)
        self.assertEqual(len(self.github.uploads), 9)

    def test_legacy_tag_draft_and_published_release_keep_all_their_assets(self):
        self.select_tag()
        release = self.legacy_release("0.9.0")
        for published in (False, True):
            release["draft"] = not published
            before = copy.deepcopy(release)
            self.run_draft()
            self.assertEqual(release, before)
            self.assertEqual(len(release["assets"]), 13)
            self.assertFalse(self.github.mutations)
            self.assertFalse(self.github.uploads)


class GitHubCommandTests(unittest.TestCase):
    def test_api_payload_uses_json_stdin_and_utf8_without_shell_interpretation(self):
        payload = {"body": "line one\n$literal @file {branch} äudio", "draft": True}
        response = subprocess.CompletedProcess([], 0, stdout=json.dumps(payload), stderr="")
        with patch.object(draft.subprocess, "run", return_value=response) as run:
            self.assertEqual(draft.api("repos/example/musicalc/releases", "POST", payload), payload)
        self.assertEqual(run.call_args.args[0], ["gh", "api", "repos/example/musicalc/releases",
                                               "--method", "POST", "--input", "-"])
        self.assertEqual(json.loads(run.call_args.kwargs["input"]), payload)
        self.assertEqual(run.call_args.kwargs["encoding"], "utf-8")
        self.assertNotIn("shell", run.call_args.kwargs)

    def test_cli_errors_including_404_propagate_without_a_create_fallback(self):
        for message in ("HTTP 403", "HTTP 404", "network timeout"):
            response = subprocess.CompletedProcess([], 1, stdout="", stderr=message)
            with self.subTest(message=message), patch.object(draft.subprocess, "run", return_value=response) as run:
                with self.assertRaisesRegex(RuntimeError, message):
                    draft.api("repos/example/musicalc/releases")
                run.assert_called_once()


if __name__ == "__main__":
    unittest.main()
