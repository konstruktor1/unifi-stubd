package device

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed profiles/*/profile.yaml
var embeddedProfiles embed.FS

var builtinProfiles = loadBuiltinProfileRecords()

// builtinProfileRecords returns the decoded built-in profile records.
func builtinProfileRecords() []record {
	return builtinProfiles
}

// loadBuiltinProfileRecords decodes checked-in profile YAML in stable order.
func loadBuiltinProfileRecords() []record {
	paths, err := fs.Glob(embeddedProfiles, "profiles/*/profile.yaml")
	if err != nil {
		panic(fmt.Sprintf("glob embedded profiles: %v", err))
	}
	sort.Strings(paths)
	records := make([]record, 0, len(paths))
	for _, path := range paths {
		data, err := embeddedProfiles.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf("read embedded profile %s: %v", path, err))
		}
		decoded, err := decodeConfigRecord(data)
		if err != nil {
			panic(fmt.Sprintf("load profile %s: %v", path, err))
		}
		if err := registerRecord(&records, path, decoded.order, decoded.profile, decoded.document, true); err != nil {
			panic(err)
		}
	}
	return records
}
