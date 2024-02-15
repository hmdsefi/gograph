package path

import (
	"math"

	"github.com/hmdsefi/gograph"
)

// FloydWarshall finds the shortest paths between all pairs of vertices in a
// weighted graph, even in the presence of negative weight edges (as long as
// there are no negative weight cycles). It was proposed by Robert Floyd and
// Stephen Warshall.
//
// Steps:
//
//  1. Initialization: Create a distance matrix D[][] where D[i][j] represents the
//     shortest distance between vertex i and vertex j. Initialize this matrix with
//     the weights of the edges between vertices if there is an edge, otherwise set
//     the value to infinity. Also, set the diagonal elements D[i][i] to 0.
//
//  2. Shortest Path Calculation: Iterate through all vertices as intermediate vertices.
//     For each pair of vertices (i, j), check if going through the current intermediate
//     vertex k leads to a shorter path than the current known distance from i to j. If so,
//     update the distance matrix D[i][j] to the new shorter distance D[i][k] + D[k][j].
//
//  3. Detection of Negative Cycles: After the iterations, if any diagonal element D[i][i]
//     of the distance matrix is negative, it indicates the presence of a negative weight cycle
//     in the graph.
//
//  4. Output: The resulting distance matrix D[][] will contain the shortest path distances
//     between all pairs of vertices. If there is a negative weight cycle, it might not produce
//     the correct shortest paths, but it can still detect the presence of such cycles.
//
// The time complexity of the Floyd-Warshall algorithm is O(V^3), where V is the
// number of vertices in the graph. Despite its cubic time complexity, it is often
// preferred over other algorithms like Bellman-Ford for dense graphs or when the
// graph has negative weight edges and no negative weight cycles, as it calculates
// shortest paths between all pairs of vertices in one go.
func FloydWarshall[T comparable](g gograph.Graph[T]) (map[T]map[T]float64, error) {
	if !g.IsWeighted() {
		return nil, ErrNotWeighted
	}

	if !g.IsDirected() {
		return nil, ErrNotDirected
	}

	vertices := g.GetAllVertices()

	dist := make(map[T]map[T]float64)
	maxValue := math.Inf(1)
	for _, source := range vertices {
		for _, dest := range vertices {
			destMap, ok := dist[source.Label()]
			if !ok {
				destMap = make(map[T]float64)
			}

			destMap[dest.Label()] = maxValue
			if dest.Label() == source.Label() {
				destMap[dest.Label()] = 0
			}

			if edge := g.GetEdge(source, dest); edge != nil {
				destMap[dest.Label()] = edge.Weight()
			}

			dist[source.Label()] = destMap
		}
	}

	for _, intermediate := range vertices {
		for _, source := range vertices {
			for _, dest := range vertices {
				weight := dist[source.Label()][intermediate.Label()] + dist[intermediate.Label()][dest.Label()]
				if weight < dist[source.Label()][dest.Label()] {
					dist[source.Label()][dest.Label()] = weight
				}
			}
		}
	}

	edges := g.AllEdges()
	for _, v := range vertices {
		for _, edge := range edges {
			if dist[v.Label()][edge.Source().Label()] != maxValue &&
				dist[v.Label()][edge.Source().Label()]+edge.Weight() < dist[v.Label()][edge.Destination().Label()] {
				return nil, ErrNegativeWeightCycle
			}
		}
	}

	return dist, nil
}
