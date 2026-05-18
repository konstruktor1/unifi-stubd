#!/usr/bin/env python3
"""Build a GPT disk image for UDM Pro SE QEMU VM boot attempts."""

from __future__ import annotations

import argparse
import os
import shutil
import struct
import subprocess
import uuid
import zlib
from pathlib import Path


SECTOR_SIZE = 512
ENTRY_SIZE = 128
ENTRY_COUNT = 128
ENTRY_LBAS = (ENTRY_SIZE * ENTRY_COUNT + SECTOR_SIZE - 1) // SECTOR_SIZE
FIRST_PARTITION_LBA = 2048
LINUX_FS_GUID = uuid.UUID("0fc63daf-8483-4772-8e79-3d69d8477de4")


def align(value: int, unit: int) -> int:
    return ((value + unit - 1) // unit) * unit


def run(command: list[str]) -> None:
    subprocess.run(command, check=True)


def make_ext4(path: Path, label: str, size_mb: int, source_dir: Path | None = None) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    if path.exists():
        path.unlink()
    with path.open("wb") as handle:
        handle.truncate(size_mb * 1024 * 1024)

    command = ["mke2fs", "-q", "-t", "ext4", "-F", "-L", label]
    if source_dir is not None:
        command.extend(["-d", str(source_dir)])
    command.append(str(path))
    run(command)


def utf16_partition_name(name: str) -> bytes:
    raw = name.encode("utf-16le")
    return raw + b"\0" * (72 - len(raw))


def write_gpt_header(
    disk: bytearray,
    current_lba: int,
    backup_lba: int,
    first_usable: int,
    last_usable: int,
    disk_guid: uuid.UUID,
    entries_lba: int,
    entries_crc: int,
) -> None:
    header = bytearray(SECTOR_SIZE)
    struct.pack_into(
        "<8sIIIIQQQQ16sQIII",
        header,
        0,
        b"EFI PART",
        0x00010000,
        92,
        0,
        0,
        current_lba,
        backup_lba,
        first_usable,
        last_usable,
        disk_guid.bytes_le,
        entries_lba,
        ENTRY_COUNT,
        ENTRY_SIZE,
        entries_crc,
    )
    crc = zlib.crc32(header[:92]) & 0xFFFFFFFF
    struct.pack_into("<I", header, 16, crc)
    offset = current_lba * SECTOR_SIZE
    disk[offset : offset + SECTOR_SIZE] = header


def build_disk(out_path: Path, partitions: list[tuple[str, Path]]) -> None:
    partition_entries = bytearray(ENTRY_LBAS * SECTOR_SIZE)
    partition_layout = []
    next_lba = FIRST_PARTITION_LBA

    for index, (name, image_path) in enumerate(partitions):
        size = image_path.stat().st_size
        sectors = align(size, SECTOR_SIZE) // SECTOR_SIZE
        first_lba = align(next_lba, 2048)
        last_lba = first_lba + sectors - 1
        partition_layout.append((name, image_path, first_lba, last_lba))
        next_lba = last_lba + 1

        entry_offset = index * ENTRY_SIZE
        unique_guid = uuid.uuid5(uuid.NAMESPACE_URL, f"unifi-stubd-udm-pro-se-vm-{name}")
        struct.pack_into(
            "<16s16sQQQ72s",
            partition_entries,
            entry_offset,
            LINUX_FS_GUID.bytes_le,
            unique_guid.bytes_le,
            first_lba,
            last_lba,
            0,
            utf16_partition_name(name),
        )

    backup_entries_lba = align(next_lba, 2048)
    backup_lba = backup_entries_lba + ENTRY_LBAS
    total_lbas = backup_lba + 1
    first_usable = FIRST_PARTITION_LBA
    last_usable = backup_entries_lba - 1
    entries_crc = zlib.crc32(partition_entries) & 0xFFFFFFFF
    disk_guid = uuid.uuid5(uuid.NAMESPACE_URL, "unifi-stubd-udm-pro-se-vm")

    out_path.parent.mkdir(parents=True, exist_ok=True)
    disk = bytearray(SECTOR_SIZE * 34)

    protective_mbr = bytearray(SECTOR_SIZE)
    protective_mbr[446 + 4] = 0xEE
    struct.pack_into("<I", protective_mbr, 446 + 8, 1)
    struct.pack_into("<I", protective_mbr, 446 + 12, min(total_lbas - 1, 0xFFFFFFFF))
    protective_mbr[510:512] = b"\x55\xaa"
    disk[0:SECTOR_SIZE] = protective_mbr
    disk[2 * SECTOR_SIZE : 2 * SECTOR_SIZE + len(partition_entries)] = partition_entries
    write_gpt_header(disk, 1, backup_lba, first_usable, last_usable, disk_guid, 2, entries_crc)

    with out_path.open("wb") as handle:
        handle.truncate(total_lbas * SECTOR_SIZE)
        handle.seek(0)
        handle.write(disk)

        for _name, image_path, first_lba, _last_lba in partition_layout:
            handle.seek(first_lba * SECTOR_SIZE)
            with image_path.open("rb") as part:
                shutil.copyfileobj(part, handle, 1024 * 1024)

        handle.seek(backup_entries_lba * SECTOR_SIZE)
        handle.write(partition_entries)
        backup_header = bytearray(SECTOR_SIZE)
        # Reuse the header writer on a sector-sized buffer by temporarily
        # composing the final sectors in memory.
        tail = bytearray((ENTRY_LBAS + 1) * SECTOR_SIZE)
        tail[0 : len(partition_entries)] = partition_entries
        write_gpt_header(tail, ENTRY_LBAS, 1, first_usable, last_usable, disk_guid, 0, entries_crc)
        backup_header[:] = tail[ENTRY_LBAS * SECTOR_SIZE : (ENTRY_LBAS + 1) * SECTOR_SIZE]
        struct.pack_into("<Q", backup_header, 24, backup_lba)
        struct.pack_into("<Q", backup_header, 32, 1)
        struct.pack_into("<Q", backup_header, 72, backup_entries_lba)
        struct.pack_into("<I", backup_header, 16, 0)
        crc = zlib.crc32(backup_header[:92]) & 0xFFFFFFFF
        struct.pack_into("<I", backup_header, 16, crc)
        handle.seek(backup_lba * SECTOR_SIZE)
        handle.write(backup_header)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--artifacts", required=True, type=Path)
    args = parser.parse_args()

    if shutil.which("mke2fs") is None:
        raise SystemExit("mke2fs is required to build the VM disk image")

    artifacts = args.artifacts
    parts = artifacts / "parts"
    staging = artifacts / "disk-staging"
    if staging.exists():
        shutil.rmtree(staging)
    (staging / "boot").mkdir(parents=True)
    (staging / "root").mkdir(parents=True)

    shutil.copy2(artifacts / "kernel.fit", staging / "boot" / "uImage")
    shutil.copy2(artifacts / "rootfs.squashfs", staging / "root" / "rootfs")

    make_ext4(parts / "boot.ext4", "boot", 96, staging / "boot")
    make_ext4(parts / "recovery.ext4", "recovery", 96)
    make_ext4(parts / "root.ext4", "root", 1400, staging / "root")
    make_ext4(parts / "config.ext4", "config", 32)
    make_ext4(parts / "log.ext4", "log", 256)
    make_ext4(parts / "persistent.ext4", "persistent", 512)
    make_ext4(parts / "overlay.ext4", "overlay", 1024)

    build_disk(
        artifacts / "vm-disk.raw",
        [
            ("boot", parts / "boot.ext4"),
            ("recovery", parts / "recovery.ext4"),
            ("root", parts / "root.ext4"),
            ("config", parts / "config.ext4"),
            ("log", parts / "log.ext4"),
            ("persistent", parts / "persistent.ext4"),
            ("overlay", parts / "overlay.ext4"),
        ],
    )
    print(f"wrote {artifacts / 'vm-disk.raw'}")


if __name__ == "__main__":
    main()
