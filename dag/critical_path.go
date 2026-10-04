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
// A path can start and end at any vertex. An empty graph returns a nil path
// and a cost of 0. If several paths have the same cost, the one found first
// in topological order wins, so the same graph gives the same path on every
// run. The returned vertices are the graph's own pointers.
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

	out := make(map[T][]*gograph.Edge[T], len(vertices))
	for _, e := range g.AllEdges() {
		label := e.Source().Label()
		out[label] = append(out[label], e)
	}

	// best is the cost of the most expensive path that ends at a vertex, and
	// prev is the vertex before it on that path. A path may start anywhere,
	// so every vertex begins with just its own cost.
	best := make(map[T]float64, len(vertices))
	prev := make(map[T]*gograph.Vertex[T], len(vertices))
	waiting := make(map[T]int, len(vertices))
	var queue []*gograph.Vertex[T]
	for _, v := range vertices {
		best[v.Label()] = vertexCost(v)
		waiting[v.Label()] = v.InDegree()
		if v.InDegree() == 0 {
			queue = append(queue, v)
		}
	}

	for i := 0; i < len(queue); i++ {
		from := queue[i]
		for _, e := range out[from.Label()] {
			to := g.GetVertexByID(e.Destination().Label())
			if cost := best[from.Label()] + edgeCost(e) + vertexCost(to); cost > best[to.Label()] {
				best[to.Label()] = cost
				prev[to.Label()] = from
			}

			waiting[to.Label()]--
			if waiting[to.Label()] == 0 {
				queue = append(queue, to)
			}
		}
	}

	if len(queue) != len(vertices) {
		return nil, 0, gograph.ErrDAGHasCycle
	}

	end := queue[0]
	for _, v := range queue[1:] {
		if best[v.Label()] > best[end.Label()] {
			end = v
		}
	}

	var path []*gograph.Vertex[T]
	for v := end; v != nil; v = prev[v.Label()] {
		path = append(path, v)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path, best[end.Label()], nil
}
