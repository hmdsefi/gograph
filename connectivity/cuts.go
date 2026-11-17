package connectivity

import "github.com/hmdsefi/gograph"

// Bridges returns the edges whose removal increases the number of connected
// components of an undirected graph. Each bridge appears once, oriented from its
// DFS parent to its child, in DFS completion order. The edges belong to g.
// Self-loops are never bridges; empty graphs return an empty slice.
// Directed graphs return gograph.ErrNotUndirected. It runs in O(V+E) time and space.
func Bridges[T comparable](g gograph.Graph[T]) ([]*gograph.Edge[T], error) {
	bridges, _, err := gograph.UndirectedCuts(g)
	return bridges, err
}

// ArticulationPoints returns the vertices whose removal increases the number of
// connected components of an undirected graph, in GetAllVertices order.
// The vertices belong to g. Empty graphs return an empty slice; isolated vertices
// and self-loops are not articulation points. Directed graphs return
// gograph.ErrNotUndirected. It runs in O(V+E) time and space.
func ArticulationPoints[T comparable](g gograph.Graph[T]) ([]*gograph.Vertex[T], error) {
	_, points, err := gograph.UndirectedCuts(g)
	return points, err
}
