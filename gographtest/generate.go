package gographtest

import (
	"math"
	"math/rand/v2"

	"github.com/hmdsefi/gograph"
)

// Path returns the path 0 -> 1 -> ... -> n-1, built with gograph.New and
// the given options. It has n vertices and n-1 edges, and n of 0 or less
// gives an empty graph. In an undirected graph, each edge joins two
// neighbors on the path. It takes O(n) time.
func Path(n int, opts ...gograph.GraphOptionFunc) gograph.Graph[int] {
	g, v := withVertices(n, opts)
	for i := 1; i < len(v); i++ {
		_, _ = g.AddEdge(v[i-1], v[i])
	}
	return g
}

// Cycle returns the path from Path with one more edge, n-1 -> 0, so it has
// n vertices and n edges. Cycle(1) is the self-loop 0 -> 0. In an
// undirected graph, Cycle(2) has the one edge between 0 and 1, which the
// closing edge repeats. In a graph created with gograph.Acyclic, the closing
// edge would close a cycle, so the graph refuses it and the result is the
// path. It takes O(n) time.
func Cycle(n int, opts ...gograph.GraphOptionFunc) gograph.Graph[int] {
	g := Path(n, opts...)
	if n > 0 {
		_, _ = g.AddEdge(g.GetVertexByID(n-1), g.GetVertexByID(0))
	}
	return g
}

// Complete returns the graph with an edge between every two of its n
// vertices, and no self-loops. A directed graph has both directions of each
// pair, n(n-1) edges in all, and an undirected graph has each pair once. In
// a graph created with gograph.Acyclic, the edges from a higher label to a
// lower one would close cycles, so the result keeps the edges from lower
// labels to higher ones. It takes O(n^2) time.
func Complete(n int, opts ...gograph.GraphOptionFunc) gograph.Graph[int] {
	g, v := withVertices(n, opts)
	directed := g.IsDirected()
	for i := range v {
		for j := range v {
			if j > i || (directed && j != i) {
				_, _ = g.AddEdge(v[i], v[j])
			}
		}
	}
	return g
}

// Grid returns a grid of rows by cols vertices, where the vertex in row r
// and column c has the label r*cols + c. Each vertex has an edge to the
// next vertex in its row and to the next one in its column, so a directed
// grid points right and down. A grid has rows*(cols-1) + (rows-1)*cols
// edges, and rows or cols of 0 or less gives an empty graph. It takes
// O(rows*cols) time.
func Grid(rows, cols int, opts ...gograph.GraphOptionFunc) gograph.Graph[int] {
	if rows <= 0 || cols <= 0 {
		return gograph.New[int](opts...)
	}

	g, v := withVertices(rows*cols, opts)
	for r := range rows {
		for c := range cols {
			i := r*cols + c
			if c+1 < cols {
				_, _ = g.AddEdge(v[i], v[i+1])
			}
			if r+1 < rows {
				_, _ = g.AddEdge(v[i], v[i+cols])
			}
		}
	}
	return g
}

// Random returns an Erdos-Renyi G(n, p) graph: each pair of its n vertices
// gets an edge with probability p, independently of the others. A directed
// graph picks each ordered pair, and an undirected graph each unordered
// pair. There are no self-loops. A p of 0 or less, or NaN, gives no edges,
// and a p of 1 or more gives the edges of Complete.
//
// It reads rng only when p is strictly between 0 and 1, so the same seed
// gives the same graph. In a graph created with gograph.Acyclic, an edge
// that would close a cycle is refused and left out.
//
// It jumps from one picked pair to the next by drawing the length of each
// gap, so it takes O(n + m) expected time for the m edges it adds, instead
// of a draw for each of the n^2 pairs.
func Random(n int, p float64, rng *rand.Rand, opts ...gograph.GraphOptionFunc) gograph.Graph[int] {
	g, v := withVertices(n, opts)
	if g.IsDirected() {
		// pair k is (u, the k-th vertex after skipping u) in the row of u
		u, rowStart := 0, 0
		eachPair(len(v)*(len(v)-1), p, rng, func(k int) {
			for k >= rowStart+len(v)-1 {
				rowStart += len(v) - 1
				u++
			}
			w := k - rowStart
			if w >= u {
				w++
			}
			_, _ = g.AddEdge(v[u], v[w])
		})
		return g
	}

	addLowerToHigher(g, v, p, rng)
	return g
}

// RandomDAG returns a directed graph with n vertices where each pair gets
// an edge from its lower label to its higher label with probability p, so
// it has no cycles. It's built with gograph.Directed, without
// gograph.Acyclic, so it keeps no topological order of its own. A p of 0 or
// less, or NaN, gives no edges, and a p of 1 or more gives all n(n-1)/2.
//
// It reads rng only when p is strictly between 0 and 1, so the same seed
// gives the same graph. Like Random, it takes O(n + m) expected time for
// the m edges it adds.
func RandomDAG(n int, p float64, rng *rand.Rand) gograph.Graph[int] {
	g, v := withVertices(n, []gograph.GraphOptionFunc{gograph.Directed()})
	addLowerToHigher(g, v, p, rng)
	return g
}

// withVertices returns a graph built with opts and its vertices 0 to n-1,
// added in order.
func withVertices(n int, opts []gograph.GraphOptionFunc) (gograph.Graph[int], []*gograph.Vertex[int]) {
	g := gograph.New[int](opts...)
	v := make([]*gograph.Vertex[int], max(n, 0))
	for i := range v {
		v[i] = g.AddVertexByLabel(i)
	}
	return g, v
}

// addLowerToHigher adds each edge u -> w with u < w with probability p.
func addLowerToHigher(g gograph.Graph[int], v []*gograph.Vertex[int], p float64, rng *rand.Rand) {
	// pair k is (u, u+1+k-rowStart), where the row of u has len(v)-1-u pairs
	u, rowStart := 0, 0
	eachPair(len(v)*(len(v)-1)/2, p, rng, func(k int) {
		for k >= rowStart+len(v)-1-u {
			rowStart += len(v) - 1 - u
			u++
		}
		_, _ = g.AddEdge(v[u], v[u+1+k-rowStart])
	})
}

// eachPair calls add with the index of each of the pairs 0 to total-1 that
// it picks, in increasing order, picking each one with probability p. The
// gap before the next picked pair follows a geometric distribution, so it
// takes one draw from rng for each pick, plus one at the end.
func eachPair(total int, p float64, rng *rand.Rand, add func(k int)) {
	switch {
	case total <= 0 || !(p > 0):
		return
	case p >= 1:
		for k := range total {
			add(k)
		}
		return
	}

	logMiss := math.Log1p(-p)
	for k := -1; ; {
		gap := math.Floor(math.Log1p(-rng.Float64()) / logMiss)
		if gap >= float64(total-1-k) {
			return
		}
		k += 1 + int(gap)
		add(k)
	}
}
