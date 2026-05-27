package inform

import (
	"encoding/binary"
	"fmt"
	"net"
)

// Decode unwraps a UniFi inform packet and returns decoded JSON payload bytes.
func Decode(data []byte, key []byte) (*Packet, []byte, error) {
	if len(data) < 40 {
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
	payloadLen := int(binary.BigEndian.Uint32(data[36:40]))
	if len(data) < 40+payloadLen {
		return nil, nil, fmt.Errorf("truncated payload")
	}

	p := &Packet{
		MAC:     append(net.HardwareAddr{}, data[8:14]...),
		Flags:   binary.BigEndian.Uint16(data[14:16]),
		IV:      append([]byte{}, data[16:32]...),
		Payload: append([]byte{}, data[40:40+payloadLen]...),
	}

	body := append([]byte{}, p.Payload...)
	body, err := decryptPayload(p, body, key, data[:40])
	if err != nil {
		return nil, nil, err
	}

	if p.Flags&FlagZlib != 0 {
		body, err = decompressPayload(body)
		if err != nil {
			return nil, nil, err
		}
	}

	return p, body, nil
}
