// Package utils. sort_utils implements sorting algorithms.
package utils

// SortOrder defines the direction of sorting.
const (
	SortAscending = iota
	SortDescending
)

// LessFunc is a comparator: returns true if a < b.
type LessFunc[T any] func(a, b T) bool

// BubbleSort sorts a slice in-place using the provided comparator.
// The comparator should return true if a < b (for ascending order).
func BubbleSort[T any](items []T, order int, less LessFunc[T]) {
	length := len(items)
	lastIndex := length - 1

	for i := 0; i < lastIndex; i++ {
		unsortedEnd := lastIndex - i

		for leftIndex := 0; leftIndex < unsortedEnd; leftIndex++ {
			rightIndex := leftIndex + 1

			asc := order == SortAscending && !less(items[lastIndex], items[rightIndex])
			desc := order == SortDescending && less(items[lastIndex], items[rightIndex])
			shouldSwap := asc || desc
			if shouldSwap {
				items[lastIndex], items[rightIndex] = items[rightIndex], items[lastIndex]
			}
		}
	}
}
