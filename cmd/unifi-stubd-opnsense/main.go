// unifi-stubd-opnsense generates reviewable unifi-stubd YAML from read-only
// OPNsense API data. It is intentionally separate from the unifi-stubd daemon.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/opnsense"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "unifi-stubd-opnsense: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var configPath string
	var sourcePath string
	var outputPath string
	var validateOnly bool
	flag.StringVar(&configPath, "config", appconfig.DefaultPath, "base unifi-stubd YAML config")
	flag.StringVar(&sourcePath, "source", "", "OPNsense generator source YAML")
	flag.StringVar(&outputPath, "out", "", "write generated unifi-stubd YAML to this path; default prints to stdout")
	flag.BoolVar(&validateOnly, "validate", false, "validate source and credentials without writing generated YAML")
	flag.Parse()

	if sourcePath == "" {
		return fmt.Errorf("-source is required")
	}
	base, err := appconfig.Load(configPath)
	if err != nil {
		return fmt.Errorf("load base config %s: %w", configPath, err)
	}
	source, err := opnsense.LoadSourceConfig(sourcePath)
	if err != nil {
		return fmt.Errorf("load OPNsense source config %s: %w", sourcePath, err)
	}
	credentials, err := opnsense.LoadCredentials(source)
	if err != nil {
		return fmt.Errorf("load OPNsense credentials: %w", err)
	}
	if validateOnly {
		fmt.Printf("opnsense source valid: mappings=%d base_url=%s\n", len(source.Interfaces), source.BaseURL)
		return nil
	}
	client, err := opnsense.NewClient(source, credentials)
	if err != nil {
		return fmt.Errorf("create OPNsense API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), source.Timeout()+time.Second)
	defer cancel()
	interfaces, err := readInterfaces(ctx, client, source)
	if err != nil {
		return fmt.Errorf("read OPNsense interfaces: %w", err)
	}
	gateways := map[string]opnsense.GatewayStatus{}
	if source.GatewayStatus {
		gateways, err = client.GatewayStatus(ctx)
		if err != nil {
			return fmt.Errorf("read OPNsense gateway status: %w", err)
		}
	}
	generated := opnsense.GenerateConfig(base, source, interfaces, gateways)
	data, err := opnsense.MarshalConfig(generated)
	if err != nil {
		return fmt.Errorf("marshal generated config: %w", err)
	}
	if outputPath == "" {
		_, err = os.Stdout.Write(data)
		if err != nil {
			return fmt.Errorf("write generated config to stdout: %w", err)
		}
		return nil
	}
	if err := os.WriteFile(outputPath, data, 0o600); err != nil {
		return fmt.Errorf("write generated config %s: %w", outputPath, err)
	}
	return nil
}

func readInterfaces(ctx context.Context, client *opnsense.Client, source opnsense.SourceConfig) (map[string]opnsense.InterfaceStatus, error) {
	interfaces, err := client.InterfacesInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("read overview: %w", err)
	}
	for _, mapping := range source.Interfaces {
		if _, ok := interfaces[strings.ToLower(mapping.Interface)]; ok {
			continue
		}
		status, err := client.Interface(ctx, mapping.Interface)
		if err != nil {
			return nil, fmt.Errorf("read interface %s: %w", mapping.Interface, err)
		}
		interfaces[mapping.Interface] = status
	}
	return interfaces, nil
}
