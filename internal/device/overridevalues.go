package device

func clampFloat64(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func cloneBoolRef(value *bool) *bool {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}

func cloneFloat64Ref(value *float64) *float64 {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}

func nonNegativeInt(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
