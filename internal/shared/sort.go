package shared

import "slices"

type SortCol int8

const (
	SortID SortCol = iota
	SortRecords
	SortBytes
)

func (s SortCol) Next() SortCol  { return (s + 1) % 3 }
func (s SortCol) String() string { return [...]string{"id", "records", "bytes"}[s] }

// SortBy sorts components in-place, preserving kind grouping order.
func SortBy(cs []ComponentMetrics, col SortCol, asc bool) {
	less := map[SortCol]func(a, b ComponentMetrics) int{
		SortID:      func(a, b ComponentMetrics) int { return cmp(a.ID, b.ID) },
		SortRecords: func(a, b ComponentMetrics) int { return cmp(a.Records(), b.Records()) },
		SortBytes:   func(a, b ComponentMetrics) int { return cmp(a.Bytes(), b.Bytes()) },
	}[col]

	slices.SortStableFunc(cs, func(a, b ComponentMetrics) int {
		if a.Kind != b.Kind {
			return int(a.Kind - b.Kind)
		}
		c := less(a, b)
		if !asc {
			return -c
		}
		return c
	})
}

func cmp[T ~string | ~int64](a, b T) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
