package device

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LoadPath loads one profile YAML file or all profile YAML files in a directory.
func (r *ProfileRegistry) LoadPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return &ProfilePathError{Path: path, Kind: ErrorKindIO, Err: err}
	}
	if !info.IsDir() {
		return r.LoadFile(path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return &ProfilePathError{Path: path, Kind: ErrorKindIO, Err: err}
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".yaml" || ext == ".yml" {
			paths = append(paths, filepath.Join(path, name))
		}
	}
	sort.Strings(paths)
	for _, profilePath := range paths {
		if err := r.LoadFile(profilePath); err != nil {
			return err
		}
	}
	return nil
}

// LoadFile loads one external profile YAML file into r.
func (r *ProfileRegistry) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return &ProfilePathError{Path: path, Kind: ErrorKindIO, Err: err}
	}
	decoded, err := r.decodeExternalConfigRecord(data)
	if err != nil {
		return &ProfilePathError{Path: path, Kind: classifyProfileError(err), Err: err}
	}
	if err := r.register(path, decoded.order, decoded.profile, decoded.document, false); err != nil {
		return &ProfilePathError{Path: path, Kind: ErrorKindValidation, Err: err}
	}
	return nil
}
