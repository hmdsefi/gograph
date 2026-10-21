package partition

import (
	"errors"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/connectivity"
)

// GirvanNewman implements the Girvan-Newman community detection algorithm for
// undirected graphs. It partitions a graph into communities by iteratively
// removing edges with the highest betweenness centrality until either the
// specified number of connected components `k` is reached, or no edges remain
// if k <= 0.
//
// The algorithm proceeds as follows:
//  1. Compute edge betweenness centrality for all edges using Brandes algorithm.
//     - BFS is used from each vertex to determine the shortest paths.
//     - Dependencies are accumulated to assign betweenness values to edges.
//  2. Identify the edge with the highest betweenness. When several edges tie,
//     take the first one in the order of AllEdges(). Values within one part
//     in a billion of the highest count as tied, since sums of the same
//     paths added in a different order can round differently.
//  3. Remove that edge.
//  4. Update connected components using a non-recursive BFS traversal.
//  5. Repeat steps 1-4 until the desired number of components (`k`) is reached.
//     If k <= 0, continue until all edges are removed.
//
// Removing one edge adds at most one component, so the result has exactly k
// communities when g has at least k vertices and at most k connected
// components. If g already has more than k connected components, they are
// returned as they are. If g has fewer than k vertices, every vertex ends up
// in its own community.
//
// Important details:
//   - High betweenness edges typically act as bridges between clusters and are
//     removed first, effectively separating communities.
//   - BFS is used exclusively for both shortest path computation and component
//     identification to avoid recursion and stack overflow issues.
//   - The input graph `g` is cloned internally to prevent mutation.
//
// Parameters:
//   - g: An instance of gograph.Graph[T] representing the undirected input graph.
//   - k: The desired number of communities (connected components). If k <= 0,
//     the algorithm continues removing edges until no edges remain.
//
// Returns:
//   - A slice of gograph.Graph[T], each representing a connected component (community).
//     The order of the communities and of the vertices in them is stable: a
//     graph built the same way gives the same result on every run.
//     Each community has every edge of g between its vertices, including
//     edges the algorithm removed while splitting the graph. The communities
//     keep the vertex and edge weights, and are weighted if g is weighted.
//   - An error if the operation fails.
//
// Time Complexity:
//   - Each iteration requires computing betweenness for all edges using BFS from each node.
//   - Let V = number of vertices, E = number of edges.
//   - Computing betweenness: O(V * (V + E)) per iteration.
//   - In the worst case, up to O(E) edges can be removed one by one.
//   - Total worst-case time complexity: O(E * V * (V + E)).
//
// Space Complexity:
//   - Storing the cloned graph: O(V + E)
//   - BFS traversal queues and dependency maps: O(V + E)
//   - Storing betweenness values for edges: O(E)
//   - Total space complexity: O(V + E)
func GirvanNewman[T comparable](g gograph.Graph[T], k int) ([]gograph.Graph[T], error) {
	if g == nil {
		return nil, errors.New("input graph is nil")
	}

	// Clone graph
	working := cloneGraph(g)

	components := connectivity.ConnectedComponents(working)
	for (k > 0 && len(components) < k) || (k <= 0 && working.Size() > 0) {
		// Compute edge betweenness
		betweenness := calculateEdgeBetweenness(working)
		if len(betweenness) == 0 {
			break
		}

		// Only one edge is removed per iteration: removing every tied edge at
		// once can split the graph into more than k components.
		edgeToRemove := mostCentralEdge(working, betweenness)
		if edgeToRemove == nil {
			break
		}

		working.RemoveEdges(edgeToRemove)

		components = connectivity.ConnectedComponents(working)
		if k > 0 && len(components) >= k {
			break
		}
	}

	var options []gograph.GraphOptionFunc
	if g.IsWeighted() {
		options = append(options, gograph.Weighted())
	}

	// Convert components to Graph[T] objects
	result := make([]gograph.Graph[T], len(components))
	for i, comp := range components {
		subgraph := gograph.New[T](options...)
		// Add vertices
		for _, v := range comp {
			subgraph.AddVertexByLabel(v.Label(), gograph.WithVertexWeight(v.Weight()))
		}
		// Add edges
		for _, v := range comp {
			for _, e := range g.EdgesOf(v) {
				if subgraph.ContainsVertex(e.Source()) && subgraph.ContainsVertex(e.Destination()) {
					_, _ = subgraph.AddEdge(
						subgraph.GetVertexByID(e.Source().Label()),
						subgraph.GetVertexByID(e.Destination().Label()),
						gograph.WithEdgeWeight(e.Weight()),
						gograph.WithEdgeLabel(e.Label()),
					)
				}
			}
		}
		result[i] = subgraph
	}

	return result, nil
}

// tieTolerance is the relative difference within which two betweenness
// values count as tied.
const tieTolerance = 1e-9

// mostCentralEdge returns the first edge in AllEdges order whose betweenness
// is within tieTolerance of the highest, or nil when no edge has a value.
// Going through the edges in graph order keeps the result independent of
// map iteration. An undirected edge gets the sum of its two stored
// directions, so both directions have the same value.
func mostCentralEdge[T comparable](g gograph.Graph[T], betweenness map[*gograph.Edge[T]]float64) *gograph.Edge[T] {
	edges := g.AllEdges()
	values := make([]float64, len(edges))
	maxVal := -1.0
	for i, e := range edges {
		values[i] = -1
		if forward, ok := betweenness[e]; ok {
			values[i] = forward + betweenness[g.GetEdge(e.Destination(), e.Source())]
		}
		if values[i] > maxVal {
			maxVal = values[i]
		}
	}
	if maxVal < 0 {
		return nil
	}

	for i, e := range edges {
		if values[i] >= maxVal*(1-tieTolerance) {
			return e
		}
	}
	return nil
}

// cloneGraph deep-copies the graph
func cloneGraph[T comparable](g gograph.Graph[T]) gograph.Graph[T] {
	clone := gograph.New[T]()
	vertexMap := make(map[T]*gograph.Vertex[T])
	for _, v := range g.GetAllVertices() {
		vClone := clone.AddVertexByLabel(v.Label(), gograph.WithVertexWeight(v.Weight()))
		vertexMap[v.Label()] = vClone
	}
	for _, e := range g.AllEdges() {
		_, _ = clone.AddEdge(
			vertexMap[e.Source().Label()],
			vertexMap[e.Destination().Label()],
			gograph.WithEdgeWeight(e.Weight()),
			gograph.WithEdgeLabel(e.Label()),
		)
	}
	return clone
}

// calculateEdgeBetweenness computes edge betweenness centrality using Brandes' algorithm
func calculateEdgeBetweenness[T comparable](g gograph.Graph[T]) map[*gograph.Edge[T]]float64 {
	betweenness := make(map[*gograph.Edge[T]]float64)
	vertices := g.GetAllVertices()

	for _, s := range vertices {
		// BFS structures
		distance := make(map[*gograph.Vertex[T]]int)
		sigma := make(map[*gograph.Vertex[T]]float64)
		pred := make(map[*gograph.Vertex[T]][]*gograph.Vertex[T])

		for _, v := range vertices {
			distance[v] = -1
			sigma[v] = 0
			pred[v] = []*gograph.Vertex[T]{}
		}
		sigma[s] = 1
		distance[s] = 0

		queue := []*gograph.Vertex[T]{s}
		var stack []*gograph.Vertex[T]

		// BFS
		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			stack = append(stack, v)

			for _, e := range g.EdgesOf(v) {
				w := e.OtherVertex(v.Label())
				if w == nil {
					continue
				}
				if distance[w] < 0 {
					distance[w] = distance[v] + 1
					queue = append(queue, w)
				}
				if distance[w] == distance[v]+1 {
					sigma[w] += sigma[v]
					pred[w] = append(pred[w], v)
				}
			}
		}

		// Accumulation
		delta := make(map[*gograph.Vertex[T]]float64)
		for len(stack) > 0 {
			w := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, v := range pred[w] {
				c := (sigma[v] / sigma[w]) * (1 + delta[w])
				edge := g.GetEdge(v, w)
				if edge != nil {
					betweenness[edge] += c
				}
				delta[v] += c
			}
		}
	}

	// Halve the counts for undirected edges
	for e := range betweenness {
		betweenness[e] /= 2
	}

	return betweenness
}
