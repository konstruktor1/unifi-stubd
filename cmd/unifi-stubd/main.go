// Command unifi-stubd emulates a minimal UniFi device for controller lab work.
package main

import "log"

var version = "dev"

func main() {
	if err := serveSwitchEmulation(); err != nil {
		log.Fatal(err)
	}
}
