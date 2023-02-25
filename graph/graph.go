package graph

import "errors"

var (
	ErrNilVertices       = errors.New("vertices are nil")
	ErrEdgeAlreadyExists = errors.New("edge already exists")
	ErrDAGCycle          = errors.New("edges would create cycle")
	ErrDAGHasCycle       = errors.New("the graph contains a cycle")
)

type Graph[T comparable] interface {
	AddEdge(from, to *Vertex[T]) (*Edge[T], error)
	GetAllEdges(from, to *Vertex[T]) []*Edge[T]
	GetEdge(from, to *Vertex[T]) *Edge[T]
	EdgesOf(v *Vertex[T]) []*Edge[T]
	RemoveEdges(edges ...*Edge[T])

	AddVertexByLabel(label T) *Vertex[T]
	AddVertex(v *Vertex[T])
	GetVertexByID(label T) *Vertex[T]
	GetAllVerticesByID(label ...T) []*Vertex[T]
	GetAllVertices() []*Vertex[T]
	RemoveVertices(vertices ...*Vertex[T])

	ContainsEdge(from, to *Vertex[T]) bool
	ContainsVertex(v *Vertex[T]) bool
}

func New[T comparable](options ...GraphOptionFunc) Graph[T] {
	var properties GraphProperties
	for _, option := range options {
		option(&properties)
	}

	base := &baseGraph[T]{
		vertices:   make(map[T]*Vertex[T]),
		edges:      make(map[T]map[T]*Edge[T]),
		properties: properties,
	}

	switch {
	case !properties.isDirected:
		return &undirected[T]{baseGraph: base}
	default:
		return base
	}
}

// Edge represents an edges in a graph. It contains start and end points.
type Edge[T comparable] struct {
	source *Vertex[T] // start point of the edges
	dest   *Vertex[T] // destination or end point of the edges
}

func NewEdge[T comparable](source *Vertex[T], dest *Vertex[T]) *Edge[T] {
	return &Edge[T]{source: source, dest: dest}
}

// Vertex represents a node or point in a graph
type Vertex[T comparable] struct {
	label     T            // uniquely identifies each vertex
	neighbors []*Vertex[T] //stores pointers to its neighbors
	inDegree  int          // number of incoming edges to this vertex
}

func NewVertex[T comparable](label T) *Vertex[T] {
	return &Vertex[T]{label: label}
}

// NeighborByID iterates over the neighbor slice and returns the
// vertex which its label is equal to the input label.
//
// It returns nil if there is no neighbor with that label.
func (v *Vertex[T]) NeighborByID(label T) *Vertex[T] {
	for i := range v.neighbors {
		if v.neighbors[i].label == label {
			return v.neighbors[i]
		}
	}

	return nil
}

// HasNeighbor checks if the input vertex is the neighbor of the
// current node or not. It returns 'true' if it finds the input
// in the neighbors. Otherwise, returns 'false'.
func (v *Vertex[T]) HasNeighbor(vertex *Vertex[T]) bool {
	return v.NeighborByID(vertex.label) != nil
}

// InDegree returns the number of incoming edges to the current vertex.
func (v *Vertex[T]) InDegree() int {
	return v.inDegree
}

// OutDegree returns the number of outgoing edges to the current vertex.
func (v *Vertex[T]) OutDegree() int {
	return v.inDegree
}

// Neighbors returns a copy of neighbor slice. If the caller changed the
// result slice, it won't impact the graph or the vertex.
func (v *Vertex[T]) Neighbors() []*Vertex[T] {
	var neighbors []*Vertex[T]
	for i := range v.neighbors {
		clone := &Vertex[T]{}
		*clone = *v.neighbors[i]
		neighbors = append(neighbors, clone)
	}

	return neighbors
}

// Label returns vertex label.
func (v *Vertex[T]) Label() T {
	return v.label
}
