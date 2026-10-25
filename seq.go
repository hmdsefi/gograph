package gograph

import "iter"

// Vertices returns the vertices of g in the order of GetAllVertices. For a
// graph from New, it reads them in place, without the slice GetAllVertices
// builds. Each vertex is the graph's own vertex. Changing g while ranging
// over the sequence isn't supported.
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

// Edges returns the edges of g in the order of AllEdges. For a graph from
// New, it collects the edges of one vertex at a time, without the slice
// AllEdges builds. Changing g while ranging over the sequence isn't
// supported.
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

		var edges []*Edge[T]
		for _, v := range bg.order {
			if v == nil {
				continue
			}
			edges = bg.outgoingEdges(v, edges[:0])
			for _, e := range edges {
				if !yield(e) {
					return
				}
			}
		}
	}
}
