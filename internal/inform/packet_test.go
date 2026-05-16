package inform

import (
	"bytes"
	"net"
	"testing"
)

func TestEncodeDecodeJSONCBCZlib(t *testing.T) {
	mac, err := net.ParseMAC("02:11:22:33:44:55")
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"mac":"02:11:22:33:44:55","model":"US8"}`)
	encoded, err := EncodeJSON(mac, DefaultAuthKey(), payload, Options{Zlib: true})
	if err != nil {
		t.Fatal(err)
	}

	packet, decoded, err := Decode(encoded, DefaultAuthKey())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(packet.MAC, mac) {
		t.Fatalf("MAC mismatch: got %s want %s", packet.MAC, mac)
	}
	if !bytes.Equal(decoded, payload) {
		t.Fatalf("payload mismatch: got %s want %s", decoded, payload)
	}
}
