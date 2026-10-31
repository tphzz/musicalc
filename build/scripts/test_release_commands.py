import importlib.util
import json
import os
from pathlib import Path
import sys
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch


def load(name, filename):
    spec = importlib.util.spec_from_file_location(name, Path(__file__).with_name(filename))
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


platform = load("build_platform", "build-platform.py")
draft = load("draft_release", "draft-release.py")
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

    def test_draft_checksum_rerun_excludes_itself(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary)
            for suffix in (".exe",".deb",".rpm",".tar.gz",".dmg",".app.zip",".apk"):
                (root/("musicalc_0.9.0"+suffix)).write_bytes(b"content")
            result=SimpleNamespace(returncode=0,stdout=json.dumps({"draft":True}),stderr="")
            with patch.object(sys,"argv",["draft-release.py",str(root)]), \
                 patch.dict(os.environ,{"GITHUB_REPOSITORY":"example/musicalc"}), \
                 patch.object(draft,"current",return_value=DATA), \
                 patch.object(draft.subprocess,"run",return_value=result), \
                 patch.object(draft,"gh") as gh:
                draft.main()
                first=(root/"musicalc_0.9.0_SHA256SUMS.txt").read_text()
                draft.main()
            self.assertEqual(first,(root/"musicalc_0.9.0_SHA256SUMS.txt").read_text())
            self.assertEqual(len(first.splitlines()),7)
            self.assertNotIn("SHA256SUMS",first)
            for call in gh.call_args_list:
                self.assertEqual(len(call.args),len(set(call.args)))

    def test_published_release_is_never_updated(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary)
            for suffix in (".exe",".deb",".rpm",".tar.gz",".dmg",".app.zip",".apk"):
                (root/("musicalc_0.9.0"+suffix)).write_bytes(b"content")
            result=SimpleNamespace(returncode=0,stdout=json.dumps({"draft":False}),stderr="")
            with patch.object(sys,"argv",["draft-release.py",str(root)]), \
                 patch.dict(os.environ,{"GITHUB_REPOSITORY":"example/musicalc"}), \
                 patch.object(draft,"current",return_value=DATA), \
                 patch.object(draft.subprocess,"run",return_value=result), \
                 patch.object(draft,"gh") as gh, self.assertRaises(SystemExit):
                draft.main()
            gh.assert_not_called()


if __name__ == "__main__":
    unittest.main()
