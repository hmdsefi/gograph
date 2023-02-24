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

	BaseGraph[T]
}

type BaseGraph[T comparable] interface {
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
