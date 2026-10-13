package dag

import "github.com/hmdsefi/gograph"

// Sources returns the vertices of a directed graph with no incoming edges, in
// the order of GetAllVertices, as the graph's own vertex pointers.
//
// It works on any directed graph, with or without cycles. A vertex on a cycle
// is never a source. Isolated vertices have no incoming edges, so they are
// included. A self-loop counts as an incoming edge, so a vertex whose only edge
// is a self-loop is not a source. An empty graph returns an empty slice and no
// error.
//
// It returns gograph.ErrNotDirected for undirected graphs. It runs in O(V)
// time.
func Sources[T comparable](g gograph.Graph[T]) ([]*gograph.Vertex[T], error) {
	if !g.IsDirected() {
		return nil, gograph.ErrNotDirected
	}

	var sources []*gograph.Vertex[T]
	for _, v := range g.GetAllVertices() {
		if v.InDegree() == 0 {
			sources = append(sources, v)
		}
	}
	if sources == nil {
		sources = []*gograph.Vertex[T]{}
	}

	return sources, nil
}

// Sinks returns the vertices of a directed graph with no outgoing edges, in
// the order of GetAllVertices, as the graph's own vertex pointers.
//
// It works on any directed graph, with or without cycles. A vertex on a cycle
// is never a sink. Isolated vertices have no outgoing edges, so they are
// included. A self-loop counts as an outgoing edge, so a vertex whose only edge
// is a self-loop is not a sink. An empty graph returns an empty slice and no
// error.
//
// It returns gograph.ErrNotDirected for undirected graphs. It runs in O(V)
// time.
func Sinks[T comparable](g gograph.Graph[T]) ([]*gograph.Vertex[T], error) {
	if !g.IsDirected() {
		return nil, gograph.ErrNotDirected
	}

	var sinks []*gograph.Vertex[T]
	for _, v := range g.GetAllVertices() {
		if v.OutDegree() == 0 {
			sinks = append(sinks, v)
		}
	}
	if sinks == nil {
		sinks = []*gograph.Vertex[T]{}
	}

	return sinks, nil
}
