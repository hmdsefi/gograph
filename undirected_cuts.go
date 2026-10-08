package gograph

// UndirectedCuts returns the bridges and articulation points of an undirected
// graph. Directed graphs return ErrNotUndirected.
//
// Each bridge appears once, oriented from its DFS parent to its child, in DFS
// completion order. Articulation points are in GetAllVertices order. The
// returned edges and vertices belong to g. Self-loops are never bridges, and
// isolated vertices and self-loops are not articulation points. Empty graphs
// return empty slices. It runs in O(V+E) time and space.
//
// connectivity.Bridges and connectivity.ArticulationPoints call this.
func UndirectedCuts[T comparable](g Graph[T]) ([]*Edge[T], []*Vertex[T], error) {
	bg, ok := g.(*baseGraph[T])
	if !ok || bg == nil {
		panic("gograph: UndirectedCuts requires a graph from New")
	}
	if bg.IsDirected() {
		return nil, nil, ErrNotUndirected
	}
	bridges, points := undirectedCuts(bg)
	return bridges, points, nil
}

// undirectedCuts walks the neighbor lists stored on the vertices. Those lists
// are the graph's own vertices, in the order they were added, so the search
// does not copy a vertex per edge.
func undirectedCuts[T comparable](g *baseGraph[T]) ([]*Edge[T], []*Vertex[T]) {
	n := len(g.order)
	// Discovery times start at 1, leaving 0 for vertices not visited yet.
	disc := make([]int, n)
	low := make([]int, n)
	isPoint := make([]bool, n)
	bridges := make([]*Edge[T], 0)
	points := make([]*Vertex[T], 0)
	type frame struct {
		v        int
		next     int
		children int
	}
	var stack []frame
	clock := 0
	for _, root := range g.order {
		if root == nil || disc[root.position] != 0 {
			continue
		}
		clock++
		disc[root.position], low[root.position] = clock, clock
		stack = append(stack, frame{v: root.position})
		for len(stack) > 0 {
			i := len(stack) - 1
			v := g.order[stack[i].v]
			if stack[i].next < len(v.neighbors) {
				w := v.neighbors[stack[i].next]
				stack[i].next++
				wi := w.position
				if len(stack) > 1 && wi == stack[len(stack)-2].v {
					continue
				}
				if disc[wi] == 0 {
					stack[i].children++
					clock++
					disc[wi], low[wi] = clock, clock
					stack = append(stack, frame{v: wi})
				} else {
					vi := stack[i].v
					low[vi] = min(low[vi], disc[wi])
				}
				continue
			}
			children := stack[i].children
			vi := stack[i].v
			stack = stack[:i]
			if len(stack) == 0 {
				isPoint[vi] = children > 1
				continue
			}
			p := stack[len(stack)-1].v
			low[p] = min(low[p], low[vi])
			if low[vi] > disc[p] {
				parent, child := g.order[p], g.order[vi]
				bridges = append(bridges, g.edges[parent.label][child.label])
			}
			if len(stack) > 1 && low[vi] >= disc[p] {
				isPoint[p] = true
			}
		}
	}
	for _, v := range g.order {
		if v != nil && isPoint[v.position] {
			points = append(points, v)
		}
	}
	return bridges, points
}
