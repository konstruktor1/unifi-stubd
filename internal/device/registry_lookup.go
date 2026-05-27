package device

import (
	"sort"
	"strings"
)

// Profiles returns a copy of the built-in device profiles.
func Profiles() []Profile {
	return NewProfileRegistry().Profiles()
}

// Profiles returns a copy of all profiles in r.
func (r ProfileRegistry) Profiles() []Profile {
	records := cloneRecords(r.records)
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].order != records[j].order {
			return records[i].order < records[j].order
		}
		return records[i].profile.Name < records[j].profile.Name
	})
	out := make([]Profile, 0, len(records))
	for _, record := range records {
		out = append(out, profileWithSource(record))
	}
	return out
}

// LookupProfile returns a built-in profile by profile name or model identifier.
func LookupProfile(name string) (Profile, bool) {
	return NewProfileRegistry().LookupProfile(name)
}

// LookupProfile returns a profile by profile name or model identifier.
func (r ProfileRegistry) LookupProfile(name string) (Profile, bool) {
	record, ok := r.lookupRecord(name)
	if !ok {
		return Profile{}, false
	}
	return profileWithSource(record), true
}

// lookupRecord resolves either a profile name or model identifier and returns a
// detached record copy for inheritance or payload use.
func (r ProfileRegistry) lookupRecord(name string) (record, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, record := range r.records {
		profile := record.profile
		if strings.ToLower(profile.Name) == name || strings.ToLower(profile.Model) == name {
			return recordEntry(record.source, record.order, profile, record.document, record.builtin), true
		}
	}
	return record{}, false
}
