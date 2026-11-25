package gograph

import "iter"

// Vertices returns the vertices of g in the order of GetAllVertices. Each
// vertex is the graph's own vertex. If g changes during the loop, vertices
// can be skipped or repeated.
//
// For a graph from New, it reads the vertices in place: ranging over all of
// them takes O(V) time and allocates nothing. For another implementation of
// Graph, it calls GetAllVertices before the first vertex.
func Vertices[T comparable](g Graph[T]) iter.Seq[*Vertex[T]] {
	return func(yield func(*Vertex[T]) bool) {
		bg, ok := g.(*baseGraph[T])
		if !ok {
			for _, v := range g.GetAllVertices() {
				if !yield(v) {
					return
				}
			}
			return
		}

		for _, v := range bg.order {
			if v != nil && !yield(v) {
				return
			}
		}
	}
}

// Edges returns the edges of g in the order of AllEdges. In an undirected
// graph, an edge between two vertices comes once in each direction, and a
// self-loop once. If g changes during the loop, edges can be skipped or
// repeated.
//
// For a graph from New, it reads the edges in place: ranging over all of
// them takes O(V+E) time and allocates nothing. For another implementation
// of Graph, it calls AllEdges before the first edge.
func Edges[T comparable](g Graph[T]) iter.Seq[*Edge[T]] {
	return func(yield func(*Edge[T]) bool) {
		bg, ok := g.(*baseGraph[T])
		if !ok {
			for _, e := range g.AllEdges() {
				if !yield(e) {
					return
				}
			}
			return
		}

		for _, v := range bg.order {
			if v != nil && !bg.yieldOutgoing(v, yield) {
				return
			}
		}
	}
}

// yieldOutgoing yields the outgoing edges of v in the order of
// outgoingEdges, and reports false once yield does.
func (g *baseGraph[T]) yieldOutgoing(v *Vertex[T], yield func(*Edge[T]) bool) bool {
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

		if edge, ok := destMap[neighbor.label]; ok && !yield(edge) {
			return false
		}
	}

	return true
}
