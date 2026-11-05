#!/usr/bin/env python3
"""Create immutable tag drafts or refresh one snapshot draft per branch."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
from urllib.parse import quote

from version import current


def gh(*args, input=None):
    result = subprocess.run(["gh", *args], input=input, encoding="utf-8", capture_output=True)
    if result.returncode:
        raise RuntimeError(result.stderr.strip() or "GitHub CLI command failed")
    return result.stdout


def api(endpoint, method="GET", payload=None):
    command = ["api", endpoint, "--method", method]
    if payload is not None:
        command += ["--input", "-"]
    output = gh(*command, input=json.dumps(payload) if payload is not None else None)
    return json.loads(output) if output.strip() else None


def pages(endpoint):
    page = 1
    while True:
        values = api(f"{endpoint}?per_page=100&page={page}")
        if not isinstance(values, list):
            raise ValueError("Expected a paginated GitHub list")
        yield from values
        if len(values) < 100:
            return
        page += 1


def find_release(base, tag):
    # The release-by-tag endpoint only finds published releases, not drafts.
    matches = [release for release in pages(f"{base}/releases") if release["tag_name"] == tag]
    if len(matches) > 1:
        raise ValueError(f"Multiple releases match {tag}; reconcile the duplicate drafts manually")
    return matches[0] if matches else None


def remote_tag(base, tag):
    refs = api(f"{base}/git/matching-refs/tags/{quote(tag, safe='')}")
    return next((ref["object"] for ref in refs if ref["ref"] == f"refs/tags/{tag}"), None)


def verify_tag_commit(base, tag, commit):
    obj = remote_tag(base, tag)
    seen = set()
    while obj and obj["type"] == "tag":
        if obj["sha"] in seen:
            raise ValueError("Cyclic annotated tag response")
        seen.add(obj["sha"])
        obj = api(f"{base}/git/tags/{obj['sha']}")["object"]
    if not obj or obj["type"] != "commit" or obj["sha"] != commit:
        raise ValueError(f"Remote tag {tag} does not identify the built commit {commit}")


def require_draft(endpoint):
    release = api(endpoint)
    if not release["draft"]:
        raise ValueError("Published releases are preserved; no further changes are allowed")
    return release


def prepare_assets(directory, version):
    checksum = directory / f"musicalc_{version}_SHA256SUMS.txt"
    paths = sorted(path for path in directory.rglob("*") if path.is_file() and path != checksum)
    names = [path.name for path in paths]
    if len(names) != len(set(names)):
        raise ValueError("Duplicate release artifact names")
    # Match the existing package names, including GoReleaser's x64 CPU suffix.
    expected = {
        f"musicalc_{version}_windows_amd64.exe", f"musicalc_{version}_windows_arm64.exe",
        f"musicalc-setup-{version}.exe", f"musicalc_{version}_linux_amd64v3.tar.gz",
        f"musicalc_{version}_linux_arm64.tar.gz", f"MusiCalc_{version}_universal.app.zip",
        f"musicalc_{version}_universal.dmg", f"musicalc_{version}_android_arm64.apk",
        *(f"musicalc_{version}_{arch}.{extension}" for arch in ("amd64", "arm64")
          for extension in ("deb", "rpm")),
    }
    missing, unexpected = expected - set(names), set(names) - expected
    if missing or unexpected:
        raise ValueError(f"Invalid release artifacts: missing={sorted(missing)}, unexpected={sorted(unexpected)}")
    inventory = {}
    for path in paths:
        if path.is_symlink() or not path.stat().st_size:
            raise ValueError(f"Release artifact must be a nonempty regular file: {path}")
        with path.open("rb") as stream:
            digest = hashlib.file_digest(stream, "sha256").hexdigest()
        inventory[path.name] = {"size": path.stat().st_size, "digest": "sha256:" + digest}
    if checksum.is_symlink():
        raise ValueError("Checksum artifact cannot be a symbolic link")
    checksum.write_text("".join(f"{inventory[path.name]['digest'][7:]}  {path.name}\n" for path in paths),
                        encoding="utf-8")
    inventory[checksum.name] = {"size": checksum.stat().st_size,
        "digest": "sha256:" + hashlib.sha256(checksum.read_bytes()).hexdigest()}
    return [*paths, checksum], inventory


def verify_assets(endpoint, expected, complete=False):
    require_draft(endpoint)
    assets = list(pages(f"{endpoint}/assets"))
    actual = {asset["name"]: asset for asset in assets}
    if len(actual) != len(assets) or expected.keys() - actual.keys():
        raise RuntimeError("Remote release asset inventory is incomplete or contains duplicate names")
    if complete and actual.keys() != expected.keys():
        raise RuntimeError("Remote release still contains obsolete assets")
    for name, info in expected.items():
        asset = actual[name]
        if asset["state"] != "uploaded" or asset["size"] != info["size"]:
            raise RuntimeError(f"Remote release asset is incomplete: {name}")
        if asset.get("digest") and asset["digest"] != info["digest"]:
            raise RuntimeError(f"Remote release asset digest mismatch: {name}")
    return assets


def metadata(data, ref, notes, ready):
    title = f"MusiCalc {data['version']}" if data["release"] else f"MusiCalc {ref} snapshot"
    server = os.environ.get("GITHUB_SERVER_URL", "https://github.com")
    repository = os.environ["GITHUB_REPOSITORY"]
    body = f"{notes}\n\nVersion: `{data['version']}`\nCommit: `{data['commit']}`\n"
    if not data["release"]:
        body += f"Branch: `{ref}`\n"
    if os.environ.get("GITHUB_RUN_ID"):
        body += f"Build: {server}/{repository}/actions/runs/{os.environ['GITHUB_RUN_ID']}\n"
    if not ready:
        title += " [incomplete]"
        body = "**Refresh incomplete. Do not publish.**\n\n" + body
    return dict(name=title, body=body.strip(), draft=True, prerelease=not data["release"],
                target_commitish=data["commit"])


def report(action, release):
    message = f"{action}: {release['html_url']}"
    print(message)
    if os.environ.get("GITHUB_STEP_SUMMARY"):
        with open(os.environ["GITHUB_STEP_SUMMARY"], "a", encoding="utf-8") as stream:
            stream.write(message + "\n")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("artifacts", type=Path)
    args = parser.parse_args()
    data = current()
    base = f"repos/{os.environ['GITHUB_REPOSITORY']}"
    ref = data["version"] if data["release"] else os.environ.get("GITHUB_REF_NAME")
    if not ref:
        raise ValueError("GITHUB_REF_NAME is required for a branch draft")
    tag = ref if data["release"] else f"snapshot/{ref}"
    release = find_release(base, tag)
    if data["release"] and release:
        report("Preserved existing tag release", release)
        return
    if not data["release"]:
        if release and not release["draft"]:
            raise ValueError("Published branch snapshots are preserved; choose an unpublished draft identity")
        if remote_tag(base, tag):
            raise ValueError(f"Snapshot tag already exists: {tag}; no Git tags will be moved or deleted")

    paths, inventory = prepare_assets(args.artifacts, data["version"])
    notes = "Rolling branch snapshot. Replaced by the next successful full build."
    if data["release"]:
        verify_tag_commit(base, tag, data["commit"])
        notes = api(f"{base}/releases/generate-notes", "POST", {
            "tag_name": tag, "target_commitish": data["commit"]})["body"]
    created = release is None
    pending = metadata(data, ref, notes, ready=False)
    if created:
        release = api(f"{base}/releases", "POST", dict(pending, tag_name=tag))
    endpoint = f"{base}/releases/{release['id']}"
    try:
        if not created:
            require_draft(endpoint)
            api(endpoint, "PATCH", pending)
        require_draft(endpoint)
        command = ["release", "upload", tag, "--repo", os.environ["GITHUB_REPOSITORY"]]
        if not data["release"]:
            command.append("--clobber")
        gh(*command, *map(str, paths))
        assets = verify_assets(endpoint, inventory)
        for asset in assets:
            if asset["name"] not in inventory:
                require_draft(endpoint)
                api(f"{base}/releases/assets/{asset['id']}", "DELETE")
        verify_assets(endpoint, inventory, complete=True)
        require_draft(endpoint)
        release = api(endpoint, "PATCH", metadata(data, ref, notes, ready=True))
    except Exception:
        if created and data["release"]:
            try:
                require_draft(endpoint)
                api(endpoint, "DELETE")
            except (ValueError, RuntimeError, OSError) as exc:
                print(f"Could not remove the new incomplete draft ({release['html_url']}): {exc}. "
                      "Reconcile it manually before retrying; its Git tag is unchanged.", file=sys.stderr)
        else:
            print(f"Snapshot refresh incomplete: {release['html_url']}. Rerun the draft job to repair it.",
                  file=sys.stderr)
        raise
    report("Created draft" if created else "Refreshed branch draft", release)


if __name__ == "__main__":
    try:
        main()
    except (ValueError, RuntimeError, OSError, subprocess.CalledProcessError) as exc:
        sys.exit(str(exc))
