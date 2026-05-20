#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path


def main() -> int:
    if len(sys.argv) < 4:
        print("usage: assert-events.py <events.jsonl> <start-line> [--min-count N] <mac> [<mac> ...]", file=sys.stderr)
        return 2
    events_path = Path(sys.argv[1])
    start_line = int(sys.argv[2])
    min_count = 1
    mac_args = sys.argv[3:]
    if len(mac_args) >= 2 and mac_args[0] == "--min-count":
        min_count = int(mac_args[1])
        mac_args = mac_args[2:]
    expected = {mac.lower() for mac in mac_args}
    if not expected:
        print("no MAC addresses supplied", file=sys.stderr)
        return 2
    found = {mac: 0 for mac in expected}
    if events_path.exists():
        for line_number, line in enumerate(events_path.read_text(encoding="utf-8").splitlines(), start=1):
            if line_number <= start_line or not line.strip():
                continue
            event = json.loads(line)
            if event.get("event") != "request":
                continue
            tnbu = event.get("tnbu") or {}
            if not tnbu.get("present"):
                continue
            mac = str(tnbu.get("mac", "")).lower()
            if mac in expected:
                found[mac] += 1
    missing = sorted(mac for mac, count in found.items() if count < min_count)
    if missing:
        details = ", ".join(f"{mac}={found[mac]}/{min_count}" for mac in missing)
        print(f"missing inform request events for MACs: {details}", file=sys.stderr)
        return 1
    details = ", ".join(f"{mac}={found[mac]}" for mac in sorted(found))
    print(f"found inform request events for MACs: {details}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
