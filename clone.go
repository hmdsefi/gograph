package gograph

import "sync/atomic"

// Clone returns an independent copy of g, preserving its directed, weighted,
// and acyclic properties, vertex and edge weights, labels, edge insertion
// indexes, and iteration order. Vertices and edges are newly allocated.
// Metadata is copied as an interface value; reference values stored in
// metadata still refer to the same data. A nil graph panics. The graph must
// not be modified while it is being copied.
//
// The copy keeps its vertices, its edges, its neighbor lists and the lists of
// edges that end at each vertex in one array each, so a vertex or edge of the
// copy that is still referenced keeps that storage alive.
//
// Time and space are O(V + E).
func Clone[T comparable](g Graph[T]) Graph[T] {
	if g == nil {
		panic("gograph: Clone requires a non-nil graph")
	}
	if bg, ok := g.(*baseGraph[T]); ok {
		return bg.clone()
	}

	return cloneGraph(g)
}

// clone copies g by vertex position, without looking up any label.
func (g *baseGraph[T]) clone() *baseGraph[T] {
	slot := make([]int, len(g.order))
	vertexCount, neighborCount, edgeCount := 0, 0, 0
	for _, v := range g.order {
		if v == nil {
			continue
		}
		slot[v.position] = vertexCount
		vertexCount++
		neighborCount += len(v.neighbors)
		edgeCount += len(g.edges[v.label])
	}

	out := &baseGraph[T]{
		vertices:      make(map[T]*Vertex[T], vertexCount),
		edges:         make(map[T]map[T]*Edge[T], len(g.edges)),
		order:         make([]*Vertex[T], vertexCount),
		properties:    g.properties,
		verticesCount: atomic.LoadUint32(&g.verticesCount),
		edgesCount:    atomic.LoadUint32(&g.edgesCount),
		nextEdge:      g.nextEdge,
	}

	vertices := make([]Vertex[T], vertexCount)
	incoming := make([]*Edge[T], edgeCount)
	for _, v := range g.order {
		if v == nil {
			continue
		}
		i := slot[v.position]
		vertices[i] = Vertex[T]{
			label:      v.label,
			properties: v.properties,
			metadata:   v.metadata,
			stored:     true,
			position:   i,
		}
		if k := len(v.incoming); k > 0 {
			vertices[i].incoming, incoming = incoming[:0:k], incoming[k:]
		}
		out.order[i] = &vertices[i]
		out.vertices[v.label] = &vertices[i]
	}

	neighbors := make([]*Vertex[T], neighborCount)
	edges := make([]Edge[T], edgeCount)
	for _, v := range g.order {
		if v == nil {
			continue
		}
		from := &vertices[slot[v.position]]
		if k := len(v.neighbors); k > 0 {
			from.neighbors, neighbors = neighbors[:k:k], neighbors[k:]
			for i, neighbor := range v.neighbors {
				from.neighbors[i] = &vertices[slot[neighbor.position]]
			}
		}

		dests := g.edges[v.label]
		if len(dests) == 0 {
			continue
		}
		outgoing := make(map[T]*Edge[T], len(dests))
		for label, e := range dests {
			to := &vertices[slot[e.dest.position]]
			edges[0] = Edge[T]{
				source:     from,
				dest:       to,
				properties: e.properties,
				metadata:   e.metadata,
				added:      e.added,
			}
			outgoing[label] = &edges[0]
			to.incoming = append(to.incoming, &edges[0])
			edges = edges[1:]
		}
		out.edges[v.label] = outgoing
	}

	return out
}

// cloneGraph copies a Graph that New didn't build, through its methods.
func cloneGraph[T comparable](g Graph[T]) *baseGraph[T] {
	all := g.GetAllVertices()
	allEdges := g.AllEdges()
	out := &baseGraph[T]{
		vertices: make(map[T]*Vertex[T], len(all)),
		edges:    make(map[T]map[T]*Edge[T]),
		order:    make([]*Vertex[T], len(all)),
		properties: GraphProperties{
			isDirected: g.IsDirected(),
			isWeighted: g.IsWeighted(),
			isAcyclic:  g.IsAcyclic(),
		},
	}

	vertices := make([]Vertex[T], len(all))
	for i, v := range all {
		vertices[i] = Vertex[T]{
			label:      v.label,
			properties: v.properties,
			metadata:   v.metadata,
			stored:     true,
			position:   i,
		}
		out.order[i] = &vertices[i]
		out.vertices[v.label] = &vertices[i]
		out.verticesCount++
	}

	outDegree := make([]int, len(all))
	inDegree := make([]int, len(all))
	for _, e := range allEdges {
		outDegree[out.vertices[e.source.label].position]++
		inDegree[out.vertices[e.dest.label].position]++
	}
	neighbors := make([]*Vertex[T], len(allEdges))
	incoming := make([]*Edge[T], len(allEdges))
	for i, k := range outDegree {
		if k > 0 {
			out.order[i].neighbors, neighbors = neighbors[:0:k], neighbors[k:]
			out.edges[out.order[i].label] = make(map[T]*Edge[T], k)
		}
		if k := inDegree[i]; k > 0 {
			out.order[i].incoming, incoming = incoming[:0:k], incoming[k:]
		}
	}

	edges := make([]Edge[T], len(allEdges))
	for i, e := range allEdges {
		from, to := out.vertices[e.source.label], out.vertices[e.dest.label]
		from.neighbors = append(from.neighbors, to)
		edges[i] = Edge[T]{
			source:     from,
			dest:       to,
			properties: e.properties,
			metadata:   e.metadata,
			added:      e.added,
		}
		to.incoming = append(to.incoming, &edges[i])
		out.edges[from.label][to.label] = &edges[i]
		out.nextEdge = max(out.nextEdge, e.added)
		out.edgesCount++
	}

	return out
}
