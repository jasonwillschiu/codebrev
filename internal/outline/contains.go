package outline

import "slices"

func containsString(haystack []string, needle string) bool {
	return slices.Contains(haystack, needle)
}
