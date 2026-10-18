package gograph

// ConnectedComponents returns the connected components of g. For a directed
// graph, edge directions are ignored, so it returns the weakly connected
// components. The vertices in the result are the vertices stored in g, not
// copies, and an empty graph returns an empty slice.
//
// The components are in the order of their first vertex in GetAllVertices.
// In an undirected graph, the vertices of a component are in breadth-first
// order from that vertex, following the order the edges were added. In a
// directed graph, they are in GetAllVertices order. The components share one
// backing array, but appending to one doesn't change the others.
//
// A graph from New is read from the neighbor lists stored on its vertices.
// An undirected one takes O(V+E) time. A directed one takes O((V+E)·α(V))
// time, where α is the inverse Ackermann function, which is below 5 for any
// graph that fits in memory. Both use O(V) extra space. Other directed Graph
// implementations are read through GetAllVertices and AllEdges, also in
// O((V+E)·α(V)) time. Other undirected Graph implementations are read through
// GetAllVertices and EdgesOf in O(V²) time. It panics if g is nil.
//
// connectivity.ConnectedComponents calls this.
func ConnectedComponents[T comparable](g Graph[T]) [][]*Vertex[T] {
	bg, ok := g.(*baseGraph[T])
	if g == nil || (ok && bg == nil) {
		panic("gograph: ConnectedComponents requires a non-nil graph")
	}
	if ok {
		if bg.properties.isDirected {
			return bg.weakComponents()
		}
		return bg.components()
	}
	return componentsByEdges(g)
}

// componentsByEdges searches through EdgesOf. EdgesOf walks every vertex,
// so this is O(V²).
func componentsByEdges[T comparable](g Graph[T]) [][]*Vertex[T] {
	if g.IsDirected() {
		return directedComponentsByEdges(g)
	}

	visited := make(map[*Vertex[T]]bool)
	components := make([][]*Vertex[T], 0)

	for _, root := range g.GetAllVertices() {
		if visited[root] {
			continue
		}

		visited[root] = true
		queue := []*Vertex[T]{root}
		component := make([]*Vertex[T], 0)

		for len(queue) > 0 {
			vertex := queue[0]
			queue = queue[1:]
			component = append(component, vertex)

			for _, edge := range g.EdgesOf(vertex) {
				neighbor := edge.OtherVertex(vertex.Label())
				if neighbor == nil || visited[neighbor] {
					continue
				}

				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}

		components = append(components, component)
	}

	return components
}

// directedComponentsByEdges joins the endpoints of every edge, ignoring its
// direction, then groups vertices in GetAllVertices order.
func directedComponentsByEdges[T comparable](g Graph[T]) [][]*Vertex[T] {
	vertices := g.GetAllVertices()
	indices := make(map[*Vertex[T]]int, len(vertices))
	parent := make([]int, len(vertices))
	for i, vertex := range vertices {
		indices[vertex] = i
		parent[i] = -1
	}

	for _, edge := range g.AllEdges() {
		from, fromOK := indices[edge.Source()]
		to, toOK := indices[edge.Destination()]
		if !fromOK || !toOK {
			continue
		}

		a, b := findRoot(parent, from), findRoot(parent, to)
		if a == b {
			continue
		}

		// Attach the smaller set to the larger one.
		if parent[a] > parent[b] {
			a, b = b, a
		}
		parent[a] += parent[b]
		parent[b] = a
	}

	count := 0
	for _, p := range parent {
		if p < 0 {
			count++
		}
	}

	// index is one more than the component of each root, or 0 until the
	// first vertex of the set is placed.
	index := make([]int, len(vertices))
	members := make([]*Vertex[T], len(vertices))
	components := make([][]*Vertex[T], 0, count)
	next := 0
	for i, vertex := range vertices {
		root := findRoot(parent, i)
		if index[root] == 0 {
			size := -parent[root]
			components = append(components, members[next:next:next+size])
			next += size
			index[root] = len(components)
		}

		c := index[root] - 1
		components[c] = append(components[c], vertex)
	}

	return components
}

// components runs a breadth-first search over the neighbor lists stored on
// the vertices. An undirected graph stores each edge in both directions, so
// the neighbor list of a vertex has every vertex it shares an edge with.
func (g *baseGraph[T]) components() [][]*Vertex[T] {
	// comp is the component of each vertex, by position, or -1 until the
	// search reaches it.
	comp := make([]int, len(g.order))
	for i := range comp {
		comp[i] = -1
	}

	// The search order is the result: each component is a run of the queue.
	queue := make([]*Vertex[T], 0, len(g.vertices))
	count := 0
	for _, root := range g.order {
		if root == nil || comp[root.position] >= 0 {
			continue
		}

		comp[root.position] = count
		queue = append(queue, root)
		for head := len(queue) - 1; head < len(queue); head++ {
			for _, neighbor := range queue[head].neighbors {
				if comp[neighbor.position] < 0 {
					comp[neighbor.position] = count
					queue = append(queue, neighbor)
				}
			}
		}
		count++
	}

	components := make([][]*Vertex[T], 0, count)
	start := 0
	for end := 1; end <= len(queue); end++ {
		if end == len(queue) || comp[queue[end].position] != comp[queue[start].position] {
			components = append(components, queue[start:end:end])
			start = end
		}
	}

	return components
}

// weakComponents joins the two ends of every edge with union-find, so it
// doesn't need a list of the edges that end at each vertex.
func (g *baseGraph[T]) weakComponents() [][]*Vertex[T] {
	// parent links each vertex, by position, toward the root of its set. A
	// root holds minus the size of its set.
	parent := make([]int, len(g.order))
	for i := range parent {
		parent[i] = -1
	}

	for _, v := range g.order {
		if v == nil {
			continue
		}

		for _, neighbor := range v.neighbors {
			a, b := findRoot(parent, v.position), findRoot(parent, neighbor.position)
			if a == b {
				continue
			}

			// attach the smaller set to the larger one
			if parent[a] > parent[b] {
				a, b = b, a
			}
			parent[a] += parent[b]
			parent[b] = a
		}
	}

	count := 0
	for i, v := range g.order {
		if v != nil && parent[i] < 0 {
			count++
		}
	}

	// index is one more than the component of each root, or 0 until the
	// first vertex of the set is placed.
	index := make([]int, len(g.order))
	vertices := make([]*Vertex[T], len(g.vertices))
	components := make([][]*Vertex[T], 0, count)
	next := 0
	for i, v := range g.order {
		if v == nil {
			continue
		}

		root := findRoot(parent, i)
		if index[root] == 0 {
			size := -parent[root]
			components = append(components, vertices[next:next:next+size])
			next += size
			index[root] = len(components)
		}

		c := index[root] - 1
		components[c] = append(components[c], v)
	}

	return components
}

// findRoot returns the root of the set that x is in, and halves the path
// from x to it.
func findRoot(parent []int, x int) int {
	for parent[x] >= 0 {
		if grandparent := parent[parent[x]]; grandparent >= 0 {
			parent[x] = grandparent
		}
		x = parent[x]
	}

	return x
}
