# Decode Inform Tool

This command decodes a single UniFi inform packet body and prints a JSON view
of both transport metadata and decoded payload content. It uses the same
`internal/inform` package as the Go daemon tests, so packet handling stays
aligned with production code instead of living in a separate research script.

Run it from the repository root:

```sh
go run ./lab/gateway-profiles/uxgpro/tools/decode-inform FILE
```

By default it uses the UniFi default inform key. Pass `-key-hex` only for
sanitized lab keys. The command is meant for temporary local captures; commit
only reduced findings or sanitized fixtures, never raw private captures or real
controller keys.
