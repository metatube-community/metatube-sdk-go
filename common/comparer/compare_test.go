package comparer

import "testing"

func TestCompare(t *testing.T) {
	for _, unit := range []struct {
		a, b string
	}{
		{"ABP-030", "ABP-030"},
		{"abp-030", "ABP-030"},
		{"ABS-030", "ABP-030"},
		{"AABP-030", "ABP-030"},
		{"KABP-030", "ABP-030"},
		{"ABP-030SP", "ABP-030"},
		{"松下紗栄子", "松下紗栄"},
		{"松下紗栄子", "松下栄子"},
		{"つぼみ", "Bae Bom"},
		{"松下紗栄子", "つぼみ"},
		{"つ", "つぼみ"},
		{"木村夏菜子", "木村夏"},
		{"木村夏菜子", "夏菜子"},
		{"葵", "葵千恵"},
		{"葵", "葵千"},
		{"葵", "葵つかさ"},
	} {
		t.Log(unit.a, unit.b, Compare(unit.a, unit.b))
	}
}
