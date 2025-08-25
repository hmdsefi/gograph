package partition

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/hmdsefi/gograph"
)

// KCutGomoryHu computes an exact k-cut using the Gomory–Hu tree approach
// with Dinic's max-flow algorithm for min-cut computations.
//
// Time Complexity: O(n * maxflow(n, m)) worst-case (Dinic)
// Space Complexity: O(n^2 + m)
//
// Parameters:
//
//	g - input graph implementing Graph[T]
//	k - number of supernodes desired
//
// Returns:
//
//	*KCutResult[T] - supernodes + cut edges
//	error          - if invalid k or empty graph
func KCutGomoryHu[T comparable](g gograph.Graph[T], k int) (*KCutResult[T], error) {
	if g.Order() == 0 {
		return nil, errors.New("graph is empty")
	}
	if k < 2 || k > int(g.Order()) {
		return nil, errors.New("invalid k")
	}

	vertices := g.GetAllVertices()
	treeEdges := gomoryHuTree(g, vertices)

	// pick k-1 smallest weight edges
	sortedEdges := make([]*ghTreeEdge[T], len(treeEdges))
	copy(sortedEdges, treeEdges)
	sort.Slice(
		sortedEdges, func(i, j int) bool {
			return sortedEdges[i].weight < sortedEdges[j].weight
		},
	)
	cutEdgesTree := sortedEdges[:k-1]

	supernodes := reconstructSupernodes(vertices, treeEdges, cutEdgesTree)

	cutEdges := findCutEdges(g, supernodes)

	return &KCutResult[T]{
		Supernodes: supernodes,
		CutEdges:   cutEdges,
	}, nil
}

// ---------------------- PRIVATE HELPERS ----------------------

type ghTreeEdge[T comparable] struct {
	u      *gograph.Vertex[T]
	v      *gograph.Vertex[T]
	weight float64
}

func gomoryHuTree[T comparable](g gograph.Graph[T], vertices []*gograph.Vertex[T]) []*ghTreeEdge[T] {
	n := len(vertices)
	if n <= 1 {
		return []*ghTreeEdge[T]{}
	}

	parent := make([]*gograph.Vertex[T], n)
	for i := 1; i < n; i++ {
		parent[i] = vertices[0]
	}

	var treeEdges []*ghTreeEdge[T]
	for i := 1; i < n; i++ {
		s := vertices[i]
		t := parent[i]

		cutWeight, reachable := minCutDinic(g, s, t)

		treeEdges = append(treeEdges, &ghTreeEdge[T]{u: s, v: t, weight: cutWeight})

		for j := i + 1; j < n; j++ {
			if containsVertex(reachable, vertices[j]) && parent[j] == t {
				parent[j] = s
			}
		}
	}
	return treeEdges
}

// ---------------------- Dinic max-flow ----------------------

type residualGraph[T comparable] struct {
	capacity map[T]map[T]float64
	vertices map[T]*gograph.Vertex[T]
}

func newResidualGraph[T comparable](g gograph.Graph[T]) *residualGraph[T] {
	res := &residualGraph[T]{
		capacity: make(map[T]map[T]float64),
		vertices: make(map[T]*gograph.Vertex[T]),
	}
	for _, v := range g.GetAllVertices() {
		res.vertices[v.Label()] = v
		res.capacity[v.Label()] = make(map[T]float64)
	}

	edges := g.AllEdges()
	for _, e := range edges {
		wt := 1.0
		if g.IsWeighted() {
			wt = e.Weight()
		}
		res.capacity[e.Source().Label()][e.Destination().Label()] += wt
		if !g.IsDirected() {
			res.capacity[e.Destination().Label()][e.Source().Label()] += wt
		}
	}
	return res
}

// minCutDinic computes min s-t cut using Dinic's max-flow
func minCutDinic[T comparable](g gograph.Graph[T], s, t *gograph.Vertex[T]) (float64, []*gograph.Vertex[T]) {
	res := newResidualGraph(g)
	flow := dinicMaxFlow(res, s.Label(), t.Label())
	reachable := residualReachable(res, s.Label())
	return flow, reachable
}

func dinicMaxFlow[T comparable](res *residualGraph[T], s, t T) float64 {
	level := make(map[T]int)
	adj := make(map[T][]T)
	for u, m := range res.capacity {
		for v := range m {
			adj[u] = append(adj[u], v)
		}
	}

	var bfs func() bool
	bfs = func() bool {
		for k := range level {
			level[k] = -1
		}
		queue := []T{s}
		level[s] = 0
		for len(queue) > 0 {
			u := queue[0]
			queue = queue[1:]
			for _, v := range adj[u] {
				if res.capacity[u][v] > 0 && level[v] < 0 {
					level[v] = level[u] + 1
					queue = append(queue, v)
				}
			}
		}
		return level[t] >= 0
	}

	var dfs func(u T, f float64) float64
	dfs = func(u T, f float64) float64 {
		if u == t {
			return f
		}
		for _, v := range adj[u] {
			if res.capacity[u][v] > 0 && level[v] == level[u]+1 {
				df := dfs(v, min(f, res.capacity[u][v]))
				if df > 0 {
					res.capacity[u][v] -= df
					res.capacity[v][u] += df
					return df
				}
			}
		}
		return 0
	}

	var totalFlow float64
	for bfs() {
		for {
			f := dfs(s, math.MaxInt)
			if f == 0 {
				break
			}
			totalFlow += f
		}
	}
	return totalFlow
}

// vertices reachable from s in the residual graph
func residualReachable[T comparable](res *residualGraph[T], s T) []*gograph.Vertex[T] {
	reached := make(map[T]bool)
	queue := []T{s}
	reached[s] = true
	var result []*gograph.Vertex[T]
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for v, cap := range res.capacity[u] {
			if cap > 0 && !reached[v] {
				reached[v] = true
				queue = append(queue, v)
			}
		}
	}
	for label := range reached {
		result = append(result, res.vertices[label])
	}
	return result
}

// ---------------------- Union-Find ----------------------

func reconstructSupernodes[T comparable](
	vertices []*gograph.Vertex[T],
	treeEdges, cutEdges []*ghTreeEdge[T],
) [][]*gograph.Vertex[T] {
	parent := make(map[*gograph.Vertex[T]]*gograph.Vertex[T])
	for _, v := range vertices {
		parent[v] = v
	}

	cutSet := make(map[*ghTreeEdge[T]]bool)
	for _, e := range cutEdges {
		cutSet[e] = true
	}

	for _, e := range treeEdges {
		if !cutSet[e] {
			union(parent, e.u, e.v)
		}
	}

	groups := make(map[*gograph.Vertex[T]][]*gograph.Vertex[T])
	for _, v := range vertices {
		root := find(parent, v)
		groups[root] = append(groups[root], v)
	}

	supernodes := make([][]*gograph.Vertex[T], 0, len(groups))
	for _, grp := range groups {
		supernodes = append(supernodes, grp)
	}
	return supernodes
}

func find[T comparable](parent map[*gograph.Vertex[T]]*gograph.Vertex[T], v *gograph.Vertex[T]) *gograph.Vertex[T] {
	if parent[v] != v {
		parent[v] = find[T](parent, parent[v])
	}
	return parent[v]
}

func union[T comparable](parent map[*gograph.Vertex[T]]*gograph.Vertex[T], u, v *gograph.Vertex[T]) {
	uRoot := find(parent, u)
	vRoot := find(parent, v)
	if uRoot != vRoot {
		parent[vRoot] = uRoot
	}
}

func findCutEdges[T comparable](g gograph.Graph[T], supernodes [][]*gograph.Vertex[T]) []*gograph.Edge[T] {
	vertexToGroup := make(map[T]int)
	for i, group := range supernodes {
		for _, v := range group {
			vertexToGroup[v.Label()] = i
		}
	}

	var cutEdges []*gograph.Edge[T]
	seen := make(map[string]bool)
	for _, e := range g.AllEdges() {
		u := vertexToGroup[e.Source().Label()]
		v := vertexToGroup[e.Destination().Label()]
		if u != v {
			key := fmt.Sprintf("%d-%d", min(u, v), max(u, v))
			if !seen[key] {
				cutEdges = append(cutEdges, e)
				seen[key] = true
			}
		}
	}
	return cutEdges
}

func containsVertex[T comparable](slice []*gograph.Vertex[T], v *gograph.Vertex[T]) bool {
	for _, x := range slice {
		if x == v {
			return true
		}
	}
	return false
}
