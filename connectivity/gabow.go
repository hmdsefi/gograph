package connectivity

import "github.com/hmdsefi/gograph"

// Gabow's path-based algorithm finds the strongly connected components of a
// directed graph in O(V+E) time with a single depth-first search. Instead of
// the lowLink values of Tarjan's algorithm, it keeps two stacks:
//
//   - stack S holds the visited vertices that aren't in a component yet, in
//     the order the search reached them.
//   - stack P holds the vertices on the search path that can still be the
//     root of a component.
//
// When the search reaches a vertex, it's pushed on both stacks. When an edge
// leads to a vertex w that is still on S, every vertex on the path after w is
// in the same component as w, so P is popped until its top was reached no
// later than w. When the search is done with a vertex that is still on top of
// P, that vertex is the root of a component: it's popped from P, and S is
// popped down to it to collect the component.

// gabowVertex wraps the gograph.Vertex struct to add the fields of the search.
type gabowVertex[T comparable] struct {
	*gograph.Vertex[T]      // the vertex being wrapped.
	preorder           int  // the order in which the search reached the vertex, -1 before that.
	onStack            bool // true while the vertex is on stack S.
}

// gabowFrame is a vertex on the current depth-first search path, with its
// neighbors and the index of the next neighbor to explore.
type gabowFrame[T comparable] struct {
	vertex    *gabowVertex[T]
	neighbors []*gograph.Vertex[T]
	next      int
}

// Gabow runs Gabow's path-based algorithm, and returns a list of strongly
// connected components, where each component is represented as an
// array of pointers to vertex structs.
//
// It keeps the search path in a slice instead of recursing, so the depth
// of the graph doesn't grow the goroutine stack.
//
// The order of the components and of the vertices in them is stable: it
// only depends on the order of the vertices and edges in the graph, so a
// graph built the same way gives the same result on every run.
func Gabow[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T] {
	graphVertices := g.GetAllVertices()
	vertices := make(map[T]*gabowVertex[T], len(graphVertices))
	for _, v := range graphVertices {
		vertices[v.Label()] = &gabowVertex[T]{Vertex: v, preorder: -1}
	}

	var (
		counter    int
		components [][]*gograph.Vertex[T]
		stack      []*gabowVertex[T] // S
		roots      []*gabowVertex[T] // P
		path       []gabowFrame[T]
	)

	enter := func(v *gabowVertex[T]) {
		v.preorder = counter
		counter++
		v.onStack = true
		stack = append(stack, v)
		roots = append(roots, v)
		path = append(path, gabowFrame[T]{vertex: v, neighbors: v.Neighbors()})
	}

	for _, root := range graphVertices {
		if vertices[root.Label()].preorder != -1 {
			continue
		}

		enter(vertices[root.Label()])
		for len(path) > 0 {
			frame := &path[len(path)-1]
			if frame.next < len(frame.neighbors) {
				w := vertices[frame.neighbors[frame.next].Label()]
				frame.next++
				if w.preorder == -1 {
					enter(w)
				} else if w.onStack {
					for roots[len(roots)-1].preorder > w.preorder {
						roots = roots[:len(roots)-1]
					}
				}
				continue
			}

			v := frame.vertex
			path = path[:len(path)-1]
			if roots[len(roots)-1] != v {
				continue
			}

			roots = roots[:len(roots)-1]
			var component []*gograph.Vertex[T]
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				w.onStack = false
				component = append(component, w.Vertex)
				if w == v {
					break
				}
			}
			components = append(components, component)
		}
	}

	return components
}
