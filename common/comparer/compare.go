package comparer

import (
	"github.com/adrg/strutil/metrics"
)

// Compare returns the similarity between two strings.
func Compare(a, b string) float64 {
	m := &metrics.Levenshtein{
		CaseSensitive: false,
		InsertCost:    1,
		DeleteCost:    1,
		ReplaceCost:   2,
	}
	return m.Compare(a, b)
}
