#!/usr/bin/env python3
"""Build and test one native desktop architecture."""
import argparse
import os
from pathlib import Path
import shutil
import subprocess
import sys
from version import ROOT, current, git


def run(*args, env=None):
    print("+", " ".join(map(str, args)), flush=True)
    subprocess.run(list(map(str, args)), cwd=ROOT, env=env, check=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("platform", choices=("linux", "windows"))
    parser.add_argument("arch", choices=("amd64", "arm64"))
    args = parser.parse_args()
    data = current()
    env = os.environ.copy()
    native = subprocess.check_output(["go", "env", "GOOS", "GOARCH"], text=True).split()
    if native != [args.platform, args.arch]:
        parser.error(f"Expected native {args.platform}/{args.arch}, found {'/'.join(native)}")
    env["CGO_ENABLED"] = "1"
    env["MUSICALC_VERSION"] = data["version"]
    if data["release"]:
        env["GORELEASER_CURRENT_TAG"] = data["version"]
    if not data["release"]:
        env["MUSICALC_SNAPSHOT_VERSION"] = data["version"]
    if args.platform == "windows":
        if not shutil.which("zig"):
            parser.error("Missing Zig; install 0.17.0 or 0.16.0 and add it to the runner PATH (README-BUILD.md)")
        version = subprocess.check_output(["zig", "version"], text=True).strip()
        if version not in ("0.16.0", "0.17.0"):
            parser.error(f"Expected Zig 0.16.0 or 0.17.0; found {version}")
        # Go 1.26 needs an explicit Git checkout when .git is a file.
        env.update(GIT_DIR=git("rev-parse", "--absolute-git-dir"),
                   GIT_WORK_TREE=git("rev-parse", "--show-toplevel"))
        target = "x86_64" if args.arch == "amd64" else "aarch64"
        env.update(CC=f"zig cc -target {target}-windows-gnu",
                   CXX=f"zig c++ -target {target}-windows-gnu",
                   ZIG_GLOBAL_CACHE_DIR=str(ROOT / "build/.cache/zig-global"),
                   ZIG_LOCAL_CACHE_DIR=str(ROOT / "build/.cache/zig-local"))
    else:
        run("pkg-config", "--print-errors", "--exists",
            "gl", "x11", "xcursor", "xrandr", "xinerama", "xi", "xxf86vm", "alsa",
            "wayland-client", "wayland-cursor", "wayland-egl", "xkbcommon")
    test = ["go", "test", "./..."]
    if args.platform == "linux":
        test = ["xvfb-run", "-a", *test]
    run(*test, env=env)
    config = ROOT / "build/release" / (
        "goreleaser-win-all.yaml" if args.platform == "windows"
        else f"goreleaser-linux-{args.arch}.yaml")
    run("goreleaser", "check", "--config", config, env=env)
    command = ["goreleaser", "build" if args.platform == "windows" else "release",
               "--clean", "--config", config]
    if args.platform == "windows":
        command += ["--id", "win-" + args.arch]
    else:
        command += ["--skip=publish"]
    if not data["release"]:
        command += ["--snapshot"]
    run(*command, env=env)
    if args.platform == "windows":
        source = ROOT / "build/dist" / ("musicalc_x64.exe" if args.arch == "amd64" else "musicalc_arm64.exe")
        target = ROOT / "build/dist" / f"musicalc_{data['version']}_windows_{args.arch}.exe"
        shutil.copy2(source, target)


if __name__ == "__main__":
    try:
        main()
    except (ValueError, OSError, subprocess.CalledProcessError) as exc:
        sys.exit(str(exc))
