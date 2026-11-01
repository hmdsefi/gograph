package gograph

// HasCycle reports whether g contains a cycle. A self-loop is a cycle. In a
// directed graph, a cycle follows the edge directions. In an undirected
// graph, the two stored directions of an edge count as one edge, so any other
// cycle needs at least three vertices.
//
// A graph created with Acyclic refuses edges that would create a cycle, so it
// returns false without a search. A graph from New takes O(V+E) time and O(V)
// extra space, and the search stops at the first cycle it finds. Any other
// graph is read through its methods: TopologySort when it is directed, and
// AllNeighbors when it is undirected. It panics if g is nil.
//
// connectivity.HasCycle calls this.
func HasCycle[T comparable](g Graph[T]) bool {
	bg, ok := g.(*baseGraph[T])
	if g == nil || (ok && bg == nil) {
		panic("gograph: HasCycle requires a non-nil graph")
	}
	if ok {
		switch {
		case bg.properties.isAcyclic:
			return false
		case bg.properties.isDirected:
			return bg.hasDirectedCycle()
		default:
			return bg.hasUndirectedCycle()
		}
	}
	return hasCycleByMethods(g)
}

// hasCycleByMethods checks a graph through TopologySort and Neighbors.
func hasCycleByMethods[T comparable](g Graph[T]) bool {
	if g.IsAcyclic() {
		return false
	}
	if g.IsDirected() {
		_, err := TopologySort(g)
		return err != nil
	}

	visited := make(map[T]bool)
	type entry struct {
		vertex    *Vertex[T]
		parent    T
		hasParent bool
	}

	for _, root := range g.GetAllVertices() {
		if visited[root.Label()] {
			continue
		}

		visited[root.Label()] = true
		stack := []entry{{vertex: root}}
		for len(stack) > 0 {
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			for neighbor := range current.vertex.AllNeighbors() {
				label := neighbor.Label()
				if !visited[label] {
					visited[label] = true
					stack = append(stack, entry{
						vertex:    neighbor,
						parent:    current.vertex.Label(),
						hasParent: true,
					})
					continue
				}
				if !current.hasParent || label != current.parent {
					return true
				}
			}
		}
	}

	return false
}

// hasDirectedCycle runs a depth-first search over the neighbor lists stored
// on the vertices. An edge to a vertex that is still on the search path
// closes a cycle.
func (g *baseGraph[T]) hasDirectedCycle() bool {
	const (
		notVisited = iota
		onPath
		finished
	)
	state := make([]uint8, len(g.order))

	type frame struct {
		vertex *Vertex[T]
		next   int // index of the next neighbor to follow
	}
	var path []frame

	for _, root := range g.order {
		if root == nil || state[root.position] != notVisited {
			continue
		}

		state[root.position] = onPath
		path = append(path, frame{vertex: root})
		for len(path) > 0 {
			top := &path[len(path)-1]
			if top.next == len(top.vertex.neighbors) {
				state[top.vertex.position] = finished
				path = path[:len(path)-1]
				continue
			}

			neighbor := top.vertex.neighbors[top.next]
			top.next++
			switch state[neighbor.position] {
			case onPath:
				return true
			case notVisited:
				state[neighbor.position] = onPath
				path = append(path, frame{vertex: neighbor})
			}
		}
	}

	return false
}

// hasUndirectedCycle runs a breadth-first search over the neighbor lists
// stored on the vertices. Each edge is stored in both directions, so the edge
// back to the vertex the search came from is skipped. Any other edge to a
// vertex the search has already reached closes a cycle.
func (g *baseGraph[T]) hasUndirectedCycle() bool {
	const (
		notReached = -2
		searchRoot = -1
	)

	// parent is the position of the vertex the search reached each vertex
	// from, by position.
	parent := make([]int, len(g.order))
	for i := range parent {
		parent[i] = notReached
	}

	queue := make([]*Vertex[T], 0, len(g.vertices))
	for _, root := range g.order {
		if root == nil || parent[root.position] != notReached {
			continue
		}

		parent[root.position] = searchRoot
		queue = append(queue[:0], root)
		for head := 0; head < len(queue); head++ {
			v := queue[head]
			for _, neighbor := range v.neighbors {
				switch {
				case parent[neighbor.position] == notReached:
					parent[neighbor.position] = v.position
					queue = append(queue, neighbor)
				case neighbor.position != parent[v.position]:
					return true
				}
			}
		}
	}

	return false
}
