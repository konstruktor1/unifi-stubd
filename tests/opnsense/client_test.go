package opnsense_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/opnsense"
)

func TestClientUsesBasicAuthAndGETOnly(t *testing.T) {
	t.Parallel()

	var method string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		if r.URL.Path != "/api/interfaces/overview/interfaces_info" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		key, secret, ok := r.BasicAuth()
		if !ok || key != "api-key" || secret != "api-secret" {
			t.Fatalf("basic auth = %q/%q ok=%t", key, secret, ok)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"interfaces": map[string]any{
				testInterfaceIXL0: map[string]any{"interface": testInterfaceIXL0, "status": "up"},
			},
		})
	}))
	defer server.Close()

	client, err := opnsense.NewClient(opnsense.SourceConfig{
		BaseURL:   server.URL,
		TimeoutMS: 1000,
		Interfaces: []opnsense.InterfaceMapping{
			{Port: 1, Interface: testInterfaceIXL0},
		},
	}, opnsense.Credentials{Key: "api-key", Secret: "api-secret"})
	if err != nil {
		t.Fatal(err)
	}
	interfaces, err := client.InterfacesInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if method != http.MethodGet {
		t.Fatalf("method = %q, want GET", method)
	}
	if _, ok := interfaces[testInterfaceIXL0]; !ok {
		t.Fatalf("%s not decoded: %+v", testInterfaceIXL0, interfaces)
	}
}

func TestClientErrorDoesNotLeakCredentials(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := opnsense.NewClient(opnsense.SourceConfig{
		BaseURL:   server.URL,
		TimeoutMS: 1000,
		Interfaces: []opnsense.InterfaceMapping{
			{Port: 1, Interface: testInterfaceIXL0},
		},
	}, opnsense.Credentials{Key: "secret-key", Secret: "secret-value"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.InterfacesInfo(context.Background())
	if err == nil {
		t.Fatal("InterfacesInfo error = nil")
	}
	text := err.Error()
	if strings.Contains(text, "secret-key") || strings.Contains(text, "secret-value") {
		t.Fatalf("error leaked credentials: %s", text)
	}
}
