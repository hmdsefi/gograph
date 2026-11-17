package path

import (
	"math/big"
	"slices"

	"github.com/hmdsefi/gograph"
)

// NegativeCycle returns one cycle whose edge weights sum to less than zero.
// The slice does not repeat the first vertex. The edge from the last vertex
// back to the first closes the cycle. A negative self-loop is a cycle of one
// vertex. The vertices are the graph's own pointers.
//
// A nil slice and a nil error mean the search found no negative cycle. When
// every finite weight is a whole number no larger than 2^53/(V+E) in absolute
// value, that answer is exact.
//
// The search covers the whole graph, not only what one start vertex reaches.
// When several cycles exist, the order of g.AllEdges decides which one is
// returned, so the same graph gives the same cycle on every run.
//
// It returns ErrNotWeighted or ErrNotDirected for a graph that BellmanFord
// doesn't accept either.
//
// An edge with a NaN weight is skipped, as if it weren't in the graph. A path
// through a +Inf weight is as long as no path. A cycle through a -Inf weight is
// a negative cycle.
//
// A cycle's weights are added without rounding to decide if it's negative.
// Weights such as 0.1 are stored rounded, so a cycle that sums to zero in
// decimal can land slightly above or below zero.
//
// Distances are float64. Inside one strongly connected component, finite
// weights that add up past the float64 range, or that are large enough to
// swallow smaller ones, can hide any cycle, even a negative self-loop, and
// BellmanFord can report one where NegativeCycle returns nil.
//
// Time complexity is O(V*E) when E >= V, and it usually stops after a few
// passes when there is a negative cycle. Space complexity is O(V+E).
func NegativeCycle[T comparable](g gograph.Graph[T]) ([]*gograph.Vertex[T], error) {
	if !g.IsWeighted() {
		return nil, ErrNotWeighted
	}

	if !g.IsDirected() {
		return nil, ErrNotDirected
	}

	vertices, _, arcs, negInf := indexEdges(g)
	offset, target, component := allComponents(len(vertices), arcs)

	// A cycle never leaves its strongly connected component. Skipping the arcs
	// between components stops a huge weight on the way into a cycle from
	// swallowing the cycle's own weights.
	arcs = slices.DeleteFunc(arcs, func(a arc) bool { return component[a.from] != component[a.to] })

	cycle := finiteNegativeCycle(len(vertices), arcs)
	if cycle == nil {
		cycle = negativeInfinityCycle(negInf, component, offset, target)
	}
	if cycle == nil {
		return nil, nil
	}

	out := make([]*gograph.Vertex[T], len(cycle))
	for i, v := range cycle {
		out[i] = vertices[v]
	}
	return out, nil
}

// allComponents labels the strongly connected component of each of the n
// vertices. A temporary vertex n with an arc to every vertex lets one search
// reach them all. The adjacency lists it returns include that vertex, which
// nothing leads back to.
func allComponents(n int, arcs []arc) (offset, target, component []int) {
	all := slices.Grow(slices.Clip(arcs), n)
	for v := range n {
		all = append(all, arc{from: n, to: v})
	}

	offset, target = adjacency(n+1, all)
	return offset, target, strongComponents(n, offset, target)
}

// finiteNegativeCycle runs the passes of Bellman-Ford, and after each pass that
// changes a distance looks for a negative cycle in the parent links. A cycle
// with a -Inf weight read as 0 that is negative is negative with -Inf too.
func finiteNegativeCycle(n int, arcs []arc) []int {
	// Starting every distance at 0 is the same as a temporary source with a
	// zero-weight edge to every vertex, so no vertex is out of reach.
	dist := make([]float64, n)
	parent := make([]int, n)
	weight := make([]float64, n)
	seen := make([]int, n)
	for v := range parent {
		parent[v] = -1
	}

	for range n {
		changed := false
		for _, a := range arcs {
			if d := dist[a.from] + a.weight; d < dist[a.to] {
				dist[a.to] = d
				parent[a.to] = a.from
				weight[a.to] = a.weight
				changed = true
			}
		}
		if !changed {
			return nil
		}

		if cycle := parentCycle(parent, weight, seen); cycle != nil {
			return cycle
		}
	}
	return nil
}

// parentCycle returns the first loop in the parent links, in edge order, whose
// weights sum to less than zero. weight[v] is the weight of the arc from
// parent[v] to v. Rounding can close a loop that isn't negative, so each one
// is checked. seen is scratch space of the same length as parent.
func parentCycle(parent []int, weight []float64, seen []int) []int {
	clear(seen)
	for start := range parent {
		v := start
		for v != -1 && seen[v] == 0 {
			seen[v] = start + 1
			v = parent[v]
		}
		if v == -1 || seen[v] != start+1 {
			continue
		}

		cycle, weights := []int{v}, []float64{weight[v]}
		for u := parent[v]; u != v; u = parent[u] {
			cycle = append(cycle, u)
			weights = append(weights, weight[u])
		}
		if exactlyNegative(weights) {
			slices.Reverse(cycle)
			return cycle
		}
	}
	return nil
}

// exactlyNegative reports whether weights sum to less than zero when they are
// added without rounding. 2,200 bits hold the sum of float64 values exactly.
func exactlyNegative(weights []float64) bool {
	sum, x := new(big.Float).SetPrec(2200), new(big.Float).SetPrec(2200)
	for _, w := range weights {
		sum.Add(sum, x.SetFloat64(w))
	}
	return sum.Sign() < 0
}

// negativeInfinityCycle returns a cycle through the first -Inf arc whose ends
// are in the same strongly connected component: the arc's source, then the
// shortest path from its destination back to the source, in hops.
func negativeInfinityCycle(negInf []arc, component, offset, target []int) []int {
	for _, a := range negInf {
		if component[a.from] != component[a.to] {
			continue
		}

		previous := make([]int, len(offset)-1)
		for v := range previous {
			previous[v] = -1
		}
		previous[a.to] = a.to
		queue := []int{a.to}
		for head := 0; previous[a.from] < 0; head++ {
			v := queue[head]
			for _, w := range target[offset[v]:offset[v+1]] {
				if previous[w] < 0 {
					previous[w] = v
					queue = append(queue, w)
				}
			}
		}

		cycle := []int{a.from}
		for v := a.from; v != a.to; {
			v = previous[v]
			cycle = append(cycle, v)
		}
		slices.Reverse(cycle[1:])
		return cycle
	}
	return nil
}
