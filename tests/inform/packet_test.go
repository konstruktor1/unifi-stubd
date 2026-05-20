// Inform packet tests cover TNBU framing, client response limits, and fuzz
// entrypoints for controller-facing protocol robustness.
package inform_test

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

func TestEncodeDecodeJSONCBCZlib(t *testing.T) {
	mac, err := net.ParseMAC("02:11:22:33:44:55")
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"mac":"02:11:22:33:44:55","model":"US8"}`)
	encoded, err := inform.EncodeJSON(mac, inform.DefaultAuthKey(), payload, inform.Options{Zlib: true})
	if err != nil {
		t.Fatal(err)
	}

	packet, decoded, err := inform.Decode(encoded, inform.DefaultAuthKey())
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

func TestClientRejectsOversizedResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("0123456789"))
	}))
	defer server.Close()

	mac, err := net.ParseMAC("02:11:22:33:44:55")
	if err != nil {
		t.Fatal(err)
	}
	_, err = inform.Client{
		URL:              server.URL + "/inform",
		MAC:              mac,
		MaxResponseBytes: 8,
	}.Send([]byte(`{"mac":"02:11:22:33:44:55"}`))
	if err == nil {
		t.Fatal("Send succeeded; want oversized response error")
	}
	if !strings.Contains(err.Error(), "inform response body exceeds 8 bytes") {
		t.Fatalf("error = %v", err)
	}
}

func FuzzDecode(f *testing.F) {
	mac, err := net.ParseMAC("02:11:22:33:44:55")
	if err != nil {
		f.Fatal(err)
	}
	encoded, err := inform.EncodeJSON(mac, inform.DefaultAuthKey(), []byte(`{"model":"US8"}`), inform.Options{Zlib: true})
	if err != nil {
		f.Fatal(err)
	}
	f.Add(encoded)
	f.Add([]byte("TNBU"))
	f.Add([]byte{})

	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _, _ = inform.Decode(data, inform.DefaultAuthKey())
	})
}
