package connectivity

import "github.com/hmdsefi/gograph"

// Bridges returns the edges whose removal increases the number of connected
// components of an undirected graph. Each bridge appears once, oriented from its
// DFS parent to its child, in DFS completion order. The edges belong to g.
// Self-loops are never bridges; empty graphs return an empty slice.
// Directed graphs return gograph.ErrNotUndirected. It runs in O(V+E) time and space.
func Bridges[T comparable](g gograph.Graph[T]) ([]*gograph.Edge[T], error) {
	bridges, _, err := cuts(g)
	return bridges, err
}

// ArticulationPoints returns the vertices whose removal increases the number of
// connected components of an undirected graph, in GetAllVertices order.
// The vertices belong to g. Empty graphs return an empty slice; isolated vertices
// and self-loops are not articulation points. Directed graphs return
// gograph.ErrNotUndirected. It runs in O(V+E) time and space.
func ArticulationPoints[T comparable](g gograph.Graph[T]) ([]*gograph.Vertex[T], error) {
	_, points, err := cuts(g)
	return points, err
}

type cutFrame[T comparable] struct {
	vertex    int
	neighbors []*gograph.Vertex[T]
	next      int
	children  int
}

func cuts[T comparable](g gograph.Graph[T]) ([]*gograph.Edge[T], []*gograph.Vertex[T], error) {
	if g.IsDirected() {
		return nil, nil, gograph.ErrNotUndirected
	}
	vertices := g.GetAllVertices()
	index := make(map[T]int, len(vertices))
	for i, v := range vertices {
		index[v.Label()] = i
	}
	// Discovery times start at 1, leaving 0 for vertices not visited yet.
	disc := make([]int, len(vertices))
	low := make([]int, len(vertices))
	isPoint := make([]bool, len(vertices))
	bridges := make([]*gograph.Edge[T], 0)
	points := make([]*gograph.Vertex[T], 0)
	clock := 0
	var stack []cutFrame[T]
	enter := func(i int) {
		clock++
		disc[i], low[i] = clock, clock
		stack = append(stack, cutFrame[T]{vertex: i, neighbors: vertices[i].Neighbors()})
	}
	for root := range vertices {
		if disc[root] != 0 {
			continue
		}
		enter(root)
		for len(stack) > 0 {
			frame := &stack[len(stack)-1]
			v := frame.vertex
			if frame.next < len(frame.neighbors) {
				w := index[frame.neighbors[frame.next].Label()]
				frame.next++
				if len(stack) > 1 && w == stack[len(stack)-2].vertex {
					continue
				}
				if disc[w] == 0 {
					frame.children++
					enter(w)
				} else {
					low[v] = min(low[v], disc[w])
				}
				continue
			}
			children := frame.children
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				isPoint[v] = children > 1
				continue
			}
			p := stack[len(stack)-1].vertex
			low[p] = min(low[p], low[v])
			if low[v] > disc[p] {
				bridges = append(bridges, g.GetEdge(vertices[p], vertices[v]))
			}
			if len(stack) > 1 && low[v] >= disc[p] {
				isPoint[p] = true
			}
		}
	}
	for i, v := range vertices {
		if isPoint[i] {
			points = append(points, v)
		}
	}
	return bridges, points, nil
}
