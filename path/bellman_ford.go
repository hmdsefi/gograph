package path

import (
	"errors"
	"math"

	"github.com/hmdsefi/gograph"
)

var (
	ErrNegativeWeightCycle = errors.New("graph contains negative weight cycle")
	ErrNotDirected         = gograph.ErrNotDirected
	ErrNotWeighted         = errors.New("graph is not weighted")
)

// BellmanFord finds the shortest path from a source vertex to all other vertices
// in a weighted graph, even in the presence of negative weight edges, as long as
// there are no negative weight cycle.
//
// The graph must have the Weighted option or ErrNotWeighted is returned.
// Unreachable vertices have distance math.Inf(1). Only negative cycles
// reachable from the source, and the overflow described below, produce
// ErrNegativeWeightCycle.
//
// Steps:
//
//		1.Initialization: Start by setting the distance of the source vertex to itself
//	    as 0, and the distance of all other vertices to infinity.
//
//		2. Relaxation: Iterate through all edges in the graph |V| - 1 times, where |V|
//		is the number of vertices. In each iteration, attempt to improve the shortest
//		path estimates for all vertices. This is done by relaxing each edge: if the distance
//		to the destination vertex through the current edge is shorter than the current
//		estimate, update the estimate with the shorter distance.
//
//		3.Detection of Negative Cycles: After the |V| - 1 iterations, perform an additional
//		iteration. If during this iteration, any of the distances are further reduced, it
//		indicates the presence of a negative weight cycle in the graph. This is because if a
//		vertex's distance can still be improved after |V| - 1 iterations, it means there's a
//		negative weight cycle that can be traversed indefinitely to reduce the distance further.
//
//		4.Output: If there is no negative weight cycle, the algorithm outputs the shortest path
//		distances from the source vertex to all other vertices. If there is a negative weight
//		cycle, the algorithm typically returns an indication of this fact.
//
// An edge with a NaN weight is skipped, as if it weren't in the graph. A path
// through a +Inf weight is as long as no path. A path through a -Inf weight
// has a distance of -Inf, and a cycle through a -Inf weight is a negative
// cycle. Finite weights that add up below the float64 range give a distance
// of -Inf, which returns ErrNegativeWeightCycle.
//
// The time complexity of the Bellman-Ford algorithm is O(V*E), where V is the number of vertices
// and E is the number of edges. The passes stop after one that changes no distance, so when
// every shortest path has at most k edges, they take O(k*E) time. With -Inf weights, finding the
// cycles through them and the vertices they lead to takes O(V+E) more time. It uses O(V+E)
// space.
//
// If the start vertex isn't in the graph, it returns gograph.ErrVertexDoesNotExist.
func BellmanFord[T comparable](g gograph.Graph[T], start T) (map[T]float64, error) {
	if !g.IsWeighted() {
		return nil, ErrNotWeighted
	}

	if !g.IsDirected() {
		return nil, ErrNotDirected
	}

	if g.GetVertexByID(start) == nil {
		return nil, gograph.ErrVertexDoesNotExist
	}

	vertices, index, arcs, negInf := indexEdges(g)

	dist := make([]float64, len(vertices))
	for i := range dist {
		dist[i] = math.Inf(1)
	}
	dist[index[start]] = 0

	// With each -Inf weight read as 0, a -Inf distance comes from finite
	// weights that overflow, and every cycle that stays negative has finite
	// weights only.
	if !relax(dist, arcs) || hasNegativeInfinity(dist) {
		return nil, ErrNegativeWeightCycle
	}

	if !spreadNegativeInfinity(dist, index[start], arcs, negInf) {
		return nil, ErrNegativeWeightCycle
	}

	out := make(map[T]float64, len(vertices))
	for i, v := range vertices {
		out[v.Label()] = dist[i]
	}

	return out, nil
}

// relax runs the passes of Bellman-Ford over arcs, and reports false when a
// pass after the first len(dist)-1 still shortens a distance, which means a
// negative cycle. It stops after a pass that changes nothing.
func relax(dist []float64, arcs []arc) bool {
	inf := math.Inf(1)
	for range len(dist) - 1 {
		changed := false
		for _, a := range arcs {
			if d := dist[a.from]; d != inf && d+a.weight < dist[a.to] {
				dist[a.to] = d + a.weight
				changed = true
			}
		}

		if !changed {
			return true
		}
	}

	for _, a := range arcs {
		if d := dist[a.from]; d != inf && d+a.weight < dist[a.to] {
			return false
		}
	}

	return true
}

// spreadNegativeInfinity takes the distances computed with each -Inf weight
// read as 0, and sets the distance of every vertex that a path from start
// through a -Inf weight reaches to -Inf. It reports false when one of those
// -Inf weights is on a cycle: both ends of the edge are then in the same
// strongly connected component. It takes O(V+E) time and space when a -Inf
// weight is reachable, and O(len(negInf)) otherwise.
func spreadNegativeInfinity(dist []float64, start int, arcs, negInf []arc) bool {
	inf := math.Inf(1)
	var heads []int
	for _, a := range negInf {
		if dist[a.from] != inf {
			heads = append(heads, a.to)
		}
	}
	if len(heads) == 0 {
		return true
	}

	offset, target := adjacency(len(dist), arcs)
	component := strongComponents(start, offset, target)
	for _, a := range negInf {
		if dist[a.from] != inf && component[a.from] == component[a.to] {
			return false
		}
	}

	negInfinity := math.Inf(-1)
	stack := heads[:0]
	for _, v := range heads {
		if dist[v] != negInfinity {
			dist[v] = negInfinity
			stack = append(stack, v)
		}
	}
	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, w := range target[offset[v]:offset[v+1]] {
			if dist[w] != negInfinity {
				dist[w] = negInfinity
				stack = append(stack, w)
			}
		}
	}

	return true
}

// adjacency returns the arcs as lists of destinations: the arcs from vertex
// v end at target[offset[v]:offset[v+1]], in the order of arcs.
func adjacency(n int, arcs []arc) (offset, target []int) {
	offset = make([]int, n+1)
	for _, a := range arcs {
		offset[a.from+1]++
	}
	for v := range n {
		offset[v+1] += offset[v]
	}

	target = make([]int, len(arcs))
	for _, a := range arcs {
		target[offset[a.from]] = a.to
		offset[a.from]++
	}
	copy(offset[1:], offset[:n])
	offset[0] = 0

	return offset, target
}

// strongComponents returns the strongly connected component of each vertex
// that start reaches, as a number shared by the vertices of one component,
// and -1 for the other vertices. It runs Tarjan's algorithm with an explicit
// stack in O(V+E) time and space.
func strongComponents(start int, offset, target []int) []int {
	n := len(offset) - 1
	component := make([]int, n)
	discovered := make([]int, n) // 0 until found, then the order found, from 1
	low := make([]int, n)
	for v := range component {
		component[v] = -1
	}

	type frame struct{ v, next int }
	frames := []frame{{v: start, next: offset[start]}}
	open := []int{start}
	discovered[start], low[start] = 1, 1
	found, components := 1, 0

	for len(frames) > 0 {
		f := &frames[len(frames)-1]
		v := f.v
		if f.next < offset[v+1] {
			w := target[f.next]
			f.next++
			switch {
			case discovered[w] == 0:
				found++
				discovered[w], low[w] = found, found
				open = append(open, w)
				frames = append(frames, frame{v: w, next: offset[w]})
			case component[w] < 0:
				low[v] = min(low[v], discovered[w])
			}
			continue
		}

		frames = frames[:len(frames)-1]
		if len(frames) > 0 {
			parent := frames[len(frames)-1].v
			low[parent] = min(low[parent], low[v])
		}

		if low[v] == discovered[v] {
			for {
				w := open[len(open)-1]
				open = open[:len(open)-1]
				component[w] = components
				if w == v {
					break
				}
			}
			components++
		}
	}

	return component
}
