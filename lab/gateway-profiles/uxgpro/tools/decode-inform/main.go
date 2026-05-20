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

// decodedInform is the JSON output produced for one decoded packet.
type decodedInform struct {
	// Packet contains transport-level inform packet metadata.
	Packet packetSummary `json:"packet"`
	// Payload contains the decoded JSON body or raw text when it is not JSON.
	Payload any `json:"payload"`
}

// packetSummary keeps transport metadata separate from the decoded JSON body.
type packetSummary struct {
	// MAC is the device address from the inform packet header.
	MAC string `json:"mac"`
	// Flags contains the raw inform feature bitmask.
	Flags uint16 `json:"flags"`
	// Encrypted reports whether the encrypted flag is set.
	Encrypted bool `json:"encrypted"`
	// Zlib reports whether the zlib compression flag is set.
	Zlib bool `json:"zlib"`
	// Snappy reports whether the snappy compression flag is set.
	Snappy bool `json:"snappy"`
	// EncryptedGCM reports whether AES-GCM was used.
	EncryptedGCM bool `json:"encrypted_gcm"`
	// IVHex is the packet IV or nonce encoded as hex.
	IVHex string `json:"iv_hex"`
	// PayloadBytes is the encoded payload length in bytes.
	PayloadBytes int `json:"payload_bytes"`
	// DecodedBytes is the decoded payload length in bytes.
	DecodedBytes int `json:"decoded_bytes"`
	// PacketVersion is the inform header version decoded by the tool.
	PacketVersion uint32 `json:"packet_version"`
	// PayloadVersion is the inform payload version decoded by the tool.
	PayloadVersion uint32 `json:"payload_version"`
}

func main() {
	var keyHex string
	flag.StringVar(&keyHex, "key-hex", "", "16-byte inform auth key as hex; defaults to UniFi default key")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage: go run ./lab/gateway-profiles/uxgpro/tools/decode-inform [-key-hex HEX] FILE\n")
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

// fail reports a CLI error in a stable one-line format and exits non-zero.
func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
