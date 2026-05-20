#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import ssl
import sys
import time
import urllib.error
import urllib.request
from http.cookiejar import CookieJar


def main() -> int:
    parser = argparse.ArgumentParser(description="Drive lab UniFi adoption over the controller API.")
    parser.add_argument(
        "action",
        choices=["version", "forget", "wait-clean", "wait-present", "wait-pending", "adopt", "wait-adopted"],
    )
    parser.add_argument("--base-url", default=os.environ.get("UNIFI_STUB_LAB_API_URL", "https://127.0.0.1:8443"))
    parser.add_argument("--site", default=os.environ.get("UNIFI_STUB_LAB_SITE", "default"))
    parser.add_argument("--username", default=os.environ.get("UNIFI_STUB_LAB_ADMIN_USER", "admin"))
    parser.add_argument("--password", default=os.environ.get("UNIFI_STUB_LAB_ADMIN_PASSWORD", "admin"))
    parser.add_argument("--mac")
    parser.add_argument("--expected-version")
    parser.add_argument("--timeout", type=int, default=60)
    args = parser.parse_args()

    client = ControllerClient(args.base_url, args.username, args.password)
    if args.action == "version":
        print(controller_version_summary(client, args.expected_version))
        return 0

    if not args.mac:
        parser.error(f"--mac is required for action {args.action}")
    client.login()

    mac = args.mac.lower()
    if args.action == "forget":
        client.command(args.site, "sitemgr", {"cmd": "delete-device", "mac": mac}, allow_error=True)
        print(f"controller forget requested: mac={mac}")
        return 0
    if args.action == "wait-clean":
        wait_until_clean(client, args.site, mac, args.timeout)
        print(f"controller device not adopted: mac={mac}")
        return 0
    if args.action == "adopt":
        client.command(args.site, "devmgr", {"cmd": "adopt", "macs": [mac]})
        print(f"controller adopt requested: mac={mac}")
        return 0
    if args.action == "wait-present":
        device = wait_for_device(client, args.site, mac, args.timeout, adopted=None)
        print(device_summary("controller device visible", device))
        return 0
    if args.action == "wait-pending":
        device = wait_for_device(client, args.site, mac, args.timeout, adopted=False)
        print(device_summary("controller device pending", device))
        return 0
    if args.action == "wait-adopted":
        device = wait_for_device(client, args.site, mac, args.timeout, adopted=True)
        print(device_summary("controller device adopted", device))
        return 0
    raise AssertionError(args.action)


class ControllerClient:
    def __init__(self, base_url: str, username: str, password: str) -> None:
        self.base_url = base_url.rstrip("/")
        self.username = username
        self.password = password
        self.opener = urllib.request.build_opener(
            urllib.request.HTTPSHandler(context=ssl._create_unverified_context()),
            urllib.request.HTTPCookieProcessor(CookieJar()),
        )

    def login(self) -> None:
        payload = {"username": self.username, "password": self.password, "remember": True}
        data = self.post_json("/api/login", payload)
        if data.get("meta", {}).get("rc") != "ok":
            raise RuntimeError(f"controller login failed: {data.get('meta')}")

    def command(
        self,
        site: str,
        manager: str,
        payload: dict[str, object],
        allow_error: bool = False,
    ) -> dict[str, object]:
        data = self.post_json(f"/api/s/{site}/cmd/{manager}", payload)
        rc = data.get("meta", {}).get("rc")
        if rc != "ok" and not allow_error:
            raise RuntimeError(f"controller command failed: {data.get('meta')}")
        return data

    def devices(self, site: str) -> list[dict[str, object]]:
        data = self.get_json(f"/api/s/{site}/stat/device")
        if data.get("meta", {}).get("rc") != "ok":
            raise RuntimeError(f"controller device list failed: {data.get('meta')}")
        devices = data.get("data", [])
        if not isinstance(devices, list):
            raise RuntimeError("controller device list returned non-list data")
        return [device for device in devices if isinstance(device, dict)]

    def status(self) -> dict[str, object]:
        data = self.get_json("/status")
        if data.get("meta", {}).get("rc") != "ok":
            raise RuntimeError(f"controller status failed: {data.get('meta')}")
        return data

    def get_json(self, path: str) -> dict[str, object]:
        with self.opener.open(self.base_url + path, timeout=10) as response:
            return json.loads(response.read().decode("utf-8"))

    def post_json(self, path: str, payload: dict[str, object]) -> dict[str, object]:
        request = urllib.request.Request(
            self.base_url + path,
            data=json.dumps(payload).encode("utf-8"),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        try:
            with self.opener.open(request, timeout=10) as response:
                return json.loads(response.read().decode("utf-8"))
        except urllib.error.HTTPError as exc:
            body = exc.read().decode("utf-8", errors="replace")
            raise RuntimeError(f"controller API HTTP {exc.code}: {body}") from exc


def wait_for_device(
    client: ControllerClient,
    site: str,
    mac: str,
    timeout: int,
    adopted: bool | None,
) -> dict[str, object]:
    deadline = time.time() + timeout
    while time.time() <= deadline:
        for device in client.devices(site):
            if str(device.get("mac", "")).lower() != mac:
                continue
            if adopted is None or bool(device.get("adopted")) is adopted:
                return device
        time.sleep(2)
    wanted = "present" if adopted is None else f"adopted={adopted}"
    raise RuntimeError(f"timed out waiting for controller device {mac} ({wanted})")


def wait_until_clean(client: ControllerClient, site: str, mac: str, timeout: int) -> None:
    deadline = time.time() + timeout
    while time.time() <= deadline:
        devices = [device for device in client.devices(site) if str(device.get("mac", "")).lower() == mac]
        if not devices or not bool(devices[0].get("adopted")):
            return
        time.sleep(2)
    raise RuntimeError(f"timed out waiting for controller device {mac} to leave adopted state")


def controller_version_summary(client: ControllerClient, expected_version: str | None) -> str:
    status = client.status()
    meta = status.get("meta", {})
    if not isinstance(meta, dict):
        raise RuntimeError("controller status returned non-object meta")
    version = str(meta.get("server_version", ""))
    if not version:
        raise RuntimeError(f"controller status did not include server_version: {meta}")
    if expected_version and version != expected_version:
        raise RuntimeError(f"controller version {version} != expected {expected_version}")
    return f"controller version: server_version={version} up={meta.get('up')}"


def device_summary(prefix: str, device: dict[str, object]) -> str:
    fields = {
        "mac": device.get("mac"),
        "model": device.get("model"),
        "name": device.get("name"),
        "ip": device.get("ip"),
        "adopted": device.get("adopted"),
        "state": device.get("state"),
        "adoption_completed": device.get("adoption_completed"),
    }
    return prefix + ": " + " ".join(f"{key}={value}" for key, value in fields.items())


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"controller adoption failed: {exc}", file=sys.stderr)
        raise SystemExit(1)
