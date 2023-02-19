package graph

// Edge represents an edges in a graph. It contains start and end points.
type Edge struct {
	source *Vertex // start point of the edges
	dest   *Vertex // destination or end point of the edges
}

func NewEdge(source *Vertex, dest *Vertex) *Edge {
	return &Edge{source: source, dest: dest}
}
