#!/usr/bin/env python3
"""Extract UDM Pro SE VM boot artifacts from the vendor firmware image."""

from __future__ import annotations

import argparse
import gzip
import hashlib
import json
import struct
from pathlib import Path


EXPECTED_SHA256 = "7cf58f4563522220716f5025a7b2954b070df6be9364d7f60af0bc644512bce4"
UBOOT_CODE_OFFSET = 352
KERNEL_FILE_HEADER_OFFSET = 1_499_240
KERNEL_FIT_OFFSET = 1_499_296
ROOTFS_FILE_HEADER_OFFSET = 16_141_086
ROOTFS_OFFSET = 16_141_142

FDT_BEGIN_NODE = 1
FDT_END_NODE = 2
FDT_PROP = 3
FDT_NOP = 4
FDT_END = 9


def align4(value: int) -> int:
    return (value + 3) & ~3


def c_string(data: bytes, offset: int) -> str:
    end = data.index(b"\0", offset)
    return data[offset:end].decode("utf-8", "replace")


def parse_fdt_properties(blob: bytes) -> dict[tuple[str, str], bytes]:
    header = struct.unpack(">10I", blob[:40])
    magic, total_size, off_struct, off_strings, _off_rsvmap, _version, _last, _bootcpu, size_strings, size_struct = header
    if magic != 0xD00DFEED:
        raise ValueError(f"not an FDT/FIT blob: magic=0x{magic:08x}")

    struct_block = blob[off_struct : off_struct + size_struct]
    strings = blob[off_strings : off_strings + size_strings]
    if total_size > len(blob):
        raise ValueError(f"truncated FDT/FIT blob: header size {total_size}, bytes {len(blob)}")

    props: dict[tuple[str, str], bytes] = {}
    stack: list[str] = []
    pos = 0
    while pos < len(struct_block):
        token = struct.unpack(">I", struct_block[pos : pos + 4])[0]
        pos += 4
        if token == FDT_BEGIN_NODE:
            end = struct_block.index(b"\0", pos)
            stack.append(struct_block[pos:end].decode("utf-8", "replace"))
            pos = align4(end + 1)
        elif token == FDT_END_NODE:
            stack.pop()
        elif token == FDT_PROP:
            length, name_offset = struct.unpack(">II", struct_block[pos : pos + 8])
            pos += 8
            value = struct_block[pos : pos + length]
            pos = align4(pos + length)
            name = c_string(strings, name_offset)
            path = "/" + "/".join(part for part in stack if part)
            props[(path, name)] = value
        elif token == FDT_NOP:
            continue
        elif token == FDT_END:
            break
        else:
            raise ValueError(f"unknown FDT token {token} at structure offset {pos - 4}")
    return props


def text_prop(props: dict[tuple[str, str], bytes], path: str, name: str) -> str:
    return props.get((path, name), b"").split(b"\0", 1)[0].decode("utf-8", "replace")


def write_image(props: dict[tuple[str, str], bytes], image: str, out_dir: Path) -> dict[str, object]:
    path = f"/images/{image}"
    data = props[(path, "data")]
    image_name = image.replace("@", "-")
    compression = text_prop(props, path, "compression")
    image_type = text_prop(props, path, "type")
    description = text_prop(props, path, "description")

    if image == "kernel@1":
        compressed_path = out_dir / "kernel.Image.gz"
        raw_path = out_dir / "kernel.Image"
    elif image == "fdt@1":
        compressed_path = out_dir / "udm-pro-se.dtb"
        raw_path = compressed_path
    elif image == "ramdisk@1":
        compressed_path = out_dir / "initramfs.cpio.gz"
        raw_path = out_dir / "initramfs.cpio"
    else:
        compressed_path = out_dir / image_name
        raw_path = compressed_path

    compressed_path.write_bytes(data)
    raw_bytes = data
    if compression == "gzip":
        raw_bytes = gzip.decompress(data)
        raw_path.write_bytes(raw_bytes)

    return {
        "image": image,
        "description": description,
        "type": image_type,
        "compression": compression or "none",
        "compressed_path": str(compressed_path),
        "compressed_bytes": len(data),
        "raw_path": str(raw_path),
        "raw_bytes": len(raw_bytes),
    }


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--firmware", required=True, type=Path)
    parser.add_argument("--out", required=True, type=Path)
    parser.add_argument("--no-sha-check", action="store_true")
    args = parser.parse_args()

    firmware = args.firmware.read_bytes()
    sha256 = hashlib.sha256(firmware).hexdigest()
    if not args.no_sha_check and sha256 != EXPECTED_SHA256:
        raise SystemExit(f"firmware sha256 mismatch: got {sha256}, expected {EXPECTED_SHA256}")

    args.out.mkdir(parents=True, exist_ok=True)
    (args.out / "firmware.sha256").write_text(f"{sha256}  {args.firmware.name}\n", encoding="utf-8")

    kernel_fit = firmware[KERNEL_FIT_OFFSET:ROOTFS_FILE_HEADER_OFFSET]
    props = parse_fdt_properties(kernel_fit)
    fit_total_size = struct.unpack(">I", kernel_fit[4:8])[0]
    kernel_fit = kernel_fit[:fit_total_size]

    (args.out / "kernel.fit").write_bytes(kernel_fit)
    (args.out / "rootfs.squashfs").write_bytes(firmware[ROOTFS_OFFSET:])
    (args.out / "uboot.bin").write_bytes(firmware[UBOOT_CODE_OFFSET:KERNEL_FILE_HEADER_OFFSET])

    images = [
        write_image(props, "kernel@1", args.out),
        write_image(props, "fdt@1", args.out),
        write_image(props, "ramdisk@1", args.out),
    ]

    summary = {
        "firmware": str(args.firmware),
        "sha256": sha256,
        "offsets": {
            "uboot_code": UBOOT_CODE_OFFSET,
            "kernel_fit": KERNEL_FIT_OFFSET,
            "rootfs_squashfs": ROOTFS_OFFSET,
        },
        "images": images,
    }
    (args.out / "extract-summary.json").write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
