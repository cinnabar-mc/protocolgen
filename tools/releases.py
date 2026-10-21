#!/usr/bin/env python3
"""Check that the release index agrees with every committed codec."""
import json
from pathlib import Path


def read(path):
    """Read a repository JSON document."""
    return json.loads(path.read_text())


def validate(root):
    """Reject missing snapshots and inconsistent targets."""
    index = read(root / "data/source/releases.json")
    releases = index["releases"]
    if index["schema_version"] != 1:
        raise ValueError("unsupported release index schema")
    if index["default_protocol"] not in releases:
        raise ValueError("unknown default protocol")
    committed = {p.parent.name for p in (root / "generated").glob("*/manifest.json")}
    if committed != set(releases):
        raise ValueError(f"release index differs from committed snapshots: {committed ^ set(releases)}")
    for snapshot, release in releases.items():
        target = read(root / "generated" / snapshot / "manifest.json")["target"]
        if (release["snapshot"] != snapshot or target["minecraft_version"] != snapshot
                or target["protocol_version"] != release["protocol_version"]
                or release["channel"] not in ("retail", "preview")
                or release["minecraft_version"] not in release["aliases"]):
            raise ValueError(f"inconsistent release {snapshot}")
    lock = read(root / "data/source/lock.json")
    if lock["release"] not in releases:
        raise ValueError("data lock selects an unknown release")
    print(f"Validated {len(releases)} releases")


if __name__ == "__main__":
    validate(Path(__file__).resolve().parent.parent)
