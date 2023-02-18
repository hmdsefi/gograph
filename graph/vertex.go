package graph

// Vertex represents a node or point in a graph
type Vertex struct {
	// ID uniquely identifies each vertex
	ID int

	// Neighbors stores pointers to its neighbors
	Neighbors []*Vertex
}
