#!/usr/bin/env python3
"""Attach validated platform outputs to an exact-tag draft GitHub release."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
from version import current


def gh(*args):
    return subprocess.check_output(["gh", *args], text=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("artifacts", type=Path)
    args = parser.parse_args()
    data = current()
    if not data["release"]:
        parser.error("Draft releases require a numeric tag")
    repository = os.environ["GITHUB_REPOSITORY"]
    checksum = args.artifacts / f"musicalc_{data['version']}_SHA256SUMS.txt"
    assets = sorted(p for p in args.artifacts.rglob("*") if p.is_file() and p != checksum)
    expected = (".exe", ".deb", ".rpm", ".tar.gz", ".dmg", ".app.zip", ".apk")
    names = [p.name for p in assets]
    if not assets or len(names) != len(set(names)):
        parser.error("Missing artifacts or duplicate asset names")
    if not all(any(name.endswith(suffix) for name in names) for suffix in expected):
        parser.error("Required platform assets are incomplete")
    if any(data["version"] not in p.name or p.suffix in (".ipa", ".p8", ".mobileprovision") for p in assets):
        parser.error("Unexpected or incorrectly versioned release asset")
    # Distinguish a missing release (404) from authentication/network failures.
    result = subprocess.run(["gh", "api", f"repos/{repository}/releases/tags/{data['version']}"],
                            text=True, capture_output=True)
    if result.returncode == 0:
        release = json.loads(result.stdout)
        if not release["draft"]:
            parser.error("Published releases are immutable; choose a new tag")
        # Uploads below replace only same-named assets on the existing draft.
    elif "HTTP 404" in result.stderr:
        gh("release", "create", data["version"], "--repo", repository, "--verify-tag",
           "--draft", "--title", f"MusiCalc {data['version']}", "--generate-notes")
    else:
        raise RuntimeError(result.stderr.strip())
    with checksum.open("w", encoding="utf-8") as out:
        for path in assets:
            with path.open("rb") as stream:
                digest = hashlib.file_digest(stream, "sha256").hexdigest()
            out.write(f"{digest}  {path.name}\n")
    gh("release", "upload", data["version"], "--repo", repository, "--clobber",
       *map(str, [*assets, checksum]))
    print(f"Draft ready: https://github.com/{repository}/releases/tag/{data['version']}")


if __name__ == "__main__":
    try:
        main()
    except (ValueError, RuntimeError, OSError, subprocess.CalledProcessError) as exc:
        sys.exit(str(exc))
