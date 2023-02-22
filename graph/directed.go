package graph

import "errors"

var (
	ErrNilVertices = errors.New("vertices are nil")
)

// directedGraph represents a directed graph. It stores a slice of
// pointers to all vertices.
type directedGraph struct {
	*base
}

func NewDirectedGraph() Graph {
	return newDirectedGraph()
}

func newDirectedGraph() *directedGraph {
	return &directedGraph{
		base: newBaseGraph(),
	}
}

// AddEdge adds a directed edges from the vertex with the 'from' id to
// the vertex with the 'to' id by appending the 'to' vertex to the
// 'neighbors' slice of the 'from' vertex.
//
// It creates the input vertices if they don't exist in the graph.
// If any of the specified vertices is nil, returns nil.
func (g *directedGraph) AddEdge(from, to *Vertex) (*Edge, error) {
	if from == nil || to == nil {
		return nil, ErrNilVertices
	}

	if g.findVertex(from.id) == nil {
		g.AddVertex(from)
	}

	if g.findVertex(to.id) == nil {
		g.AddVertex(to)
	}

	from.neighbors = append(from.neighbors, to)
	from.outDegree++
	to.inDegree++

	return g.addToEdgeMap(from, to), nil
}
