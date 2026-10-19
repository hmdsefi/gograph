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
// A graph from New is searched by vertex position. Any other Graph is read
// through GetAllVertices and GetEdge, with its vertices looked up by label.
// It panics if g is nil.
//
// connectivity.Bridges and connectivity.ArticulationPoints call this.
func UndirectedCuts[T comparable](g Graph[T]) ([]*Edge[T], []*Vertex[T], error) {
	bg, ok := g.(*baseGraph[T])
	if g == nil || (ok && bg == nil) {
		panic("gograph: UndirectedCuts requires a non-nil graph")
	}
	if g.IsDirected() {
		return nil, nil, ErrNotUndirected
	}
	if ok {
		bridges, points := undirectedCuts(bg.order, nil, func(from, to *Vertex[T]) *Edge[T] {
			return bg.edges[from.label][to.label]
		})
		return bridges, points, nil
	}

	vertices := g.GetAllVertices()
	index := make(map[T]int, len(vertices))
	for i, v := range vertices {
		index[v.label] = i
	}
	bridges, points := undirectedCuts(vertices, index, g.GetEdge)
	return bridges, points, nil
}

// undirectedCuts walks the neighbor lists stored on the vertices, so the
// search does not copy a vertex per edge. A neighbor is found by its position
// in order when index is nil, and by its label otherwise. Neighbors missing
// from index are skipped. order may contain nil entries for removed vertices.
func undirectedCuts[T comparable](
	order []*Vertex[T],
	index map[T]int,
	edge func(from, to *Vertex[T]) *Edge[T],
) ([]*Edge[T], []*Vertex[T]) {
	n := len(order)
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
	for ri, root := range order {
		if root == nil || disc[ri] != 0 {
			continue
		}
		clock++
		disc[ri], low[ri] = clock, clock
		stack = append(stack, frame{v: ri})
		for len(stack) > 0 {
			i := len(stack) - 1
			v := order[stack[i].v]
			if stack[i].next < len(v.neighbors) {
				w := v.neighbors[stack[i].next]
				stack[i].next++
				wi := w.position
				if index != nil {
					var found bool
					if wi, found = index[w.label]; !found {
						continue
					}
				}
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
				bridges = append(bridges, edge(order[p], order[vi]))
			}
			if len(stack) > 1 && low[vi] >= disc[p] {
				isPoint[p] = true
			}
		}
	}
	for i, v := range order {
		if v != nil && isPoint[i] {
			points = append(points, v)
		}
	}
	return bridges, points
}
