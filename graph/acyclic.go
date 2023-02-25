package graph

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
