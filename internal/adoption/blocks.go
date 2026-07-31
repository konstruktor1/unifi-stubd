package adoption

// BlockSelection is the side-effect-free reporting policy learned from noop.
// Include and Exclude remain independent producer predicates; UDAPI descriptor
// IDs are status metadata and do not select ordinary inform fields.
type BlockSelection struct {
	Include []string
	Exclude []string
}

// BlockSelection returns a detached policy snapshot from a response.
func (response ControllerResponse) BlockSelection() BlockSelection {
	return BlockSelection{
		Include: append([]string(nil), response.IncludeBlocks...),
		Exclude: append([]string(nil), response.ExcludeBlocks...),
	}
}

// Includes reports whether one exact producer was enabled by a string entry.
func (selection BlockSelection) Includes(name string) bool {
	return name != "" && containsBlock(selection.Include, name)
}

// Excludes reports whether one exact producer was disabled by a string entry.
func (selection BlockSelection) Excludes(name string) bool {
	return name != "" && containsBlock(selection.Exclude, name)
}

func containsBlock(values []string, name string) bool {
	for _, value := range values {
		if value == name {
			return true
		}
	}
	return false
}
