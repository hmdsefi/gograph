package connectivity

import "github.com/hmdsefi/gograph"

// ConnectedComponents returns the connected components of g. For a directed
// graph, edge directions are ignored, so it returns the weakly connected
// components.
//
// The vertices in the result are the vertices stored in g, not copies. The
// order of the components and of the vertices in them is stable: it only
// depends on the order of the vertices and edges in the graph.
//
// It runs in O(V + E) time and uses O(V) additional space.
func ConnectedComponents[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T] {
	visited := make(map[*gograph.Vertex[T]]bool)
	components := make([][]*gograph.Vertex[T], 0)

	for _, root := range g.GetAllVertices() {
		if visited[root] {
			continue
		}

		visited[root] = true
		queue := []*gograph.Vertex[T]{root}
		component := make([]*gograph.Vertex[T], 0)

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
