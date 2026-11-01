package connectivity

import (
	"slices"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/traverse"
)

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

// FindCycle returns the vertices of one cycle of g in order, without
// repeating the first vertex at the end: [a b c] means a -> b -> c -> a. It
// returns nil and false when g has no cycle, so it reports a cycle exactly
// when HasCycle does.
//
// A self-loop is a cycle of one vertex. In a directed graph, a cycle
// follows the edge directions. In an undirected graph, the two stored
// directions of an edge count as one edge, so any other cycle has at least
// three vertices.
//
// The cycle is the first one that traverse.DFSEventsAll finds: the first
// back edge u -> v closes it, and it runs from v down the search path to u.
// That is one cycle of g, and a shorter one can exist.
//
// A graph created with gograph.Acyclic returns false without a search. Any
// other graph takes O(V+E) time, and the search stops at the first back
// edge. It keeps O(V) state, plus the neighbors still to look at of each
// vertex on the search path, which is at most O(V+E). HasCycle gives the
// answer alone faster, since it skips the edge events. It panics if g is
// nil.
func FindCycle[T comparable](g gograph.Graph[T]) ([]*gograph.Vertex[T], bool) {
	if g.IsAcyclic() {
		return nil, false
	}

	var path []*gograph.Vertex[T]
	for e := range traverse.DFSEventsAll(g) {
		switch e.Kind {
		case traverse.DiscoverVertex:
			path = append(path, e.Vertex)
		case traverse.FinishVertex:
			path = path[:len(path)-1]
		case traverse.BackEdge:
			// the source of a back edge is the last vertex of path
			to := e.Edge.Destination()
			for i := len(path) - 1; i >= 0; i-- {
				if path[i] == to {
					return slices.Clone(path[i:]), true
				}
			}
		}
	}

	return nil, false
}
