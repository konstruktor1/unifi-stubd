package device

import "testing"

func TestTargetAddressFromInformURL(t *testing.T) {
	got, err := targetAddress("http://10.10.0.30:8080/inform")
	if err != nil {
		t.Fatal(err)
	}
	if got != "10.10.0.30:8080" {
		t.Fatalf("targetAddress = %q, want 10.10.0.30:8080", got)
	}
}

func TestTargetAddressFromRawIP(t *testing.T) {
	got, err := targetAddress("10.10.0.30")
	if err != nil {
		t.Fatal(err)
	}
	if got != "10.10.0.30:9" {
		t.Fatalf("targetAddress = %q, want 10.10.0.30:9", got)
	}
}
