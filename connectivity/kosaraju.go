package connectivity

import "github.com/hmdsefi/gograph"

// Kosaraju's Algorithm: This algorithm is also based on depth-first search
// and is used to find strongly connected components in a graph. The algorithm
// has a time complexity of O(V+E) and is considered to be one of the most
// efficient algorithms for finding strongly connected components.
//
// Note that this implementation assumes that the graph is connected, and
// may not work correctly for disconnected graphs. It also does not handle
// cases where the graph contains self-loops or parallel edges.

// kosarajuDFS exposes two different depth-first search methods.
type kosarajuDFS[T comparable] struct {
	visited map[T]bool
}

func newKosarajuSCCS[T comparable]() *kosarajuDFS[T] {
	return &kosarajuDFS[T]{
		visited: make(map[T]bool),
	}
}

// implements Kosaraju's Algorithm. It performs a depth-first search of
// the graph to create a stack of vertices, and then performs a second
// depth-first search on the transposed graph to identify the strongly
// connected components.
//
// The function returns a slice of slices, where each slice represents
// a strongly connected component and contains the vertices that belong
// to that component.
func kosaraju[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T] {
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
			sccs = append(sccs, scc)
		}
	}

	return sccs
}

// dfs1 creates the stack of vertices.
func (k *kosarajuDFS[T]) dfs1(v *gograph.Vertex[T], stack *[]T) {
	k.visited[v.Label()] = true
	neighbors := v.Neighbors()
	for _, neighbor := range neighbors {
		if !k.visited[neighbor.Label()] {
			k.dfs1(neighbor, stack)
		}
	}
	*stack = append(*stack, v.Label())
}

// dfs2 explores the strongly connected components.
func (k *kosarajuDFS[T]) dfs2(v *gograph.Vertex[T], scc *[]*gograph.Vertex[T]) {
	k.visited[v.Label()] = true
	*scc = append(*scc, v)
	neighbors := v.Neighbors()
	for _, neighbor := range neighbors {
		if !k.visited[neighbor.Label()] {
			k.dfs2(neighbor, scc)
		}
	}
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
