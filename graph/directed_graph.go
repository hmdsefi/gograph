package graph

import "errors"

var (
	ErrNilVertices = errors.New("vertices are nil")
)

// directedGraph represents a directed graph. It stores a slice of
// pointers to all vertices.
type directedGraph struct {
	// vertices is a map of vertices of the graph. the key of the map
	// is the vertex id.
	vertices map[int]*Vertex

	// edges represents the edge between two vertices. The key of the
	// first map is the id of source vertex and the key of the inner
	// map is the id of destination vertex.
	edges map[int]map[int]*Edge
}

func NewDirectedGraph() Graph {
	return newDirectedGraph()
}

func newDirectedGraph() *directedGraph {
	return &directedGraph{
		vertices: make(map[int]*Vertex),
		edges:    make(map[int]map[int]*Edge),
	}
}

// AddVertexWithID adds a new vertex with the given id to the graph.
// If there is a vertex with the same id in the graph, returns nil.
// Otherwise, returns the created vertex.
func (g *directedGraph) AddVertexWithID(id int) *Vertex {
	v := g.addVertex(&Vertex{id: id})

	return v
}

// AddVertex adds the input vertex to the graph. It doesn't add
// vertex to the graph if the input vertex id is already exists
// in the graph.
func (g *directedGraph) AddVertex(v *Vertex) {
	if v == nil {
		return
	}

	g.addVertex(v)
}

func (g *directedGraph) addVertex(v *Vertex) *Vertex {
	if _, ok := g.vertices[v.id]; ok {
		return nil
	}

	g.vertices[v.id] = v
	return v
}

// AddEdge adds a directed edges from the vertex with the 'from' id to
// the vertex with the 'to' id by appending the 'to' vertex to the
// 'neighbors' slice of the 'from' vertex.
//
// It creates the input vertices if they don't exist in the graph.
// If any of the specified vertices is nil, returns nil.
func (g *directedGraph) AddEdge(from, to *Vertex) (*Edge, error) {
	if from == nil || to == nil {
		return nil, ErrNilVertices
	}

	if g.findVertex(from.id) == nil {
		g.AddVertex(from)
	}

	if g.findVertex(to.id) == nil {
		g.AddVertex(to)
	}

	edge := NewEdge(from, to)

	from.neighbors = append(from.neighbors, to)
	from.outDegree++
	to.inDegree++

	if _, ok := g.edges[from.id]; !ok {
		g.edges[from.id] = map[int]*Edge{to.id: edge}
	} else {
		g.edges[from.id][to.id] = edge
	}

	return edge, nil
}

func (g *directedGraph) findVertex(id int) *Vertex {
	return g.vertices[id]
}

// GetAllEdges returns a slice of all edges connecting source vertex to
// target vertex if such vertices exist in this graph.
//
// If any of the specified vertices is nil, returns nil.
// If any of the vertices does not exist, returns nil.
// If both vertices exist but no edges found, returns an empty set.
func (g *directedGraph) GetAllEdges(from, to *Vertex) []*Edge {
	if from == nil || to == nil {
		return nil
	}

	if g.findVertex(from.id) == nil {
		return nil
	}

	if g.findVertex(to.id) == nil {
		return nil
	}

	var edges []*Edge
	destMap, ok := g.edges[from.id]
	if !ok {
		return edges
	}

	for destID := range destMap {
		edges = append(edges, destMap[destID])
	}

	return edges
}

// GetEdge returns an edge connecting source vertex to target vertex
// if such vertices and such edge exist in this graph.
//
// If any of the specified vertices is nil, returns nil.
// If edge does not exist, returns nil.
func (g *directedGraph) GetEdge(from, to *Vertex) *Edge {
	if from == nil || to == nil {
		return nil
	}

	if g.findVertex(from.id) == nil {
		return nil
	}

	if g.findVertex(to.id) == nil {
		return nil
	}

	if destMap, ok := g.edges[from.id]; ok {
		return destMap[to.id]
	}

	return nil
}

// EdgesOf returns a slice of all edges touching the specified vertex.
// If no edges are touching the specified vertex returns an empty set.
//
// If the input vertex is nil, returns nil.
// If the input vertex does not exist, returns nil.
func (g *directedGraph) EdgesOf(v *Vertex) []*Edge {
	if v == nil {
		return nil
	}

	if g.findVertex(v.id) == nil {
		return nil
	}

	var edges []*Edge

	// find all the edges that start from the input vertex
	if destMap, ok := g.edges[v.id]; ok {
		for destID := range destMap {
			edges = append(edges, destMap[destID])
		}
	}

	// find all the edges that the input vertex is the
	// destination of the edge
	for sourceID, destMap := range g.edges {
		if sourceID == v.id {
			continue
		}

		for destID := range destMap {
			if destID == v.id {
				edges = append(edges, destMap[destID])
			}
		}
	}

	return edges
}

// RemoveEdges removes input edges from the graph if they exist.
func (g *directedGraph) RemoveEdges(edges ...*Edge) {
	for i := range edges {
		g.removeEdge(edges[i])
	}
}

// removeEdge removes the edge from edges destination map, if size of
// the internal map is zero, removes the source id from the edges.
func (g *directedGraph) removeEdge(edge *Edge) {
	if edge == nil {
		return
	}

	if g.findVertex(edge.source.id) == nil {
		return
	}

	if g.findVertex(edge.dest.id) == nil {
		return
	}

	if destMap, ok := g.edges[edge.source.id]; ok {
		delete(destMap, edge.dest.id)

		// remove the neighbor vertex from the source neighbors slice.
		g.removeNeighbor(edge.source.id, edge.dest.id)

		// remove the source vertex id from the edge map, if it
		// doesn't have any edges.
		if len(destMap) == 0 {
			delete(g.edges, edge.source.id)
		}
	}
}

func (g *directedGraph) removeNeighbor(sourceID, neighborID int) {
	source := g.findVertex(sourceID)
	for i := range source.neighbors {
		if source.neighbors[i].id == neighborID {
			source.neighbors[i].inDegree--
			source.outDegree--

			if i == 0 {
				source.neighbors = source.neighbors[1:]
			} else if i == len(source.neighbors)-1 {
				source.neighbors = source.neighbors[:len(source.neighbors)-1]
			} else {
				source.neighbors = append(source.neighbors[:i], source.neighbors[i+1:]...)
			}

			break
		}
	}
}

// GetVertexByID returns the vertex with the input id.
//
// If vertex doesn't exist, returns nil.
func (g *directedGraph) GetVertexByID(id int) *Vertex {
	return g.findVertex(id)
}

// GetAllVerticesByID returns a slice of vertices with the input id list.
//
// If vertex doesn't exist, doesn't add nil to the output list.
func (g *directedGraph) GetAllVerticesByID(idList ...int) []*Vertex {
	var vertices []*Vertex
	for _, id := range idList {
		v := g.GetVertexByID(id)
		if v != nil {
			vertices = append(vertices, v)
		}
	}

	return vertices
}

// GetAllVertices returns a slice of all existing vertices in the graph.
func (g *directedGraph) GetAllVertices() []*Vertex {
	var vertices []*Vertex
	for _, vertex := range g.vertices {
		vertices = append(vertices, vertex)
	}

	return vertices
}

// RemoveVertices removes all the specified vertices from this graph including
// all its touching edges if present.
func (g *directedGraph) RemoveVertices(vertices ...*Vertex) {
	for i := range vertices {
		g.removeVertex(vertices[i])
	}
}

func (g *directedGraph) removeVertex(in *Vertex) {
	if in == nil {
		return
	}

	v := g.findVertex(in.id)
	if v == nil {
		return
	}

	for i := range v.neighbors {
		v.neighbors[i].inDegree--
	}

	for sourceID := range g.edges {
		if edge, ok := g.edges[sourceID][v.id]; ok {
			g.removeEdge(edge)
			delete(g.edges[sourceID], v.id)
		}
	}

	delete(g.edges, v.id)
	delete(g.vertices, v.id)
}

// ContainsEdge returns 'true' if and only if this graph contains an edge
// going from the source vertex to the target vertex.
//
// If any of the specified vertices does not exist in the graph, or if is nil,
// returns 'false'.
func (g *directedGraph) ContainsEdge(from, to *Vertex) bool {
	if from == nil || to == nil {
		return false
	}

	if g.findVertex(from.id) == nil {
		return false
	}

	if g.findVertex(to.id) == nil {
		return false
	}

	if destMap, ok := g.edges[from.id]; ok {
		if _, ok = destMap[to.id]; ok {
			return true
		}
	}

	return false
}

// ContainsVertex returns 'true' if this graph contains the specified vertex.
//
// If the specified vertex is nil, returns 'false'.
func (g *directedGraph) ContainsVertex(v *Vertex) bool {
	if v == nil {
		return false
	}

	return g.findVertex(v.id) != nil
}
