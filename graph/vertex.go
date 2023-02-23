package graph

// Vertex represents a node or point in a graph
type Vertex[T comparable] struct {
	label     T            // uniquely identifies each vertex
	neighbors []*Vertex[T] //stores pointers to its neighbors
	inDegree  int          // number of incoming edges to this vertex
	outDegree int          // number of outgoing edges from this vertex
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
