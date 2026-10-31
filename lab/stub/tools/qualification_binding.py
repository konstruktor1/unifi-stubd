"""Validate candidate SSH inputs against the existing native product inventory."""

import argparse
import ipaddress
import os
import re
import sys
import tomllib
from pathlib import Path


def physical_file(path: Path) -> bool:
    """Accept an existing regular file through physical ancestors only."""
    return path.is_absolute() and path.is_file() and path.resolve(strict=True) == path


def validate_binding(
    root: Path, data_root: Path, host_service: str, requested_ssh: dict[str, str | int]
) -> None:
    """Reject transport or source bindings outside the assigned product role."""
    if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]*", host_service):
        raise ValueError("invalid managed host service")
    if requested_ssh.keys() != {
        "host",
        "port",
        "user",
        "identity_file",
        "known_hosts_file",
    }:
        raise ValueError("supply all native SSH transport fields")
    machine = data_root / "tier3/machines/unifi-stubd-lab"
    if not machine.is_dir() or machine.resolve(strict=True) != machine:
        raise ValueError("assigned product machine must be a physical directory")
    inventory = (
        root
        / ".consumer-data/build-runner/test-runner/inventory"
        / f"{host_service}.toml"
    )
    if not physical_file(inventory) or inventory.stat().st_size > 65536:
        raise ValueError("native product inventory must be a bounded physical file")
    with inventory.open("rb") as stream:
        hosts = tomllib.load(stream).get("hosts", {})
    host = hosts.get(host_service, {}) if isinstance(hosts, dict) else {}
    expected_source = data_root / "tier3/test-runner/_managed" / host_service
    if (
        not isinstance(host, dict)
        or host.get("provider") != "qemu"
        or host.get("profile") != "linux-arm64"
        or host.get("platform") != "linux/arm64"
        or host.get("status") not in ("experimental", "supported")
        or host.get("directory") != str(expected_source)
    ):
        raise ValueError("native source does not bind the assigned ARM64 product role")
    transport = host.get("ssh", {})
    if (
        not isinstance(transport, dict)
        or transport.get("host_key_checking") != "yes"
        or transport.get("shell") != "posix"
        or any(transport.get(key) != value for key, value in requested_ssh.items())
        or transport.get("known_hosts_file")
        != str(machine / "identity/known_hosts.test-runner")
    ):
        raise ValueError("SSH inputs differ from the native product transport")
    for key in ("identity_file", "known_hosts_file"):
        if not physical_file(Path(transport[key])):
            raise ValueError("SSH input must be an existing physical regular file")
    seed_path = machine / "base/seed.iso"
    if not physical_file(seed_path):
        raise ValueError("assigned product machine seed must be a physical file")
    seed = f"file={seed_path}"
    qemu = host.get("qemu", {})
    extra_args = qemu.get("extra_args", []) if isinstance(qemu, dict) else []
    if not isinstance(extra_args, list):
        raise ValueError("native product QEMU arguments must be an array")
    if not any(
        isinstance(value, str) and seed in value.split(",") for value in extra_args
    ):
        raise ValueError("native source has no assigned product machine seed binding")


def main() -> int:
    """Check the current native record without reading key bytes or changing it."""
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--data-root", required=True, type=Path)
    parser.add_argument("--host-service", required=True)
    parser.add_argument("--candidate-ip", required=True, type=ipaddress.IPv4Address)
    parser.add_argument("--host", required=True)
    parser.add_argument("--port", required=True, type=int)
    parser.add_argument("--user", required=True)
    parser.add_argument("--identity-file", required=True)
    parser.add_argument("--known-hosts-file", required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[3]
    if os.environ.get("PROVIDER_CONSUMER_ROOT") != str(root):
        parser.error("inherit the native Core consumer root")
    try:
        validate_binding(
            root,
            args.data_root,
            args.host_service,
            {
                "host": args.host,
                "port": args.port,
                "user": args.user,
                "identity_file": args.identity_file,
                "known_hosts_file": args.known_hosts_file,
            },
        )
    except (OSError, ValueError) as error:
        print(f"qualification binding: {error}", file=sys.stderr)
        return 1
    print("native product source and SSH binding valid")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
