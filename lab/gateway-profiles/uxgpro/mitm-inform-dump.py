"""Capture UniFi inform requests and responses in the UXG-Pro lab."""

from __future__ import annotations

import hashlib
import json
import os
import time
from pathlib import Path

from mitmproxy import ctx, http


CAPTURE_DIR = Path(os.environ.get("MITM_CAPTURE_DIR", "/captures"))
EVENTS_PATH = CAPTURE_DIR / "events.jsonl"


def load(_loader) -> None:
    CAPTURE_DIR.mkdir(parents=True, exist_ok=True)
    ctx.log.info(f"writing inform captures to {CAPTURE_DIR}")


def request(flow: http.HTTPFlow) -> None:
    if flow.request.path != "/inform":
        return

    event_id = _event_id(flow)
    flow.metadata["inform_event_id"] = event_id
    body_path = CAPTURE_DIR / f"{event_id}-request.bin"
    body_path.write_bytes(flow.request.raw_content or b"")

    _write_event(
        {
            "event": "request",
            "id": event_id,
            "timestamp": _timestamp(),
            "client": _client(flow),
            "method": flow.request.method,
            "url": flow.request.pretty_url,
            "http_version": flow.request.http_version,
            "headers": _headers(flow.request.headers),
            "body_path": str(body_path),
            "body_bytes": len(flow.request.raw_content or b""),
            "body_sha256": _sha256(flow.request.raw_content or b""),
            "tnbu": _tnbu_summary(flow.request.raw_content or b""),
        }
    )


def response(flow: http.HTTPFlow) -> None:
    if flow.request.path != "/inform" or flow.response is None:
        return

    event_id = flow.metadata.get("inform_event_id") or _event_id(flow)
    body_path = CAPTURE_DIR / f"{event_id}-response.bin"
    body_path.write_bytes(flow.response.raw_content or b"")

    _write_event(
        {
            "event": "response",
            "id": event_id,
            "timestamp": _timestamp(),
            "status_code": flow.response.status_code,
            "reason": flow.response.reason,
            "headers": _headers(flow.response.headers),
            "body_path": str(body_path),
            "body_bytes": len(flow.response.raw_content or b""),
            "body_sha256": _sha256(flow.response.raw_content or b""),
            "tnbu": _tnbu_summary(flow.response.raw_content or b""),
        }
    )


def _event_id(flow: http.HTTPFlow) -> str:
    started = int((flow.request.timestamp_start or time.time()) * 1000)
    return f"{started}-{flow.id[:8]}"


def _timestamp() -> str:
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())


def _client(flow: http.HTTPFlow) -> str:
    peer = flow.client_conn.peername
    if peer is None:
        return ""
    return f"{peer[0]}:{peer[1]}"


def _headers(headers: http.Headers) -> dict[str, str]:
    return {key: value for key, value in headers.items()}


def _sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def _tnbu_summary(data: bytes) -> dict[str, int | str | bool]:
    if len(data) < 40 or data[:4] != b"TNBU":
        return {"present": False}
    return {
        "present": True,
        "packet_version": int.from_bytes(data[4:8], "big"),
        "mac": data[8:14].hex(":"),
        "flags": int.from_bytes(data[14:16], "big"),
        "iv_hex": data[16:32].hex(),
        "payload_version": int.from_bytes(data[32:36], "big"),
        "payload_bytes": int.from_bytes(data[36:40], "big"),
    }


def _write_event(event: dict[str, object]) -> None:
    with EVENTS_PATH.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(event, sort_keys=True) + "\n")
    ctx.log.info(
        f"inform {event['event']} id={event['id']} "
        f"bytes={event['body_bytes']} tnbu={event['tnbu']}"
    )
