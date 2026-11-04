#!/usr/bin/env python3
"""Check a Windows executable's architecture and GUI subsystem."""
import argparse
from pathlib import Path
import struct


MACHINES = {"amd64": 0x8664, "arm64": 0xaa64}


def verify_windows_binary(path, arch):
    path = Path(path)
    if arch not in MACHINES:
        raise ValueError(f"Unsupported Windows architecture: {arch}")
    with path.open("rb") as binary:
        dos = binary.read(64)
        if len(dos) != 64 or dos[:2] != b"MZ":
            raise ValueError(f"{path}: invalid DOS header")
        offset = struct.unpack_from("<I", dos, 0x3c)[0]
        if offset < 64:
            raise ValueError(f"{path}: invalid PE header offset")
        binary.seek(offset)
        header = binary.read(24)
        if len(header) != 24 or header[:4] != b"PE\0\0":
            raise ValueError(f"{path}: invalid PE header")
        machine = struct.unpack_from("<H", header, 4)[0]
        if machine != MACHINES[arch]:
            raise ValueError(f"{path}: PE machine {machine:#06x}; expected {arch} ({MACHINES[arch]:#06x})")
        size = struct.unpack_from("<H", header, 20)[0]
        optional = binary.read(size)
        if size < 112 or len(optional) != size or struct.unpack_from("<H", optional)[0] != 0x20b:
            raise ValueError(f"{path}: expected a complete PE32+ optional header")
        subsystem = struct.unpack_from("<H", optional, 68)[0]
        if subsystem != 2:
            raise ValueError(f"{path}: expected Windows GUI subsystem (2), found {subsystem}")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("exe", type=Path)
    parser.add_argument("arch", choices=MACHINES)
    args = parser.parse_args()
    try:
        verify_windows_binary(args.exe, args.arch)
    except OSError as error:
        parser.exit(1, f"{args.exe}: {error.strerror}\n")
    except ValueError as error:
        parser.exit(1, f"{error}\n")
    print(f"Verified Windows {args.arch} GUI executable: {args.exe}")


if __name__ == "__main__":
    main()
