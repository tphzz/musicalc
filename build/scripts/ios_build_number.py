#!/usr/bin/env python3
"""Generate a stateless UTC iPhone build number for a serialized build attempt."""
import argparse
from datetime import datetime, timezone
import time


def current_build():
    # The whole iPhone job is serialized. Waiting before sampling separates even immediate retries.
    time.sleep(1)
    return int(datetime.now(timezone.utc).strftime("%Y%m%d%H%M%S"))


if __name__ == "__main__":
    argparse.ArgumentParser(description=__doc__).parse_args()
    print(current_build())
