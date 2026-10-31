#!/usr/bin/env python3
"""Verify LOAD-segment alignment from an Android shared library."""
import argparse
import re
import subprocess


def verify(output):
    alignments = []
    for line in output.splitlines():
        if not re.match(r"\s*LOAD\s", line):
            continue
        value = line.split()[-1]
        if not re.fullmatch(r"0x[0-9a-fA-F]+", value):
            raise ValueError("Malformed LOAD alignment")
        alignment = int(value, 16)
        if alignment < 16384 or alignment & (alignment - 1):
            raise ValueError("Native LOAD alignment must be a power of two of at least 16 KB")
        alignments.append(alignment)
    if not alignments:
        raise ValueError("readelf returned no LOAD segments")
    return alignments


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("readelf")
    parser.add_argument("library")
    args = parser.parse_args()
    try:
        output = subprocess.check_output([args.readelf, "-lW", args.library], text=True)
        print("Verified LOAD alignments:", verify(output))
    except (ValueError, OSError, subprocess.CalledProcessError) as exc:
        parser.exit(1, f"Android alignment error: {exc}\n")


if __name__ == "__main__":
    main()
