#!/usr/bin/env python3
"""Build and test one native desktop architecture."""
import argparse
import os
from pathlib import Path
import shutil
import subprocess
import sys
from version import ROOT, current


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
    if args.platform == "windows" and args.arch == "amd64":
        env.update(CC="clang --target=x86_64-w64-windows-gnu",
                   CXX="clang++ --target=x86_64-w64-windows-gnu",
                   GOFLAGS="-ldflags=-extld=x86_64-w64-mingw32-gcc")
        for tool in ("clang", "x86_64-w64-mingw32-gcc"):
            if not shutil.which(tool):
                parser.error(f"Missing {tool}; install LLVM and MinGW as described in README-BUILD.md")
    elif args.platform == "windows":
        if not shutil.which("zig"):
            parser.error("Missing Zig 0.16.0; add it to the runner PATH")
        version = subprocess.check_output(["zig", "version"], text=True).strip()
        if version != "0.16.0":
            parser.error(f"Expected Zig 0.16.0; found {version}")
        env.update(CC="zig cc -target aarch64-windows-gnu", CXX="zig c++ -target aarch64-windows-gnu")
    else:
        run("pkg-config", "--exists", "gl", "x11", "xcursor", "xrandr", "xinerama", "xi", "xxf86vm", "alsa")
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
