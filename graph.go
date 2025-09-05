package gograph

import (
	"errors"
)

var (
	ErrNilVertices        = errors.New("vertices are nil")
	ErrVertexDoesNotExist = errors.New("vertex does not exist")
	ErrEdgeAlreadyExists  = errors.New("edge already exists")
	ErrDAGCycle           = errors.New("edges would create cycle")
	ErrDAGHasCycle        = errors.New("the graph contains a cycle")
)

// Graph defines methods for managing a graph with vertices and edges. It is the
// base interface in the graph hierarchy. Each graph object contains a set of
// vertices and edges.
//
// Through generics, a graph can be typed to specific classes for vertices' label T.
type Graph[T comparable] interface {
	GraphType

	// AddEdge adds an edge from the vertex with the 'from' label to
	// the vertex with the 'to' label by appending the 'to' vertex to the
	// 'neighbors' slice of the 'from' vertex, in directed graph.
	//
	// In undirected graph, it also adds an edge from the vertex with
	// the 'to' label to the vertex with the 'from' label by appending
	// the 'from' vertex to the 'neighbors' slice of the 'to' vertex. it
	// means that it create the edges in both direction between the specified
	// vertices.
	//
	// This method accepts additional edge options such as weight and adds
	// them to the new edge.
	//
	//
	// It creates the input vertices if they don't exist in the graph.
	// If any of the specified vertices is nil, returns nil.
	// If edge already exist, returns error.
	AddEdge(from, to *Vertex[T], options ...EdgeOptionFunc) (*Edge[T], error)

	// GetAllEdges returns a slice of all edges connecting source vertex to
	// target vertex if such vertices exist in this graph.
	//
	// In directed graph, it returns a single edge.
	//
	// If any of the specified vertices is nil, returns nil.
	// If any of the vertices does not exist, returns nil.
	// If both vertices exist but no edges found, returns an empty set.
	GetAllEdges(from, to *Vertex[T]) []*Edge[T]

	// AllEdges returns all the edges in the graph.
	AllEdges() []*Edge[T]

	// GetEdge returns an edge connecting source vertex to target vertex
	// if such vertices and such edge exist in this graph.
	//
	// In undirected graph, returns only the edge from the "from" vertex to
	// the "to" vertex.
	//
	// If any of the specified vertices is nil, returns nil.
	// If edge does not exist, returns nil.
	GetEdge(from, to *Vertex[T]) *Edge[T]

	// EdgesOf returns a slice of all edges touching the specified vertex.
	// If no edges are touching the specified vertex returns an empty slice.
	//
	// If the input vertex is nil, returns nil.
	// If the input vertex does not exist, returns nil.
	EdgesOf(v *Vertex[T]) []*Edge[T]

	// RemoveEdges removes input edges from the graph from the specified
	// slice of edges, if they exist. In undirected graph, removes edges
	// in both directions.
	RemoveEdges(edges ...*Edge[T])

	// AddVertexByLabel adds a new vertex with the given label to the graph.
	// Label of the vertex is a comparable type. This method also accepts the
	// vertex properties such as weight.
	//
	// If there is a vertex with the same label in the graph, returns nil.
	// Otherwise, returns the created vertex.
	AddVertexByLabel(label T, options ...VertexOptionFunc) *Vertex[T]

	// AddVertex adds the input vertex to the graph. It doesn't add
	// vertex to the graph if the input vertex label is already exists
	// in the graph.
	AddVertex(v *Vertex[T])

	// GetVertexByID returns the vertex with the input label.
	//
	// If vertex doesn't exist, returns nil.
	GetVertexByID(label T) *Vertex[T]

	// GetAllVerticesByID returns a slice of vertices with the specified label list.
	//
	// If vertex doesn't exist, doesn't add nil to the output list.
	GetAllVerticesByID(label ...T) []*Vertex[T]

	// GetAllVertices returns a slice of all existing vertices in the graph.
	GetAllVertices() []*Vertex[T]

	// RemoveVertices removes all the specified vertices from this graph including
	// all its touching edges if present.
	RemoveVertices(vertices ...*Vertex[T])

	// ContainsEdge returns 'true' if and only if this graph contains an edge
	// going from the source vertex to the target vertex.
	//
	// If any of the specified vertices does not exist in the graph, or if is nil,
	// returns 'false'.
	ContainsEdge(from, to *Vertex[T]) bool

	// ContainsVertex returns 'true' if this graph contains the specified vertex.
	//
	// If the specified vertex is nil, returns 'false'.
	ContainsVertex(v *Vertex[T]) bool

	// Order returns the number of vertices in the graph.
	Order() uint32

	// Size returns the number of edges in the graph
	Size() uint32
}

// New creates a new instance of base graph that implemented the Graph interface.
func New[T comparable](options ...GraphOptionFunc) Graph[T] {
	return newBaseGraph[T](newProperties(options...))
}

// Edge represents an edges in a graph. It contains start and end points.
type Edge[T comparable] struct {
	source     *Vertex[T] // start point of the edges
	dest       *Vertex[T] // destination or end point of the edges
	properties EdgeProperties
	metadata   any // optional metadata associated with the edge
}

func NewEdge[T comparable](source *Vertex[T], dest *Vertex[T], options ...EdgeOptionFunc) *Edge[T] {
	var properties EdgeProperties
	for _, option := range options {
		option(&properties)
	}

	return &Edge[T]{
		source:     source,
		dest:       dest,
		properties: properties,
	}
}

// Weight returns the weight of the edge.
func (e *Edge[T]) Weight() float64 {
	return e.properties.weight
}

// OtherVertex accepts the label of one the vertices of the edge
// and returns the other one. If the input label doesn't match to
// either of the vertices, returns nil.
func (e *Edge[T]) OtherVertex(v T) *Vertex[T] {
	if e.source.label == v {
		return e.dest
	}

	if e.dest.label == v {
		return e.source
	}

	return nil
}

// Source returns edge source vertex
func (e *Edge[T]) Source() *Vertex[T] {
	return e.source
}

// Destination returns edge dest vertex
func (e *Edge[T]) Destination() *Vertex[T] {
	return e.dest
}

// Metadata returns the metadata associated with the edge.
func (e Edge[T]) Metadata() any {
	return e.metadata
}

// Vertex represents a node or point in a graph
type Vertex[T comparable] struct {
	label      T            // uniquely identifies each vertex
	neighbors  []*Vertex[T] // stores pointers to its neighbors
	inDegree   int          // number of incoming edges to this vertex
	properties VertexProperties
	metadata   any // optional metadata associated with the vertex
}

func NewVertex[T comparable](label T, options ...VertexOptionFunc) *Vertex[T] {
	return &Vertex[T]{label: label}
}

// NeighborByLabel iterates over the neighbor slice and returns the
// vertex which its label is equal to the input label.
//
// It returns nil if there is no neighbor with that label.
func (v *Vertex[T]) NeighborByLabel(label T) *Vertex[T] {
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
	return v.NeighborByLabel(vertex.label) != nil
}

// InDegree returns the number of incoming edges to the current vertex.
func (v *Vertex[T]) InDegree() int {
	return v.inDegree
}

// OutDegree returns the number of outgoing edges to the current vertex.
func (v *Vertex[T]) OutDegree() int {
	return len(v.neighbors)
}

// Degree returns the total degree of the vertex which is the sum of
// in and out degrees.
func (v *Vertex[T]) Degree() int {
	return v.inDegree + v.OutDegree()
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

// Weight returns vertex label.
func (v *Vertex[T]) Weight() float64 {
	return v.properties.weight
}

// Metadata returns the metadata associated with the vertex.
func (v *Vertex[T]) Metadata() any {
	return v.metadata
}
