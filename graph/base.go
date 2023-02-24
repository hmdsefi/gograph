package graph

// base represents a basic graph. It stores a slice of
// pointers to all vertices.
type base[T comparable] struct {
	// vertices is a map of vertices of the graph. the key of the map
	// is the vertex label.
	vertices map[T]*Vertex[T]

	// edges represents the edge between two vertices. The key of the
	// first map is the label of source vertex and the key of the inner
	// map is the label of destination vertex.
	edges map[T]map[T]*Edge[T]
}

func newBaseGraph[T comparable]() *base[T] {
	return &base[T]{
		vertices: make(map[T]*Vertex[T]),
		edges:    make(map[T]map[T]*Edge[T]),
	}
}

// addToEdgeMap creates a new edge struct and adds it to the edges map inside
// the base struct. Note that it doesn't add the neighbor to the source vertex.
//
// It returns the created edge.
func (g *base[T]) addToEdgeMap(from, to *Vertex[T]) *Edge[T] {
	edge := NewEdge(from, to)
	if _, ok := g.edges[from.label]; !ok {
		g.edges[from.label] = map[T]*Edge[T]{to.label: edge}
	} else {
		g.edges[from.label][to.label] = edge
	}

	return edge
}

// AddVertexByLabel adds a new vertex with the given label to the graph.
// If there is a vertex with the same label in the graph, returns nil.
// Otherwise, returns the created vertex.
func (g *base[T]) AddVertexByLabel(label T) *Vertex[T] {
	v := g.addVertex(&Vertex[T]{label: label})

	return v
}

// AddVertex adds the input vertex to the graph. It doesn't add
// vertex to the graph if the input vertex label is already exists
// in the graph.
func (g *base[T]) AddVertex(v *Vertex[T]) {
	if v == nil {
		return
	}

	g.addVertex(v)
}

func (g *base[T]) addVertex(v *Vertex[T]) *Vertex[T] {
	if _, ok := g.vertices[v.label]; ok {
		return nil
	}

	g.vertices[v.label] = v
	return v
}

func (g *base[T]) findVertex(label T) *Vertex[T] {
	return g.vertices[label]
}

// GetAllEdges returns a slice of all edges connecting source vertex to
// target vertex if such vertices exist in this graph.
//
// If any of the specified vertices is nil, returns nil.
// If any of the vertices does not exist, returns nil.
// If both vertices exist but no edges found, returns an empty set.
func (g *base[T]) GetAllEdges(from, to *Vertex[T]) []*Edge[T] {
	if from == nil || to == nil {
		return nil
	}

	if g.findVertex(from.label) == nil {
		return nil
	}

	if g.findVertex(to.label) == nil {
		return nil
	}

	var edges []*Edge[T]
	destMap, ok := g.edges[from.label]
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
func (g *base[T]) GetEdge(from, to *Vertex[T]) *Edge[T] {
	if from == nil || to == nil {
		return nil
	}

	if g.findVertex(from.label) == nil {
		return nil
	}

	if g.findVertex(to.label) == nil {
		return nil
	}

	if destMap, ok := g.edges[from.label]; ok {
		return destMap[to.label]
	}

	return nil
}

// EdgesOf returns a slice of all edges touching the specified vertex.
// If no edges are touching the specified vertex returns an empty set.
//
// If the input vertex is nil, returns nil.
// If the input vertex does not exist, returns nil.
func (g *base[T]) EdgesOf(v *Vertex[T]) []*Edge[T] {
	if v == nil {
		return nil
	}

	if g.findVertex(v.label) == nil {
		return nil
	}

	var edges []*Edge[T]

	// find all the edges that start from the input vertex
	if destMap, ok := g.edges[v.label]; ok {
		for destID := range destMap {
			edges = append(edges, destMap[destID])
		}
	}

	// find all the edges that the input vertex is the
	// destination of the edge
	for sourceID, destMap := range g.edges {
		if sourceID == v.label {
			continue
		}

		for destID := range destMap {
			if destID == v.label {
				edges = append(edges, destMap[destID])
			}
		}
	}

	return edges
}

// RemoveEdges removes input edges from the graph if they exist.
func (g *base[T]) RemoveEdges(edges ...*Edge[T]) {
	for i := range edges {
		g.removeEdge(edges[i])
	}
}

// removeEdge removes the edge from edges destination map, if size of
// the internal map is zero, removes the source label from the edges.
func (g *base[T]) removeEdge(edge *Edge[T]) {
	if edge == nil {
		return
	}

	if g.findVertex(edge.source.label) == nil {
		return
	}

	if g.findVertex(edge.dest.label) == nil {
		return
	}

	if destMap, ok := g.edges[edge.source.label]; ok {
		delete(destMap, edge.dest.label)

		// remove the neighbor vertex from the source neighbors slice.
		g.removeNeighbor(edge.source.label, edge.dest.label)

		// remove the source vertex label from the edge map, if it
		// doesn't have any edges.
		if len(destMap) == 0 {
			delete(g.edges, edge.source.label)
		}
	}
}

func (g *base[T]) removeNeighbor(sourceID, neighborLbl T) {
	source := g.findVertex(sourceID)
	for i := range source.neighbors {
		if source.neighbors[i].label == neighborLbl {
			source.neighbors[i].inDegree--

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

// GetVertexByID returns the vertex with the input label.
//
// If vertex doesn't exist, returns nil.
func (g *base[T]) GetVertexByID(label T) *Vertex[T] {
	return g.findVertex(label)
}

// GetAllVerticesByID returns a slice of vertices with the input label list.
//
// If vertex doesn't exist, doesn't add nil to the output list.
func (g *base[T]) GetAllVerticesByID(idList ...T) []*Vertex[T] {
	var vertices []*Vertex[T]
	for _, label := range idList {
		v := g.GetVertexByID(label)
		if v != nil {
			vertices = append(vertices, v)
		}
	}

	return vertices
}

// GetAllVertices returns a slice of all existing vertices in the graph.
func (g *base[T]) GetAllVertices() []*Vertex[T] {
	var vertices []*Vertex[T]
	for _, vertex := range g.vertices {
		vertices = append(vertices, vertex)
	}

	return vertices
}

// RemoveVertices removes all the specified vertices from this graph including
// all its touching edges if present.
func (g *base[T]) RemoveVertices(vertices ...*Vertex[T]) {
	for i := range vertices {
		g.removeVertex(vertices[i])
	}
}

func (g *base[T]) removeVertex(in *Vertex[T]) {
	if in == nil {
		return
	}

	v := g.findVertex(in.label)
	if v == nil {
		return
	}

	for i := range v.neighbors {
		v.neighbors[i].inDegree--
	}

	for sourceID := range g.edges {
		if edge, ok := g.edges[sourceID][v.label]; ok {
			g.removeEdge(edge)
			delete(g.edges[sourceID], v.label)
		}
	}

	delete(g.edges, v.label)
	delete(g.vertices, v.label)
}

// ContainsEdge returns 'true' if and only if this graph contains an edge
// going from the source vertex to the target vertex.
//
// If any of the specified vertices does not exist in the graph, or if is nil,
// returns 'false'.
func (g *base[T]) ContainsEdge(from, to *Vertex[T]) bool {
	if from == nil || to == nil {
		return false
	}

	if g.findVertex(from.label) == nil {
		return false
	}

	if g.findVertex(to.label) == nil {
		return false
	}

	if destMap, ok := g.edges[from.label]; ok {
		if _, ok = destMap[to.label]; ok {
			return true
		}
	}

	return false
}

// ContainsVertex returns 'true' if this graph contains the specified vertex.
//
// If the specified vertex is nil, returns 'false'.
func (g *base[T]) ContainsVertex(v *Vertex[T]) bool {
	if v == nil {
		return false
	}

	return g.findVertex(v.label) != nil
}
