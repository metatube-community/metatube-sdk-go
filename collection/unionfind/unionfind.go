// Package unionfind implements union-find data structures and algorithms.
// It supports union and find queries.
//
// The union-find (a.k.a. disjoint-sets) data type is collection of n elements.
// Intially, each element belongs to exactly one set (n sets initially).
// Each set is represented by one element (canonical element, root, identifier, leader, or set representative).
// The union operation merges the set containing the element p with the set containing the element q.
// The find operation returns the canonical element of the set containing the element p.
//
// Elements in one set are considered connected to each other.
// "p is connected to q" is an equivalence relation:
//
//	Reflexive: p is connected to p.
//	Symmetric: If p is connected to q, then q is connected to p.
//	Transitive: If p is connected to q and q is connected to r, then p is connected to r.
//
// An equivalence relation partitions the objects into equivalence classes.
package unionfind

// UnionFind is the interface a union-find data type.
type UnionFind interface {
	Union(int, int)
	Find(int) (int, bool)
	IsConnected(int, int) bool
	Count() int
}

type quickFind struct {
	count int   // number of components (equivalence classes)
	id    []int // determines component IDs (class representatives)
}

// NewQuickFind creates a new union-find data structure with quick find.
func NewQuickFind(n int) UnionFind {
	id := make([]int, n)
	for i := 0; i < n; i++ {
		id[i] = i
	}

	return &quickFind{
		count: n,
		id:    id,
	}
}

func (u *quickFind) isValid(i int) bool {
	return 0 <= i && i < len(u.id)
}

func (u *quickFind) Union(p, q int) {
	if !u.isValid(p) || !u.isValid(q) {
		return
	}

	pid, _ := u.Find(p)
	qid, _ := u.Find(q)

	if pid == qid {
		return
	}

	// Rename p's component to q's id
	for i := range u.id {
		if u.id[i] == pid {
			u.id[i] = qid
		}
	}

	u.count--
}

func (u *quickFind) Find(p int) (int, bool) {
	if !u.isValid(p) {
		return -1, false
	}

	return u.id[p], true
}

func (u *quickFind) IsConnected(p, q int) bool {
	if !u.isValid(p) || !u.isValid(q) {
		return false
	}

	pid, _ := u.Find(p)
	qid, _ := u.Find(q)

	return pid == qid
}

func (u *quickFind) Count() int {
	return u.count
}

type quickUnion struct {
	count int   // number of components (equivalence classes)
	root  []int // determines component parents (class representatives)
}

// NewQuickUnion creates a new union-find data structure with quick union.
func NewQuickUnion(n int) UnionFind {
	root := make([]int, n)
	for i := 0; i < n; i++ {
		root[i] = i
	}

	return &quickUnion{
		count: n,
		root:  root,
	}
}

func (u *quickUnion) isValid(i int) bool {
	return 0 <= i && i < len(u.root)
}

func (u *quickUnion) Union(p, q int) {
	if !u.isValid(p) || !u.isValid(q) {
		return
	}

	proot, _ := u.Find(p)
	qroot, _ := u.Find(q)

	if proot == qroot {
		return
	}

	u.root[proot] = qroot
	u.count--
}

func (u *quickUnion) Find(p int) (int, bool) {
	if !u.isValid(p) {
		return -1, false
	}

	for p != u.root[p] {
		p = u.root[p]
	}

	return p, true
}

func (u *quickUnion) IsConnected(p, q int) bool {
	if !u.isValid(p) || !u.isValid(q) {
		return false
	}

	proot, _ := u.Find(p)
	qroot, _ := u.Find(q)

	return proot == qroot
}

func (u *quickUnion) Count() int {
	return u.count
}

type weightedQuickUnion struct {
	count int   // number of components (equivalence classes)
	root  []int // determines component parents (class representatives)
	size  []int // number of elements in component (class) rooted at i
}

// NewWeightedQuickUnion creates a new weighted union-find data structure with quick union.
func NewWeightedQuickUnion(n int) UnionFind {
	root := make([]int, n)
	size := make([]int, n)
	for i := 0; i < n; i++ {
		root[i] = i
		size[i] = 1
	}

	return &weightedQuickUnion{
		count: n,
		root:  root,
		size:  size,
	}
}

func (u *weightedQuickUnion) isValid(i int) bool {
	return 0 <= i && i < len(u.root)
}

func (u *weightedQuickUnion) Union(p, q int) {
	if !u.isValid(p) || !u.isValid(q) {
		return
	}

	proot, _ := u.Find(p)
	qroot, _ := u.Find(q)

	if proot == qroot {
		return
	}

	// make smaller root point to larger one
	if u.size[proot] < u.size[qroot] {
		u.root[proot] = qroot
		u.size[qroot] += u.size[proot]
	} else {
		u.root[qroot] = proot
		u.size[proot] += u.size[qroot]
	}

	u.count--
}

func (u *weightedQuickUnion) Find(p int) (int, bool) {
	if !u.isValid(p) {
		return -1, false
	}

	for p != u.root[p] {
		p = u.root[p]
	}

	return p, true
}

func (u *weightedQuickUnion) IsConnected(p, q int) bool {
	if !u.isValid(p) || !u.isValid(q) {
		return false
	}

	proot, _ := u.Find(p)
	qroot, _ := u.Find(q)

	return proot == qroot
}

func (u *weightedQuickUnion) Count() int {
	return u.count
}
