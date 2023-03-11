package connectivity

import "github.com/hmdsefi/gograph"

// Tarjan's algorithm is based on depth-first search and is widely used for
// finding strongly connected components in a graph. The algorithm is efficient
// and has a time complexity of O(V+E), where V is the number of vertices and E
// is the number of edges in the graph.

// tarjanVertex wraps the gograph.Vertex struct to add new fields to it.
type tarjanVertex[T comparable] struct {
	*gograph.Vertex[T]      // the vertex that being wrapped.
	index              int  // represents the order in which a vertex is visited during the DFS search.
	lowLink            int  // the minimum index of any vertex reachable from the vertex during the search.
	onStack            bool // a boolean flag that shows if the vertex is in the stack or not.
}

func newTarjanVertex[T comparable](vertex *gograph.Vertex[T]) *tarjanVertex[T] {
	return &tarjanVertex[T]{
		Vertex: vertex,
		index:  -1,
	}
}

type tarjanSCCS[T comparable] struct {
	vertices map[T]*tarjanVertex[T]
}

func newTarjanSCCS[T comparable](vertices map[T]*tarjanVertex[T]) *tarjanSCCS[T] {
	return &tarjanSCCS[T]{vertices: vertices}
}

// tarjan  is the entry point to the algorithm. It initializes the index,
// stack, and sccs variables and then loops through all the vertices in
// the graph. It returns a slice of vertices' slice, where each inner
// slice represents a strongly connected component of the graph.
func tarjan[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T] {
	var (
		index     int
		stack     []*tarjanVertex[T]
		tvertices = make(map[T]*tarjanVertex[T])
		sccs      [][]*tarjanVertex[T]
	)

	vertices := g.GetAllVertices()

	for _, v := range vertices {
		tv := newTarjanVertex(v)
		tvertices[tv.Label()] = tv
	}

	tarj := newTarjanSCCS(tvertices)

	for _, v := range tvertices {
		if v.index < 0 {
			tarj.visit(v, &index, &stack, &sccs)
		}
	}

	result := make([][]*gograph.Vertex[T], len(sccs))
	for i, list := range sccs {
		result[i] = make([]*gograph.Vertex[T], len(list))
		for j := range list {
			result[i][j] = list[j].Vertex
		}
	}
	return result
}

// visit updates the index and lowLink values of the vertex, adds it to
// the stack, and recursively calls itself on each of its neighbors. If
// a neighbor has not been visited before, its index and lowLink values
// are updated, and the recursion continues. If a neighbor has already
// been visited and is still on the stack, its lowLink value is updated.
func (t *tarjanSCCS[T]) visit(
	v *tarjanVertex[T],
	index *int,
	stack *[]*tarjanVertex[T],
	sccs *[][]*tarjanVertex[T],
) {
	v.index = *index
	v.lowLink = *index
	*index++
	*stack = append(*stack, v)
	v.onStack = true

	neighbors := v.Neighbors()
	for _, w := range neighbors {
		tv := t.vertices[w.Label()]
		if tv.index == -1 {
			t.visit(tv, index, stack, sccs)
			v.lowLink = min(v.lowLink, tv.lowLink)
		} else if tv.onStack {
			v.lowLink = min(v.lowLink, tv.index)
		}
	}

	if v.lowLink == v.index {
		var scc []*tarjanVertex[T]
		for {
			w := (*stack)[len(*stack)-1]
			*stack = (*stack)[:len(*stack)-1]
			w.onStack = false
			scc = append(scc, w)
			if w == v {
				break
			}
		}
		*sccs = append(*sccs, scc)
	}
}

// min is a helper function that returns the minimum of two integers.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
