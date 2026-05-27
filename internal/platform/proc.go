package platform

import (
	"context"
	"fmt"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/observe"
	"github.com/prometheus/procfs"
)

// Proc reads Linux procfs counters when enabled; unsupported or disabled
// sources return warnings instead of installing dependencies.
func (p hostPlatform) Proc(_ context.Context, cfg ProcConfig) (ProcSnapshot, []error) {
	source := normalizedSource(cfg.Source)
	if source == SourceOff {
		return ProcSnapshot{}, nil
	}
	if source != ProcSourceProcFS {
		return ProcSnapshot{}, []error{fmt.Errorf("unsupported proc source %q", source)}
	}
	root := strings.TrimSpace(cfg.Root)
	if root == "" {
		root = "/proc"
	}
	fs, err := procfs.NewFS(root)
	if err != nil {
		return ProcSnapshot{}, []error{fmt.Errorf("open procfs %s: %w", root, err)}
	}
	netdev, err := fs.NetDev()
	if err != nil {
		return ProcSnapshot{}, []error{fmt.Errorf("read procfs netdev: %w", err)}
	}
	out := ProcSnapshot{Interfaces: map[string]observe.InterfaceStats{}}
	for name, line := range netdev {
		out.Interfaces[name] = observe.InterfaceStats{
			RXBytes:   int64(line.RxBytes),
			TXBytes:   int64(line.TxBytes),
			RXPackets: int64(line.RxPackets),
			TXPackets: int64(line.TxPackets),
			RXErrors:  int64(line.RxErrors),
			TXErrors:  int64(line.TxErrors),
		}
	}
	return out, nil
}
