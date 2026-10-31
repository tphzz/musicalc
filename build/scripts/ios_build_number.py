#!/usr/bin/env python3
"""Reserve durable TestFlight build numbers; never roll reservations back."""
import argparse
from contextlib import contextmanager
import json
import os
from pathlib import Path
import tempfile
from version import RELEASE_RE


@contextmanager
def locked(path):
    with open(path, "a+b") as stream:
        if os.name == "nt":
            import msvcrt
            stream.seek(0)
            msvcrt.locking(stream.fileno(), msvcrt.LK_LOCK, 1)
        else:
            import fcntl
            fcntl.flock(stream, fcntl.LOCK_EX)
        try:
            yield
        finally:
            if os.name == "nt":
                stream.seek(0)
                msvcrt.locking(stream.fileno(), msvcrt.LK_UNLCK, 1)
            else:
                fcntl.flock(stream, fcntl.LOCK_UN)


def reserve(path, bundle, version, seed=None, *, minimum=None):
    path = Path(path)
    if not RELEASE_RE.fullmatch(version):
        raise ValueError("Marketing version must be MAJOR.MINOR.PATCH")
    if minimum is not None and (type(minimum) is not int or not 0 <= minimum <= 9999):
        raise ValueError("App Store Connect build number must be 0..9999")
    if seed is not None and minimum is not None:
        raise ValueError("Cannot combine manual seeding and automatic allocation")
    path.parent.mkdir(parents=True, exist_ok=True)
    with locked(path.with_suffix(".lock")):
        if path.exists():
            state = json.loads(path.read_text(encoding="utf-8"))
            if (not isinstance(state, dict)
                    or not isinstance(state.get("bundle"), str)
                    or not isinstance(state.get("versions"), dict)):
                raise ValueError(
                    f"Invalid build counter {path}: expected a bundle string and versions object; "
                    "restore a verified backup or reconcile with App Store Connect"
                )
            if state["bundle"] != bundle:
                raise ValueError(
                    f"Build counter {path} belongs to {state['bundle']!r}, "
                    f"but the configured bundle is {bundle!r}. "
                    "Preserve this counter and reconcile the bundle ID before building"
                )
        elif seed is None and minimum is None:
            raise ValueError("Counter missing; seed explicitly from App Store Connect before building")
        else:
            state = {"bundle": bundle, "versions": {}}
        previous = state["versions"].get(version)
        if minimum is not None:
            if version in state["versions"] and (type(previous) is not int or not 0 <= previous <= 9999):
                raise ValueError("Corrupt local reservation; restore a backup before building")
            next_value = max(previous if previous is not None else 0, minimum) + 1
            if next_value > 9999:
                raise ValueError("Build numbers exhausted; select a new marketing version")
        elif seed is not None:
            if type(seed) is not int or not 0 <= seed < 9999 or (previous is not None and seed < previous):
                raise ValueError("Seed must be 0..9998 and cannot decrease a reservation")
            next_value = seed
        else:
            if type(previous) is not int or not 0 <= previous < 9999:
                raise ValueError("Version is unseeded, corrupt, or exhausted; reconcile with App Store Connect")
            next_value = previous + 1
        state["versions"][version] = next_value
        descriptor, temporary = tempfile.mkstemp(dir=path.parent, prefix=path.name + ".")
        try:
            with os.fdopen(descriptor, "w", encoding="utf-8") as out:
                json.dump(state, out, sort_keys=True)
                out.write("\n")
                out.flush()
                os.fsync(out.fileno())
            os.replace(temporary, path)
            if os.name != "nt":
                directory = os.open(path.parent, os.O_RDONLY)
                try:
                    os.fsync(directory)
                finally:
                    os.close(directory)
        finally:
            if os.path.exists(temporary):
                os.unlink(temporary)
        return next_value


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--file", type=Path, required=True)
    parser.add_argument("--bundle", required=True)
    parser.add_argument("--version", required=True)
    parser.add_argument("--seed", type=int)
    args = parser.parse_args()
    try:
        print(reserve(args.file, args.bundle, args.version, args.seed))
    except (ValueError, OSError) as exc:
        parser.exit(1, f"Build counter error: {exc}\n")


if __name__ == "__main__":
    main()
