package path

import "github.com/hmdsefi/gograph"

func Dijkstra[T comparable](g gograph.Graph[T], start T) map[T]float64 {
	const Inf float64 = 1<<31 - 1

	dist := make(map[T]float64)

	startVertex := g.GetVertexByID(start)
	if startVertex == nil {
		return dist
	}

	vertices := g.GetAllVertices()
	for _, v := range vertices {
		dist[v.Label()] = Inf
	}

	dist[start] = 0
	visited := make(map[T]bool)
	for len(visited) < len(vertices) {
		var u *gograph.Vertex[T]
		for _, v := range vertices {
			if !visited[v.Label()] && (u == nil || dist[v.Label()] < dist[u.Label()]) {
				u = v
			}
		}
		visited[u.Label()] = true
		neighbors := u.Neighbors()
		for _, neighbor := range neighbors {
			edge := g.GetEdge(u, neighbor)
			if alt := dist[u.Label()] + edge.Weight(); alt < dist[edge.Destination().Label()] {
				dist[edge.Destination().Label()] = alt
			}
		}
	}
	return dist
}
