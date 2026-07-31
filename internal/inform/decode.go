package inform

import (
	"encoding/binary"
	"fmt"
	"net"
)

// Decode unwraps a UniFi inform packet and returns decoded JSON payload bytes.
// It is direction-neutral: device requests may carry zlib-compressed JSON.
// Controller responses use the stricter DecodeControllerResponse policy.
func Decode(data []byte, key []byte) (*Packet, []byte, error) {
	p, header, err := parsePacket(data)
	if err != nil {
		return nil, nil, err
	}
	if p.Flags&FlagSnappy != 0 {
		return nil, nil, fmt.Errorf("snappy inform payload is unsupported")
	}

	body, err := decryptPayload(p, append([]byte{}, p.Payload...), key, header)
	if err != nil {
		return nil, nil, err
	}
	if p.Flags&FlagZlib != 0 {
		body, err = decompressPayload(body)
		if err != nil {
			return nil, nil, err
		}
	}
	if err := validateDecodedPayloadSize(body); err != nil {
		return p, nil, err
	}
	return p, body, nil
}

// DecodeControllerResponse binds the controller response cipher to the mode
// already selected for the request. Decrypted controller text is returned
// directly for JSON processing; response-side zlib and Snappy are unsupported.
//
// Once the fixed envelope is valid, Packet is retained on policy or decryption
// failures so callers can expose bounded, non-secret TNBU diagnostics.
func DecodeControllerResponse(data []byte, key []byte, expectedGCM bool) (*Packet, []byte, error) {
	p, header, err := parsePacket(data)
	if err != nil {
		return p, nil, err
	}

	usesGCM := p.Flags&FlagEncryptedGCM != 0
	if usesGCM != expectedGCM {
		expected := "AES-CBC"
		actual := "AES-GCM"
		if expectedGCM {
			expected = "AES-GCM"
			actual = "AES-CBC"
		}
		return p, nil, fmt.Errorf(
			"controller response cipher mismatch: expected %s, got %s",
			expected,
			actual,
		)
	}

	body, err := decryptPayload(p, append([]byte{}, p.Payload...), key, header)
	if err != nil {
		return p, nil, err
	}
	if p.Flags&FlagZlib != 0 {
		return p, nil, fmt.Errorf("zlib-compressed controller responses are unsupported")
	}
	if p.Flags&FlagSnappy != 0 {
		return p, nil, fmt.Errorf("snappy-compressed controller responses are unsupported")
	}
	if err := validateDecodedPayloadSize(body); err != nil {
		return p, nil, err
	}
	return p, body, nil
}

func validateDecodedPayloadSize(body []byte) error {
	if len(body) > MaxDecodedPayloadBytes {
		return fmt.Errorf("decoded inform payload exceeds %d bytes", MaxDecodedPayloadBytes)
	}
	return nil
}

// parsePacket validates and copies the fixed TNBU envelope without decrypting
// its payload. Errors found after the exact envelope length is established may
// return Packet so the response path can retain bounded header metadata.
func parsePacket(data []byte) (*Packet, []byte, error) {
	if len(data) < informHeaderBytes {
		return nil, nil, fmt.Errorf("inform packet too short")
	}
	if string(data[:4]) != Magic {
		return nil, nil, fmt.Errorf("invalid magic")
	}
	if binary.BigEndian.Uint32(data[4:8]) != PacketVersion {
		return nil, nil, fmt.Errorf("unsupported packet version")
	}
	if binary.BigEndian.Uint32(data[32:36]) != PayloadVersion {
		return nil, nil, fmt.Errorf("unsupported payload version")
	}
	payloadLen := uint64(binary.BigEndian.Uint32(data[36:40]))
	if payloadLen > uint64(MaxEncodedPayloadBytes) {
		return nil, nil, fmt.Errorf(
			"inform payload length %d exceeds %d bytes",
			payloadLen,
			MaxEncodedPayloadBytes,
		)
	}
	packetPayloadLen := uint64(len(data) - informHeaderBytes)
	if payloadLen != packetPayloadLen {
		return nil, nil, fmt.Errorf(
			"inform payload length mismatch: header says %d bytes, packet has %d",
			payloadLen,
			packetPayloadLen,
		)
	}

	p := &Packet{
		MAC:     append(net.HardwareAddr{}, data[8:14]...),
		Flags:   binary.BigEndian.Uint16(data[14:16]),
		IV:      append([]byte{}, data[16:32]...),
		Payload: append([]byte{}, data[informHeaderBytes:]...),
	}
	if err := validateEnvelopeFlags(p.Flags); err != nil {
		return p, nil, err
	}
	if p.Flags&FlagEncryptedGCM != 0 && len(p.Payload) < GCMTagSize {
		return p, nil, fmt.Errorf("AES-GCM payload is shorter than its %d-byte authentication tag", GCMTagSize)
	}
	return p, data[:informHeaderBytes], nil
}

// validateEnvelopeFlags rejects envelopes that could otherwise turn an
// authenticated GCM body into unauthenticated plaintext by changing only the
// TNBU flags. All supported inform bodies are encrypted and unknown flags are
// rejected. Known representation flags are checked by the directional decoder
// after any required GCM header authentication.
func validateEnvelopeFlags(flags uint16) error {
	const knownFlags = FlagEncrypted | FlagZlib | FlagSnappy | FlagEncryptedGCM
	if flags&FlagEncrypted == 0 {
		return fmt.Errorf("unencrypted inform payload is not accepted")
	}
	if unknown := flags &^ knownFlags; unknown != 0 {
		return fmt.Errorf("unknown inform flags 0x%04x", unknown)
	}
	return nil
}
