package sort

type Interface interface {
	Len() int           // Number of elements
	Less(i, j int) bool // Checks two elements
	Swap(i, j int)      // Swaps two elements
}
