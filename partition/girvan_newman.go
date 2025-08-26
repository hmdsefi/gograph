package partition

import (
	"errors"

	"github.com/hmdsefi/gograph"
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
//  2. Identify edges with maximum betweenness.
//  3. Remove one or more edges with the highest betweenness.
//  4. Update connected components using a non-recursive BFS traversal.
//  5. Repeat steps 1-4 until the desired number of components (`k`) is reached.
//     If k <= 0, continue until all edges are removed.
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

	components := getConnectedComponents(working)
	for (k > 0 && len(components) < k) || (k <= 0 && working.Size() > 0) {
		// Compute edge betweenness
		betweenness := calculateEdgeBetweenness(working)
		if len(betweenness) == 0 {
			break
		}

		// Find max betweenness
		maxVal := -1.0
		var edgesToRemove []*gograph.Edge[T]
		for e, val := range betweenness {
			if val > maxVal {
				maxVal = val
				edgesToRemove = []*gograph.Edge[T]{e}
			} else if val == maxVal {
				edgesToRemove = append(edgesToRemove, e)
			}
		}

		// Remove edges
		working.RemoveEdges(edgesToRemove...)

		components = getConnectedComponents(working)
		if k > 0 && len(components) >= k {
			break
		}
	}

	// Convert components to Graph[T] objects
	result := make([]gograph.Graph[T], len(components))
	for i, comp := range components {
		subgraph := gograph.New[T]()
		// Add vertices
		for _, v := range comp {
			subgraph.AddVertexByLabel(v.Label(), func(p *gograph.VertexProperties) {})
		}
		// Add edges
		for _, v := range comp {
			for _, e := range g.EdgesOf(v) {
				if subgraph.ContainsVertex(e.Source()) && subgraph.ContainsVertex(e.Destination()) {
					_, _ = subgraph.AddEdge(
						subgraph.GetVertexByID(e.Source().Label()),
						subgraph.GetVertexByID(e.Destination().Label()),
						gograph.WithEdgeWeight(e.Weight()),
					)
				}
			}
		}
		result[i] = subgraph
	}

	return result, nil
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
		)
	}
	return clone
}

// getConnectedComponents returns slices of vertices representing each connected component (non-recursive)
func getConnectedComponents[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T] {
	visited := make(map[*gograph.Vertex[T]]bool)
	var components [][]*gograph.Vertex[T]
	var queue []*gograph.Vertex[T]

	for _, v := range g.GetAllVertices() {
		if !visited[v] {
			var comp []*gograph.Vertex[T]
			queue = append(queue, v)
			visited[v] = true
			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]
				comp = append(comp, curr)
				for _, e := range g.EdgesOf(curr) {
					neighbor := e.OtherVertex(curr.Label())
					if neighbor != nil && !visited[neighbor] {
						visited[neighbor] = true
						queue = append(queue, neighbor)
					}
				}
			}
			components = append(components, comp)
		}
	}

	return components
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
