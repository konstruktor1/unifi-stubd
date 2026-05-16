.PHONY: test run dry-run fmt

test:
	go test ./...

run:
	go run ./cmd/unifi-stubd

dry-run:
	go run ./cmd/unifi-stubd -dry-run

fmt:
	gofmt -w cmd internal
