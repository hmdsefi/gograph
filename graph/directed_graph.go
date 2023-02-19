package graph

import "errors"

var (
	ErrNilVertices = errors.New("vertices are nil")
)

// DirectedGraph represents a directed graph. It stores a slice of
// pointers to all vertices.
type DirectedGraph struct {
	// vertices is a map of vertices of the graph. the key of the map
	// is the vertex id.
	vertices map[int]*Vertex

	// edges represents the edge between two vertices. The key of the
	// first map is the id of source vertex and the key of the inner
	// map is the id of destination vertex.
	edges map[int]map[int]*Edge
}

// AddVertexWithID adds a new vertex with the given id to the graph.
// If there is a vertex with the same id in the graph, returns nil.
// Otherwise, returns the created vertex.
func (g *DirectedGraph) AddVertexWithID(id int) *Vertex {
	v := g.addVertex(&Vertex{id: id})

	return v
}

// AddVertex adds the input vertex to the graph. It doesn't add
// vertex to the graph if the input vertex id is already exists
// in the graph.
func (g *DirectedGraph) AddVertex(v *Vertex) {
	g.addVertex(v)
}

func (g *DirectedGraph) addVertex(v *Vertex) *Vertex {
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
func (g *DirectedGraph) AddEdge(from, to *Vertex) (*Edge, error) {
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
	if _, ok := g.edges[from.id]; !ok {
		g.edges[from.id] = map[int]*Edge{to.id: edge}
	} else {
		g.edges[from.id][to.id] = edge
	}

	return edge, nil
}

func (g *DirectedGraph) findVertex(id int) *Vertex {
	return g.vertices[id]
}

// GetAllEdges returns a slice of all edges connecting source vertex to
// target vertex if such vertices exist in this graph.
//
// If any of the specified vertices is nil, returns nil.
// If any of the vertices does not exist, returns nil.
// If both vertices exist but no edges found, returns an empty set.
func (g *DirectedGraph) GetAllEdges(from, to *Vertex) []*Edge {
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
func (g *DirectedGraph) GetEdge(from, to *Vertex) *Edge {
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
func (g *DirectedGraph) EdgesOf(v *Vertex) []*Edge {
	if v == nil {
		return nil
	}

	if g.findVertex(v.id) == nil {
		return nil
	}

	var edges []*Edge

	if destMap, ok := g.edges[v.id]; ok {
		for destID := range destMap {
			edges = append(edges, destMap[destID])
		}
	}

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

func (g *DirectedGraph) RemoveEdges(edges ...*Edge) {
	//TODO implement me
	panic("implement me")
}

func (g *DirectedGraph) GetVertexByID(id int) *Vertex {
	//TODO implement me
	panic("implement me")
}

func (g *DirectedGraph) GetAllVerticesByID(id ...int) []*Vertex {
	//TODO implement me
	panic("implement me")
}

func (g *DirectedGraph) GetAllVertices() []*Vertex {
	//TODO implement me
	panic("implement me")
}

func (g *DirectedGraph) RemoveVertices(vertices ...*Vertex) {
	//TODO implement me
	panic("implement me")
}

func (g *DirectedGraph) ContainsEdge(from, to *Vertex) bool {
	//TODO implement me
	panic("implement me")
}

func (g *DirectedGraph) ContainsVertex(v *Vertex) bool {
	//TODO implement me
	panic("implement me")
}
