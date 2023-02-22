package graph

import "errors"

var (
	ErrDAGCycle    = errors.New("edges would create cycle")
	ErrDAGHasCycle = errors.New("the graph contains a cycle")
)

// dag represents a directed graph that has no cycles. It is a graph
// where there is no path that starts and ends at the same vertex.
type dag struct {
	*base
}

func NewDAG() Graph {
	return &dag{
		base: newBaseGraph(),
	}
}

// AddEdge adds a directed edges from the vertex with the 'from' id to
// the vertex with the 'to' id, after checking if the edges would create
// a cycle.
//
// AddEdge guarantees that the graph remain dag after adding new edges.
//
// It returns error if it finds a cycle between 'from' and 'to'.
func (g *dag) AddEdge(from, to *Vertex) (*Edge, error) {
	// Add the new edges
	from.neighbors = append(from.neighbors, to)
	from.outDegree++
	to.inDegree++

	// If topological sort returns an error, new edges created a cycle
	_, err := TopologySort(g)
	if err != nil {
		// Remove the new edges
		from.neighbors = from.neighbors[:len(from.neighbors)-1]
		to.inDegree--
		from.outDegree--

		return nil, ErrDAGCycle
	}

	return g.addToEdgeMap(from, to), nil
}

// TopologySort performs a topological sort of the graph using
// Kahn's algorithm. If the sorted list of vertices does not contain
// all vertices in the graph, it means there is a cycle in the graph.
//
// It returns error if it finds a cycle in the graph.
func TopologySort(g Graph) ([]*Vertex, error) {
	// Initialize a map to store the inDegree of each vertex
	inDegrees := make(map[*Vertex]int)
	vertices := g.GetAllVertices()
	for _, v := range vertices {
		inDegrees[v] = v.inDegree
	}

	// Initialize a queue with vertices of inDegrees zero
	queue := make([]*Vertex, 0)
	for v, inDegree := range inDegrees {
		if inDegree == 0 {
			queue = append(queue, v)
		}
	}

	// Initialize the sorted list of vertices
	sortedVertices := make([]*Vertex, 0)

	// Loop through the vertices with inDegree zero
	for len(queue) > 0 {
		// Get the next vertex with inDegree zero
		curr := queue[0]
		queue = queue[1:]

		// Add the vertex to the sorted list
		sortedVertices = append(sortedVertices, curr)

		// Decrement the inDegree of each of the vertex's neighbors
		for _, neighbor := range curr.neighbors {
			inDegrees[neighbor]--
			if inDegrees[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If the sorted list does not contain all vertices, there is a cycle
	if len(sortedVertices) != len(vertices) {
		return nil, ErrDAGHasCycle
	}

	return sortedVertices, nil
}
