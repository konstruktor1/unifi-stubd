package wanhealth

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Ping executes one OS ping. A context timeout bounds platforms whose ping
// flags differ, so no raw ICMP privileges or platform-specific timeout flags
// are required here.
func (CommandRunner) Ping(ctx context.Context, host string, _ time.Duration) ([]byte, time.Duration, error) {
	start := time.Now()
	cmd := exec.CommandContext(ctx, "ping", "-c", "1", host)
	output, err := cmd.CombinedOutput()
	elapsed := time.Since(start)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return output, elapsed, fmt.Errorf("ping timeout: %w", ctxErr)
	}
	if err != nil {
		message := firstNonEmptyLine(string(output))
		if message != "" {
			return output, elapsed, fmt.Errorf("%w: %s", err, message)
		}
		return output, elapsed, fmt.Errorf("run ping: %w", err)
	}
	return output, elapsed, nil
}
