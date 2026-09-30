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

// Tarjan is the entry point to the algorithm. It initializes the index,
// stack, and sccs variables and then loops through all the vertices in
// the graph. It returns a slice of vertices' slice, where each inner
// slice represents a strongly connected component of the graph.
//
// The order of the components and of the vertices in them is stable: it
// only depends on the order of the vertices and edges in the graph, so a
// graph built the same way gives the same result on every run.
func Tarjan[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T] {
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

	for _, v := range vertices {
		if tv := tvertices[v.Label()]; tv.index < 0 {
			tarj.visit(tv, &index, &stack, &sccs)
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

// tarjanFrame is a vertex on the current depth-first search path, with
// its neighbors and the index of the next neighbor to explore.
type tarjanFrame[T comparable] struct {
	vertex    *tarjanVertex[T]
	neighbors []*gograph.Vertex[T]
	next      int
}

// visit runs the depth-first search from the root vertex. Each vertex it
// enters gets an index and lowLink value and is pushed on the stack. If
// a neighbor has not been visited before, the search continues from it,
// and its lowLink value is taken into account when it's finished. If a
// neighbor has already been visited and is still on the stack, its index
// is taken into account.
//
// It keeps the search path in a slice instead of recursing, so the depth
// of the graph doesn't grow the goroutine stack.
func (t *tarjanSCCS[T]) visit(
	root *tarjanVertex[T],
	index *int,
	stack *[]*tarjanVertex[T],
	sccs *[][]*tarjanVertex[T],
) {
	var path []tarjanFrame[T]
	enter := func(v *tarjanVertex[T]) {
		v.index = *index
		v.lowLink = *index
		*index++
		*stack = append(*stack, v)
		v.onStack = true
		path = append(path, tarjanFrame[T]{vertex: v, neighbors: v.Neighbors()})
	}

	enter(root)
	for len(path) > 0 {
		frame := &path[len(path)-1]
		v := frame.vertex

		if frame.next < len(frame.neighbors) {
			tv := t.vertices[frame.neighbors[frame.next].Label()]
			frame.next++
			if tv.index == -1 {
				enter(tv)
			} else if tv.onStack {
				v.lowLink = min(v.lowLink, tv.index)
			}
			continue
		}

		path = path[:len(path)-1]
		if len(path) > 0 {
			parent := path[len(path)-1].vertex
			parent.lowLink = min(parent.lowLink, v.lowLink)
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
}

// min is a helper function that returns the minimum of two integers.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
