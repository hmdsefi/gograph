package graph

// Graph represents a directed graph. It stores a slice of
// pointers to all vertices.
type Graph struct {
	// Vertices represents nodes or points in the graph
	Vertices []*Vertex
}

// AddVertex adds a new vertex with the given id to the graph.
func (g *Graph) AddVertex(id int) {
	v := &Vertex{ID: id}
	g.Vertices = append(g.Vertices, v)
}

// AddEdge adds a directed edge from the vertex with the 'from' id to
// the vertex with the 'to' id by appending the 'to' vertex to the
// 'Neighbors' slice of the 'from' vertex.
func (g *Graph) AddEdge(from, to int) {
	v1 := g.findVertex(from)
	v2 := g.findVertex(to)
	v1.Neighbors = append(v1.Neighbors, v2)
}

// findVertex searches for the given id in the vertices. It returns
// a pointer to the vertex if it finds it. Otherwise, returns nil.
func (g *Graph) findVertex(id int) *Vertex {
	for _, v := range g.Vertices {
		if v.ID == id {
			return v
		}
	}
	return nil
}
