# Generic Stub Compose Lab

This is the main Docker lab for the Go `unifi-stubd` daemon. It starts a UniFi
Network Application, MongoDB, an inform MITM, and one `stub` container built
from the repository root. Use it when validating discovery, inform, adoption
state, or profile payload behavior in the Go service.

The lab is deliberately named `stub` everywhere: Compose service, container
name, hostname, and persistent volume. That keeps the generic daemon lab
separate from firmware research directories, where containers are wrappers
around extracted vendor root filesystems.

Captured inform traffic is local output and belongs in the ignored
`captures/` directory. Do not commit raw controller captures, adoption keys,
tokens, private URLs, or device-specific data from this lab.
