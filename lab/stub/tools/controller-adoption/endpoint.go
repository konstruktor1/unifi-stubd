package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const guestLoopbackHost = "127.0.0.1"

// setLoopbackInform confines the disposable controller's durable inform URL to
// the guest's authenticated reverse SSH tunnel. Readback prevents a silently
// ignored controller setting from being advertised as a working endpoint.
func setLoopbackInform(ctx context.Context, client *controllerClient) error {
	if err := client.login(ctx); err != nil {
		return err
	}
	if _, err := client.doJSON(ctx, http.MethodPost, "/api/s/"+controllerSite+"/set/setting/super_mgmt", map[string]any{
		"override_inform_host":          true,
		"override_inform_host_location": guestLoopbackHost,
	}); err != nil {
		return fmt.Errorf("set guest loopback inform endpoint: %w", err)
	}
	envelope, err := client.doJSON(ctx, http.MethodGet, "/api/s/"+controllerSite+"/get/setting", nil)
	if err != nil {
		return fmt.Errorf("read guest loopback inform setting: %w", err)
	}
	var settings []struct {
		Key                string `json:"key"`
		OverrideInformHost bool   `json:"override_inform_host"`
		InformHostLocation string `json:"override_inform_host_location"`
	}
	if err := json.Unmarshal(envelope.Data, &settings); err != nil {
		return fmt.Errorf("decode inform setting: %w", err)
	}
	var observed []string
	for _, setting := range settings {
		if setting.Key == "super_mgmt" && setting.OverrideInformHost && setting.InformHostLocation == guestLoopbackHost {
			fmt.Println("controller inform endpoint: guest-loopback port=8080")
			return nil
		}
		if setting.Key == "mgmt" || setting.Key == "super_mgmt" {
			observed = append(observed, fmt.Sprintf("%s:override=%t,host_present=%t,loopback=%t", setting.Key,
				setting.OverrideInformHost, setting.InformHostLocation != "", setting.InformHostLocation == guestLoopbackHost))
		}
	}
	return fmt.Errorf("controller did not retain the guest loopback inform override (%s)", strings.Join(observed, ";"))
}
