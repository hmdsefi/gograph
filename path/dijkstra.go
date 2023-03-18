package path

import (
	"math"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/util"
)

// DijkstraSimple is a simple implementation of Dijkstra's algorithm. It's using
// a simple slice to keep track of unvisited vertices, and selected the vertex
// with the smallest tentative distance using linear search.
//
// The time complexity of the simple Dijkstra's algorithm implementation is O(V^2).
//
// It returns the shortest distances from the starting vertex to all other vertices
// in the graph.
func DijkstraSimple[T comparable](g gograph.Graph[T], start T) map[T]float64 {
	dist := make(map[T]float64)

	startVertex := g.GetVertexByID(start)
	if startVertex == nil {
		return dist
	}

	vertices := g.GetAllVertices()
	for _, v := range vertices {
		dist[v.Label()] = math.MaxFloat64
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

// dVertex represents dijkstra vertex.
type dVertex[T comparable] struct {
	label   T
	dist    float64
	visited bool
	prev    T
}

func newDVertex[T comparable](label T) *dVertex[T] {
	return &dVertex[T]{
		label:   label,
		dist:    math.MaxFloat64,
		visited: false,
	}
}

// Dijkstra is a standard implementation of Dijkstra's Algorithm that uses a
// min heap as a priority queue to find the shortest path between start vertex
// and all other vertices in the specified graph.
//
// The time complexity of the standard Dijkstra's algorithm with a min heap is O((E+V)logV).
//
// It returns the shortest distances from the starting vertex to all other vertices
// in the graph.
func Dijkstra[T comparable](g gograph.Graph[T], start T) map[T]float64 {
	startVertex := g.GetVertexByID(start)
	if startVertex == nil {
		return make(map[T]float64)
	}

	// Initialize the heap and the visited map
	pq := util.NewVertexPriorityQueue[T]()
	visited := make(map[T]bool)

	// Initialize the start vertex
	verticesMap := make(map[T]*dVertex[T])
	vertices := g.GetAllVertices()
	for _, v := range vertices {
		verticesMap[v.Label()] = newDVertex(v.Label())
	}

	verticesMap[start].dist = 0

	// Add the start vertex to the heap
	pq.Push(util.NewVertexWithPriority(g.GetVertexByID(start), verticesMap[start].dist))

	// Main loop
	for pq.Len() > 0 {
		// Extract the vertex with the smallest tentative distance from the heap
		curr := pq.Pop()
		visited[curr.Vertex().Label()] = true

		// Update the distances of its neighbors
		neighbors := curr.Vertex().Neighbors()
		for _, v := range neighbors {
			if !visited[v.Label()] {
				neighbor := verticesMap[v.Label()]
				newDist := curr.Priority() + g.GetEdge(curr.Vertex(), v).Weight()
				if newDist < neighbor.dist {
					neighbor.dist = newDist
					neighbor.prev = curr.Vertex().Label()
					pq.Push(util.NewVertexWithPriority(v, verticesMap[v.Label()].dist))
				}
			}
		}
	}

	// Return the distances from the start vertex to each other vertex
	distances := make(map[T]float64)
	for _, v := range verticesMap {
		distances[v.label] = v.dist
	}

	return distances
}
