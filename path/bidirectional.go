package path

import (
	"container/heap"
	"errors"
	"math"

	"github.com/hmdsefi/gograph"
)

// ErrUnreachable is returned when no path leads from start to goal.
var ErrUnreachable = errors.New("goal is not reachable from start")

// Bidirectional returns the cheapest path from start to goal, and its cost.
//
// cost returns the cost of using an edge. A nil cost uses the edge weight,
// and an edge without a weight costs 0. Costs must be zero or positive: a
// negative cost on an edge the search reaches returns ErrNegativeWeight. An
// edge whose cost is NaN is skipped, and a path through a +Inf cost is as long
// as no path.
//
// One search runs forward from start and one backward from goal. On a
// directed graph the backward search walks incoming edges, and an undirected
// edge is usable in either direction. They stop once the cheapest meeting
// point is settled, so a goal close to start doesn't scan the whole graph.
//
// A start or goal that is not in the graph returns
// gograph.ErrVertexDoesNotExist, and a goal that start can't reach returns
// ErrUnreachable. When start equals goal, the path is that one vertex and the
// cost is 0. The path vertices are the graph's own pointers, start first and
// goal last. Ties break by the order vertices and edges were added, so the
// same graph gives the same path on every run.
//
// Time is O((V+E) log V) in the worst case. Extra space is O(V+E) in the
// worst case, because the queues can hold stale entries.
func Bidirectional[T comparable](
	g gograph.Graph[T],
	start, goal T,
	cost func(*gograph.Edge[T]) float64,
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

	forward, backward := newHalfSearch(from, true), newHalfSearch(to, false)
	best := math.Inf(1)
	var meet *gograph.Vertex[T]

	for forward.clean() && backward.clean() && forward.top()+backward.top() < best {
		side, other := forward, backward
		if backward.top() < forward.top() {
			side, other = backward, forward
		}

		v := side.pop()
		for _, e := range g.EdgesOf(v) {
			next, ok := side.neighbor(e, v)
			if !ok {
				continue
			}
			c := cost(e)
			if c < 0 {
				return nil, 0, ErrNegativeWeight
			}
			if math.IsNaN(c) || math.IsInf(c, 1) || !side.improve(v, next, c) {
				continue
			}
			if d, ok := other.dist[next]; ok && side.dist[next]+d < best {
				best = side.dist[next] + d
				meet = next
			}
		}
	}

	if meet == nil {
		return nil, 0, ErrUnreachable
	}

	path := forward.pathTo(meet)
	var total float64
	for _, v := range path[1:] {
		total += forward.edgeCost[v]
	}
	for v := meet; v != to; v = backward.pred[v] {
		total += backward.edgeCost[v]
		path = append(path, backward.pred[v])
	}
	return path, total, nil
}

// halfSearch is one side of Bidirectional. pred[v] is the vertex before v on
// the cheapest known path from the side's root, toward start for the forward
// side and toward goal for the backward side, and edgeCost[v] is the cost of
// the edge between them.
type halfSearch[T comparable] struct {
	forward  bool
	dist     map[*gograph.Vertex[T]]float64
	pred     map[*gograph.Vertex[T]]*gograph.Vertex[T]
	edgeCost map[*gograph.Vertex[T]]float64
	settled  map[*gograph.Vertex[T]]bool
	queue    halfQueue[T]
	pushed   int
}

func newHalfSearch[T comparable](root *gograph.Vertex[T], forward bool) *halfSearch[T] {
	s := &halfSearch[T]{
		forward:  forward,
		dist:     map[*gograph.Vertex[T]]float64{root: 0},
		pred:     map[*gograph.Vertex[T]]*gograph.Vertex[T]{},
		edgeCost: map[*gograph.Vertex[T]]float64{},
		settled:  map[*gograph.Vertex[T]]bool{},
	}
	s.push(root)
	return s
}

// neighbor returns the vertex e leads to from v in this side's direction, and
// false when e doesn't lead away from v that way.
func (s *halfSearch[T]) neighbor(e *gograph.Edge[T], v *gograph.Vertex[T]) (*gograph.Vertex[T], bool) {
	if s.forward {
		return e.Destination(), e.Source() == v
	}
	return e.Source(), e.Destination() == v
}

// improve records the path to next through v when it is strictly cheaper.
func (s *halfSearch[T]) improve(v, next *gograph.Vertex[T], c float64) bool {
	if s.settled[next] {
		return false
	}
	d := s.dist[v] + c
	if old, ok := s.dist[next]; ok && d >= old {
		return false
	}
	s.dist[next] = d
	s.pred[next] = v
	s.edgeCost[next] = c
	s.push(next)
	return true
}

func (s *halfSearch[T]) push(v *gograph.Vertex[T]) {
	heap.Push(&s.queue, halfEntry[T]{vertex: v, dist: s.dist[v], order: s.pushed})
	s.pushed++
}

// clean drops queue entries for settled vertices and for distances that have
// since improved, and reports whether an entry is left.
func (s *halfSearch[T]) clean() bool {
	for len(s.queue) > 0 {
		e := s.queue[0]
		if !s.settled[e.vertex] && e.dist == s.dist[e.vertex] {
			return true
		}
		heap.Pop(&s.queue)
	}
	return false
}

func (s *halfSearch[T]) top() float64 { return s.queue[0].dist }

func (s *halfSearch[T]) pop() *gograph.Vertex[T] {
	e, _ := heap.Pop(&s.queue).(halfEntry[T])
	v := e.vertex
	s.settled[v] = true
	return v
}

// pathTo returns the forward path from the root to v, root first.
func (s *halfSearch[T]) pathTo(v *gograph.Vertex[T]) []*gograph.Vertex[T] {
	var rev []*gograph.Vertex[T]
	for ; v != nil; v = s.pred[v] {
		rev = append(rev, v)
	}
	path := make([]*gograph.Vertex[T], len(rev))
	for i, x := range rev {
		path[len(rev)-1-i] = x
	}
	return path
}

// halfEntry is a queue entry. Entries with the same distance come out in the
// order they were pushed.
type halfEntry[T comparable] struct {
	vertex *gograph.Vertex[T]
	dist   float64
	order  int
}

type halfQueue[T comparable] []halfEntry[T]

func (q halfQueue[T]) Len() int { return len(q) }

func (q halfQueue[T]) Less(i, j int) bool {
	if q[i].dist != q[j].dist {
		return q[i].dist < q[j].dist
	}
	return q[i].order < q[j].order
}

func (q halfQueue[T]) Swap(i, j int) { q[i], q[j] = q[j], q[i] }

func (q *halfQueue[T]) Push(x any) {
	if e, ok := x.(halfEntry[T]); ok {
		*q = append(*q, e)
	}
}

func (q *halfQueue[T]) Pop() any {
	old := *q
	e := old[len(old)-1]
	*q = old[:len(old)-1]
	return e
}
