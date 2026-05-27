package device

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// DecodeConfig decodes and validates one standalone profile YAML document.
func DecodeConfig(data []byte) (Profile, int, error) {
	decoded, err := decodeConfigRecord(data)
	if err != nil {
		return Profile{}, 0, err
	}
	return decoded.profile, decoded.order, nil
}

// DecodeExternalConfig decodes a profile YAML document with optional inheritance.
func (r *ProfileRegistry) DecodeExternalConfig(data []byte) (Profile, int, error) {
	decoded, err := r.decodeExternalConfigRecord(data)
	if err != nil {
		return Profile{}, 0, err
	}
	return decoded.profile, decoded.order, nil
}

// decodedProfile carries both the typed profile and original YAML document
// needed for inheritance and registry storage.
type decodedProfile struct {
	profile  Profile
	order    int
	document *yaml.Node
}

// decodeConfigRecord handles built-in profiles, which must be standalone YAML
// documents with defaults and validation applied during init registration.
func decodeConfigRecord(data []byte) (decodedProfile, error) {
	document, err := decodeProfileDocument(data)
	if err != nil {
		return decodedProfile{}, err
	}
	profile, err := decodeProfileYAML(document)
	if err != nil {
		return decodedProfile{}, err
	}
	applyProfileDefaults(&profile)
	if err := validateProfile(profile); err != nil {
		return decodedProfile{}, err
	}
	return decodedProfile{profile: profile, order: profile.Order, document: document}, nil
}

// decodeExternalConfigRecord applies optional YAML-level inheritance before the
// final strict typed decode, preserving explicit zero-value overrides.
func (r *ProfileRegistry) decodeExternalConfigRecord(data []byte) (decodedProfile, error) {
	document, err := decodeProfileDocument(data)
	if err != nil {
		return decodedProfile{}, err
	}
	profile, err := decodeProfileYAML(document)
	if err != nil {
		return decodedProfile{}, err
	}
	order := profile.Order
	mergedDocument := document
	if strings.TrimSpace(profile.Extends) != "" {
		base, ok := r.lookupRecord(profile.Extends)
		if !ok {
			return decodedProfile{}, fmt.Errorf("extends %q not found", profile.Extends)
		}
		mergedDocument = mergeProfileDocuments(base.document, document)
		profile, err = decodeProfileYAML(mergedDocument)
		if err != nil {
			return decodedProfile{}, err
		}
	}
	applyProfileDefaults(&profile)
	if err := validateProfile(profile); err != nil {
		return decodedProfile{}, err
	}
	return decodedProfile{profile: profile, order: order, document: mergedDocument}, nil
}

// decodeProfileDocument keeps a YAML node tree so external profiles can merge
// inheritance before strict typed decoding.
func decodeProfileDocument(data []byte) (*yaml.Node, error) {
	var doc yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode profile YAML: %w", err)
	}
	if yamlDocumentContent(&doc) == nil {
		return nil, fmt.Errorf("decode profile YAML: empty document")
	}
	return cloneYAMLNode(&doc), nil
}

// decodeProfileYAML performs the strict KnownFields decode that rejects
// misspelled profile keys.
func decodeProfileYAML(document *yaml.Node) (Profile, error) {
	var profile Profile
	data, err := yaml.Marshal(yamlDocumentContent(document))
	if err != nil {
		return Profile{}, fmt.Errorf("decode profile YAML: marshal merged YAML: %w", err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&profile); err != nil {
		return Profile{}, fmt.Errorf("decode profile YAML: %w", err)
	}
	normalizeProfile(&profile)
	return profile, nil
}
