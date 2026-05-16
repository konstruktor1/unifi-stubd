package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

func sendInform(mac net.HardwareAddr, url string, store adoption.Store, payload []byte) (*inform.Response, bool, error) {
	key, err := authKeyBytes(store.AuthKey)
	if err != nil {
		log.Printf("invalid adoption authkey, falling back to default key: %v", err)
		key = nil
	}

	options := []inform.Options{{Zlib: true}}
	if store.AuthKey != "" {
		if store.UseAESGCM {
			options = []inform.Options{{Zlib: true, GCM: true}}
		} else {
			options = []inform.Options{
				{Zlib: true, GCM: true},
				{Zlib: true},
			}
		}
	}

	var lastErr error
	var lastResp *inform.Response
	var lastUsedGCM bool
	for _, opts := range options {
		resp, err := inform.Client{
			URL:     url,
			MAC:     mac,
			Key:     key,
			Options: opts,
		}.Send(payload)
		if err == nil {
			lastResp = resp
			lastUsedGCM = opts.GCM
			if resp.StatusCode == http.StatusOK {
				return resp, opts.GCM, nil
			}
			continue
		}
		lastErr = err
	}
	if lastResp != nil {
		return lastResp, lastUsedGCM, nil
	}
	return nil, false, lastErr
}

func authKeyBytes(authKey string) ([]byte, error) {
	if authKey == "" {
		return nil, nil
	}
	if len(authKey) == 16 {
		return []byte(authKey), nil
	}
	key, err := hex.DecodeString(authKey)
	if err != nil {
		return nil, err
	}
	if len(key) != 16 {
		return nil, fmt.Errorf("decoded authkey has %d bytes, want 16", len(key))
	}
	return key, nil
}
