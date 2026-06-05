package config

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// MigrationResult describes a conservative config schema normalization.
type MigrationResult struct {
	Data     []byte
	Changed  bool
	Actions  []string
	Warnings []string
}

// MigrateData normalizes known legacy YAML aliases without changing runtime
// behavior.
func MigrateData(data []byte) (MigrationResult, error) {
	doc, root, err := parseMigrationDoc(data)
	if err != nil {
		return MigrationResult{}, err
	}
	result := MigrationResult{Data: data}
	if err := migrateScalarAlias(root, "controller", "controller_url", &result); err != nil {
		return MigrationResult{}, err
	}
	if err := migrateScalarAlias(root, "inform_url", "controller_url", &result); err != nil {
		return MigrationResult{}, err
	}
	if migrateOperationMode(root, &result) {
		result.Changed = true
	}
	if err := migrateBridgeAlias(root, "observe_bridge", "bridge", &result); err != nil {
		return MigrationResult{}, err
	}
	if err := migrateBridgeAlias(root, "observe_interface", "uplink_interface", &result); err != nil {
		return MigrationResult{}, err
	}
	if err := migrateNodeAlias(root, "port_map", "port_mappings", &result); err != nil {
		return MigrationResult{}, err
	}
	if !result.Changed {
		return result, nil
	}
	out, err := encodeMigrationDoc(doc)
	if err != nil {
		return MigrationResult{}, err
	}
	if _, err := Decode(out); err != nil {
		return MigrationResult{}, fmt.Errorf("validate migrated config: %w", err)
	}
	result.Data = out
	return result, nil
}

func parseMigrationDoc(data []byte) (*yaml.Node, *yaml.Node, error) {
	var doc yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&doc); err != nil {
		return nil, nil, fmt.Errorf("parse config for migration: %w", err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil, nil, fmt.Errorf("config migration requires one YAML document")
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, nil, fmt.Errorf("config migration requires a YAML mapping document")
	}
	return &doc, root, nil
}

func encodeMigrationDoc(doc *yaml.Node) ([]byte, error) {
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(doc); err != nil {
		_ = encoder.Close()
		return nil, fmt.Errorf("encode migrated config: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("encode migrated config: %w", err)
	}
	return out.Bytes(), nil
}

func migrateScalarAlias(root *yaml.Node, alias, target string, result *MigrationResult) error {
	aliasKey, aliasValue, aliasIndex := mappingPair(root, alias)
	if aliasKey == nil {
		return nil
	}
	if aliasValue.Kind != yaml.ScalarNode {
		return fmt.Errorf("%s must be a scalar value to migrate to %s", alias, target)
	}
	aliasText := strings.TrimSpace(aliasValue.Value)
	targetKey, targetValue, targetIndex := mappingPair(root, target)
	if targetKey == nil {
		aliasKey.Value = target
		result.Changed = true
		result.Actions = append(result.Actions, fmt.Sprintf("%s -> %s", alias, target))
		return nil
	}
	if targetValue.Kind != yaml.ScalarNode {
		return fmt.Errorf("%s already exists and is not a scalar value", target)
	}
	targetText := strings.TrimSpace(targetValue.Value)
	switch {
	case targetText == "" && aliasText != "":
		root.Content[targetIndex+1] = cloneNode(aliasValue)
		removeMappingPair(root, aliasIndex)
		result.Changed = true
		result.Actions = append(result.Actions, fmt.Sprintf("%s -> %s", alias, target))
	case targetText == aliasText:
		removeMappingPair(root, aliasIndex)
		result.Changed = true
		result.Actions = append(result.Actions, fmt.Sprintf("remove duplicate %s", alias))
	default:
		return fmt.Errorf("%s and %s contain different values; migrate manually", alias, target)
	}
	return nil
}

func migrateOperationMode(root *yaml.Node, result *MigrationResult) bool {
	_, value, _ := mappingPair(root, "operation_mode")
	if value == nil || value.Kind != yaml.ScalarNode || strings.TrimSpace(value.Value) != "observe" {
		return false
	}
	value.Value = "bridge-observe"
	result.Actions = append(result.Actions, "operation_mode observe -> bridge-observe")
	return true
}

func migrateBridgeAlias(root *yaml.Node, alias, target string, result *MigrationResult) error {
	_, aliasValue, aliasIndex := mappingPair(root, alias)
	if aliasValue == nil {
		return nil
	}
	if aliasValue.Kind != yaml.ScalarNode {
		return fmt.Errorf("%s must be a scalar value to migrate to bridge_observe.%s", alias, target)
	}
	aliasText := strings.TrimSpace(aliasValue.Value)
	if aliasText == "" {
		return nil
	}
	bridge, err := bridgeObserveNode(root)
	if err != nil {
		return err
	}
	_, targetValue, targetIndex := mappingPair(bridge, target)
	if targetValue == nil {
		appendMappingPair(bridge, target, cloneNode(aliasValue))
		removeMappingPair(root, aliasIndex)
		result.Changed = true
		result.Actions = append(result.Actions, fmt.Sprintf("%s -> bridge_observe.%s", alias, target))
		return nil
	}
	if targetValue.Kind != yaml.ScalarNode {
		return fmt.Errorf("bridge_observe.%s already exists and is not a scalar value", target)
	}
	targetText := strings.TrimSpace(targetValue.Value)
	switch targetText {
	case "":
		bridge.Content[targetIndex+1] = cloneNode(aliasValue)
		removeMappingPair(root, aliasIndex)
		result.Changed = true
		result.Actions = append(result.Actions, fmt.Sprintf("%s -> bridge_observe.%s", alias, target))
	case aliasText:
		removeMappingPair(root, aliasIndex)
		result.Changed = true
		result.Actions = append(result.Actions, fmt.Sprintf("remove duplicate %s", alias))
	default:
		return fmt.Errorf("%s and bridge_observe.%s contain different values; migrate manually", alias, target)
	}
	return nil
}

func migrateNodeAlias(root *yaml.Node, alias, target string, result *MigrationResult) error {
	aliasKey, aliasValue, aliasIndex := mappingPair(root, alias)
	if aliasKey == nil {
		return nil
	}
	_, targetValue, _ := mappingPair(root, target)
	if targetValue == nil {
		aliasKey.Value = target
		result.Changed = true
		result.Actions = append(result.Actions, fmt.Sprintf("%s -> %s", alias, target))
		return nil
	}
	if sameNode(aliasValue, targetValue) {
		removeMappingPair(root, aliasIndex)
		result.Changed = true
		result.Actions = append(result.Actions, fmt.Sprintf("remove duplicate %s", alias))
		return nil
	}
	return fmt.Errorf("%s and %s contain different values; migrate manually", alias, target)
}

func bridgeObserveNode(root *yaml.Node) (*yaml.Node, error) {
	_, value, _ := mappingPair(root, "bridge_observe")
	if value != nil {
		if value.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("bridge_observe already exists and is not a mapping")
		}
		return value, nil
	}
	node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	appendMappingPair(root, "bridge_observe", node)
	return node, nil
}

func mappingPair(root *yaml.Node, key string) (*yaml.Node, *yaml.Node, int) {
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == key {
			return root.Content[i], root.Content[i+1], i
		}
	}
	return nil, nil, -1
}

func appendMappingPair(root *yaml.Node, key string, value *yaml.Node) {
	root.Content = append(root.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		value,
	)
}

func removeMappingPair(root *yaml.Node, index int) {
	root.Content = append(root.Content[:index], root.Content[index+2:]...)
}

func cloneNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}
	clone := *node
	clone.Content = make([]*yaml.Node, len(node.Content))
	for i, child := range node.Content {
		clone.Content[i] = cloneNode(child)
	}
	return &clone
}

func sameNode(left, right *yaml.Node) bool {
	leftData, leftErr := yaml.Marshal(left)
	rightData, rightErr := yaml.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftData, rightData)
}
