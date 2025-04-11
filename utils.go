package main

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

// Check if two status codes are equivalent
func isStatusEquivalent(code1, code2 int) bool {
	for _, equivalentCodes := range config.EquivalentStatusCodes.IntSliceSlice() {
		if contains(equivalentCodes, code1) && contains(equivalentCodes, code2) {
			return true
		}
	}
	return code1 == code2
}

// Check if an slice contains a specific value
func contains[T comparable](slice []T, value T) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// Check if two slices of strings are equal
func stringSliceEqual(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}
