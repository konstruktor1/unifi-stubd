package main

import (
	"encoding/hex"
	"fmt"
)

func printDryRun(packet, payload []byte) {
	fmt.Println("discovery_packet_hex:")
	fmt.Println(hex.EncodeToString(packet))
	fmt.Println()
	fmt.Println("minimal_inform_payload_json:")
	fmt.Println(string(payload))
}
