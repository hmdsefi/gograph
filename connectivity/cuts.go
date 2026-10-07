package connectivity

import (
	"errors"

	"github.com/hmdsefi/gograph"
)

// ErrNotUndirected indicates that an undirected graph is required.
var ErrNotUndirected = errors.New("graph is not undirected")

// Bridges returns the edges whose removal increases the number of connected
// components of an undirected graph. Each bridge appears once, oriented from its
// DFS parent to its child, in DFS completion order. The edges belong to g.
// Self-loops are never bridges; empty graphs return an empty slice.
// Directed graphs return ErrNotUndirected. It runs in O(V+E) time and space.
func Bridges[T comparable](g gograph.Graph[T]) ([]*gograph.Edge[T], error) {
	bridges, _, err := cuts(g)
	return bridges, err
}

// ArticulationPoints returns the vertices whose removal increases the number of
// connected components of an undirected graph, in GetAllVertices order.
// The vertices belong to g. Empty graphs return an empty slice; isolated vertices
// and self-loops are not articulation points. Directed graphs return
// ErrNotUndirected. It runs in O(V+E) time and space.
func ArticulationPoints[T comparable](g gograph.Graph[T]) ([]*gograph.Vertex[T], error) {
	_, points, err := cuts(g)
	return points, err
}

type cutFrame[T comparable] struct {
	vertex    *gograph.Vertex[T]
	neighbors []*gograph.Vertex[T]
	next      int
	children  int
}

func cuts[T comparable](g gograph.Graph[T]) ([]*gograph.Edge[T], []*gograph.Vertex[T], error) {
	if g.IsDirected() {
		return nil, nil, ErrNotUndirected
	}
	vertices := g.GetAllVertices()
	index := make(map[T]int, len(vertices))
	low := make(map[T]int, len(vertices))
	isPoint := make(map[T]bool)
	bridges := make([]*gograph.Edge[T], 0)
	points := make([]*gograph.Vertex[T], 0)
	clock := 0
	var stack []cutFrame[T]
	enter := func(v *gograph.Vertex[T]) {
		clock++
		index[v.Label()], low[v.Label()] = clock, clock
		stack = append(stack, cutFrame[T]{vertex: v, neighbors: v.Neighbors()})
	}
	for _, root := range vertices {
		if index[root.Label()] != 0 {
			continue
		}
		enter(root)
		for len(stack) > 0 {
			frame := &stack[len(stack)-1]
			label := frame.vertex.Label()
			if frame.next < len(frame.neighbors) {
				neighbor := frame.neighbors[frame.next].Label()
				frame.next++
				if len(stack) > 1 && neighbor == stack[len(stack)-2].vertex.Label() {
					continue
				}
				if index[neighbor] == 0 {
					frame.children++
					enter(g.GetVertexByID(neighbor))
				} else {
					low[label] = min(low[label], index[neighbor])
				}
				continue
			}
			vertex, children := frame.vertex, frame.children
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				isPoint[label] = children > 1
				continue
			}
			parent := stack[len(stack)-1].vertex
			pl := parent.Label()
			low[pl] = min(low[pl], low[label])
			if low[label] > index[pl] {
				bridges = append(bridges, g.GetEdge(parent, vertex))
			}
			if len(stack) > 1 && low[label] >= index[pl] {
				isPoint[pl] = true
			}
		}
	}
	for _, v := range vertices {
		if isPoint[v.Label()] {
			points = append(points, v)
		}
	}
	return bridges, points, nil
}
