package graph

// directedGraph represents a directed graph. It stores a slice of
// pointers to all vertices.
type directedGraph[T comparable] struct {
	*base[T]
}

func NewDirectedGraph[T comparable]() Graph[T] {
	return newDirectedGraph[T]()
}

func newDirectedGraph[T comparable]() *directedGraph[T] {
	return &directedGraph[T]{
		base: newBaseGraph[T](),
	}
}

// AddEdge adds a directed edges from the vertex with the 'from' label to
// the vertex with the 'to' label by appending the 'to' vertex to the
// 'neighbors' slice of the 'from' vertex.
//
// It creates the input vertices if they don't exist in the graph.
// If any of the specified vertices is nil, returns nil.
// If edge already exist, returns error.
func (g *directedGraph[T]) AddEdge(from, to *Vertex[T]) (*Edge[T], error) {
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

	from.neighbors = append(from.neighbors, to)
	from.outDegree++
	to.inDegree++

	return g.addToEdgeMap(from, to), nil
}
