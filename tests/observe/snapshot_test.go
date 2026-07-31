package observe_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

func TestReadLinuxInterfaceSnapshot(t *testing.T) {
	root := t.TempDir()
	values := map[string]string{
		"statistics/rx_bytes":   "1200",
		"statistics/tx_bytes":   "3400",
		"statistics/rx_packets": "12",
		"statistics/tx_packets": "34",
		"statistics/rx_errors":  "1",
		"statistics/tx_errors":  "2",
		"speed":                 "2500",
	}
	for name, value := range values {
		path := filepath.Join(root, "class", "net", "eno1", name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	stats, err := observe.ReadInterfaceStats(root, "eno1")
	if err != nil {
		t.Fatal(err)
	}
	if !stats.CountersSet || stats.RXBytes != 1200 || stats.TXBytes != 3400 || stats.SpeedMbps != 2500 {
		t.Fatalf("interface snapshot = %+v", stats)
	}
	if stats.RXPackets != 12 || stats.TXPackets != 34 || stats.RXErrors != 1 || stats.TXErrors != 2 {
		t.Fatalf("packet/error snapshot = %+v", stats)
	}
}
