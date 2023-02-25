package graph

type undirected[T comparable] struct {
	*baseGraph[T]
}

func (u *undirected[T]) RemoveEdges(edges ...*Edge[T]) {
	for _, edge := range edges {
		u.baseGraph.removeEdge(edge)
		u.baseGraph.removeEdge(NewEdge(edge.dest, edge.source))
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

	for i := range v.neighbors {
		v.neighbors[i].inDegree--
	}

	for sourceID := range u.edges {
		if edge, ok := u.edges[sourceID][v.label]; ok {
			u.baseGraph.removeEdge(edge)
			u.baseGraph.removeEdge(NewEdge(edge.dest, edge.source))
			delete(u.edges[sourceID], v.label)
		}
	}

	delete(u.edges, v.label)
	delete(u.vertices, v.label)
}
