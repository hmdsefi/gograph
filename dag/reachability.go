package dag

import (
	"fmt"

	"github.com/hmdsefi/gograph"
)

// Descendants returns the vertices reachable from the vertex with the given
// label, nearest first. The vertex itself is only included if it's on a
// cycle.
//
// It runs in O(V + E) time for the part of the graph it reaches.
func Descendants[T comparable](g gograph.Graph[T], label T) ([]*gograph.Vertex[T], error) {
	if !g.IsDirected() {
		return nil, gograph.ErrNotDirected
	}

	start, err := lookup(g, label)
	if err != nil {
		return nil, err
	}

	next := outgoing(g)
	return reachable(next(start), next), nil
}

// Ancestors returns the vertices that can reach the vertex with the given
// label, nearest first. The vertex itself is only included if it's on a
// cycle.
//
// The graph doesn't index incoming edges, so it reads all edges once and
// runs in O(V + E) time.
func Ancestors[T comparable](g gograph.Graph[T], label T) ([]*gograph.Vertex[T], error) {
	if !g.IsDirected() {
		return nil, gograph.ErrNotDirected
	}

	start, err := lookup(g, label)
	if err != nil {
		return nil, err
	}

	next := incoming(g)
	return reachable(next(start), next), nil
}

// Affected returns the changed vertices and every vertex reachable from any
// of them. The changed vertices come first, in the order given, followed by
// the others, nearest first. A label that is given more than once, or that
// is reachable from another changed vertex, is returned once.
//
// It runs in O(V + E) time for the part of the graph it reaches.
func Affected[T comparable](g gograph.Graph[T], changed ...T) ([]*gograph.Vertex[T], error) {
	if !g.IsDirected() {
		return nil, gograph.ErrNotDirected
	}

	start := make([]*gograph.Vertex[T], 0, len(changed))
	for _, label := range changed {
		v, err := lookup(g, label)
		if err != nil {
			return nil, err
		}
		start = append(start, v)
	}

	return reachable(start, outgoing(g)), nil
}

// lookup returns the vertex of g with the given label.
func lookup[T comparable](g gograph.Graph[T], label T) (*gograph.Vertex[T], error) {
	v := g.GetVertexByID(label)
	if v == nil {
		return nil, fmt.Errorf("%w: %v", gograph.ErrVertexDoesNotExist, label)
	}

	return v, nil
}

// outgoing returns a function that gives the vertices v has edges to, as
// the graph's own pointers.
func outgoing[T comparable](g gograph.Graph[T]) func(v *gograph.Vertex[T]) []*gograph.Vertex[T] {
	return func(v *gograph.Vertex[T]) []*gograph.Vertex[T] {
		neighbors := v.Neighbors()
		for i, n := range neighbors {
			neighbors[i] = g.GetVertexByID(n.Label())
		}
		return neighbors
	}
}

// incoming returns a function that gives the vertices with an edge to v. It
// indexes the edges of g once.
func incoming[T comparable](g gograph.Graph[T]) func(v *gograph.Vertex[T]) []*gograph.Vertex[T] {
	sources := make(map[T][]*gograph.Vertex[T])
	for _, e := range g.AllEdges() {
		dest := e.Destination().Label()
		sources[dest] = append(sources[dest], e.Source())
	}

	return func(v *gograph.Vertex[T]) []*gograph.Vertex[T] {
		return sources[v.Label()]
	}
}

// reachable returns first, without repeats, followed by the vertices
// reachable from it through next, in breadth-first order.
func reachable[T comparable](
	first []*gograph.Vertex[T],
	next func(v *gograph.Vertex[T]) []*gograph.Vertex[T],
) []*gograph.Vertex[T] {
	seen := make(map[T]bool)
	var queue []*gograph.Vertex[T]
	visit := func(v *gograph.Vertex[T]) {
		if !seen[v.Label()] {
			seen[v.Label()] = true
			queue = append(queue, v)
		}
	}

	for _, v := range first {
		visit(v)
	}
	for i := 0; i < len(queue); i++ {
		for _, v := range next(queue[i]) {
			visit(v)
		}
	}

	return queue
}
