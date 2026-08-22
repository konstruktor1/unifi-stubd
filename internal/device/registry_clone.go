package device

// cloneProfile detaches profile slices before registry callers can mutate them.
func cloneProfile(profile Profile) Profile {
	profile.PortGroups = cloneNonEmptySlice(profile.PortGroups)
	profile.PortNames = cloneNonEmptySlice(profile.PortNames)
	profile.PortRoles = cloneNonEmptySlice(profile.PortRoles)
	profile.PortNetworkGroups = cloneNonEmptySlice(profile.PortNetworkGroups)
	profile.ValidatedControllerVersions = cloneNonEmptySlice(profile.ValidatedControllerVersions)
	return profile
}

// profileWithSource attaches registry provenance to a detached profile copy for
// status, list, and export views.
func profileWithSource(record record) Profile {
	profile := cloneProfile(record.profile)
	profile.Source = record.source
	if record.builtin {
		profile.SourceType = sourceTypeBuiltIn
	} else {
		profile.SourceType = sourceTypeExternal
	}
	return profile
}

// cloneRecords detaches registry entries and their YAML documents.
func cloneRecords(records []record) []record {
	if len(records) == 0 {
		return nil
	}
	out := make([]record, len(records))
	for index, record := range records {
		out[index] = record
		out[index].profile = cloneProfile(record.profile)
		out[index].document = cloneYAMLNode(record.document)
	}
	return out
}

// cloneNonEmptySlice preserves nil for absent profile slices while copying
// populated values.
func cloneNonEmptySlice[T any](values []T) []T {
	if len(values) == 0 {
		return nil
	}
	out := make([]T, len(values))
	copy(out, values)
	return out
}
