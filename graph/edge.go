package graph

// Edge represents an edges in a graph. It contains start and end points.
type Edge[T comparable] struct {
	source *Vertex[T] // start point of the edges
	dest   *Vertex[T] // destination or end point of the edges
}

func NewEdge[T comparable](source *Vertex[T], dest *Vertex[T]) *Edge[T] {
	return &Edge[T]{source: source, dest: dest}
}
