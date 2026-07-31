package inform_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

const (
	headerBytes = 40
	plainJSON   = `{"_type":"noop"}`
	cbcWireHex  = "544e42550000000002005e0053010001" +
		"000102030405060708090a0b0c0d0e0f" +
		"0000000100000020" +
		"1dceec939d8b40363e346590b0c94fa3" +
		"9086870ca8f2496b449b7740a54044c1"
	gcmWireHex = "544e42550000000002005e0053010009" +
		"000102030405060708090a0b0c0d0e0f" +
		"0000000100000020" +
		"d2351fb27f11bf501a5ceec4a4880c3d" +
		"7810e55cde1eb71543dd9c67b7719d6f"
)

func wire(t *testing.T, value string) []byte {
	t.Helper()
	decoded, err := hex.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return decoded
}

// TestDefaultAdoptionKey fixes the public pending-adoption key to its recovered
// bytes and verifies callers cannot mutate the next returned value.
func TestDefaultAdoptionKey(t *testing.T) {
	want := wire(t, "ba86f2bbe107c7c57eb5f2690775c712")
	got := inform.DefaultAuthKey()
	if !bytes.Equal(got, want) || !inform.IsDefaultAuthKey(got) {
		t.Fatalf("default authkey = %x; want %x", got, want)
	}
	got[0] ^= 0xff
	if next := inform.DefaultAuthKey(); !bytes.Equal(next, want) {
		t.Fatalf("caller mutation changed default authkey to %x", next)
	}
}

// TestRecoveredInformWireImages decodes independently calculated CBC and GCM
// packets, avoiding a self-confirming product encode/decode round trip.
func TestRecoveredInformWireImages(t *testing.T) {
	for _, test := range []struct {
		name  string
		hex   string
		flags uint16
	}{
		{name: "AES-CBC", hex: cbcWireHex, flags: inform.FlagEncrypted},
		{name: "AES-GCM", hex: gcmWireHex, flags: inform.FlagEncrypted | inform.FlagEncryptedGCM},
	} {
		t.Run(test.name, func(t *testing.T) {
			packet, payload, err := inform.Decode(wire(t, test.hex), inform.DefaultAuthKey())
			if err != nil {
				t.Fatal(err)
			}
			if packet.MAC.String() != "02:00:5e:00:53:01" || packet.Flags != test.flags {
				t.Fatalf("decoded envelope = MAC %s flags 0x%04x", packet.MAC, packet.Flags)
			}
			if string(payload) != plainJSON {
				t.Fatalf("decoded payload = %q; want %q", payload, plainJSON)
			}
		})
	}
}

// TestGCMAuthenticatesTheEnvelope proves identity, mode, nonce, ciphertext,
// and tag are one authenticated product boundary.
func TestGCMAuthenticatesTheEnvelope(t *testing.T) {
	original := wire(t, gcmWireHex)
	for _, test := range []struct {
		name   string
		offset int
	}{
		{name: "device identity", offset: 8},
		{name: "wire mode", offset: 15},
		{name: "nonce", offset: 16},
		{name: "ciphertext", offset: headerBytes},
		{name: "authentication tag", offset: len(original) - 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			mutated := bytes.Clone(original)
			mutated[test.offset] ^= 0x02
			_, plaintext, err := inform.Decode(mutated, inform.DefaultAuthKey())
			if err == nil || plaintext != nil || !strings.Contains(err.Error(), "decrypt AES-GCM payload") {
				t.Fatalf("tampered %s produced plaintext %q and error %v", test.name, plaintext, err)
			}
		})
	}
}

// TestGCMRejectsAnotherDeviceKey proves a valid wire image is bound to the
// individual device key, not merely to a structurally valid TNBU envelope.
func TestGCMRejectsAnotherDeviceKey(t *testing.T) {
	wrongKey := wire(t, "00112233445566778899aabbccddeeff")
	_, plaintext, err := inform.Decode(wire(t, gcmWireHex), wrongKey)
	if err == nil || plaintext != nil || !strings.Contains(err.Error(), "decrypt AES-GCM payload") {
		t.Fatalf("wrong key produced plaintext %q and error %v", plaintext, err)
	}
}

// TestInformHeaderAndSizeBoundaries binds TNBU framing and both allocation
// limits without exercising HTTP clients or scheduler policy.
func TestInformHeaderAndSizeBoundaries(t *testing.T) {
	packet := wire(t, cbcWireHex)
	if string(packet[:4]) != inform.Magic || binary.BigEndian.Uint32(packet[4:8]) != inform.PacketVersion {
		t.Fatalf("invalid TNBU prefix: %x", packet[:8])
	}
	if binary.BigEndian.Uint32(packet[32:36]) != inform.PayloadVersion {
		t.Fatalf("payload version = %d", binary.BigEndian.Uint32(packet[32:36]))
	}
	if !bytes.Equal(packet[8:14], []byte{2, 0, 94, 0, 83, 1}) ||
		!bytes.Equal(packet[16:32], wire(t, "000102030405060708090a0b0c0d0e0f")) {
		t.Fatalf("identity or IV changed in TNBU header: %x", packet[8:32])
	}
	if got := binary.BigEndian.Uint32(packet[36:40]); got != uint32(len(packet)-headerBytes) {
		t.Fatalf("declared payload length = %d; packet carries %d", got, len(packet)-headerBytes)
	}

	short := packet[:headerBytes-1]
	mismatch := bytes.Clone(packet)
	binary.BigEndian.PutUint32(mismatch[36:40], uint32(len(packet)-headerBytes+1))
	oversized := bytes.Clone(packet[:headerBytes])
	binary.BigEndian.PutUint32(oversized[36:40], uint32(inform.MaxEncodedPayloadBytes+1))
	for name, invalid := range map[string][]byte{
		"short header":    short,
		"length mismatch": mismatch,
		"encoded limit":   oversized,
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := inform.Decode(invalid, inform.DefaultAuthKey()); err == nil {
				t.Fatal("invalid envelope was accepted")
			}
		})
	}
	if _, err := inform.EncodeJSON([]byte{2, 0, 94, 0, 83, 1}, inform.DefaultAuthKey(),
		make([]byte, inform.MaxDecodedPayloadBytes+1), inform.Options{}); err == nil {
		t.Fatal("oversized plaintext was accepted")
	}
}
