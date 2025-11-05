package unionfind

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	findTest struct {
		name       string
		p          int
		expectedID int
		expectedOK bool
	}

	connectivityTest struct {
		name                string
		p, q                int
		expectedIsConnected bool
	}

	unionFindTest struct {
		name              string
		n                 int
		unions            [][2]int
		expectedCount     int
		findTests         []findTest
		connectivityTests []connectivityTest
	}
)

var runTest = func(t *testing.T, tc unionFindTest, uf UnionFind) {
	for _, u := range tc.unions {
		uf.Union(u[0], u[1])
	}

	t.Run("Count", func(t *testing.T) {
		assert.Equal(t, tc.expectedCount, uf.Count())
	})

	t.Run("Find", func(t *testing.T) {
		for _, tc := range tc.findTests {
			t.Run(tc.name, func(t *testing.T) {
				id, ok := uf.Find(tc.p)
				assert.Equal(t, tc.expectedID, id)
				assert.Equal(t, tc.expectedOK, ok)
			})
		}
	})

	t.Run("IsConnected", func(t *testing.T) {
		for _, tc := range tc.connectivityTests {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, tc.expectedIsConnected, uf.IsConnected(tc.p, tc.q))
			})
		}
	})
}

func TestQuickFind(t *testing.T) {
	tests := []unionFindTest{
		{
			name: "OK",
			n:    10,
			unions: [][2]int{
				{-1, -2}, // invalid
				{0, 1},
				{2, 3},
				{2, 4},
				{3, 4}, // alredy connected
				{5, 6},
				{6, 7},
				{8, 7},
				{8, 6},   // alredy connected
				{11, 12}, // invalid
			},
			expectedCount: 4,
			findTests: []findTest{
				{
					name:       "Invalid",
					p:          -1,
					expectedID: -1,
					expectedOK: false,
				},
				{
					name:       "First",
					p:          1,
					expectedID: 1,
					expectedOK: true,
				},
				{
					name:       "Second",
					p:          4,
					expectedID: 4,
					expectedOK: true,
				},
				{
					name:       "Third",
					p:          8,
					expectedID: 7,
					expectedOK: true,
				},
				{
					name:       "Fourth",
					p:          9,
					expectedID: 9,
					expectedOK: true,
				},
			},
			connectivityTests: []connectivityTest{
				{
					name:                "Invalid",
					p:                   -1,
					q:                   11,
					expectedIsConnected: false,
				},
				{
					name:                "Connected#1",
					p:                   0,
					q:                   1,
					expectedIsConnected: true,
				},
				{
					name:                "Connected#2",
					p:                   2,
					q:                   4,
					expectedIsConnected: true,
				},
				{
					name:                "Connected#3",
					p:                   6,
					q:                   8,
					expectedIsConnected: true,
				},
				{
					name:                "Disconnected#1",
					p:                   1,
					q:                   3,
					expectedIsConnected: false,
				},
				{
					name:                "Disconnected#2",
					p:                   3,
					q:                   5,
					expectedIsConnected: false,
				},
				{
					name:                "Disconnected#3",
					p:                   7,
					q:                   9,
					expectedIsConnected: false,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uf := NewQuickFind(tc.n)
			runTest(t, tc, uf)
		})
	}
}

func TestQuickUnion(t *testing.T) {
	tests := []unionFindTest{
		{
			name: "OK",
			n:    10,
			unions: [][2]int{
				{-1, -2}, // invalid
				{0, 1},
				{2, 3},
				{2, 4},
				{3, 4}, // alredy connected
				{5, 6},
				{6, 7},
				{8, 7},
				{8, 6},   // alredy connected
				{11, 12}, // invalid
			},
			expectedCount: 4,
			findTests: []findTest{
				{
					name:       "Invalid",
					p:          -1,
					expectedID: -1,
					expectedOK: false,
				},
				{
					name:       "First",
					p:          1,
					expectedID: 1,
					expectedOK: true,
				},
				{
					name:       "Second",
					p:          4,
					expectedID: 4,
					expectedOK: true,
				},
				{
					name:       "Third",
					p:          8,
					expectedID: 7,
					expectedOK: true,
				},
				{
					name:       "Fourth",
					p:          9,
					expectedID: 9,
					expectedOK: true,
				},
			},
			connectivityTests: []connectivityTest{
				{
					name:                "Invalid",
					p:                   -1,
					q:                   11,
					expectedIsConnected: false,
				},
				{
					name:                "Connected#1",
					p:                   0,
					q:                   1,
					expectedIsConnected: true,
				},
				{
					name:                "Connected#2",
					p:                   2,
					q:                   4,
					expectedIsConnected: true,
				},
				{
					name:                "Connected#3",
					p:                   6,
					q:                   8,
					expectedIsConnected: true,
				},
				{
					name:                "Disconnected#1",
					p:                   1,
					q:                   3,
					expectedIsConnected: false,
				},
				{
					name:                "Disconnected#2",
					p:                   3,
					q:                   5,
					expectedIsConnected: false,
				},
				{
					name:                "Disconnected#3",
					p:                   7,
					q:                   9,
					expectedIsConnected: false,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uf := NewQuickUnion(tc.n)
			runTest(t, tc, uf)
		})
	}
}

func TestWeightedQuickUnion(t *testing.T) {
	tests := []unionFindTest{
		{
			name: "OK",
			n:    10,
			unions: [][2]int{
				{-1, -2}, // invalid
				{0, 1},
				{2, 3},
				{2, 4},
				{3, 4}, // alredy connected
				{5, 6},
				{6, 7},
				{8, 7},
				{8, 6},   // alredy connected
				{11, 12}, // invalid
			},
			expectedCount: 4,
			findTests: []findTest{
				{
					name:       "Invalid",
					p:          -1,
					expectedID: -1,
					expectedOK: false,
				},
				{
					name:       "First",
					p:          1,
					expectedID: 0,
					expectedOK: true,
				},
				{
					name:       "Second",
					p:          4,
					expectedID: 2,
					expectedOK: true,
				},
				{
					name:       "Third",
					p:          8,
					expectedID: 5,
					expectedOK: true,
				},
				{
					name:       "Fourth",
					p:          9,
					expectedID: 9,
					expectedOK: true,
				},
			},
			connectivityTests: []connectivityTest{
				{
					name:                "Invalid",
					p:                   -1,
					q:                   11,
					expectedIsConnected: false,
				},
				{
					name:                "Connected#1",
					p:                   0,
					q:                   1,
					expectedIsConnected: true,
				},
				{
					name:                "Connected#2",
					p:                   2,
					q:                   4,
					expectedIsConnected: true,
				},
				{
					name:                "Connected#3",
					p:                   6,
					q:                   8,
					expectedIsConnected: true,
				},
				{
					name:                "Disconnected#1",
					p:                   1,
					q:                   3,
					expectedIsConnected: false,
				},
				{
					name:                "Disconnected#2",
					p:                   3,
					q:                   5,
					expectedIsConnected: false,
				},
				{
					name:                "Disconnected#3",
					p:                   7,
					q:                   9,
					expectedIsConnected: false,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uf := NewWeightedQuickUnion(tc.n)
			runTest(t, tc, uf)
		})
	}
}
