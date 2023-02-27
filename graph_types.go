package gograph

// GraphType defines methods to determine the type of graph.
// A graph can have multiple types. e.g., a directed graph
// can be a weighted or acyclic.
type GraphType interface {
	// IsDirected returns true if the graph is directed, false otherwise.
	IsDirected() bool

	// IsAcyclic returns true if the graph is acyclic, false otherwise.
	IsAcyclic() bool

	// IsWeighted returns true if the graph is weighted, false otherwise.
	IsWeighted() bool
}

// IsDirected returns true if the graph is directed, false otherwise.
func (g *baseGraph[T]) IsDirected() bool {
	return g.properties.isDirected
}

// IsAcyclic returns true if the graph is acyclic, false otherwise.
func (g *baseGraph[T]) IsAcyclic() bool {
	return g.properties.isAcyclic
}

// IsWeighted returns true if the graph is weighted, false otherwise.
func (g *baseGraph[T]) IsWeighted() bool {
	return g.properties.isWeighted
}
