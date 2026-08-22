// Inform sending selects one adoption key and packet cipher for each heartbeat.
// The chosen mode is recorded so logs and status can explain the actual
// controller exchange without a response-driven duplicate request.
package main

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

// informCipherStatus records the selected request cipher and sanitized response
// metadata for last_inform.
type informCipherStatus struct {
	RequestFlags          uint16
	KeySource             string
	ResponseBodyPresent   bool
	ResponsePacketDecoded bool
	ResponseFlags         uint16
}

// sendInform chooses exactly one adoption key and cipher for one heartbeat.
// HTTP status, transport, or response failures flow through the scheduler;
// none of them may duplicate the heartbeat under another cipher.
func sendInform(mac net.HardwareAddr, url string, store adoption.Store, payload []byte, sourceIP net.IP) (*inform.Response, informCipherStatus, error) {
	key, err := informAuthKey(store)
	if err != nil {
		return nil, informCipherStatus{}, err
	}
	keySource := "default"
	if store.AuthKey != "" {
		keySource = "adopted"
	}

	options := inform.Options{Zlib: true, GCM: store.AuthKey != "" && store.UseAESGCM}
	status := informCipherStatus{
		RequestFlags: requestFlags(options),
		KeySource:    keySource,
	}
	resp, sendErr := inform.Client{
		URL:       url,
		MAC:       mac,
		Key:       key,
		Options:   options,
		LocalAddr: sourceIP,
	}.Send(payload)
	if resp != nil {
		status.ResponseBodyPresent = len(resp.RawBody) > 0
		if resp.Packet != nil {
			status.ResponseFlags = resp.Packet.Flags
			// A policy-rejected response may expose bounded TNBU header
			// metadata without having produced authenticated plaintext.
			if sendErr == nil {
				status.ResponsePacketDecoded = true
			}
		}
	}
	if sendErr != nil {
		return resp, status, fmt.Errorf("send inform: %w", sendErr)
	}
	return resp, status, nil
}

// informAuthKey makes use of the public default key an explicit lifecycle
// decision. Missing or invalid managed key material never reaches HTTP.
func informAuthKey(store adoption.Store) ([]byte, error) {
	if store.AuthKey != "" {
		key, err := authKeyBytes(store.AuthKey)
		if err != nil {
			return nil, fmt.Errorf("invalid persisted adoption authkey: %w", err)
		}
		return key, nil
	}
	if store.UseAESGCM {
		return nil, fmt.Errorf("default authkey is not permitted with adopted cipher state")
	}
	switch store.State {
	case "", adoption.StateFactory, adoption.StateDiscovered, adoption.StateProvisioning:
		return inform.DefaultAuthKey(), nil
	case adoption.StateAdopting, adoption.StateConnected, adoption.StateFailed:
		return nil, fmt.Errorf("default authkey is not permitted for adoption state %q", store.State)
	default:
		return nil, fmt.Errorf("default authkey is not permitted for unknown adoption state")
	}
}

// requestFlags returns the exact TNBU flags emitted for one client attempt.
// Keeping this derivation next to cipher selection lets sanitized runtime
// status prove the actual request mode without retaining packet bodies.
func requestFlags(options inform.Options) uint16 {
	flags := inform.FlagEncrypted
	if options.Zlib {
		flags |= inform.FlagZlib
	}
	if options.GCM {
		flags |= inform.FlagEncryptedGCM
	}
	return flags
}

// authKeyBytes accepts both historical raw 16-byte keys and hex-encoded keys
// stored in adoption.env.
func authKeyBytes(authKey string) ([]byte, error) {
	if authKey == "" {
		return nil, nil
	}
	if len(authKey) == 16 {
		key := []byte(authKey)
		if inform.IsDefaultAuthKey(key) {
			return nil, fmt.Errorf("individual authkey equals public default key")
		}
		return key, nil
	}
	key, err := hex.DecodeString(authKey)
	if err != nil {
		return nil, fmt.Errorf("decode adoption authkey: %w", err)
	}
	if len(key) != 16 {
		return nil, fmt.Errorf("decoded authkey has %d bytes, want 16", len(key))
	}
	if inform.IsDefaultAuthKey(key) {
		return nil, fmt.Errorf("individual authkey equals public default key")
	}
	return key, nil
}
