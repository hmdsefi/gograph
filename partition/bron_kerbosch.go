package partition

import (
	"container/heap"
	"math/bits"

	"github.com/hmdsefi/gograph"
)

// MaximalCliques finds all maximal cliques in the input graph using the
// Bron–Kerbosch algorithm with pivot selection, degeneracy ordering, and bitsets.
//
// A **clique** is a subset of vertices where every two distinct vertices are
// connected by an edge. A **maximal clique** is a clique that cannot be extended
// by adding another adjacent vertex.
//
// This implementation is optimized for performance:
//  1. **Degeneracy ordering**: processes vertices in a specific order to reduce
//     recursive calls and improve efficiency.
//  2. **Pivot selection**: selects a pivot vertex at each recursive call to
//     reduce the number of branches.
//  3. **Bitsets**: represents candidate sets (P, X) efficiently using []uint64
//     to speed up set operations on large graphs.
//
// Parameters:
//   - g: a gograph.Graph[T] representing the graph. T must be a comparable type.
//     Each vertex in the graph can be accessed via g.GetAllVertices() and
//     neighbors via Vertex.Neighbors().
//
// Returns:
//   - [][]*gograph.Vertex[T]: a slice of maximal cliques. Each clique is a slice
//     of pointers to Vertex[T]. Vertices in a clique are guaranteed to be fully
//     connected, and no clique is a subset of another.
//
// Complexity:
//   - **Time Complexity**: O(3^(n/3)) in the worst case for general graphs,
//     where n is the number of vertices. This is the known bound for enumerating
//     all maximal cliques. In practice, degeneracy ordering + pivoting reduces
//     the number of recursive calls significantly on sparse graphs.
//   - **Space Complexity**: O(n^2 / 64) for bitsets plus O(k*n) for storing cliques,
//     where k is the number of maximal cliques. Additional recursion stack space
//     is O(n) in depth.
//
// Example usage:
//
//	g := gograph.New[string]()
//	a := g.AddVertexByLabel("A")
//	b := g.AddVertexByLabel("B")
//	c := g.AddVertexByLabel("C")
//	_, _ = g.AddEdge(a, b)
//	_, _ = g.AddEdge(b, c)
//	_, _ = g.AddEdge(c, a)
//
//	cliques := MaximalCliques(g)
//	for _, clique := range cliques {
//	    for _, v := range clique {
//	        fmt.Print(v.Label(), " ")
//	    }
//	    fmt.Println()
//	}
//
// Notes:
//   - The function returns the actual Vertex pointers from the input graph;
//     do not modify the vertices while iterating the results.
//   - The order of cliques or vertices within a clique is not guaranteed.
//     If deterministic ordering is required, use a normalization function
//     (e.g., sort by vertex label).
func MaximalCliques[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T] {
	vertices := g.GetAllVertices()
	n := len(vertices)
	if n == 0 {
		return nil
	}

	// label -> index
	indexOf := make(map[T]int, n)
	for i, v := range vertices {
		indexOf[v.Label()] = i
	}

	// adjacency list (indices) and adjacency bitsets
	adj := make([][]int, n)
	neighborsBits := make([][]uint64, n)
	words := wordLen(n)
	for i := 0; i < n; i++ {
		neighborsBits[i] = make([]uint64, words)
	}

	for i, v := range vertices {
		for _, nb := range v.Neighbors() {
			if j, ok := indexOf[nb.Label()]; ok {
				adj[i] = append(adj[i], j)
				setBit(neighborsBits[i], j)
			}
		}
	}

	// degeneracy ordering (returns vertices removed low-degree first)
	order := degeneracyOrder(adj, n)

	// posInOrder: index -> position in order (used to split neighbors into P/X)
	posInOrder := make([]int, n)
	for pos, idx := range order {
		posInOrder[idx] = pos
	}

	// We'll collect cliques as slices of int indices first
	var cliquesIdx [][]int
	// scratch P/X for top-level calls
	P := make([]uint64, words)
	X := make([]uint64, words)

	// For each vertex v in degeneracy order:
	for _, v := range order {
		// reset P and X
		for i := range P {
			P[i] = 0
			X[i] = 0
		}
		// Build P = N(v) ∩ {vertices after v in order}
		// Build X = N(v) ∩ {vertices before v in order}
		for _, w := range adj[v] {
			if posInOrder[w] > posInOrder[v] {
				setBit(P, w)
			} else {
				setBit(X, w)
			}
		}

		// Recurse with R = {v}, cloned P and X
		bronKerboschPivot([]int{v}, cloneBitset(P), cloneBitset(X), neighborsBits, n, &cliquesIdx)

		// remove v implicitly (degeneracy ensures no duplicates)
	}

	// convert index cliques to []*Vertex[T]
	result := make([][]*gograph.Vertex[T], len(cliquesIdx))
	for i, cl := range cliquesIdx {
		out := make([]*gograph.Vertex[T], len(cl))
		for j, idx := range cl {
			out[j] = vertices[idx]
		}
		result[i] = out
	}
	return result
}

func wordLen(n int) int { return (n + 63) >> 6 }

func setBit(b []uint64, i int) {
	b[i>>6] |= 1 << uint(i&63)
}

func clearBit(b []uint64, i int) {
	b[i>>6] &^= 1 << uint(i&63)
}

func cloneBitset(b []uint64) []uint64 {
	if b == nil {
		return nil
	}
	c := make([]uint64, len(b))
	copy(c, b)
	return c
}

func intersectBitset(a, b []uint64) []uint64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	res := make([]uint64, n)
	for i := 0; i < n; i++ {
		res[i] = a[i] & b[i]
	}
	return res
}

func differenceBitset(a, b []uint64) []uint64 {
	n := len(a)
	res := make([]uint64, n)
	for i := 0; i < n; i++ {
		var bi uint64
		if i < len(b) {
			bi = b[i]
		}
		res[i] = a[i] &^ bi
	}
	return res
}

func unionBitset(a, b []uint64) []uint64 {
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	res := make([]uint64, n)
	for i := 0; i < n; i++ {
		var ai, bi uint64
		if i < len(a) {
			ai = a[i]
		}
		if i < len(b) {
			bi = b[i]
		}
		res[i] = ai | bi
	}
	return res
}

func countBits(b []uint64) int {
	c := 0
	for _, w := range b {
		c += bits.OnesCount64(w)
	}
	return c
}

// forEachSetBit calls fn(i) for every set bit in the bitset b.
// If fn returns true, iteration stops early.
func forEachSetBit(b []uint64, fn func(idx int) (stop bool)) {
	for wi, word := range b {
		for word != 0 {
			t := bits.TrailingZeros64(word)
			idx := (wi << 6) + t
			if fn(idx) {
				return
			}

			// clear the least significant set bit
			word &= word - 1
		}
	}
}

// ---------------------
// Degeneracy ordering (min-heap approach)
// ---------------------

type heapItem struct {
	deg int
	v   int
	// idx field not necessary for this simple push-new-updates approach
}
type minHeap []heapItem

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool  { return h[i].deg < h[j].deg }
func (h minHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) { *h = append(*h, x.(heapItem)) }
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// degeneracyOrder returns an ordering of vertex indices (low-degree first removed).
func degeneracyOrder(adj [][]int, n int) []int {
	deg := make([]int, n)
	for i := 0; i < n; i++ {
		deg[i] = len(adj[i])
	}

	h := &minHeap{}
	heap.Init(h)
	for i := 0; i < n; i++ {
		heap.Push(h, heapItem{deg: deg[i], v: i})
	}

	removed := make([]bool, n)
	order := make([]int, 0, n)

	for h.Len() > 0 {
		it := heap.Pop(h).(heapItem)
		v := it.v
		// skip outdated entries (we push updated degs rather than decrease-key)
		if removed[v] {
			continue
		}
		removed[v] = true
		order = append(order, v)
		for _, w := range adj[v] {
			if removed[w] {
				continue
			}
			deg[w]--
			heap.Push(h, heapItem{deg: deg[w], v: w})
		}
	}

	return order
}

// bronKerboschPivot does recursion; neighborsBits is adjacency bitset per vertex.
// n is number of vertices (for word sizes and potential masking if needed).
func bronKerboschPivot(
	R []int,
	P []uint64,
	X []uint64,
	neighborsBits [][]uint64,
	n int,
	cliques *[][]int,
) {
	// if P and X are empty → R is maximal
	if countBits(P) == 0 && countBits(X) == 0 {
		c := make([]int, len(R))
		copy(c, R)
		*cliques = append(*cliques, c)
		return
	}

	// choose pivot u from P ∪ X maximizing |P ∩ N(u)|
	unionPX := unionBitset(P, X)
	u := -1
	best := -1
	forEachSetBit(
		unionPX, func(idx int) bool {
			// compute |P ∩ N(idx)|
			cnt := countBits(intersectBitset(P, neighborsBits[idx]))
			if cnt > best {
				best = cnt
				u = idx
			}
			return false
		},
	)

	// candidates = P \ N(u)
	var candidates []uint64
	if u >= 0 {
		candidates = differenceBitset(P, neighborsBits[u])
	} else {
		candidates = cloneBitset(P)
	}

	// iterate over set bits in candidates
	// We must iterate over a snapshot (indices) because we'll mutate P/X during loop.
	var candidateIndices []int
	forEachSetBit(
		candidates, func(idx int) bool {
			candidateIndices = append(candidateIndices, idx)
			return false
		},
	)

	for _, v := range candidateIndices {
		// R' = R ∪ {v}
		Rp := append(R, v)

		// P' = P ∩ N(v)
		Pp := intersectBitset(P, neighborsBits[v])

		// X' = X ∩ N(v)
		Xp := intersectBitset(X, neighborsBits[v])

		// recurse
		bronKerboschPivot(Rp, Pp, Xp, neighborsBits, n, cliques)

		// move v from P to X in the current frame
		clearBit(P, v)
		setBit(X, v)
	}
}
