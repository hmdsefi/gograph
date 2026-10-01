package connectivity

import "github.com/hmdsefi/gograph"

// Kosaraju's Algorithm: This algorithm is also based on depth-first search
// and is used to find strongly connected components in a graph. The algorithm
// has a time complexity of O(V+E) and is considered to be one of the most
// efficient algorithms for finding strongly connected components.

// kosarajuDFS exposes two different depth-first search methods.
type kosarajuDFS[T comparable] struct {
	visited map[T]bool
}

func newKosarajuSCCS[T comparable]() *kosarajuDFS[T] {
	return &kosarajuDFS[T]{
		visited: make(map[T]bool),
	}
}

// Kosaraju implements Kosaraju's Algorithm. It performs a depth-first
// search of the graph to create a stack of vertices, and then performs
// a second depth-first search on the transposed graph to identify the
// strongly connected components.
//
// The function returns a slice of slices, where each slice represents
// a strongly connected component and contains the vertices that belong
// to that component. The vertices are the ones stored in g, not copies.
//
// The order of the components and of the vertices in them is stable: it
// only depends on the order of the vertices and edges in the graph, so a
// graph built the same way gives the same result on every run.
func Kosaraju[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T] {
	vertices := g.GetAllVertices()

	// Step 1: Perform a depth-first search of the graph to create a stack of vertices
	kosar := newKosarajuSCCS[T]()
	stack := make([]T, 0, len(vertices))
	for _, v := range vertices {
		if !kosar.visited[v.Label()] {
			kosar.dfs1(v, &stack)
		}
	}

	// Step 2: Perform a second depth-first search on the transposed graph
	transposed := kosar.reverse(g)
	kosar.visited = make(map[T]bool)
	sccs := make([][]*gograph.Vertex[T], 0)
	for len(stack) > 0 {
		v := transposed.GetVertexByID(stack[len(stack)-1])
		stack = stack[:len(stack)-1]

		if !kosar.visited[v.Label()] {
			scc := make([]*gograph.Vertex[T], 0)
			kosar.dfs2(v, &scc)
			for i := range scc {
				scc[i] = g.GetVertexByID(scc[i].Label())
			}
			sccs = append(sccs, scc)
		}
	}

	return sccs
}

// kosarajuFrame is a vertex on the current depth-first search path, with
// its neighbors and the index of the next neighbor to explore.
type kosarajuFrame[T comparable] struct {
	vertex    *gograph.Vertex[T]
	neighbors []*gograph.Vertex[T]
	next      int
}

// search runs a depth-first search from the root vertex over vertices
// that aren't visited yet. It calls enter when it first reaches a vertex
// and finish when all of the vertex's neighbors are explored.
//
// It keeps the search path in a slice instead of recursing, so the depth
// of the graph doesn't grow the goroutine stack.
func (k *kosarajuDFS[T]) search(root *gograph.Vertex[T], enter, finish func(v *gograph.Vertex[T])) {
	k.visited[root.Label()] = true
	enter(root)
	path := []kosarajuFrame[T]{{vertex: root, neighbors: root.Neighbors()}}

	for len(path) > 0 {
		frame := &path[len(path)-1]
		if frame.next < len(frame.neighbors) {
			neighbor := frame.neighbors[frame.next]
			frame.next++
			if !k.visited[neighbor.Label()] {
				k.visited[neighbor.Label()] = true
				enter(neighbor)
				path = append(path, kosarajuFrame[T]{vertex: neighbor, neighbors: neighbor.Neighbors()})
			}
			continue
		}

		finish(frame.vertex)
		path = path[:len(path)-1]
	}
}

// dfs1 creates the stack of vertices.
func (k *kosarajuDFS[T]) dfs1(v *gograph.Vertex[T], stack *[]T) {
	k.search(
		v,
		func(*gograph.Vertex[T]) {},
		func(v *gograph.Vertex[T]) { *stack = append(*stack, v.Label()) },
	)
}

// dfs2 explores the strongly connected components.
func (k *kosarajuDFS[T]) dfs2(v *gograph.Vertex[T], scc *[]*gograph.Vertex[T]) {
	k.search(
		v,
		func(v *gograph.Vertex[T]) { *scc = append(*scc, v) },
		func(*gograph.Vertex[T]) {},
	)
}

func (k *kosarajuDFS[T]) reverse(g gograph.Graph[T]) gograph.Graph[T] {
	reversed := gograph.New[T](gograph.Directed())
	vertices := g.GetAllVertices()

	for _, v := range vertices {
		reversed.AddVertexByLabel(v.Label())
	}

	for i := range vertices {
		neighbors := vertices[i].Neighbors()
		for j := range neighbors {
			_, _ = reversed.AddEdge(
				reversed.GetVertexByID(neighbors[j].Label()),
				reversed.GetVertexByID(vertices[i].Label()),
			)
		}
	}
	return reversed
}
