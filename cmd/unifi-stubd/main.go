// Command unifi-stubd emulates a minimal UniFi device for controller lab work.
package main

import (
	"errors"
	"fmt"
	"os"
)

var version = "dev"

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
