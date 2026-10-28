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
// The graph must have the Weighted option or ErrNotWeighted is returned.
// Unreachable vertices have distance math.Inf(1).
//
// Steps:
//
//  1. Initialization: Create a distance matrix D[][] where D[i][j] represents the
//     shortest distance between vertex i and vertex j. Initialize this matrix with
//     the weights of the edges between vertices if there is an edge, otherwise set
//     the value to infinity. Also, set the diagonal elements D[i][i] to 0. A self-loop
//     only changes D[i][i] if its weight is negative.
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
// An edge with a NaN weight is skipped, as if it weren't in the graph. A path
// through a +Inf weight is as long as no path. A path through a -Inf weight
// has a distance of -Inf, and a cycle through a -Inf weight is a negative
// cycle. Finite weights that add up below the float64 range give a distance
// of -Inf, which returns ErrNegativeWeightCycle.
//
// The time complexity of the Floyd-Warshall algorithm is O(V^3), where V is the
// number of vertices in the graph. Despite its cubic time complexity, it is often
// preferred over other algorithms like Bellman-Ford for dense graphs or when the
// graph has negative weight edges and no negative weight cycles, as it calculates
// shortest paths between all pairs of vertices in one go. With -Inf weights, each
// vertex that one of them leads to adds O(V^2) time, so the total stays O(V^3). It
// uses O(V^2) space.
func FloydWarshall[T comparable](g gograph.Graph[T]) (map[T]map[T]float64, error) {
	if !g.IsWeighted() {
		return nil, ErrNotWeighted
	}

	if !g.IsDirected() {
		return nil, ErrNotDirected
	}

	vertices, _, arcs, negInf := indexEdges(g)
	n := len(vertices)

	// dist[i*n+j] is the distance from vertices[i] to vertices[j]. Each -Inf
	// weight is read as 0 until the negative cycles are ruled out.
	inf := math.Inf(1)
	dist := make([]float64, n*n)
	for i := range dist {
		dist[i] = inf
	}
	for i := range n {
		dist[i*n+i] = 0
	}
	for _, a := range arcs {
		if a.weight < dist[a.from*n+a.to] {
			dist[a.from*n+a.to] = a.weight
		}
	}

	for k := range n {
		through := dist[k*n : (k+1)*n]
		for i := range n {
			row := dist[i*n : (i+1)*n]
			// a path from i through k is no shorter than no path
			if row[k] == inf {
				continue
			}
			for j, d := range through {
				if weight := row[k] + d; weight < row[j] {
					row[j] = weight
				}
			}
		}
	}

	for i := range n {
		if dist[i*n+i] < 0 { // Check for negative cycles using diagonal
			return nil, ErrNegativeWeightCycle
		}
	}
	if hasNegativeInfinity(dist) {
		return nil, ErrNegativeWeightCycle
	}

	// a -Inf weight is on a cycle when its destination reaches its source
	for _, a := range negInf {
		if dist[a.to*n+a.from] != inf {
			return nil, ErrNegativeWeightCycle
		}
	}
	spreadNegativeInfinityAll(dist, n, negInf)

	out := make(map[T]map[T]float64, n)
	for i, source := range vertices {
		row := make(map[T]float64, n)
		for j, dest := range vertices {
			row[dest.Label()] = dist[i*n+j]
		}
		out[source.Label()] = row
	}

	return out, nil
}

// spreadNegativeInfinityAll sets dist[i*n+j] to -Inf when a path from
// vertex i to vertex j goes through a -Inf weight. Reachability reads as a
// distance below +Inf, which holds for -Inf too, so the rows can change in
// place. It takes O(V*len(negInf)) time, plus O(V^2) for each vertex that a
// -Inf weight leads to.
func spreadNegativeInfinityAll(dist []float64, n int, negInf []arc) {
	if len(negInf) == 0 {
		return
	}

	inf, negInfinity := math.Inf(1), math.Inf(-1)
	heads := make([]int, 0, len(negInf))
	added := make([]int, n) // i+1 once vertex is in heads for row i
	for i := range n {
		row := dist[i*n : (i+1)*n]
		heads = heads[:0]
		for _, a := range negInf {
			if row[a.from] != inf && added[a.to] != i+1 {
				added[a.to] = i + 1
				heads = append(heads, a.to)
			}
		}

		for _, w := range heads {
			for j, d := range dist[w*n : (w+1)*n] {
				if d != inf {
					row[j] = negInfinity
				}
			}
		}
	}
}
