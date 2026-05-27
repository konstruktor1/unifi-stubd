package device

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Register adds one profile to r.
func (r *ProfileRegistry) register(source string, order int, profile Profile, document *yaml.Node, builtin bool) error {
	return registerRecord(&r.records, source, order, profile, document, builtin)
}

// registerRecord enforces profile-name and model uniqueness while allowing an
// explicitly marked external profile to replace a built-in record.
func registerRecord(records *[]record, source string, order int, profile Profile, document *yaml.Node, builtin bool) error {
	for index, record := range *records {
		nameDuplicate := record.profile.Name == profile.Name
		modelDuplicate := strings.EqualFold(record.profile.Model, profile.Model)
		if !nameDuplicate && !modelDuplicate {
			continue
		}
		if profile.AllowBuiltinOverride && record.builtin {
			(*records)[index] = recordEntry(source, order, profile, document, builtin)
			return nil
		}
		switch {
		case nameDuplicate:
			return fmt.Errorf("duplicate profile name %q in %s and %s", profile.Name, record.source, source)
		default:
			return fmt.Errorf("duplicate profile model %q in %s and %s", profile.Model, record.source, source)
		}
	}
	*records = append(*records, recordEntry(source, order, profile, document, builtin))
	return nil
}

// recordEntry stores detached profile and YAML-node copies so later caller
// changes cannot mutate registered profile data.
func recordEntry(source string, order int, profile Profile, document *yaml.Node, builtin bool) record {
	return record{
		source:   source,
		builtin:  builtin,
		order:    order,
		profile:  cloneProfile(profile),
		document: cloneYAMLNode(document),
	}
}
