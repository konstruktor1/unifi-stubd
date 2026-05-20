// Package profiledata merges YAML inheritance before strict typed decoding.
// This preserves explicit zero-value overrides such as false, 0, empty strings,
// and [] while still rejecting unknown fields later.
package profiledata

import "gopkg.in/yaml.v3"

func mergeProfileDocuments(base *yaml.Node, overlay *yaml.Node) *yaml.Node {
	baseContent := cloneYAMLNode(yamlDocumentContent(base))
	overlayContent := yamlDocumentContent(overlay)
	merged := mergeYAMLNodes(baseContent, overlayContent)
	return yamlDocument(merged)
}

func mergeYAMLNodes(base *yaml.Node, overlay *yaml.Node) *yaml.Node {
	if base == nil {
		return cloneYAMLNode(overlay)
	}
	if overlay == nil {
		return cloneYAMLNode(base)
	}
	if base.Kind != yaml.MappingNode || overlay.Kind != yaml.MappingNode {
		return cloneYAMLNode(overlay)
	}
	merged := cloneYAMLNode(base)
	for index := 0; index+1 < len(overlay.Content); index += 2 {
		key := overlay.Content[index]
		value := overlay.Content[index+1]
		baseIndex := yamlMappingKeyIndex(merged, key.Value)
		if baseIndex < 0 {
			merged.Content = append(merged.Content, cloneYAMLNode(key), cloneYAMLNode(value))
			continue
		}
		merged.Content[baseIndex+1] = mergeYAMLNodes(merged.Content[baseIndex+1], value)
	}
	return merged
}

func yamlMappingKeyIndex(node *yaml.Node, key string) int {
	node = yamlDocumentContent(node)
	if node == nil || node.Kind != yaml.MappingNode {
		return -1
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == key {
			return index
		}
	}
	return -1
}

func yamlDocumentContent(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}
	for node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}
	for node.Kind == yaml.AliasNode && node.Alias != nil {
		node = node.Alias
	}
	if node.Kind == 0 {
		return nil
	}
	return node
}

func yamlDocument(content *yaml.Node) *yaml.Node {
	if content == nil {
		return nil
	}
	return &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{content},
	}
}

func cloneYAMLNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}
	clone := *node
	if len(node.Content) > 0 {
		clone.Content = make([]*yaml.Node, len(node.Content))
		for index, child := range node.Content {
			clone.Content[index] = cloneYAMLNode(child)
		}
	}
	if node.Alias != nil {
		clone.Alias = cloneYAMLNode(node.Alias)
	}
	return &clone
}
