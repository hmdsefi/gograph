package connectivity

import "github.com/hmdsefi/gograph"

// ConnectedComponents returns the connected components of g. For a directed
// graph, edge directions are ignored, so it returns the weakly connected
// components. The vertices in the result are the vertices stored in g, not
// copies, and an empty graph returns an empty slice.
//
// The components are in the order of their first vertex in GetAllVertices.
// In an undirected graph, the vertices of a component are in breadth-first
// order from that vertex, following the order the edges were added. In a
// directed graph, they are in GetAllVertices order.
//
// A graph from gograph.New takes O(V+E) time, up to the inverse Ackermann
// factor of union-find for directed graphs, and uses O(V) extra space. Any
// other Graph is read through GetAllVertices and EdgesOf. It panics if g is
// nil.
func ConnectedComponents[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T] {
	return gograph.ConnectedComponents(g)
}
