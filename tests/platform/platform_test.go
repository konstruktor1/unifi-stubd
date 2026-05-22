package platform_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

// TestParseLLDPCLIJSON verifies lldpcli JSON shape variants normalize to
// portable neighbors.
func TestParseLLDPCLIJSON(t *testing.T) {
	t.Parallel()

	data := []byte(`{
	  "lldp": {
	    "interface": {
	      "eth0": {
	        "chassis": {
	          "switch-a": {
	            "id": {"type": "mac", "value": "00:11:22:33:44:55"},
	            "name": {"value": "switch-a"},
	            "mgmt-ip": {"value": "192.0.2.2"},
	            "capability": [{"type": "Bridge"}, {"type": "Router"}]
	          }
	        },
	        "port": {
	          "Gi1/0/1": {
	            "id": {"type": "ifname", "value": "Gi1/0/1"},
	            "descr": {"value": "uplink"}
	          }
	        }
	      }
	    }
	  }
	}`)

	neighbors, err := platform.ParseLLDPCLIJSON(data)
	if err != nil {
		t.Fatalf("parse lldp json: %v", err)
	}
	if len(neighbors) != 1 {
		t.Fatalf("got %d neighbors, want 1", len(neighbors))
	}
	neighbor := neighbors[0]
	if neighbor.Interface != "eth0" {
		t.Fatalf("interface = %q, want eth0", neighbor.Interface)
	}
	if neighbor.ChassisMAC != "00:11:22:33:44:55" {
		t.Fatalf("chassis mac = %q", neighbor.ChassisMAC)
	}
	if neighbor.SystemName != "switch-a" {
		t.Fatalf("system name = %q", neighbor.SystemName)
	}
	if neighbor.PortID != "Gi1/0/1" {
		t.Fatalf("port id = %q", neighbor.PortID)
	}
	if neighbor.ManagementIP != "192.0.2.2" {
		t.Fatalf("management ip = %q", neighbor.ManagementIP)
	}
	if len(neighbor.Capabilities) != 2 {
		t.Fatalf("capabilities = %#v, want 2 entries", neighbor.Capabilities)
	}
}

// TestSyslogReaderReturnsRecentEntries verifies syslog reading keeps only the
// requested recent entries.
func TestSyslogReaderReturnsRecentEntries(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "messages")
	content := "Jan  2 03:04:05 host kernel: older line\n" +
		"Jan  2 03:04:06 host unifi-stubd: platform ready\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write syslog fixture: %v", err)
	}

	plt := platform.NewForOS("freebsd", platform.Config{})
	entries, errs := plt.Logs(context.Background(), platform.LogConfig{
		Source: platform.LogSourceSyslog,
		Path:   path,
		Lines:  1,
	})
	if len(errs) > 0 {
		t.Fatalf("unexpected syslog errors: %v", errs)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Unit != "unifi-stubd" {
		t.Fatalf("unit = %q, want unifi-stubd", entries[0].Unit)
	}
	if entries[0].Message != "platform ready" {
		t.Fatalf("message = %q, want platform ready", entries[0].Message)
	}
}

// TestFreeBSDCapabilityReportsSyslogPath verifies FreeBSD capability output can
// report configured syslog availability.
func TestFreeBSDCapabilityReportsSyslogPath(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "messages")
	if err := os.WriteFile(path, []byte("Jan  2 03:04:06 host app: ok\n"), 0o600); err != nil {
		t.Fatalf("write syslog fixture: %v", err)
	}

	cfg := platform.Config{LogSource: platform.LogSourceSyslog, SyslogPath: path}
	report := platform.NewForOS("freebsd", cfg).Capabilities(context.Background(), cfg)
	if report.GOOS != "freebsd" {
		t.Fatalf("goos = %q, want freebsd", report.GOOS)
	}
	for _, capability := range report.Capabilities {
		if capability.Name == "logs" {
			if capability.State != "available" {
				t.Fatalf("logs state = %q, want available", capability.State)
			}
			if capability.Detail != path {
				t.Fatalf("logs detail = %q, want %q", capability.Detail, path)
			}
			return
		}
	}
	t.Fatalf("logs capability not reported: %#v", report.Capabilities)
}
