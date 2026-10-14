package dag

import "github.com/hmdsefi/gograph"

// CriticalPath returns the path through a directed acyclic graph with the
// largest total cost, and that cost. The cost of a path is the sum of the
// weights of its vertices and of its edges. When tasks run in parallel, this
// is the longest chain of dependent tasks, so it sets the total time.
//
// Weights are read with Vertex.Weight and Edge.Weight, and unset weights are
// 0. The graph doesn't have to be created with Weighted(). Negative weights
// are allowed.
//
// The path starts at a vertex with no incoming edges and ends at one with no
// outgoing edges, so a vertex without a weight, like a start or finish
// milestone, stays on it. An empty graph returns a nil path and a cost of 0.
// If several paths have the same cost, the one found first in topological
// order wins, so the same graph gives the same path on every run. The
// returned vertices are the graph's own pointers.
//
// It runs in O(V + E) time.
//
// It returns gograph.ErrNotDirected for undirected graphs, and
// gograph.ErrDAGHasCycle if the graph has a cycle.
func CriticalPath[T comparable](g gograph.Graph[T]) ([]*gograph.Vertex[T], float64, error) {
	return CriticalPathFunc(g,
		func(v *gograph.Vertex[T]) float64 { return v.Weight() },
		func(e *gograph.Edge[T]) float64 { return e.Weight() },
	)
}

// CriticalPathFunc works like CriticalPath, but takes the costs from
// vertexCost and edgeCost instead of the weights stored in the graph. That
// suits costs kept outside the graph, such as measured durations. A nil
// function means a cost of 0.
func CriticalPathFunc[T comparable](
	g gograph.Graph[T],
	vertexCost func(*gograph.Vertex[T]) float64,
	edgeCost func(*gograph.Edge[T]) float64,
) ([]*gograph.Vertex[T], float64, error) {
	if !g.IsDirected() {
		return nil, 0, gograph.ErrNotDirected
	}

	if vertexCost == nil {
		vertexCost = func(*gograph.Vertex[T]) float64 { return 0 }
	}
	if edgeCost == nil {
		edgeCost = func(*gograph.Edge[T]) float64 { return 0 }
	}

	vertices := g.GetAllVertices()
	if len(vertices) == 0 {
		return nil, 0, nil
	}

	index := make(map[T]int, len(vertices))
	for i, v := range vertices {
		index[v.Label()] = i
	}
	out := make([][]*gograph.Edge[T], len(vertices))
	for _, e := range g.AllEdges() {
		i := index[e.Source().Label()]
		out[i] = append(out[i], e)
	}

	// best[i] is the cost of the most expensive path to vertices[i] from a
	// vertex with no incoming edges, and prev[i] is the index of the vertex
	// before it on that path, or -1. Until a vertex leaves the queue, best
	// holds the cost up to its incoming edge.
	best := make([]float64, len(vertices))
	prev := make([]int, len(vertices))
	waiting := make([]int, len(vertices))
	queue := make([]int, 0, len(vertices))
	for i, v := range vertices {
		prev[i] = -1
		waiting[i] = v.InDegree()
		if waiting[i] == 0 {
			queue = append(queue, i)
		}
	}

	for i := 0; i < len(queue); i++ {
		from := queue[i]
		best[from] += vertexCost(vertices[from])
		for _, e := range out[from] {
			to := index[e.Destination().Label()]
			if cost := best[from] + edgeCost(e); prev[to] == -1 || cost > best[to] {
				best[to] = cost
				prev[to] = from
			}

			waiting[to]--
			if waiting[to] == 0 {
				queue = append(queue, to)
			}
		}
	}

	if len(queue) != len(vertices) {
		return nil, 0, gograph.ErrDAGHasCycle
	}

	end := -1
	for _, i := range queue {
		if vertices[i].OutDegree() == 0 && (end == -1 || best[i] > best[end]) {
			end = i
		}
	}

	var path []*gograph.Vertex[T]
	for i := end; i != -1; i = prev[i] {
		path = append(path, vertices[i])
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path, best[end], nil
}
