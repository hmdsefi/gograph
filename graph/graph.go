package graph

type Graph interface {
	AddEdge(from, to *Vertex) (*Edge, error)

	baseGraph
}

type baseGraph interface {
	GetAllEdges(from, to *Vertex) []*Edge
	GetEdge(from, to *Vertex) *Edge
	EdgesOf(v *Vertex) []*Edge
	RemoveEdges(edges ...*Edge)

	AddVertexWithID(id int) *Vertex
	AddVertex(v *Vertex)
	GetVertexByID(id int) *Vertex
	GetAllVerticesByID(id ...int) []*Vertex
	GetAllVertices() []*Vertex
	RemoveVertices(vertices ...*Vertex)

	ContainsEdge(from, to *Vertex) bool
	ContainsVertex(v *Vertex) bool
}
