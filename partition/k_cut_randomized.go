package partition

import (
	"fmt"
	"math/rand"

	"github.com/hmdsefi/gograph"
)

type KCutResult[T comparable] struct {
	Supernodes [][]*gograph.Vertex[T]
	CutEdges   []*gograph.Edge[T]
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
//     the contracted vertex groups after k-way partitioning.
//   - The returned cut edges are edges that connect different supernodes in the
//     original graph.
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
//	g := NewBaseGraph[string](false, false) // undirected, unweighted
//	a := g.AddVertexByLabel("A")
//	b := g.AddVertexByLabel("B")
//	g.AddEdge(a, b)
//	result, err := RandomizedKCut(g, 2)
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

	// 4. Collect cut edges (edges that connect different supernodes)
	var cutEdges []*gograph.Edge[T]

	if len(supernodes) < int(g.Order()) {
		seen := make(map[string]bool) // deduplicate undirected edges
		for _, e := range g.AllEdges() {
			u := vertexToSupernode[e.Source().Label()]
			v := vertexToSupernode[e.Destination().Label()]
			if u != v {
				// canonical key for undirected edge
				var key string
				if fmt.Sprint(u) < fmt.Sprint(v) {
					key = fmt.Sprintf("%v-%v", u, v)
				} else {
					key = fmt.Sprintf("%v-%v", v, u)
				}
				if !seen[key] {
					cutEdges = append(cutEdges, e)
					seen[key] = true
				}
			}
		}
	}

	// 5. Convert supernodes map to slices
	resultSupernodes := make([][]*gograph.Vertex[T], 0, len(supernodes))
	for _, nodes := range supernodes {
		group := make([]*gograph.Vertex[T], 0, len(nodes))
		for _, v := range nodes {
			group = append(group, v)
		}
		resultSupernodes = append(resultSupernodes, group)
	}

	return &KCutResult[T]{
		Supernodes: resultSupernodes,
		CutEdges:   cutEdges,
	}, nil
}
