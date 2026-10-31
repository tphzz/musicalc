#!/usr/bin/env python3
"""Resolve one validated version for local packaging and manual workflows."""
import argparse
import json
import os
from pathlib import Path
import re
import shlex
import subprocess

NUMERIC = r"(?:0|[1-9][0-9]*)"
RELEASE_RE = re.compile(rf"{NUMERIC}\.{NUMERIC}\.{NUMERIC}", re.ASCII)
SNAPSHOT_RE = re.compile(
    rf"{NUMERIC}\.{NUMERIC}\.{NUMERIC}-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*", re.ASCII
)
ROOT = Path(__file__).resolve().parents[2]


def git(*args, required=True):
    result = subprocess.run(["git", *args], cwd=ROOT, text=True, capture_output=True)
    if required and result.returncode:
        raise ValueError(result.stderr.strip() or "Git command failed")
    return result.stdout.strip() if result.returncode == 0 else ""


def resolve(ref_type, ref_name, commit, latest_tag="", override="", build=1):
    if not re.fullmatch(r"[0-9a-f]{40}", commit):
        raise ValueError("Expected a full Git commit SHA")
    if ref_type not in ("branch", "tag"):
        raise ValueError("Expected a branch or tag ref")
    if ref_type == "tag":
        if not RELEASE_RE.fullmatch(ref_name):
            raise ValueError("Release tags must be MAJOR.MINOR.PATCH without leading zeros")
        if override:
            raise ValueError("Snapshot overrides are not allowed for release tags")
        core = ref_name
        version = core
    else:
        core = ref_name if RELEASE_RE.fullmatch(ref_name) else latest_tag
        if not RELEASE_RE.fullmatch(core):
            core = "0.0.0"
        version = override or f"{core}-snapshot.g{commit[:7]}"
        if not SNAPSHOT_RE.fullmatch(version):
            raise ValueError("Snapshot version must be a numeric version with a prerelease suffix")
        # SemVer numeric prerelease identifiers cannot have leading zeros.
        if any(p.isdigit() and len(p) > 1 and p.startswith("0") for p in version.split("-", 1)[1].split(".")):
            raise ValueError("Snapshot prerelease numbers cannot have leading zeros")
        core = version.split("-", 1)[0]
    return {"version": version, "marketing_version": core, "release": ref_type == "tag",
            "commit": commit, "build": build}


def current(ref_type=None, ref_name=None, override=None):
    commit = git("rev-parse", "HEAD")
    ref_type = ref_type or os.environ.get("GITHUB_REF_TYPE")
    ref_name = ref_name or os.environ.get("GITHUB_REF_NAME")
    if not ref_type:
        exact = git("describe", "--tags", "--exact-match", required=False)
        ref_type = "tag" if exact else "branch"
        ref_name = exact or git("branch", "--show-current") or "detached"
    if not ref_name:
        raise ValueError("Ref name is required")
    if ref_type == "tag" and git("rev-parse", f"refs/tags/{ref_name}^{{}}") != commit:
        raise ValueError("Selected tag does not point to HEAD")
    candidates = git("tag", "--merged", commit, "--sort=-version:refname").splitlines()
    latest = next((tag for tag in candidates if RELEASE_RE.fullmatch(tag)), "")
    override = os.environ.get("MUSICALC_SNAPSHOT_VERSION", "") if override is None else override
    return resolve(ref_type, ref_name, commit, latest, override, int(git("rev-list", "--count", "HEAD")))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--ref-type", choices=("branch", "tag"))
    parser.add_argument("--ref-name")
    parser.add_argument("--snapshot-version")
    parser.add_argument("--shell", action="store_true", help="Print Bash exports")
    parser.add_argument("--github", action="store_true", help="Append workflow outputs/environment")
    args = parser.parse_args()
    try:
        data = current(args.ref_type, args.ref_name, args.snapshot_version)
    except (ValueError, OSError) as exc:
        parser.exit(1, f"Version error: {exc}\n")
    env = {"MUSICALC_VERSION": data["version"], "MUSICALC_MARKETING_VERSION": data["marketing_version"],
           "MUSICALC_RELEASE": str(data["release"]).lower(), "MUSICALC_BUILD": str(data["build"]),
           "MUSICALC_COMMIT": data["commit"]}
    if not data["release"]:
        env["MUSICALC_SNAPSHOT_VERSION"] = data["version"]
    if args.github:
        for key, values in (("GITHUB_ENV", env), ("GITHUB_OUTPUT", {
            "version": data["version"], "release": str(data["release"]).lower(),
            "commit": data["commit"]})):
            with open(os.environ[key], "a", encoding="utf-8") as stream:
                stream.writelines(f"{name}={value}\n" for name, value in values.items())
    if args.shell:
        print("\n".join(f"export {key}={shlex.quote(value)}" for key, value in env.items()))
    else:
        print(json.dumps(data, sort_keys=True))


if __name__ == "__main__":
    main()
