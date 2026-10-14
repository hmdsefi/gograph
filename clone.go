package gograph

import "sync/atomic"

// Clone returns an independent copy of g, preserving its directed, weighted,
// and acyclic properties, vertex and edge weights, labels, and iteration order.
// Vertices and edges are newly allocated. Metadata is copied as an interface
// value; reference values stored in metadata still refer to the same data.
// A nil graph panics. The graph must not be modified while it is being copied.
//
// Time and space are O(V + E).
func Clone[T comparable](g Graph[T]) Graph[T] {
	if g == nil {
		panic("gograph: Clone requires a non-nil graph")
	}

	vertices := g.GetAllVertices()
	if bg, ok := g.(*baseGraph[T]); ok {
		labels := make([]T, len(vertices))
		for i, v := range vertices {
			labels[i] = v.Label()
		}
		out, err := inducedSubgraph(bg, labels)
		if err != nil {
			panic(err)
		}
		for _, e := range bg.AllEdges() {
			out.GetEdge(e.Source(), e.Destination()).added = e.InsertionIndex()
		}
		out.nextEdge = bg.nextEdge
		return out
	}

	out := newBaseGraph[T](GraphProperties{
		isDirected: g.IsDirected(),
		isWeighted: g.IsWeighted(),
		isAcyclic:  g.IsAcyclic(),
	})
	for _, v := range vertices {
		copy := out.AddVertexByLabel(v.Label(), WithVertexWeight(v.Weight()))
		copy.metadata = v.Metadata()
		out.edges[v.Label()] = make(map[T]*Edge[T])
	}
	for _, e := range g.AllEdges() {
		from := out.GetVertexByID(e.Source().Label())
		to := out.GetVertexByID(e.Destination().Label())
		putInducedEdge(out, from, to, e)
		out.GetEdge(from, to).added = e.InsertionIndex()
		out.nextEdge = max(out.nextEdge, e.InsertionIndex())
		atomic.AddUint32(&out.edgesCount, 1)
	}
	return out
}
