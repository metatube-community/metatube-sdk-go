package utils

import "testing"

func TestSimilarity(t *testing.T) {
	for _, unit := range []struct {
		a, b string
	}{
		{"ABP-030", "ABP-030"},
		{"abp-030", "ABP-030"},
		{"ABS-030", "ABP-030"},
		{"AABP-030", "ABP-030"},
		{"KABP-030", "ABP-030"},
		{"ABP-030SP", "ABP-030"},
	} {
		t.Log(unit.a, unit.b, Similarity(unit.a, unit.b))
	}
}
