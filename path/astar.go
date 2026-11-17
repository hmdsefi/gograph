package path

import (
	"container/heap"
	"math"

	"github.com/hmdsefi/gograph"
)

// AStar returns the cheapest path from start to goal, and its cost.
//
// cost returns the cost of using an edge. A nil cost uses the edge weight,
// which is 0 when the edge has none. Costs must be zero or positive: a negative
// cost on an edge the search reaches returns ErrNegativeWeight. An edge whose
// cost is NaN is skipped, and a path through a +Inf cost is as long as no path.
//
// heuristic estimates the remaining cost from a vertex to goal. A nil heuristic
// means 0, which makes the search Dijkstra's algorithm. The estimate must not
// exceed the true remaining cost, or the path returned may not be the cheapest.
// It's called at most once for each vertex, and never for goal, whose remaining
// cost is known to be 0. A NaN estimate counts as 0, and a vertex with a +Inf
// estimate is searched last.
//
// A start or goal that is not in the graph returns
// gograph.ErrVertexDoesNotExist, and a goal that start can't reach returns
// ErrUnreachable. When start equals goal, the path is that one vertex and the
// cost is 0. The path vertices are the graph's own pointers, start first and
// goal last. The search follows outgoing edges on a directed graph and stops
// when goal is settled, so it doesn't read the rest of the graph.
//
// When two vertices have the same estimated total, the one with the larger cost
// so far goes first, since it's closer to goal. After that, ties break by the
// order vertices and edges were added, so the same graph and the same functions
// give the same path on every run.
//
// For a graph from gograph.New, time is O((V+E) log V) when the heuristic is
// consistent, meaning that it never drops by more than the cost of an edge: each
// vertex is expanded at most once. With a heuristic that only stays within the
// true remaining cost, a vertex can be expanded again when a cheaper way to it
// turns up. The search then follows Martelli's algorithm B, which makes O(V²)
// expansions where plain A* can make exponentially many, for
// O(V·(V+E) log V) time. Extra space is O(V). V and E count only what the
// search reaches.
func AStar[T comparable](
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

	s := astarSearch[T]{
		goal:      to,
		heuristic: heuristic,
		nodes:     map[*gograph.Vertex[T]]*astarNode[T]{},
		byTotal:   astarQueue[T]{byTotal: true},
		floor:     math.Inf(-1),
	}
	root := &astarNode[T]{vertex: from, index: -1, estimate: s.estimate(from)}
	root.total = root.estimate
	s.nodes[from] = root
	s.enqueue(root)

	for {
		n := s.next()
		if n == nil {
			return nil, 0, ErrUnreachable
		}
		if n.vertex == to {
			return n.path(), n.cost, nil
		}

		for neighbor := range n.vertex.AllNeighbors() {
			e := g.GetEdge(n.vertex, neighbor)
			if e == nil {
				continue
			}
			c := cost(e)
			if c < 0 {
				return nil, 0, ErrNegativeWeight
			}
			if math.IsNaN(c) {
				continue
			}
			if d := n.cost + c; !math.IsInf(d, 1) {
				s.relax(n, neighbor, d)
			}
		}
	}
}

// astarSearch is the state of one AStar call. Each vertex has one node, and a
// node is in at most one queue, so the state stays within O(V).
//
// floor is the largest estimated total expanded so far. byTotal holds the nodes
// at or above it, and byCost those below it. A node falls below the floor only
// when a heuristic is inconsistent, and those nodes go first, cheapest cost
// first. That is what bounds the number of expansions. With a consistent
// heuristic the totals never fall, byCost stays empty, and this is plain A*.
type astarSearch[T comparable] struct {
	goal      *gograph.Vertex[T]
	heuristic func(T) float64
	nodes     map[*gograph.Vertex[T]]*astarNode[T]
	byTotal   astarQueue[T]
	byCost    astarQueue[T]
	floor     float64
	updates   int
}

// next removes the node to expand from its queue, and returns nil when both
// queues are empty.
func (s *astarSearch[T]) next() *astarNode[T] {
	if s.byCost.Len() > 0 {
		n, _ := heap.Pop(&s.byCost).(*astarNode[T])
		return n
	}
	if s.byTotal.Len() == 0 {
		return nil
	}
	n, _ := heap.Pop(&s.byTotal).(*astarNode[T])
	s.floor = n.total
	return n
}

// relax records the path to v through n when it costs d and is strictly cheaper
// than the cheapest path known so far.
func (s *astarSearch[T]) relax(n *astarNode[T], v *gograph.Vertex[T], d float64) {
	m, ok := s.nodes[v]
	if !ok {
		m = &astarNode[T]{vertex: v, index: -1, estimate: s.estimate(v)}
		s.nodes[v] = m
	} else if d >= m.cost {
		return
	}

	m.cost = d
	m.total = d + m.estimate
	m.pred = n
	s.enqueue(m)
}

func (s *astarSearch[T]) estimate(v *gograph.Vertex[T]) float64 {
	if s.heuristic == nil || v == s.goal {
		return 0
	}
	h := s.heuristic(v.Label())
	if math.IsNaN(h) {
		return 0
	}
	return h
}

// enqueue puts n in the queue its total belongs to, or moves it there when it's
// already queued. The total of a queued node only goes down, so a node can move
// from byTotal to byCost and never back.
func (s *astarSearch[T]) enqueue(n *astarNode[T]) {
	s.updates++
	n.order = s.updates

	q := &s.byTotal
	if n.total < s.floor {
		q = &s.byCost
	}
	if n.queue == q {
		heap.Fix(q, n.index)
		return
	}
	if n.queue != nil {
		heap.Remove(n.queue, n.index)
	}
	heap.Push(q, n)
}

// astarNode is the search state of one vertex. pred is the node before it on
// the cheapest known path from start. cost is the cost of that path, estimate
// the cached heuristic, and total their sum. order is when the node last
// improved, which breaks ties. queue and index say where the node is queued.
type astarNode[T comparable] struct {
	vertex   *gograph.Vertex[T]
	pred     *astarNode[T]
	queue    *astarQueue[T]
	cost     float64
	estimate float64
	total    float64
	order    int
	index    int
}

// path returns the vertices from start to n, start first.
func (n *astarNode[T]) path() []*gograph.Vertex[T] {
	size := 0
	for x := n; x != nil; x = x.pred {
		size++
	}
	path := make([]*gograph.Vertex[T], size)
	for x := n; x != nil; x = x.pred {
		size--
		path[size] = x.vertex
	}
	return path
}

// astarQueue is an indexed binary heap of nodes. Each node knows its index, so
// a cheaper path updates its entry in place and no stale entries build up.
type astarQueue[T comparable] struct {
	nodes   []*astarNode[T]
	byTotal bool
}

func (q *astarQueue[T]) Len() int { return len(q.nodes) }

func (q *astarQueue[T]) Less(i, j int) bool {
	a, b := q.nodes[i], q.nodes[j]
	if q.byTotal {
		if a.total != b.total {
			return a.total < b.total
		}
		if a.cost != b.cost {
			return a.cost > b.cost
		}
	} else if a.cost != b.cost {
		return a.cost < b.cost
	}
	return a.order < b.order
}

func (q *astarQueue[T]) Swap(i, j int) {
	q.nodes[i], q.nodes[j] = q.nodes[j], q.nodes[i]
	q.nodes[i].index = i
	q.nodes[j].index = j
}

func (q *astarQueue[T]) Push(x any) {
	n, _ := x.(*astarNode[T])
	n.queue = q
	n.index = len(q.nodes)
	q.nodes = append(q.nodes, n)
}

func (q *astarQueue[T]) Pop() any {
	last := len(q.nodes) - 1
	n := q.nodes[last]
	q.nodes[last] = nil
	q.nodes = q.nodes[:last]
	n.queue = nil
	n.index = -1
	return n
}
