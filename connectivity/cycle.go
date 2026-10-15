package connectivity

import "github.com/hmdsefi/gograph"

// HasCycle reports whether g contains a cycle. Directed graphs are checked
// using their edge directions. In undirected graphs, the two stored directions
// of an edge are treated as one edge. A self-loop is a cycle.
//
// A graph created with gograph.Acyclic returns false without a search. A
// graph from gograph.New takes O(V+E) time and O(V) extra space, and the
// search stops at the first cycle it finds. Any other graph is read through
// its methods. It panics if g is nil.
func HasCycle[T comparable](g gograph.Graph[T]) bool {
	return gograph.HasCycle(g)
}
