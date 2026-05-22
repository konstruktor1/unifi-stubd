// Command unifi-stubd emulates a minimal UniFi device for controller lab work.
package main

import (
	"errors"
	"fmt"
	"os"
)

// version is replaced by release builds; dev keeps local runs identifiable.
var version = "dev"

// main runs the daemon entry point and maps known CLI errors to exit codes.
func main() {
	if err := serveSwitchEmulation(); err != nil {
		var exit interface {
			ExitCode() int
		}
		code := 1
		if errors.As(err, &exit) {
			code = exit.ExitCode()
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code)
	}
}
