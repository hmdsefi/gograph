package graph

// dag represents a directed graph that has no cycles. It is a graph
// where there is no path that starts and ends at the same vertex.
type dag[T comparable] struct {
	*base[T]
}

func NewDAG[T comparable]() Graph[T] {
	return &dag[T]{
		base: newBaseGraph[T](),
	}
}

// AddEdge adds a directed edges from the vertex with the 'from' label to
// the vertex with the 'to' label, after checking if the edges would create
// a cycle.
//
// AddEdge guarantees that the graph remain dag after adding new edges.
//
// If it finds a cycle between 'from' and 'to', returns error.
// If edge already exist, returns error.
func (g *dag[T]) AddEdge(from, to *Vertex[T]) (*Edge[T], error) {
	if from == nil || to == nil {
		return nil, ErrNilVertices
	}

	if g.findVertex(from.label) == nil {
		g.AddVertex(from)
	}

	if g.findVertex(to.label) == nil {
		g.AddVertex(to)
	}

	// prevent edge-multiplicity
	if g.ContainsEdge(from, to) {
		return nil, ErrEdgeAlreadyExists
	}

	// Add the new edges
	from.neighbors = append(from.neighbors, to)
	to.inDegree++

	// If topological sort returns an error, new edges created a cycle
	_, err := TopologySort[T](g)
	if err != nil {
		// Remove the new edges
		from.neighbors = from.neighbors[:len(from.neighbors)-1]
		to.inDegree--

		return nil, ErrDAGCycle
	}

	return g.addToEdgeMap(from, to), nil
}

// TopologySort performs a topological sort of the graph using
// Kahn's algorithm. If the sorted list of vertices does not contain
// all vertices in the graph, it means there is a cycle in the graph.
//
// It returns error if it finds a cycle in the graph.
func TopologySort[T comparable](g Graph[T]) ([]*Vertex[T], error) {
	// Initialize a map to store the inDegree of each vertex
	inDegrees := make(map[*Vertex[T]]int)
	vertices := g.GetAllVertices()
	for _, v := range vertices {
		inDegrees[v] = v.inDegree
	}

	// Initialize a queue with vertices of inDegrees zero
	queue := make([]*Vertex[T], 0)
	for v, inDegree := range inDegrees {
		if inDegree == 0 {
			queue = append(queue, v)
		}
	}

	// Initialize the sorted list of vertices
	sortedVertices := make([]*Vertex[T], 0)

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
