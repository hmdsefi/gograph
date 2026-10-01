package partition

import (
	"fmt"
	"math/rand"

	"github.com/hmdsefi/gograph"
)

// KCutResult is the result of RandomizedKCut.
type KCutResult[T comparable] struct {
	// Supernodes are the groups of vertices left after contraction.
	Supernodes [][]*gograph.Vertex[T]
	// CutEdges are the edges of the graph between two different supernodes.
	CutEdges []*gograph.Edge[T]
}

// RandomizedKCut computes an approximate k-cut of an undirected graph using
// a randomized contraction algorithm (generalization of Karger’s min-cut).
//
// The algorithm works as follows:
//  1. Initialize each vertex as its own supernode.
//  2. While the number of supernodes > k:
//     a. Pick a random edge (u, v) from the remaining edges.
//     b. Contract the edge: merge vertices u and v into a single supernode.
//     - Redirect all edges of v to u.
//     - Remove self-loops.
//  3. When exactly k supernodes remain, the remaining edges between supernodes
//     form the approximate k-cut.
//  4. Return both the supernodes and the edges crossing between them.
//
// Notes:
//   - This is a randomized algorithm; different runs may produce different cuts.
//   - Probability of finding the true minimum k-cut decreases with graph size
//     and k. Repeating the algorithm multiple times improves the chances.
//   - Only applicable to undirected graphs; for directed graphs, results are
//     approximate and may not correspond to a global min-cut.
//   - The returned supernodes are slices of vertex pointers representing
//     the contracted vertex groups after k-way partitioning. Supernodes are
//     ordered by their first vertex in GetAllVertices(), and the vertices in
//     each supernode follow the same order.
//   - The returned cut edges are all edges of the original graph that connect
//     different supernodes. In an undirected graph each edge is listed once,
//     in one of its two directions. In a directed graph, u->v and v->u are
//     two edges and both are listed.
//   - Contraction only merges vertices that are connected. If g has more than
//     k connected components, the edges run out first, and the result has one
//     supernode per connected component, which is more than k.
//
// Time Complexity: O(n * m) per run, where n is the number of vertices and m
// is the number of edges in the graph.
//
// Space Complexity: O(n + m) for storing vertex sets and edge lists.
//
// Parameters:
//
//	g - The input graph implementing Graph[T] interface.
//	k - The number of supernodes desired in the partition (k ≥ 2).
//
// Returns:
//
//	*KCutResult[T] - Struct containing:
//	    - Supernodes: slice of supernodes (vertex groups) after contraction.
//	    - CutEdges: edges connecting different supernodes (the k-cut).
//	error - Non-nil if k < 2 or graph has fewer than k vertices.
//
// Example usage:
//
//	g := gograph.New[string]() // undirected, unweighted
//	a := g.AddVertexByLabel("A")
//	b := g.AddVertexByLabel("B")
//	_, _ = g.AddEdge(a, b)
//	result, err := partition.RandomizedKCut(g, 2)
//	if err != nil { log.Fatal(err) }
//	fmt.Println("Supernodes:", result.Supernodes)
//	fmt.Println("Cut edges:", result.CutEdges)
func RandomizedKCut[T comparable](g gograph.Graph[T], k int) (*KCutResult[T], error) {
	if k < 2 {
		return nil, fmt.Errorf("k must be at least 2")
	}

	if int(g.Order()) < k {
		return nil, fmt.Errorf("graph has fewer vertices (%d) than k=%d", g.Order(), k)
	}

	// 1. Initialize each vertex as its own supernode
	supernodes := make(map[T]map[T]*gograph.Vertex[T])  // supernode ID -> set of vertex labels
	vertexToSupernode := make(map[T]*gograph.Vertex[T]) // vertex -> supernode ID
	for _, v := range g.GetAllVertices() {
		vertexToSupernode[v.Label()] = v
		supernodes[v.Label()] = map[T]*gograph.Vertex[T]{v.Label(): v}
	}

	// 2. Collect all edges
	edges := g.AllEdges()
	rand.Shuffle(len(edges), func(i, j int) { edges[i], edges[j] = edges[j], edges[i] })

	// 3. Contract edges randomly until number of supernodes == k
	for len(supernodes) > k {
		if len(edges) == 0 {
			break
		}
		e := edges[0]
		edges = edges[1:]

		u := vertexToSupernode[e.Source().Label()]
		v := vertexToSupernode[e.Destination().Label()]
		if u == v {
			continue // same supernode, skip
		}

		// Merge v into u
		for label, vertex := range supernodes[v.Label()] {
			supernodes[u.Label()][label] = vertex
			vertexToSupernode[label] = u
		}
		delete(supernodes, v.Label())
	}

	// 4. Collect cut edges (edges that connect different supernodes). An
	// undirected graph stores each edge in both directions, so the reverse
	// of an edge that is already in the result is skipped.
	type labelPair struct{ from, to T }
	var cutEdges []*gograph.Edge[T]
	inCut := make(map[labelPair]bool)
	for _, e := range g.AllEdges() {
		from, to := e.Source().Label(), e.Destination().Label()
		if vertexToSupernode[from] == vertexToSupernode[to] {
			continue
		}
		if !g.IsDirected() && inCut[labelPair{to, from}] {
			continue
		}
		inCut[labelPair{from, to}] = true
		cutEdges = append(cutEdges, e)
	}

	// 5. Convert supernodes to slices, in the order of GetAllVertices()
	resultSupernodes := make([][]*gograph.Vertex[T], 0, len(supernodes))
	groupOf := make(map[*gograph.Vertex[T]]int, len(supernodes))
	for _, v := range g.GetAllVertices() {
		supernode := vertexToSupernode[v.Label()]
		i, ok := groupOf[supernode]
		if !ok {
			i = len(resultSupernodes)
			groupOf[supernode] = i
			resultSupernodes = append(resultSupernodes, nil)
		}
		resultSupernodes[i] = append(resultSupernodes[i], v)
	}

	return &KCutResult[T]{
		Supernodes: resultSupernodes,
		CutEdges:   cutEdges,
	}, nil
}
