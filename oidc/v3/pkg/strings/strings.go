package strings

import "slices"

func Contains(list []string, needle string) bool {
	return slices.Contains(list, needle)
}
