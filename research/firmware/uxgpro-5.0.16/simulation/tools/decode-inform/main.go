// Command decode-inform decodes one captured UniFi inform packet body.
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

type decodedInform struct {
	Packet  packetSummary `json:"packet"`
	Payload any           `json:"payload"`
}

type packetSummary struct {
	MAC            string `json:"mac"`
	Flags          uint16 `json:"flags"`
	Encrypted      bool   `json:"encrypted"`
	Zlib           bool   `json:"zlib"`
	Snappy         bool   `json:"snappy"`
	EncryptedGCM   bool   `json:"encrypted_gcm"`
	IVHex          string `json:"iv_hex"`
	PayloadBytes   int    `json:"payload_bytes"`
	DecodedBytes   int    `json:"decoded_bytes"`
	PacketVersion  uint32 `json:"packet_version"`
	PayloadVersion uint32 `json:"payload_version"`
}

func main() {
	var keyHex string
	flag.StringVar(&keyHex, "key-hex", "", "16-byte inform auth key as hex; defaults to UniFi default key")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage: go run ./research/firmware/uxgpro-5.0.16/simulation/tools/decode-inform [-key-hex HEX] FILE\n")
		os.Exit(2)
	}

	key := inform.DefaultAuthKey()
	if keyHex != "" {
		decoded, err := hex.DecodeString(keyHex)
		if err != nil {
			fail("decode key hex: %v", err)
		}
		key = decoded
	}

	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fail("read inform packet: %v", err)
	}

	packet, body, err := inform.Decode(data, key)
	if err != nil {
		fail("decode inform packet: %v", err)
	}

	payload := any(json.RawMessage(body))
	if !json.Valid(body) {
		payload = string(body)
	}

	out := decodedInform{
		Packet: packetSummary{
			MAC:            packet.MAC.String(),
			Flags:          packet.Flags,
			Encrypted:      packet.Flags&inform.FlagEncrypted != 0,
			Zlib:           packet.Flags&inform.FlagZlib != 0,
			Snappy:         packet.Flags&inform.FlagSnappy != 0,
			EncryptedGCM:   packet.Flags&inform.FlagEncryptedGCM != 0,
			IVHex:          hex.EncodeToString(packet.IV),
			PayloadBytes:   len(packet.Payload),
			DecodedBytes:   len(body),
			PacketVersion:  inform.PacketVersion,
			PayloadVersion: inform.PayloadVersion,
		},
		Payload: payload,
	}

	encoded, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fail("marshal decoded output: %v", err)
	}
	fmt.Println(string(encoded))
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
