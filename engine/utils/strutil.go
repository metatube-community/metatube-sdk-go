package utils

import (
	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

// Similarity returns the similarity between two strings.
func Similarity(a, b string) float64 {
	m := metrics.NewLevenshtein()
	m.CaseSensitive = false
	m.InsertCost = 1
	m.DeleteCost = 1
	m.ReplaceCost = 2
	return strutil.Similarity(a, b, m)
}
