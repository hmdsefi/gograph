package path

import (
	"math"

	"github.com/hmdsefi/gograph"
)

// IDAStar returns the cheapest path from start to goal, and its cost, using
// iterative deepening A*. It repeats a depth-first search with a rising limit
// on the estimated total cost, the cost so far plus the heuristic, and keeps
// only the current path. Each new limit is the smallest estimated total that
// was beyond the previous one. Planners use it when the graph of states is too
// large for the frontier of AStar to fit in memory.
//
// cost and heuristic follow the contract of AStar. cost returns the cost of
// using an edge. A nil cost uses the edge weight, which is 0 when the edge has
// none. Costs must be zero or positive: a negative cost on an edge the search
// follows returns ErrNegativeWeight. An edge whose cost is NaN is skipped, and
// a path through a +Inf cost is as long as no path. heuristic estimates the
// remaining cost from a vertex to goal, and a nil heuristic means 0. It's never
// called for goal, whose remaining cost is known to be 0. A NaN estimate counts
// as 0. A vertex with a +Inf estimate, or one whose estimate makes the total
// +Inf, is searched only when no path avoids it: the searches then start over
// with those estimates counted as 0, so the path is still the cheapest one
// through such vertices.
//
// The path is a cheapest one when the heuristic is admissible: it never
// exceeds the true remaining cost. The heuristic isn't checked, and with one
// that overestimates, the path returned may not be the cheapest.
//
// A start or goal that is not in the graph returns
// gograph.ErrVertexDoesNotExist, and a goal that start can't reach returns
// ErrUnreachable. When start equals goal, the path is that one vertex and the
// cost is 0. The path vertices are the graph's own pointers, start first and
// goal last. The search follows outgoing edges on a directed graph, never
// returns to a vertex on the current path, and stops at the first path to
// goal within the limit. Ties break by the order of vertices and edges: the
// search tries the neighbors of a vertex in the order of Neighbors, so the
// same graph and the same functions give the same path on every run. AStar
// takes the larger cost so far first on equal totals, so when several paths
// are cheapest the two can return different ones, and each can read a
// negative cost the other never reaches.
//
// Extra space is O(L·d), where L is the number of vertices on the deepest
// path the search follows, at most V, and d is the most neighbors a vertex on
// it has: the path, and the neighbors still to try at each vertex on it. It
// grows with how deep the search goes, not with how many vertices it reaches.
// The time is the price: each search follows every path whose estimated total
// is within the limit, so a vertex is searched again for each such path to it,
// and the searches repeat for each new limit. With costs that differ by small
// amounts there can be one search for each different estimated total below
// the answer. Where costs are 0, or goal can't be reached, a search can follow
// every simple path from start, which is exponential in the size of the
// graph. AStar is the better choice when its frontier fits in memory, because
// it settles each vertex once.
func IDAStar[T comparable](
	g gograph.Graph[T],
	start, goal T,
	cost func(*gograph.Edge[T]) float64,
	heuristic func(T) float64,
) ([]*gograph.Vertex[T], float64, error) {
	from, to := g.GetVertexByID(start), g.GetVertexByID(goal)
	if from == nil || to == nil {
		return nil, 0, gograph.ErrVertexDoesNotExist
	}
	if from == to {
		return []*gograph.Vertex[T]{from}, 0, nil
	}
	if cost == nil {
		cost = (*gograph.Edge[T]).Weight
	}

	s := idaSearch[T]{
		g:         g,
		goal:      to,
		cost:      cost,
		heuristic: heuristic,
		onPath:    map[*gograph.Vertex[T]]struct{}{},
	}
	limit := s.firstLimit(from)
	for {
		found, err := s.search(from, limit)
		if err != nil {
			return nil, 0, err
		}
		if found {
			return s.path(), s.goalCost, nil
		}
		if !math.IsInf(s.next, 1) {
			limit = s.next
			continue
		}
		if !s.prunedInf || s.infAsZero {
			return nil, 0, ErrUnreachable
		}
		// Every path left goes through a vertex whose estimated total is
		// +Inf. Start over with those estimates counted as 0.
		s.infAsZero = true
		limit = s.firstLimit(from)
	}
}

// firstLimit returns the limit of the first search. The estimate of start is
// never compared with a limit, so a +Inf one gives -Inf: the first search then
// only sets the limit to the smallest estimated total of the neighbors.
func (s *idaSearch[T]) firstLimit(start *gograph.Vertex[T]) float64 {
	if h := s.estimate(start); !math.IsInf(h, 1) {
		return h
	}
	return math.Inf(-1)
}

// idaSearch is the state of one IDAStar call. frames is the current path, and
// pending holds the neighbors of the vertices on it.
type idaSearch[T comparable] struct {
	g         gograph.Graph[T]
	goal      *gograph.Vertex[T]
	cost      func(*gograph.Edge[T]) float64
	heuristic func(T) float64

	frames  []idaFrame[T]
	pending []*gograph.Vertex[T]
	onPath  map[*gograph.Vertex[T]]struct{}

	// next is the smallest estimated total beyond the limit of the last
	// search, and prunedInf whether a vertex was left out for a +Inf one.
	// infAsZero is set once the searches start over with +Inf estimates
	// counted as 0.
	next      float64
	prunedInf bool
	infAsZero bool
	goalCost  float64
}

// idaFrame is a vertex on the path and the cost of the path up to it. Its
// neighbors are pending[start:end], and the ones from next on are still to
// try.
type idaFrame[T comparable] struct {
	vertex *gograph.Vertex[T]
	cost   float64
	start  int
	next   int
	end    int
}

// search runs one depth-first search from start, following the paths whose
// estimated total stays within limit, and reports whether it reached goal.
func (s *idaSearch[T]) search(start *gograph.Vertex[T], limit float64) (bool, error) {
	s.next = math.Inf(1)
	s.prunedInf = false
	s.frames = s.frames[:0]
	s.pending = s.pending[:0]
	clear(s.onPath)
	s.enter(start, 0)

	for len(s.frames) > 0 {
		f := &s.frames[len(s.frames)-1]
		if f.next == f.end {
			delete(s.onPath, f.vertex)
			s.pending = s.pending[:f.start]
			s.frames = s.frames[:len(s.frames)-1]
			continue
		}

		w := s.pending[f.next]
		f.next++
		e := s.g.GetEdge(f.vertex, w)
		if e == nil {
			continue
		}
		// The cost is read before the path check, so a negative cost on an
		// edge back into the path is reported, as AStar does.
		c := s.cost(e)
		if c < 0 {
			return false, ErrNegativeWeight
		}
		if _, ok := s.onPath[w]; ok || math.IsNaN(c) {
			continue
		}
		d := f.cost + c
		if math.IsInf(d, 1) {
			continue
		}

		total := d + s.estimate(w)
		if math.IsInf(total, 1) && s.infAsZero {
			total = d
		}
		if total > limit {
			s.next = min(s.next, total)
			s.prunedInf = s.prunedInf || math.IsInf(total, 1)
			continue
		}
		if w == s.goal {
			s.goalCost = d
			return true, nil
		}
		s.enter(w, d)
	}
	return false, nil
}

// enter puts v on the path with the cost of the path up to it, and queues its
// neighbors.
func (s *idaSearch[T]) enter(v *gograph.Vertex[T], cost float64) {
	start := len(s.pending)
	for n := range v.AllNeighbors() {
		s.pending = append(s.pending, n)
	}
	s.onPath[v] = struct{}{}
	s.frames = append(s.frames, idaFrame[T]{vertex: v, cost: cost, start: start, next: start, end: len(s.pending)})
}

func (s *idaSearch[T]) estimate(v *gograph.Vertex[T]) float64 {
	if s.heuristic == nil || v == s.goal {
		return 0
	}
	h := s.heuristic(v.Label())
	if math.IsNaN(h) || (s.infAsZero && math.IsInf(h, 1)) {
		return 0
	}
	return h
}

// path returns the vertices on the path of the last search, then goal.
func (s *idaSearch[T]) path() []*gograph.Vertex[T] {
	path := make([]*gograph.Vertex[T], len(s.frames)+1)
	for i, f := range s.frames {
		path[i] = f.vertex
	}
	path[len(s.frames)] = s.goal
	return path
}
