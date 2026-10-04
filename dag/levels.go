package dag

import "github.com/hmdsefi/gograph"

// Levels groups the vertices of a directed acyclic graph so that every edge
// goes from a lower level to a higher one. Level 0 has the vertices with no
// incoming edges, and the level of any other vertex is the length of the
// longest path to it from one of those. Vertices in the same level have no
// edges between them, so they can run at the same time once every level
// before theirs is done.
//
// The vertices of each level follow the order of GetAllVertices, as the
// graph's own vertex pointers. It runs in O(V + E) time.
//
// It returns gograph.ErrNotDirected for undirected graphs, and
// gograph.ErrDAGHasCycle if the graph has a cycle.
func Levels[T comparable](g gograph.Graph[T]) ([][]*gograph.Vertex[T], error) {
	if !g.IsDirected() {
		return nil, gograph.ErrNotDirected
	}

	vertices := g.GetAllVertices()
	if len(vertices) == 0 {
		return nil, nil
	}

	waiting := make(map[T]int, len(vertices))
	var queue []*gograph.Vertex[T]
	for _, v := range vertices {
		waiting[v.Label()] = v.InDegree()
		if v.InDegree() == 0 {
			queue = append(queue, v)
		}
	}

	next := outgoing(g)
	level := make(map[T]int, len(vertices))
	deepest := 0
	for i := 0; i < len(queue); i++ {
		from := queue[i].Label()
		for _, v := range next(queue[i]) {
			to := v.Label()
			level[to] = max(level[to], level[from]+1)
			waiting[to]--
			if waiting[to] == 0 {
				queue = append(queue, v)
				deepest = max(deepest, level[to])
			}
		}
	}

	if len(queue) != len(vertices) {
		return nil, gograph.ErrDAGHasCycle
	}

	levels := make([][]*gograph.Vertex[T], deepest+1)
	for _, v := range vertices {
		l := level[v.Label()]
		levels[l] = append(levels[l], v)
	}

	return levels, nil
}
