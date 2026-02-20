package utils

import (
	"gomistakes/interfaces-usecase/sort"
)

// Because we factored common behaviour, we can have utility methods like this.
func IsSorted(data sort.Interface) bool {
	n := data.Len()
	for i := n - 1; i > 0; i-- {
		if data.Less(i, i-1) {
			return false
		}
	}
	return true
}
