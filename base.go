package gograph

// baseGraph represents a basic implementation of Graph interface. It
// supports multiple types of graph.
//
// This implementation is not safe for concurrent read/write from different
// goroutines. If two goroutines try to modify the same graph it raises panic.
type baseGraph[T comparable] struct {
	// vertices is a map of vertices of the graph. the key of the map
	// is the vertex label.
	vertices map[T]*Vertex[T]

	// edges represents the edge between two vertices. The key of the
	// first map is the label of source vertex and the key of the inner
	// map is the label of destination vertex.
	edges map[T]map[T]*Edge[T]

	properties GraphProperties
}

func newBaseGraph[T comparable](properties GraphProperties) *baseGraph[T] {
	return &baseGraph[T]{
		vertices:   make(map[T]*Vertex[T]),
		edges:      make(map[T]map[T]*Edge[T]),
		properties: properties,
	}
}

// addToEdgeMap creates a new edge struct and adds it to the edges map inside
// the baseGraph struct. Note that it doesn't add the neighbor to the source vertex.
//
// It returns the created edge.
func (g *baseGraph[T]) addToEdgeMap(from, to *Vertex[T], options ...EdgeOptionFunc) *Edge[T] {
	edge := NewEdge(from, to, options...)
	if _, ok := g.edges[from.label]; !ok {
		g.edges[from.label] = map[T]*Edge[T]{to.label: edge}
	} else {
		g.edges[from.label][to.label] = edge
	}

	return edge
}

// AddEdge adds and edge from the vertex with the 'from' label to
// the vertex with the 'to' label by appending the 'to' vertex to the
// 'neighbors' slice of the 'from' vertex, in directed graph.
//
// In undirected graph, it creates edges in both directions between
// the specified vertices.
//
// It creates the input vertices if they don't exist in the graph.
// If any of the specified vertices is nil, returns nil.
// If edge already exist, returns error.
func (g *baseGraph[T]) AddEdge(from, to *Vertex[T], options ...EdgeOptionFunc) (*Edge[T], error) {
	if from == nil || to == nil {
		return nil, ErrNilVertices
	}

	if g.findVertex(from.label) == nil {
		g.AddVertex(from)
	}

	if g.findVertex(to.label) == nil {
		g.AddVertex(to)
	}

	// prevent edge-multiplicity
	if g.ContainsEdge(from, to) {
		return nil, ErrEdgeAlreadyExists
	}

	from.neighbors = append(from.neighbors, to)
	to.inDegree++

	// prevent cycle creation, if graph is acyclic
	if g.properties.isAcyclic {
		// If topological sort returns an error, new edges created a cycle
		_, err := TopologySort[T](g)
		if err != nil {
			// Remove the new edges
			from.neighbors = from.neighbors[:len(from.neighbors)-1]
			to.inDegree--

			return nil, ErrDAGCycle
		}
	}

	// add "from" to the "to" vertex neighbor slice, if graph is undirected.
	if !g.properties.isDirected {
		to.neighbors = append(to.neighbors, from)
		from.inDegree++

		g.addToEdgeMap(to, from, options...)
	}

	return g.addToEdgeMap(from, to, options...), nil
}

// AddVertexByLabel adds a new vertex with the given label to the graph.
// Label of the vertex is a comparable type. This method also accepts the
// vertex properties such as weight.
//
// If there is a vertex with the same label in the graph, returns nil.
// Otherwise, returns the created vertex.
func (g *baseGraph[T]) AddVertexByLabel(label T, options ...VertexOptionFunc) *Vertex[T] {
	var properties VertexProperties
	for _, option := range options {
		option(&properties)
	}

	v := g.addVertex(&Vertex[T]{label: label, properties: properties})

	return v
}

// AddVertex adds the input vertex to the graph. It doesn't add
// vertex to the graph if the input vertex label is already exists
// in the graph.
func (g *baseGraph[T]) AddVertex(v *Vertex[T]) {
	if v == nil {
		return
	}

	g.addVertex(v)
}

func (g *baseGraph[T]) addVertex(v *Vertex[T]) *Vertex[T] {
	if _, ok := g.vertices[v.label]; ok {
		return nil
	}

	g.vertices[v.label] = v
	return v
}

func (g *baseGraph[T]) findVertex(label T) *Vertex[T] {
	return g.vertices[label]
}

// GetAllEdges returns a slice of all edges connecting source vertex to
// target vertex if such vertices exist in this graph.
//
// In directed graph, it returns a single edge.
//
// If any of the specified vertices is nil, returns nil.
// If any of the vertices does not exist, returns nil.
// If both vertices exist but no edges found, returns an empty set.
func (g *baseGraph[T]) GetAllEdges(from, to *Vertex[T]) []*Edge[T] {
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

	if destMap, ok := g.edges[from.label]; ok {
		if edge, ok := destMap[to.label]; ok {
			edges = append(edges, edge)
		}
	}

	if !g.IsDirected() {
		if destMap, ok := g.edges[to.label]; ok {
			if edge, ok := destMap[from.label]; ok {
				edges = append(edges, edge)
			}
		}
	}

	return edges
}

// GetEdge returns an edge connecting source vertex to target vertex
// if such vertices and such edge exist in this graph.
//
// In undirected graph, returns only the edge from the "from" vertex to
// the "to" vertex.
//
// If any of the specified vertices is nil, returns nil.
// If edge does not exist, returns nil.
func (g *baseGraph[T]) GetEdge(from, to *Vertex[T]) *Edge[T] {
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
// If no edges are touching the specified vertex returns an empty slice.
//
// If the input vertex is nil, returns nil.
// If the input vertex does not exist, returns nil.
func (g *baseGraph[T]) EdgesOf(v *Vertex[T]) []*Edge[T] {
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

// RemoveEdges removes input edges from the graph from the specified
// slice of edges, if they exist.
func (g *baseGraph[T]) RemoveEdges(edges ...*Edge[T]) {
	for i := range edges {
		g.removeAllEdges(edges[i])
	}
}

// removeAllEdges removes edges in both directions between the
// source and dest vertices in the specified edge, if the graph
// is undirected. Otherwise, removes the edge from the source to
// the dest only.
func (g *baseGraph[T]) removeAllEdges(edge *Edge[T]) {
	if edge == nil {
		return
	}

	if edge.source == nil || g.findVertex(edge.source.label) == nil {
		return
	}

	if edge.dest == nil || g.findVertex(edge.dest.label) == nil {
		return
	}

	g.removeEdge(edge)

	if !g.IsDirected() {
		g.removeEdge(NewEdge(edge.dest, edge.source))
	}
}

// removeEdge removes the edge from edges destination map, if size of
// the internal map is zero, removes the source label from the edges.
func (g *baseGraph[T]) removeEdge(edge *Edge[T]) {
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

func (g *baseGraph[T]) removeNeighbor(sourceID, neighborLbl T) {
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
func (g *baseGraph[T]) GetVertexByID(label T) *Vertex[T] {
	return g.findVertex(label)
}

// GetAllVerticesByID returns a slice of vertices with the specified label list.
//
// If vertex doesn't exist, doesn't add nil to the output list.
func (g *baseGraph[T]) GetAllVerticesByID(idList ...T) []*Vertex[T] {
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
func (g *baseGraph[T]) GetAllVertices() []*Vertex[T] {
	var vertices []*Vertex[T]
	for _, vertex := range g.vertices {
		vertices = append(vertices, vertex)
	}

	return vertices
}

// RemoveVertices removes all the specified vertices from this graph including
// all its touching edges if present.
func (g *baseGraph[T]) RemoveVertices(vertices ...*Vertex[T]) {
	for i := range vertices {
		g.removeVertex(vertices[i])
	}
}

func (g *baseGraph[T]) removeVertex(in *Vertex[T]) {
	if in == nil {
		return
	}

	v := g.findVertex(in.label)
	if v == nil {
		return
	}

	if g.IsDirected() {
		for i := range v.neighbors {
			v.neighbors[i].inDegree--
		}
	}

	for sourceID := range g.edges {
		if edge, ok := g.edges[sourceID][v.label]; ok {
			g.removeAllEdges(edge)
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
func (g *baseGraph[T]) ContainsEdge(from, to *Vertex[T]) bool {
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
func (g *baseGraph[T]) ContainsVertex(v *Vertex[T]) bool {
	if v == nil {
		return false
	}

	return g.findVertex(v.label) != nil
}

func (g *baseGraph[T]) AllEdges() []*Edge[T] {
	var out []*Edge[T]
	for _, dest := range g.edges {
		for _, edge := range dest {
			out = append(out, edge)
		}
	}

	return out
}
