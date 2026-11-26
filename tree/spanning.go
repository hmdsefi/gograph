// Package tree finds spanning trees of undirected graphs.
//
// A minimum spanning tree connects every vertex of a connected graph with the
// smallest total edge weight, for example the cheapest set of cables or pipes
// that reaches every site. A graph in several pieces has a minimum spanning
// forest: one tree for each connected component.
package tree

import (
	"math"
	"slices"

	"github.com/hmdsefi/gograph"
)

// Kruskal returns a minimum spanning forest of an undirected graph: the edges
// of a minimum spanning tree of each connected component, and their total
// weight. A connected graph of V vertices gives V-1 edges, and a graph of C
// connected components gives V-C. An empty graph, or one with no edges, gives
// no edges and a weight of 0.
//
// Each undirected edge is considered once, even though it is stored in both
// directions, and self-loops are never part of a tree. The edges are the
// graph's own edges, each one in the direction that starts at the endpoint
// added to the graph first. An unweighted graph has weight 0 on every edge, so
// any spanning forest is a minimum one. An edge whose weight is NaN is left
// out. The total is the float64 sum of the weights of the forest, so an
// infinite weight, or finite weights whose sum overflows, makes it +Inf or
// -Inf, and it's NaN when both signs meet.
//
// The edges come sorted by weight. Edges of equal weight are compared by their
// endpoints in GetAllVertices order: first the endpoint added earlier, then the
// other one. That makes the minimum spanning forest unique, so the same graph
// gives the same edges on every run, and Prim returns the same edges in the
// same order.
//
// Both functions read the graph through GetAllVertices, the neighbors of each
// vertex, and GetEdge. A neighbor that isn't in GetAllVertices, or one for
// which GetEdge returns nil, is left out. Kruskal reads each edge from the
// endpoint added first. Prim reads it from the vertex that just joined its
// tree, and returns the edge from the endpoint added first, or the edge it
// read when GetEdge in that direction returns nil. The two agree on
// any graph that lists an undirected edge from both of its vertices, with the
// same weight in both directions, as the graphs from gograph.New do.
//
// A directed graph returns gograph.ErrNotUndirected.
//
// It sorts the edges and joins trees with a union-find over vertex numbers,
// with union by size and path halving. Time is O(V + E log E), since every
// vertex is read even when it has no edges, and extra space is O(V+E), for the
// sorted copy of the edges. Prim needs only O(V) extra space and is faster on
// dense graphs.
func Kruskal[T comparable](g gograph.Graph[T]) ([]*gograph.Edge[T], float64, error) {
	if g.IsDirected() {
		return nil, 0, gograph.ErrNotUndirected
	}

	vertices, index := vertexIndex(g)
	candidates := make([]weightedEdge[T], 0, g.Size()/2)
	for i, v := range vertices {
		low := int32(i) //nolint:gosec // a vertex number fits in int32
		for neighbor := range v.AllNeighbors() {
			// Each edge is read from its endpoint added first, so the second
			// direction and self-loops are skipped.
			if high, ok := index[neighbor]; ok && high > low {
				if c, ok := candidate(g, vertices, low, high); ok {
					candidates = append(candidates, c)
				}
			}
		}
	}
	slices.SortFunc(candidates, compareEdges[T])

	sets := newDisjointSets(len(vertices))
	forest := make([]*gograph.Edge[T], 0, min(len(candidates), max(len(vertices)-1, 0)))
	total := 0.0
	for _, c := range candidates {
		if len(forest) == len(vertices)-1 {
			break // one tree already spans every vertex
		}
		if sets.union(c.low, c.high) {
			forest = append(forest, c.edge)
			total += c.weight
		}
	}

	return forest, total, nil
}

// Prim returns the same minimum spanning forest as Kruskal, with the same
// edges in the same order and the same total weight. See Kruskal for the
// result, the order of the edges, how the graph is read, and the errors.
//
// It grows one tree at a time from the first vertex, in GetAllVertices order,
// that no tree reaches yet. The vertices outside the tree wait in an indexed
// d-ary heap, keyed by their cheapest edge to the tree, so a cheaper edge
// updates a vertex in place and the heap never holds more than V entries.
// The heap has d = max(2, E/V) children per node.
//
// Time is O(E log_d V + V log V), which is O(E log V) on a sparse graph and
// O(V²) on a dense one. Extra space is O(V).
func Prim[T comparable](g gograph.Graph[T]) ([]*gograph.Edge[T], float64, error) {
	if g.IsDirected() {
		return nil, 0, gograph.ErrNotUndirected
	}

	vertices, index := vertexIndex(g)
	n := len(vertices)
	// Size counts each edge between two vertices once in each direction.
	edges := int(g.Size()) / 2

	p := prim[T]{
		best:   make([]weightedEdge[T], n),
		state:  make([]uint8, n),
		heap:   newIndexedHeap(n, edges/max(n, 1)),
		forest: make([]weightedEdge[T], 0, max(n-1, 0)),
	}
	p.heap.less = func(a, b int32) bool { return compareEdges(p.best[a], p.best[b]) < 0 }

	for root := range vertices {
		if p.state[root] != waiting {
			continue
		}
		p.state[root] = inTree
		next := int32(root) //nolint:gosec // a vertex number fits in int32
		for next >= 0 {
			p.grow(g, vertices, index, next)
			next = -1
			if p.heap.len() > 0 {
				next = p.heap.pop()
				p.state[next] = inTree
				p.forest = append(p.forest, p.best[next])
			}
		}
	}

	slices.SortFunc(p.forest, compareEdges[T])
	forest := make([]*gograph.Edge[T], len(p.forest))
	total := 0.0
	for i, c := range p.forest {
		forest[i] = c.edge
		if c.edge.Source() != vertices[c.low] {
			if e := g.GetEdge(vertices[c.low], vertices[c.high]); e != nil {
				forest[i] = e
			}
		}
		total += c.weight
	}

	return forest, total, nil
}

// candidate returns the edge between the vertices numbered low and high, read
// with GetEdge from low to high, and false when there is none or its weight is
// NaN.
func candidate[T comparable](g gograph.Graph[T], vertices []*gograph.Vertex[T], low, high int32) (weightedEdge[T], bool) {
	e := g.GetEdge(vertices[low], vertices[high])
	if e == nil {
		return weightedEdge[T]{}, false
	}
	w := e.Weight()
	if math.IsNaN(w) {
		return weightedEdge[T]{}, false
	}
	return weightedEdge[T]{edge: e, weight: w, low: low, high: high}, true
}

// Vertex states in Prim.
const (
	waiting uint8 = iota // no tree reaches the vertex yet
	queued               // in the heap, with its cheapest edge in best
	inTree               // part of a finished or growing tree
)

// prim is the state of one Prim call. best holds, for each queued vertex, its
// cheapest edge to the growing tree.
type prim[T comparable] struct {
	best   []weightedEdge[T]
	state  []uint8
	heap   *indexedHeap
	forest []weightedEdge[T]
}

// grow adds the edges of vertex i, which just joined the tree, as candidates
// for its neighbors outside the tree.
func (p *prim[T]) grow(
	g gograph.Graph[T],
	vertices []*gograph.Vertex[T],
	index map[*gograph.Vertex[T]]int32,
	i int32,
) {
	for neighbor := range vertices[i].AllNeighbors() {
		j, ok := index[neighbor]
		if !ok || p.state[j] == inTree {
			continue // not in the graph, a self-loop, or an edge back into the tree
		}
		// Read the edge from the vertex that just joined, which keeps the
		// lookups in one row of the graph. The result switches to the edge
		// from the endpoint added first.
		e := g.GetEdge(vertices[i], neighbor)
		if e == nil {
			continue
		}
		w := e.Weight()
		if math.IsNaN(w) {
			continue
		}
		c := weightedEdge[T]{edge: e, weight: w, low: min(i, j), high: max(i, j)}

		switch {
		case p.state[j] == waiting:
			p.state[j] = queued
			p.best[j] = c
			p.heap.push(j)
		case compareEdges(c, p.best[j]) < 0:
			p.best[j] = c
			p.heap.up(j)
		}
	}
}

// weightedEdge is an edge with its weight and the numbers of its endpoints,
// low < high, in GetAllVertices order.
type weightedEdge[T comparable] struct {
	edge   *gograph.Edge[T]
	weight float64
	low    int32
	high   int32
}

// compareEdges orders edges by weight, then by their endpoints. No two edges
// of an undirected graph have the same endpoints, so it's a strict order.
func compareEdges[T comparable](a, b weightedEdge[T]) int {
	switch {
	case a.weight < b.weight:
		return -1
	case a.weight > b.weight:
		return 1
	case a.low != b.low:
		return int(a.low - b.low)
	default:
		return int(a.high - b.high)
	}
}

// vertexIndex returns the vertices of g in GetAllVertices order, and the
// number of each one.
func vertexIndex[T comparable](g gograph.Graph[T]) ([]*gograph.Vertex[T], map[*gograph.Vertex[T]]int32) {
	vertices := make([]*gograph.Vertex[T], 0, g.Order())
	index := make(map[*gograph.Vertex[T]]int32, g.Order())
	for v := range gograph.Vertices(g) {
		index[v] = int32(len(vertices)) //nolint:gosec // a vertex number fits in int32
		vertices = append(vertices, v)
	}
	return vertices, index
}

// disjointSets is a union-find over the numbers 0 to n-1.
type disjointSets struct {
	parent []int32
	size   []int32
}

func newDisjointSets(n int) *disjointSets {
	s := &disjointSets{parent: make([]int32, n), size: make([]int32, n)}
	for i := range s.parent {
		s.parent[i] = int32(i) //nolint:gosec // a vertex number fits in int32
		s.size[i] = 1
	}
	return s
}

// find returns the root of x, halving the path on the way.
func (s *disjointSets) find(x int32) int32 {
	for s.parent[x] != x {
		s.parent[x] = s.parent[s.parent[x]]
		x = s.parent[x]
	}
	return x
}

// union merges the sets of a and b, and reports false when they were already
// one set.
func (s *disjointSets) union(a, b int32) bool {
	a, b = s.find(a), s.find(b)
	if a == b {
		return false
	}
	if s.size[a] < s.size[b] {
		a, b = b, a
	}
	s.parent[b] = a
	s.size[a] += s.size[b]
	return true
}

// indexedHeap is a d-ary min-heap of vertex numbers. pos[v] is where v is in
// items, so a vertex whose key goes down moves up in place.
type indexedHeap struct {
	items []int32
	pos   []int32
	d     int
	less  func(a, b int32) bool
}

func newIndexedHeap(n, d int) *indexedHeap {
	return &indexedHeap{items: make([]int32, 0, n), pos: make([]int32, n), d: max(d, 2)}
}

func (h *indexedHeap) len() int { return len(h.items) }

func (h *indexedHeap) push(v int32) {
	h.pos[v] = int32(len(h.items)) //nolint:gosec // a heap index fits in int32
	h.items = append(h.items, v)
	h.up(v)
}

func (h *indexedHeap) pop() int32 {
	top := h.items[0]
	last := len(h.items) - 1
	h.items[0] = h.items[last]
	h.pos[h.items[0]] = 0
	h.items = h.items[:last]
	if last > 0 {
		h.down(0)
	}
	return top
}

// up moves v toward the root after its key went down.
func (h *indexedHeap) up(v int32) {
	i := int(h.pos[v])
	for i > 0 {
		parent := (i - 1) / h.d
		if !h.less(v, h.items[parent]) {
			break
		}
		h.items[i] = h.items[parent]
		h.pos[h.items[i]] = int32(i) //nolint:gosec // a heap index fits in int32
		i = parent
	}
	h.items[i] = v
	h.pos[v] = int32(i) //nolint:gosec // a heap index fits in int32
}

func (h *indexedHeap) down(i int) {
	v := h.items[i]
	n := len(h.items)
	for {
		first := i*h.d + 1
		if first >= n {
			break
		}
		smallest := first
		for c := first + 1; c < min(first+h.d, n); c++ {
			if h.less(h.items[c], h.items[smallest]) {
				smallest = c
			}
		}
		if !h.less(h.items[smallest], v) {
			break
		}
		h.items[i] = h.items[smallest]
		h.pos[h.items[i]] = int32(i) //nolint:gosec // a heap index fits in int32
		i = smallest
	}
	h.items[i] = v
	h.pos[v] = int32(i) //nolint:gosec // a heap index fits in int32
}
