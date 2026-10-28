package gograph

import (
	"errors"
	"iter"
)

var (
	ErrNilVertices        = errors.New("vertices are nil")
	ErrVertexDoesNotExist = errors.New("vertex does not exist")
	ErrEdgeAlreadyExists  = errors.New("edge already exists")
	ErrDAGCycle           = errors.New("edges would create cycle")
	ErrDAGHasCycle        = errors.New("the graph contains a cycle")
	ErrNotDirected        = errors.New("graph is not directed")
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
	// vertices. A self-loop is stored once.
	//
	// This method accepts additional edge options such as weight and adds
	// them to the new edge.
	//
	// In a graph created with Acyclic, an edge that would close a cycle
	// returns ErrDAGCycle without adding the edge. For a graph from New, the
	// check only visits the vertices reachable from the 'to' vertex. It takes
	// O(1) time when the 'to' vertex has no outgoing edges or the 'from'
	// vertex has no incoming edges, and O(V+E) in the worst case.
	//
	// It creates the input vertices if they don't exist in the graph, the
	// same way AddVertex does, so a vertex from another graph is copied.
	// If any of the specified vertices is nil, returns nil.
	// If edge already exist, returns error.
	AddEdge(from, to *Vertex[T], options ...EdgeOptionFunc) (*Edge[T], error)

	// GetAllEdges returns a slice of all edges connecting source vertex to
	// target vertex if such vertices exist in this graph.
	//
	// In directed graph, or if both vertices are the same, it returns a single edge.
	//
	// If any of the specified vertices is nil, returns nil.
	// If any of the vertices does not exist, returns nil.
	// If both vertices exist but no edges found, returns nil.
	GetAllEdges(from, to *Vertex[T]) []*Edge[T]

	// AllEdges returns all the edges in the graph. The edges are grouped by
	// source vertex in the order of GetAllVertices, and the edges of each
	// source vertex are in the order they were added. In an undirected
	// graph, an edge between distinct vertices is stored in both directions,
	// so it appears in the group of both its vertices. A self-loop is stored once.
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
	// If no edges are touching the specified vertex returns nil.
	//
	// An undirected edge between distinct vertices appears in both directions.
	// Use v.OutDegree() for the number of neighbors in an undirected graph.
	//
	// The edges that start from the vertex come first, in the order they
	// were added. The edges that end at the vertex follow, in the order of
	// their source vertices in GetAllVertices. For a graph from New, it takes
	// O(d log d) time for a vertex with d edges.
	//
	// If the input vertex is nil, returns nil.
	// If the input vertex does not exist, returns nil.
	EdgesOf(v *Vertex[T]) []*Edge[T]

	// RemoveEdges removes input edges from the graph from the specified
	// slice of edges, if they exist. In undirected graph, removes edges
	// in both directions. For a graph from New, removing an edge takes time
	// proportional to the degrees of its two vertices.
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
	//
	// A vertex belongs to one graph. If the input vertex has already been
	// added to a graph, including this one before it was removed, the graph
	// stores a copy with the same label, weight and metadata, and no edges.
	AddVertex(v *Vertex[T])

	// GetVertexByID returns the vertex with the input label.
	//
	// If vertex doesn't exist, returns nil.
	GetVertexByID(label T) *Vertex[T]

	// GetAllVerticesByID returns a slice of vertices with the specified label list.
	//
	// If vertex doesn't exist, doesn't add nil to the output list.
	GetAllVerticesByID(label ...T) []*Vertex[T]

	// GetAllVertices returns a slice of all existing vertices in the graph,
	// in the order they were added. A vertex that is removed and added
	// again moves to the end.
	GetAllVertices() []*Vertex[T]

	// RemoveVertices removes all the specified vertices from this graph including
	// all its touching edges if present. A removed vertex reports no neighbors
	// and an in-degree of 0. A vertex of another graph with the same label as a
	// removed one keeps its edges. For a graph from New, removing a vertex
	// takes time proportional to the degrees of its neighbors.
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

	// Size returns the number of stored edges in the graph. An undirected edge
	// between distinct vertices is stored in both directions and counts twice.
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
	added      int // 1-based order among edges added to a graph, or 0
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

// InsertionIndex reports the order in which this edge was added to its
// graph. Edges are numbered from 1 in the order they are stored. On an
// undirected graph each AddEdge stores the opposite direction first, so
// the edge AddEdge returns is not always number 1. An edge created with
// NewEdge and not added to a graph returns 0.
func (e *Edge[T]) InsertionIndex() int {
	return e.added
}

// Label returns the edge label. It returns an empty string when no label was
// set.
func (e *Edge[T]) Label() string {
	return e.properties.label
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
	incoming   []*Edge[T]   // stored edges that end at this vertex, in no particular order
	properties VertexProperties
	metadata   any    // optional metadata associated with the vertex
	stored     bool   // whether a graph has stored this vertex
	seen       uint32 // the last search of its graph that reached this vertex
	position   int    // index of the vertex in the insertion order of its graph
}

// NewVertex creates a vertex with the given label and applies the options,
// such as WithVertexWeight, to it. The vertex doesn't belong to a graph
// until it's added with AddVertex or AddEdge.
func NewVertex[T comparable](label T, options ...VertexOptionFunc) *Vertex[T] {
	var properties VertexProperties
	for _, option := range options {
		option(&properties)
	}

	return &Vertex[T]{label: label, properties: properties}
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
// in the neighbors. Otherwise, or if the input is nil, returns 'false'.
func (v *Vertex[T]) HasNeighbor(vertex *Vertex[T]) bool {
	return vertex != nil && v.NeighborByLabel(vertex.label) != nil
}

// InDegree returns the number of incoming edges to the current vertex.
func (v *Vertex[T]) InDegree() int {
	return len(v.incoming)
}

// OutDegree returns the number of outgoing edges to the current vertex.
func (v *Vertex[T]) OutDegree() int {
	return len(v.neighbors)
}

// Degree returns the total degree of the vertex which is the sum of
// in and out degrees. In an undirected graph this counts incoming and outgoing
// edges separately; use OutDegree for the number of neighbors.
func (v *Vertex[T]) Degree() int {
	return v.InDegree() + v.OutDegree()
}

// Neighbors returns a copy of neighbor slice. If the caller changed the
// result slice, it won't impact the graph or the vertex. The returned
// vertices are snapshots that share their own neighbor lists with the graph,
// so their Neighbors, OutDegree and HasNeighbor aren't reliable once the
// graph changes. Use GetVertexByID to read the current neighbors of a vertex,
// and AllNeighbors to read them without copies.
func (v *Vertex[T]) Neighbors() []*Vertex[T] {
	var neighbors []*Vertex[T]
	for i := range v.neighbors {
		clone := &Vertex[T]{}
		*clone = *v.neighbors[i]
		neighbors = append(neighbors, clone)
	}

	return neighbors
}

// AllNeighbors returns an iterator over the vertices this vertex has an edge
// to, in the order of Neighbors. For a vertex of a graph, it yields the
// graph's own vertices without copying them, so each one is the vertex
// GetVertexByID returns for its label. In an undirected graph it yields each
// neighbor once, and a vertex with a self-loop yields itself once. A removed
// vertex yields nothing.
//
// Ranging over all neighbors takes O(OutDegree) time and allocates nothing.
// If the graph changes during the loop, the iterator can skip or repeat a
// neighbor.
func (v *Vertex[T]) AllNeighbors() iter.Seq[*Vertex[T]] {
	return func(yield func(*Vertex[T]) bool) {
		for i := 0; i < len(v.neighbors); i++ {
			if !yield(v.neighbors[i]) {
				return
			}
		}
	}
}

// Label returns vertex label.
func (v *Vertex[T]) Label() T {
	return v.label
}

// Weight returns the vertex weight.
func (v *Vertex[T]) Weight() float64 {
	return v.properties.weight
}

// Metadata returns the metadata associated with the vertex.
func (v *Vertex[T]) Metadata() any {
	return v.metadata
}
