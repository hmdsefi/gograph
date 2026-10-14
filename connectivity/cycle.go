package connectivity

import "github.com/hmdsefi/gograph"

// HasCycle reports whether g contains a cycle. Directed graphs are checked
// using their edge directions. In undirected graphs, the two stored directions
// of an edge are treated as one edge.
//
// It runs in O(V + E) time and uses O(V) additional space.
func HasCycle[T comparable](g gograph.Graph[T]) bool {
	if g.IsDirected() {
		_, err := gograph.TopologySort(g)
		return err != nil
	}

	visited := make(map[T]bool)
	type entry struct {
		vertex    *gograph.Vertex[T]
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

			for _, neighbor := range current.vertex.Neighbors() {
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
