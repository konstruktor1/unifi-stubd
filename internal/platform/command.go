package platform

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// commandContext runs bounded read-only host commands for optional integrations.
func commandContext(ctx context.Context, timeout time.Duration, name string, args ...string) ([]byte, error) {
	if timeout <= 0 {
		timeout = defaultCommandTimeout
	}
	commandCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	out, err := exec.CommandContext(commandCtx, name, args...).Output()
	if errors.Is(commandCtx.Err(), context.DeadlineExceeded) {
		return out, fmt.Errorf("%s timed out after %s", name, timeout)
	}
	if err != nil {
		return out, fmt.Errorf("run %s: %w", name, err)
	}
	return out, nil
}
