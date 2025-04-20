package cluster

import (
	"cmp"

	"github.com/moorara/algo/unionfind"
)

type Group[T Locatable[T, R], R cmp.Ordered] struct {
	Items []T
}

// GroupByDistance partitions a slice of items into proximity-based groups.
// Any two items whose DistanceTo value is less than or equal to the given
// threshold are considered to belong to the same group.
func GroupByDistance[T Locatable[T, R], R cmp.Ordered](items []T, threshold R) []Group[T, R] {
	n := len(items)
	uf := unionfind.NewQuickUnion(n)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if items[i].DistanceTo(items[j]) <= threshold {
				uf.Union(i, j)
			}
		}
	}

	groupMap := make(map[int][]T)
	for i := 0; i < n; i++ {
		root, _ := uf.Find(i)
		groupMap[root] = append(groupMap[root], items[i])
	}

	var groups []Group[T, R]
	for _, group := range groupMap {
		groups = append(groups, Group[T, R]{
			Items: group,
		})
	}

	return groups
}
