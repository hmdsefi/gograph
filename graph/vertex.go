package graph

// Vertex represents a node or point in a graph
type Vertex struct {
	id        int       // uniquely identifies each vertex
	neighbors []*Vertex //stores pointers to its neighbors
	inDegree  int       // number of incoming edges to this vertex
	outDegree int       // number of outgoing edges from this vertex
}

func NewVertex(id int) *Vertex {
	return &Vertex{id: id}
}

// NeighborByID iterates over the neighbor slice and returns the
// vertex which its id is equal to the input id.
//
// It returns nil if there is no neighbor with that id.
func (v *Vertex) NeighborByID(id int) *Vertex {
	for i := range v.neighbors {
		if v.neighbors[i].id == id {
			return v.neighbors[i]
		}
	}

	return nil
}

// HasNeighbor checks if the input vertex is the neighbor of the
// current node or not. It returns 'true' if it finds the input
// in the neighbors. Otherwise, returns 'false'.
func (v *Vertex) HasNeighbor(vertex *Vertex) bool {
	return v.NeighborByID(vertex.id) != nil
}

// InDegree returns the number of incoming edges to the current vertex.
func (v *Vertex) InDegree() int {
	return v.inDegree
}

// OutDegree returns the number of outgoing edges to the current vertex.
func (v *Vertex) OutDegree() int {
	return v.inDegree
}
