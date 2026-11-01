package gograph

import (
	"cmp"
	"math"
	"slices"
	"sync/atomic"
)

// baseGraph represents a basic implementation of Graph interface. It
// supports multiple types of graph.
//
// This implementation is not safe for concurrent read/write from different
// goroutines. Use a lock around every call when any goroutine modifies the
// graph; otherwise the results are undefined and the program may crash.
type baseGraph[T comparable] struct {
	// vertices is a map of vertices of the graph. the key of the map
	// is the vertex label.
	vertices map[T]*Vertex[T]

	// edges represents the edge between two vertices. The key of the
	// first map is the label of source vertex and the key of the inner
	// map is the label of destination vertex.
	edges map[T]map[T]*Edge[T]

	// order keeps the vertices in the order they were added. A removed
	// vertex leaves a nil slot, and removedCount counts those slots
	// until removeFromOrder drops them.
	order        []*Vertex[T]
	removedCount int

	properties GraphProperties

	verticesCount uint32
	edgesCount    uint32
	nextEdge      int

	// search numbers the searches that mark vertices, and stack, behind and
	// places are their reused work lists.
	search uint32
	stack  []*Vertex[T]
	behind []*Vertex[T]
	places []int

	// ord is the topological order of an acyclic graph, indexed by vertex
	// position: every edge u -> v has ord[u.position] < ord[v.position].
	// It is in use while it has one entry per slot of order, and is built
	// again otherwise. first and last are at or beyond the lowest and the
	// highest place in use, and the order is built again before either one
	// could overflow.
	ord         []int
	first, last int
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
	g.nextEdge++
	edge.added = g.nextEdge
	dests := g.edges[from.label]
	if dests == nil {
		dests = make(map[T]*Edge[T])
		g.edges[from.label] = dests
	}
	dests[to.label] = edge

	atomic.AddUint32(&g.edgesCount, 1)
	return edge
}

// AddEdge adds and edge from the vertex with the 'from' label to
// the vertex with the 'to' label by appending the 'to' vertex to the
// 'neighbors' slice of the 'from' vertex, in directed graph.
//
// In undirected graph, it creates edges in both directions between
// the specified vertices. A self-loop is stored once.
//
// It creates the input vertices if they don't exist in the graph, the
// same way AddVertex does, so a vertex from another graph is copied.
// If any of the specified vertices is nil, returns nil.
// If edge already exist, returns error.
func (g *baseGraph[T]) AddEdge(from, to *Vertex[T], options ...EdgeOptionFunc) (*Edge[T], error) {
	if from == nil || to == nil {
		return nil, ErrNilVertices
	}

	from = g.storedVertex(from)
	to = g.storedVertex(to)

	// prevent edge-multiplicity
	if _, ok := g.edges[from.label][to.label]; ok {
		return nil, ErrEdgeAlreadyExists
	}

	// in an acyclic graph, place the new edge in the topological order, or
	// reject it if it closes a cycle
	if g.properties.isAcyclic && !g.placeEdge(from, to) {
		return nil, ErrDAGCycle
	}

	// add the edge from "to" to "from" too, if graph is undirected.
	if !g.properties.isDirected && from != to {
		g.link(to, from, options...)
	}

	return g.link(from, to, options...), nil
}

// link stores the edge from "from" to "to" and returns it.
func (g *baseGraph[T]) link(from, to *Vertex[T], options ...EdgeOptionFunc) *Edge[T] {
	from.neighbors = append(from.neighbors, to)
	edge := g.addToEdgeMap(from, to, options...)
	to.incoming = append(to.incoming, edge)

	return edge
}

// AddVertexByLabel adds a new vertex with the given label to the graph.
// Label of the vertex is a comparable type. This method also accepts the
// vertex properties such as weight.
//
// If there is a vertex with the same label in the graph, returns nil.
// Otherwise, returns the created vertex.
func (g *baseGraph[T]) AddVertexByLabel(label T, options ...VertexOptionFunc) *Vertex[T] {
	v := g.addVertex(NewVertex(label, options...))

	return v
}

// AddVertex adds the input vertex to the graph. It doesn't add
// vertex to the graph if the input vertex label is already exists
// in the graph.
//
// A vertex belongs to one graph. If the input vertex has already been
// added to a graph, including this one before it was removed, the graph
// stores a copy with the same label, weight and metadata, and no edges.
func (g *baseGraph[T]) AddVertex(v *Vertex[T]) {
	if v == nil {
		return
	}

	g.addVertex(v)
}

// storedVertex returns the vertex of the graph with the label of v. If there
// isn't one, it adds v the same way AddVertex does.
func (g *baseGraph[T]) storedVertex(v *Vertex[T]) *Vertex[T] {
	if stored := g.vertices[v.label]; stored != nil {
		return stored
	}

	return g.addVertex(v)
}

func (g *baseGraph[T]) addVertex(v *Vertex[T]) *Vertex[T] {
	if _, ok := g.vertices[v.label]; ok {
		return nil
	}

	if v.stored {
		v = &Vertex[T]{label: v.label, properties: v.properties, metadata: v.metadata}
	}
	v.stored = true

	g.vertices[v.label] = v
	v.position = len(g.order)
	g.order = append(g.order, v)
	if g.properties.isAcyclic && len(g.ord) == v.position && g.last < math.MaxInt {
		g.last++
		g.ord = append(g.ord, g.last)
	}
	atomic.AddUint32(&g.verticesCount, 1)

	return v
}

// removeFromOrder clears the slot of v in the insertion order. Once more
// than half of the slots are empty, it moves the remaining vertices to the
// front, so removals stay O(1) amortized.
func (g *baseGraph[T]) removeFromOrder(v *Vertex[T]) {
	g.order[v.position] = nil
	g.removedCount++
	if g.removedCount <= len(g.order)/2 {
		return
	}

	n := 0
	ordered := len(g.ord) == len(g.order)
	for i, u := range g.order {
		if u != nil {
			u.position = n
			g.order[n] = u
			if ordered {
				g.ord[n] = g.ord[i]
			}
			n++
		}
	}
	clear(g.order[n:])
	g.order = g.order[:n]
	if ordered {
		g.ord = g.ord[:n]
	} else {
		g.ord = nil
	}
	g.removedCount = 0
}

func (g *baseGraph[T]) findVertex(label T) *Vertex[T] {
	return g.vertices[label]
}

// GetAllEdges returns a slice of all edges connecting source vertex to
// target vertex if such vertices exist in this graph.
//
// In directed graph, or if both vertices are the same, it returns a single edge.
//
// If any of the specified vertices is nil, returns nil.
// If any of the vertices does not exist, returns nil.
// If both vertices exist but no edges found, returns nil.
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

	if !g.IsDirected() && from.label != to.label {
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
// If no edges are touching the specified vertex returns nil.
//
// The edges that start from the vertex come first, in the order they were
// added. The edges that end at the vertex follow, in the order of their
// source vertices in GetAllVertices.
//
// If the input vertex is nil, returns nil.
// If the input vertex does not exist, returns nil.
func (g *baseGraph[T]) EdgesOf(v *Vertex[T]) []*Edge[T] {
	if v == nil {
		return nil
	}

	vertex := g.findVertex(v.label)
	if vertex == nil {
		return nil
	}

	n := len(vertex.neighbors) + len(vertex.incoming)
	if n == 0 {
		return nil
	}

	// find all the edges that start from the input vertex
	edges := g.outgoingEdges(vertex, make([]*Edge[T], 0, n))

	// the edges that end at the input vertex, except a self-loop, which is
	// already listed, in the order of their source vertices
	out := len(edges)
	for _, edge := range vertex.incoming {
		if edge.source != vertex {
			edges = append(edges, edge)
		}
	}
	slices.SortFunc(edges[out:], func(a, b *Edge[T]) int {
		return cmp.Compare(a.source.position, b.source.position)
	})

	return edges
}

// outgoingEdges appends the edges that start from v to out, in the order
// of its neighbors, and returns the result.
func (g *baseGraph[T]) outgoingEdges(v *Vertex[T], out []*Edge[T]) []*Edge[T] {
	destMap := g.edges[v.label]
	selfLoop := false
	for _, neighbor := range v.neighbors {
		// list a self-loop once, even if v is in its own neighbor list twice
		if neighbor.label == v.label {
			if selfLoop {
				continue
			}
			selfLoop = true
		}

		if edge, ok := destMap[neighbor.label]; ok {
			out = append(out, edge)
		}
	}

	return out
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

	if !g.IsDirected() && edge.source.label != edge.dest.label {
		g.removeEdge(NewEdge(edge.dest, edge.source))
	}
}

// removeEdge removes the edge from edges destination map, if size of
// the internal map is zero, removes the source label from the edges.
// It does nothing if the graph doesn't contain the edge.
func (g *baseGraph[T]) removeEdge(edge *Edge[T]) {
	destMap, ok := g.edges[edge.source.label]
	if !ok {
		return
	}
	stored, ok := destMap[edge.dest.label]
	if !ok {
		return
	}

	g.unlink(stored)
}

// unlink removes a stored edge from the edge map, from the neighbors of its
// source and from the incoming edges of its destination.
func (g *baseGraph[T]) unlink(edge *Edge[T]) {
	destMap := g.edges[edge.source.label]
	delete(destMap, edge.dest.label)

	// remove the source vertex label from the edge map, if it
	// doesn't have any edges.
	if len(destMap) == 0 {
		delete(g.edges, edge.source.label)
	}
	atomic.AddUint32(&g.edgesCount, ^(uint32(1) - 1))

	removeNeighbor(edge.source, edge.dest)
	removeIncoming(edge)
}

// removeNeighbor removes neighbor from the neighbors of source, keeping the
// order of the others.
func removeNeighbor[T comparable](source, neighbor *Vertex[T]) {
	source.neighbors = removeItem(source.neighbors, neighbor)
}

// removeIncoming removes edge from the incoming edges of its destination.
func removeIncoming[T comparable](edge *Edge[T]) {
	edge.dest.incoming = removeItem(edge.dest.incoming, edge)
}

// removeItem removes the first x from list, keeping the order of the others,
// and clears the freed slot. It checks both ends first, so removing the items
// in the order they were added, or in reverse, takes O(1) time per item.
func removeItem[E comparable](list []E, x E) []E {
	var zero E
	switch n := len(list); {
	case n == 0:
		return list
	case list[0] == x:
		list[0] = zero
		return list[1:]
	case list[n-1] == x:
		list[n-1] = zero
		return list[:n-1]
	}
	if i := slices.Index(list, x); i >= 0 {
		return slices.Delete(list, i, i+1)
	}

	return list
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

// GetAllVertices returns a slice of all existing vertices in the graph, in
// the order they were added. A vertex that is removed and added again
// moves to the end.
func (g *baseGraph[T]) GetAllVertices() []*Vertex[T] {
	if len(g.vertices) == 0 {
		return nil
	}

	vertices := make([]*Vertex[T], 0, len(g.vertices))
	for _, vertex := range g.order {
		if vertex != nil {
			vertices = append(vertices, vertex)
		}
	}

	return vertices
}

// RemoveVertices removes all the specified vertices from this graph including
// all its touching edges if present. A removed vertex reports no neighbors
// and an in-degree of 0.
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

	g.removeFromOrder(v)

	// the edges that end at v, except a self-loop, which starts at v too
	for _, edge := range v.incoming {
		if edge.source != v {
			destMap := g.edges[edge.source.label]
			delete(destMap, v.label)
			if len(destMap) == 0 {
				delete(g.edges, edge.source.label)
			}
			removeNeighbor(edge.source, v)
			atomic.AddUint32(&g.edgesCount, ^(uint32(1) - 1))
		}
	}

	// the edges that start from v
	for _, edge := range g.edges[v.label] {
		if edge.dest != v {
			removeIncoming(edge)
		}
		atomic.AddUint32(&g.edgesCount, ^(uint32(1) - 1))
	}
	delete(g.edges, v.label)
	delete(g.vertices, v.label)
	atomic.AddUint32(&g.verticesCount, ^(uint32(1) - 1))

	v.neighbors = nil
	v.incoming = nil
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

// AllEdges returns all the edges in the graph. The edges are grouped by
// source vertex in the order of GetAllVertices, and the edges of each
// source vertex are in the order they were added.
func (g *baseGraph[T]) AllEdges() []*Edge[T] {
	var out []*Edge[T]
	for _, vertex := range g.order {
		if vertex != nil {
			out = g.outgoingEdges(vertex, out)
		}
	}

	return out
}

// Order returns the number of vertices in the graph.
func (g *baseGraph[T]) Order() uint32 {
	return atomic.LoadUint32(&g.verticesCount)
}

// Size returns the number of edges in the graph
func (g *baseGraph[T]) Size() uint32 {
	return atomic.LoadUint32(&g.edgesCount)
}
