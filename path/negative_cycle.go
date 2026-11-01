package path

import (
	"slices"

	"github.com/hmdsefi/gograph"
)

// NegativeCycle returns one cycle whose edge weights sum to less than zero.
// The slice does not repeat the first vertex. The edge from the last vertex
// back to the first closes the cycle. A negative self-loop is a cycle of one
// vertex. The vertices are the graph's own pointers.
//
// A nil slice and a nil error mean there is no negative cycle.
//
// The search covers the whole graph, not only what one start vertex reaches.
// When several cycles exist, the order of g.AllEdges decides which one is
// returned, so the same graph gives the same cycle on every run.
//
// It returns ErrNotWeighted or ErrNotDirected for a graph that BellmanFord
// doesn't accept either. Edges with a NaN weight are never relaxed.
//
// Time complexity is O(V*E).
func NegativeCycle[T comparable](g gograph.Graph[T]) ([]*gograph.Vertex[T], error) {
	if !g.IsWeighted() {
		return nil, ErrNotWeighted
	}

	if !g.IsDirected() {
		return nil, ErrNotDirected
	}

	vertices := g.GetAllVertices()
	index := make(map[T]int, len(vertices))
	for i, v := range vertices {
		index[v.Label()] = i
	}

	// Starting every distance at 0 is the same as a temporary source with a
	// zero-weight edge to every vertex, so no vertex is out of reach.
	dist := make([]float64, len(vertices))
	parent := make([]int, len(vertices))
	for i := range parent {
		parent[i] = -1
	}

	edges := g.AllEdges()
	for range vertices {
		changed := false
		for _, edge := range edges {
			from, to := index[edge.Source().Label()], index[edge.Destination().Label()]
			if d := dist[from] + edge.Weight(); d < dist[to] {
				dist[to] = d
				parent[to] = from
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	return vertexCycle(vertices, parent), nil
}

// vertexCycle returns the first cycle in the parent links, in edge order, or
// nil if they don't form one. A cycle in these links is always negative, and a
// pass that still relaxes an edge after V-1 passes always leaves one.
func vertexCycle[T comparable](vertices []*gograph.Vertex[T], parent []int) []*gograph.Vertex[T] {
	seenFrom := make([]int, len(parent))
	for start := range parent {
		v := start
		for v != -1 && seenFrom[v] == 0 {
			seenFrom[v] = start + 1
			v = parent[v]
		}
		if v == -1 || seenFrom[v] != start+1 {
			continue
		}

		cycle := []*gograph.Vertex[T]{vertices[v]}
		for u := parent[v]; u != v; u = parent[u] {
			cycle = append(cycle, vertices[u])
		}
		slices.Reverse(cycle)
		return cycle
	}
	return nil
}
