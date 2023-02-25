package graph

type undirected[T comparable] struct {
	*baseGraph[T]
}

func newUndirected[T comparable](graph *baseGraph[T]) *undirected[T] {
	return &undirected[T]{
		baseGraph: graph,
	}
}

func (u *undirected[T]) RemoveEdges(edges ...*Edge[T]) {
	for _, edge := range edges {
		u.baseGraph.removeEdge(edge)
		u.baseGraph.removeEdge(NewEdge(edge.dest, edge.source))
	}
}

// RemoveVertices removes all the specified vertices from this graph including
// all its touching edges if present.
func (u *undirected[T]) RemoveVertices(vertices ...*Vertex[T]) {
	for i := range vertices {
		u.removeVertex(vertices[i])
	}
}

func (u *undirected[T]) removeVertex(in *Vertex[T]) {
	if in == nil {
		return
	}

	v := u.findVertex(in.label)
	if v == nil {
		return
	}

	//for i := range v.neighbors {
	//	v.neighbors[i].inDegree--
	//}

	// clone neighbor slice
	cloneNeighbors := v.neighbors
	for i := range cloneNeighbors {
		u.baseGraph.removeEdge(NewEdge(v, cloneNeighbors[i]))
		u.baseGraph.removeEdge(NewEdge(cloneNeighbors[i], v))
	}

	delete(u.edges, v.label)
	delete(u.vertices, v.label)
}
