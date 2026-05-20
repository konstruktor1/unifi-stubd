# UBNTHAL Mock Identity

UDM firmware reads `/proc/ubnthal` for board and manufacturing metadata before
many higher-level services are ready. These files provide a deterministic lab
identity for that early userspace boundary.

`board` carries the model, board id, base MAC, and serial-like fields. 
`system.info` carries CPU, flash/RAM size, board revision, and per-port MAC
values. The values are chosen to be consistent with a UDM Pro SE shape while
remaining synthetic documentation-range lab data.

Do not replace these files with data copied from a real console. Real serials,
MAC addresses, QR IDs, and manufacturing fields would turn a reproducible lab
fixture into private device data.
