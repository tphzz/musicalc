#!/usr/bin/env python3
"""Build the checksum-pinned Fyne tool with the iPhone compatibility patch."""
import json
from pathlib import Path
import shutil
import stat
import subprocess
import tempfile
from version import ROOT

MODULE = "fyne.io/tools"
VERSION = "v1.7.3"
COMMIT = "c4bd04319042bf4ea9db2e529b8d53f93131c803"
MODULE_SUM = "h1:wd3Ywi1zkHDB1whmZAq+xSUXXyJLo3pKrcoPopb7vGw="

download = json.loads(subprocess.check_output(
    ["go", "mod", "download", "-json", MODULE + "@" + VERSION], text=True))
if download.get("Error") or download.get("Sum") != MODULE_SUM:
    raise SystemExit("Pinned Fyne module checksum does not match")
if download.get("Origin") and download["Origin"].get("Hash") != COMMIT:
    raise SystemExit("Pinned Fyne source commit does not match")
patch_tool = shutil.which("patch")
if not patch_tool:
    raise SystemExit("The standard macOS patch utility is required")
patch = ROOT / "build/patches/fyne-tools-ios.patch"
cache = ROOT / "build/.cache"
cache.mkdir(parents=True, exist_ok=True)
with tempfile.TemporaryDirectory(prefix="fyne-ios-", dir=cache) as temporary:
    source = Path(temporary) / "source"
    shutil.copytree(download["Dir"], source)
    # Modify only the disposable copy, never the module cache.
    for item in source.rglob("*"):
        if item.is_file():
            item.chmod(item.stat().st_mode | stat.S_IWUSR)
    command = [patch_tool, "-p1", "-N", "-i", str(patch)]
    subprocess.run([*command, "--dry-run"], cwd=source, check=True, stdin=subprocess.DEVNULL)
    subprocess.run(command, cwd=source, check=True, stdin=subprocess.DEVNULL)
    tool = cache / "fyne-ios"
    subprocess.run(["go", "build", "-trimpath", "-o", str(tool), "./cmd/fyne"], cwd=source, check=True, stdin=subprocess.DEVNULL)
    print(tool)
