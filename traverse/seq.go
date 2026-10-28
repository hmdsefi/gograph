package traverse

import (
	"iter"
	"slices"

	"github.com/hmdsefi/gograph"
)

// BFS returns the vertices that start reaches, in breadth-first order: the
// order of NewBreadthFirstIterator. It yields nothing if start isn't in g,
// and skips a vertex removed from g after it was queued.
//
// Ranging over all of it takes O(V+E) time and O(V) extra space for the V
// vertices and E edges that start reaches.
func BFS[T comparable](g gograph.Graph[T], start T) iter.Seq[*gograph.Vertex[T]] {
	return func(yield func(*gograph.Vertex[T]) bool) {
		for v := range BFSWithDepth(g, start) {
			if !yield(v) {
				return
			}
		}
	}
}

// BFSWithDepth returns the vertices that start reaches, in the order of BFS,
// each with its depth. start has depth 0, and every other vertex the number
// of edges on a shortest path from start. It yields nothing if start isn't
// in g, and skips a vertex removed from g after it was queued.
//
// Ranging over all of it takes O(V+E) time and O(V) extra space for the V
// vertices and E edges that start reaches.
func BFSWithDepth[T comparable](g gograph.Graph[T], start T) iter.Seq2[*gograph.Vertex[T], int] {
	return func(yield func(*gograph.Vertex[T], int) bool) {
		bfs(g, start, yield)
	}
}

func bfs[T comparable](g gograph.Graph[T], start T, yield func(*gograph.Vertex[T], int) bool) {
	first := g.GetVertexByID(start)
	if first == nil {
		return
	}

	queue := []*gograph.Vertex[T]{first}
	queued := map[*gograph.Vertex[T]]struct{}{first: {}}
	depth, levelEnd := 0, 1
	for head := 0; head < len(queue); head++ {
		if head == levelEnd {
			depth++
			levelEnd = len(queue)
		}

		v := queue[head]
		if !hasEdges(v) && !stored(g, v) {
			continue
		}
		if !yield(v, depth) {
			return
		}
		for n := range v.AllNeighbors() {
			if _, ok := queued[n]; !ok {
				queued[n] = struct{}{}
				queue = append(queue, n)
			}
		}
	}
}

// DFS returns the vertices that start reaches, in the order of
// NewDepthFirstIterator: the next vertex is always the latest one found,
// and the neighbors of a vertex are found in the order of Neighbors. It
// yields nothing if start isn't in g, and skips a vertex removed from g
// after it was found.
//
// Ranging over all of it takes O(V+E) time and O(V) extra space for the V
// vertices and E edges that start reaches.
func DFS[T comparable](g gograph.Graph[T], start T) iter.Seq[*gograph.Vertex[T]] {
	return func(yield func(*gograph.Vertex[T]) bool) {
		dfs(g, start, yield)
	}
}

func dfs[T comparable](g gograph.Graph[T], start T, yield func(*gograph.Vertex[T]) bool) {
	first := g.GetVertexByID(start)
	if first == nil {
		return
	}

	stack := []*gograph.Vertex[T]{first}
	stacked := map[*gograph.Vertex[T]]struct{}{first: {}}
	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if !hasEdges(v) && !stored(g, v) {
			continue
		}
		if !yield(v) {
			return
		}
		for n := range v.AllNeighbors() {
			if _, ok := stacked[n]; !ok {
				stacked[n] = struct{}{}
				stack = append(stack, n)
			}
		}
	}
}

// ClosestFirst returns the vertices that start reaches, by increasing
// distance from start: the order of NewClosestFirstIterator. Each vertex
// comes with its distance, the sum of the edge weights on a shortest path
// from start. It yields nothing if start isn't in g, and skips a vertex
// removed from g after it was queued.
//
// Edge weights are expected to be non-negative. An edge with a NaN weight is
// skipped, and a vertex reached only through a +Inf weight comes after
// every vertex at a finite distance. An edge added without a weight weighs
// 0, so in a graph without weights every distance is 0. BFSWithDepth gives
// the number of edges from start instead.
//
// Ranging over all of it takes O((V+E) log V) time and O(V) extra space for
// the V vertices and E edges that start reaches.
func ClosestFirst[T comparable](g gograph.Graph[T], start T) iter.Seq2[*gograph.Vertex[T], float64] {
	return func(yield func(*gograph.Vertex[T], float64) bool) {
		closestFirst(g, start, yield)
	}
}

func closestFirst[T comparable](g gograph.Graph[T], start T, yield func(*gograph.Vertex[T], float64) bool) {
	first := g.GetVertexByID(start)
	if first == nil {
		return
	}

	pq := newClosestFirstQueue[T]()
	pq.push(nil, first, 0)
	for pq.Len() > 0 {
		entry := pq.pop()
		if !hasEdges(entry.vertex) && !stored(g, entry.vertex) {
			continue
		}
		if !yield(entry.vertex, entry.priority) {
			return
		}
		pq.relax(g, entry.vertex, entry.priority)
	}
}

// Topological returns the vertices of g in the order of
// gograph.TopologySort. It returns gograph.ErrNotDirected if g isn't
// directed and gograph.ErrDAGHasCycle if g has a cycle, each with a sequence
// that yields nothing.
//
// It sorts g when it is called, in O(V+E) time and O(V) space, so breaking
// out of the loop early saves nothing, and ranging over the sequence again
// yields the same order even if g has changed.
func Topological[T comparable](g gograph.Graph[T]) (iter.Seq[*gograph.Vertex[T]], error) {
	order, err := gograph.TopologySort(g)
	if err != nil {
		return func(func(*gograph.Vertex[T]) bool) {}, err
	}

	return slices.Values(order), nil
}
