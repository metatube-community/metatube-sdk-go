package cluster

import (
	"cmp"
	"slices"
	"sort"

	weighted "github.com/metatube-community/metatube-sdk-go/collection/slices"
)

// SortGroupsBySize sorts the provided groups in-place by descending the number of items.
func SortGroupsBySize[T Locatable[T, R], R cmp.Ordered](groups []Group[T, R]) {
	sort.SliceStable(groups, func(i, j int) bool {
		return len(groups[i].Items) > len(groups[j].Items)
	})
}

// SortGroupsByWeight sorts the provided groups in descending order of total weight.
func SortGroupsByWeight[T WeightedLocatable[T, R, W], R cmp.Ordered, W cmp.Ordered](groups []Group[T, R]) {
	if len(groups) <= 1 {
		return
	}

	// group weight calculator.
	weight := func(group Group[T, R]) W {
		var sum W
		for _, item := range group.Items {
			sum += item.Weight()
		}
		return sum
	}

	// calculate weights for each group.
	weights := slices.Collect(func(yield func(W) bool) {
		for _, group := range groups {
			if !yield(weight(group)) {
				return
			}
		}
	})

	// weighted stable sort.
	weighted.NewWeightedSlice(groups, weights).SortFunc(sort.Stable)
}
