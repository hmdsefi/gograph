package connectivity

import "github.com/hmdsefi/gograph"

// Gabow's Algorithm is a linear time algorithm to find strongly connected
// components in a directed graph. It is similar to Tarjan's Algorithm in
// that it uses a stack to keep track of the current strongly connected
// component, but it uses a different approach to find the strongly connected
// components.
//
// The algorithm works by using a depth-first search, but it does not use
// recursion. Instead, it keeps track of the vertices visited so far in an
// array called visited. When a vertex is visited for the first time, it is
// added to the stack and marked as visited. Then, it visits all of its
// neighbors and adds them to the stack if they have not been visited before.
// If a neighbor is already on the stack, then it is part of the same strongly
// connected component, so Gabow's Algorithm keeps track of the minimum index
// of any vertex on the stack that can be reached from the neighbor. This
// minimum index is called the lowLink of the vertex.
//
// When the depth-first search is complete, the algorithm checks if the current
// vertex is the root of a strongly connected component. This is true if the
// lowLink of the vertex is equal to its index. If the vertex is the root of
// a strongly connected component, it pops all the vertices on the stack with
// indices greater than or equal to the vertex's index and adds them to a new
// strongly connected component.
//
// The algorithm continues in this way until all vertices have been visited.
// The resulting strongly connected components are returned by the algorithm.

// Gabow runs the Gabow's algorithm, and returns a list of strongly
// connected components, where each component is represented as an
// array of pointers to vertex structs.
//
// The order of the components and of the vertices in them is stable: it
// only depends on the order of the vertices and edges in the graph, so a
// graph built the same way gives the same result on every run.
func Gabow[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T] {
	var (
		index       int
		components  [][]*gograph.Vertex[T]
		stack       []*tarjanVertex[T]
		strongLinks func(v *tarjanVertex[T])
	)

	graphVertices := g.GetAllVertices()
	vertices := make(map[T]*tarjanVertex[T])
	for _, v := range graphVertices {
		vertices[v.Label()] = newTarjanVertex(v)
	}

	// strongLinks performs the DFS search from the root vertex and
	// identifies the strongly connected components. It keeps the search
	// path in a slice instead of recursing, so the depth of the graph
	// doesn't grow the goroutine stack.
	strongLinks = func(root *tarjanVertex[T]) {
		var path []tarjanFrame[T]
		enter := func(v *tarjanVertex[T]) {
			v.index = index
			v.lowLink = index
			index++
			stack = append(stack, v)
			v.onStack = true
			path = append(path, tarjanFrame[T]{vertex: v, neighbors: v.Neighbors()})
		}

		enter(root)
		for len(path) > 0 {
			frame := &path[len(path)-1]
			v := frame.vertex

			// The DFS search explores the neighbors of the vertex on top of
			// the path one at a time. For each neighbor w of v, the algorithm
			// either continues the search from w or updates the lowLink field
			// of v if w is already on the stack.
			if frame.next < len(frame.neighbors) {
				w := vertices[frame.neighbors[frame.next].Label()]
				frame.next++
				if w.index == -1 {
					enter(w)
				} else if w.onStack {
					if w.index < v.lowLink {
						v.lowLink = w.index
					}
				}
				continue
			}

			// All neighbors of v are explored. Its lowLink is final, so pass
			// it on to the vertex that reached v.
			path = path[:len(path)-1]
			if len(path) > 0 {
				parent := path[len(path)-1].vertex
				if v.lowLink < parent.lowLink {
					parent.lowLink = v.lowLink
				}
			}

			// v is the root of a strongly connected component when its
			// lowLink field is equal to its index field (i.e., when there
			// is no back edge to a node with a lower index).
			if v.lowLink == v.index {
				var (
					component []*tarjanVertex[T]
					w         *tarjanVertex[T]
				)
				for {
					w, stack = stack[len(stack)-1], stack[:len(stack)-1]
					w.onStack = false
					component = append(component, w)
					if w == v {
						break
					}
				}

				var temp []*gograph.Vertex[T]
				for i := range component {
					temp = append(temp, component[i].Vertex)
				}
				components = append(components, temp)
			}
		}
	}

	for _, v := range graphVertices {
		if tv := vertices[v.Label()]; tv.index == -1 {
			strongLinks(tv)
		}
	}

	return components
}
