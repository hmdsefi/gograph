package path

import (
	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/traverse"
)

var (
	// ErrNotDAG is returned when a graph contains cycles but a function requires a DAG
	ErrNotDAG = gograph.ErrDAGHasCycle
)

// TransitiveReduction computes the transitive reduction of a directed graph.
// The transitive reduction of a directed graph G is a graph G' with the same vertices
// such that there is a path from vertex u to vertex v in G' if and only if there is
// a path from u to v in G, and G' has as few edges as possible.
//
// This implementation uses a depth-first search approach inspired by NetworkX library
// to efficiently compute the transitive reduction by identifying and removing edges
// that have alternate paths.
//
// For directed acyclic graphs (DAGs), the transitive reduction can be computed efficiently
// without needing to build the full transitive closure matrix.
//
// It returns an error if the graph is not directed or if the graph contains cycles.
func TransitiveReduction[T comparable](g gograph.Graph[T]) (gograph.Graph[T], error) {
	// Transitive reduction requires a directed graph
	if !g.IsDirected() {
		return nil, ErrNotDirected
	}

	// For a general directed graph, we need to ensure it's acyclic
	if !g.IsAcyclic() {
		// If the graph is not marked as acyclic, topology sort will return an error if it contains cycles
		_, err := gograph.TopologySort(g)
		if err != nil {
			return nil, ErrNotDAG
		}
	}

	// Create a new graph for the transitive reduction with the same properties as the input
	var reducedGraph gograph.Graph[T]
	if g.IsWeighted() {
		reducedGraph = gograph.New[T](gograph.Weighted(), gograph.Directed())
	} else {
		reducedGraph = gograph.New[T](gograph.Directed())
	}

	// Add all vertices from the original graph to the reduced graph
	vertices := g.GetAllVertices()
	for _, v := range vertices {
		reducedGraph.AddVertexByLabel(v.Label())
	}

	// Map to cache descendants for vertices that we've already processed
	descendants := make(map[T]map[T]bool)

	// Process each vertex in the graph
	for _, u := range vertices {
		// Get the neighbors of the current vertex
		neighbors := make(map[T]bool)
		for _, neighbor := range u.Neighbors() {
			neighbors[neighbor.Label()] = true
		}

		// For each neighbor of u, remove its descendants from consideration
		// as direct neighbors in the transitive reduction
		for _, v := range u.Neighbors() {
			// Skip if we've already removed this neighbor
			if _, exists := neighbors[v.Label()]; !exists {
				continue
			}

			// Get or compute descendants of v
			vDescendants, exists := descendants[v.Label()]
			if !exists {
				vDescendants = findDescendants(g, v)
				descendants[v.Label()] = vDescendants
			}

			// Remove v's descendants from u's neighbors
			for desc := range vDescendants {
				delete(neighbors, desc)
			}
		}

		// Add edges from u to its remaining neighbors in the reduced graph
		uVertex := reducedGraph.GetVertexByID(u.Label())
		for neighbor := range neighbors {
			vVertex := reducedGraph.GetVertexByID(neighbor)

			// Preserve edge weight if the graph is weighted
			if g.IsWeighted() {
				originalEdge := g.GetEdge(u, g.GetVertexByID(neighbor))
				if originalEdge != nil {
					_, err := reducedGraph.AddEdge(uVertex, vVertex, gograph.WithEdgeWeight(originalEdge.Weight()))
					if err != nil {
						return nil, err
					}
				}
			} else {
				_, err := reducedGraph.AddEdge(uVertex, vVertex)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return reducedGraph, nil
}

// findDescendants returns a map of all descendants of a vertex in the graph
// using the depth-first traversal iterator from the traverse package
func findDescendants[T comparable](g gograph.Graph[T], v *gograph.Vertex[T]) map[T]bool {
	descendants := make(map[T]bool)

	// Process each neighbor of the vertex
	for _, neighbor := range v.Neighbors() {
		// Create a depth-first iterator starting from this neighbor
		dfsIter, err := traverse.NewDepthFirstIterator(g, neighbor.Label())
		if err != nil {
			continue
		}

		// Add all other reachable vertices as descendants
		for dfsIter.HasNext() {
			descendant := dfsIter.Next()
			descendants[descendant.Label()] = true
		}
	}

	return descendants
}
