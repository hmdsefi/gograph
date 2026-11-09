package path

import (
	"container/heap"
	"math"

	"github.com/hmdsefi/gograph"
)

// KShortest returns up to k cheapest simple paths from start to goal, cheapest
// first. A simple path does not repeat a vertex. Fewer than k paths is not an
// error: the slice is shorter, and it's empty when goal can't be reached.
//
// cost returns the cost of using an edge. A nil cost uses the edge weight,
// which is 0 when the edge has none. Costs must be zero or positive: a negative
// cost on an edge reachable from start returns ErrNegativeWeight. An edge whose
// cost is NaN is skipped, and a path through a +Inf cost is as long as no path.
// cost is called once for each edge that leaves a vertex start can reach,
// except self-loops, which a simple path never uses.
//
// k below 1 returns ErrInvalidK, and a start or goal that is not in the graph
// returns gograph.ErrVertexDoesNotExist. When start equals goal, the result is
// one path of that one vertex for any k, since a simple path can't leave and
// come back. The path vertices are the graph's own pointers, start first and
// goal last. Directed graphs follow outgoing edges, and an undirected edge is
// usable in either direction. Paths of equal cost come in the order the search
// finds them, which follows the order vertices and edges were added, so the
// same graph gives the same paths on every run.
//
// It's Yen's algorithm with Lawler's change: each path only branches from the
// vertex where it left the path it came from, so no path is found twice. Each
// branch search is A* guided by the exact cost to goal in the whole graph,
// found once by a backward search. At most k-1 candidate paths are kept, and a
// branch that can't beat the worst of them is skipped.
//
// Time is O(k·V·(V+E) log V) in the worst case. Extra space is O(V+E) for the
// part of the graph start can reach, plus O(k·V) for the at most k-1 candidate
// paths of up to V vertices each. V and E count only what start can reach.
func KShortest[T comparable](
	g gograph.Graph[T],
	start, goal T,
	k int,
	cost func(*gograph.Edge[T]) float64,
) ([][]*gograph.Vertex[T], error) {
	if k < 1 {
		return nil, ErrInvalidK
	}
	from, to := g.GetVertexByID(start), g.GetVertexByID(goal)
	if from == nil || to == nil {
		return nil, gograph.ErrVertexDoesNotExist
	}
	if from == to {
		return [][]*gograph.Vertex[T]{{from}}, nil
	}
	if cost == nil {
		cost = (*gograph.Edge[T]).Weight
	}

	s, err := newKShortestSearch(g, from, to, cost)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return [][]*gograph.Vertex[T]{}, nil
	}
	return s.run(k), nil
}

// kShortestSearch holds the part of the graph start can reach as arrays. The
// arcs that leave vertex v are head[v] up to head[v+1]. Vertex 0 is start.
type kShortestSearch[T comparable] struct {
	vertices []*gograph.Vertex[T]
	head     []int32
	arcTo    []int32
	arcCost  []float64
	goal     int32
	// toGoal is the cheapest cost from each vertex to goal in the whole graph,
	// +Inf when there is none. It's the exact heuristic for the branch searches:
	// a branch only removes vertices and edges, which can't make a path cheaper.
	toGoal []float64

	// Branch search state, reused by every search. A value is current when its
	// stamp equals gen, so a search doesn't have to clear the arrays.
	gen        uint32
	seen       []uint32
	blocked    []uint32
	arcBlocked []uint32
	dist       []float64
	total      []float64
	pred       []int32
	predArc    []int32
	order      []int32
	updates    int32
	queue      vertexHeap
}

// newKShortestSearch reads the part of g that from reaches, and returns nil
// when goal is not in it.
func newKShortestSearch[T comparable](
	g gograph.Graph[T],
	from, to *gograph.Vertex[T],
	cost func(*gograph.Edge[T]) float64,
) (*kShortestSearch[T], error) {
	s := &kShortestSearch[T]{vertices: []*gograph.Vertex[T]{from}, goal: -1}
	index := map[*gograph.Vertex[T]]int32{from: 0}

	for u := 0; u < len(s.vertices); u++ {
		v := s.vertices[u]
		s.head = append(s.head, int32(len(s.arcTo))) //nolint:gosec // an index fits in int32
		for next := range v.AllNeighbors() {
			if next == v {
				continue // a simple path never uses a self-loop
			}
			e := g.GetEdge(v, next)
			if e == nil {
				continue
			}
			c := cost(e)
			if c < 0 {
				return nil, ErrNegativeWeight
			}
			if math.IsNaN(c) || math.IsInf(c, 1) {
				continue
			}
			i, ok := index[next]
			if !ok {
				i = int32(len(s.vertices)) //nolint:gosec // an index fits in int32
				index[next] = i
				s.vertices = append(s.vertices, next)
			}
			s.arcTo = append(s.arcTo, i)
			s.arcCost = append(s.arcCost, c)
		}
	}
	s.head = append(s.head, int32(len(s.arcTo))) //nolint:gosec // an index fits in int32

	goal, ok := index[to]
	if !ok {
		return nil, nil
	}
	s.goal = goal
	s.costsToGoal()

	n, m := len(s.vertices), len(s.arcTo)
	s.seen = make([]uint32, n)
	s.blocked = make([]uint32, n)
	s.arcBlocked = make([]uint32, m)
	s.dist = make([]float64, n)
	s.total = make([]float64, n)
	s.pred = make([]int32, n)
	s.predArc = make([]int32, n)
	s.order = make([]int32, n)
	s.queue = vertexHeap{pos: make([]int32, n), less: s.branchLess}
	for i := range s.queue.pos {
		s.queue.pos[i] = -1
	}
	return s, nil
}

// costsToGoal sets toGoal with Dijkstra's algorithm over the reversed arcs.
func (s *kShortestSearch[T]) costsToGoal() {
	n := len(s.vertices)
	// reverse arcs, grouped by their destination
	revHead := make([]int32, n+1)
	for _, v := range s.arcTo {
		revHead[v+1]++
	}
	for v := range n {
		revHead[v+1] += revHead[v]
	}
	revFrom := make([]int32, len(s.arcTo))
	revArc := make([]int32, len(s.arcTo))
	fill := append([]int32(nil), revHead[:n]...)
	for u := range n {
		for a := s.head[u]; a < s.head[u+1]; a++ {
			v := s.arcTo[a]
			revFrom[fill[v]] = int32(u) //nolint:gosec // an index fits in int32
			revArc[fill[v]] = a
			fill[v]++
		}
	}

	s.toGoal = make([]float64, n)
	for i := range s.toGoal {
		s.toGoal[i] = math.Inf(1)
	}
	s.toGoal[s.goal] = 0

	q := vertexHeap{pos: make([]int32, n)}
	q.less = func(a, b int32) bool { return s.toGoal[a] < s.toGoal[b] }
	for i := range q.pos {
		q.pos[i] = -1
	}
	q.push(s.goal)
	for q.len() > 0 {
		v := q.pop()
		for r := revHead[v]; r < revHead[v+1]; r++ {
			u := revFrom[r]
			d := s.toGoal[v] + s.arcCost[revArc[r]]
			if d < s.toGoal[u] {
				s.toGoal[u] = d
				q.update(u)
			}
		}
	}
}

// branchLess orders the branch search queue the way A* does: by estimated
// total, then by the larger cost so far, then by when the vertex last improved.
func (s *kShortestSearch[T]) branchLess(a, b int32) bool {
	if s.total[a] != s.total[b] {
		return s.total[a] < s.total[b]
	}
	if s.dist[a] != s.dist[b] {
		return s.dist[a] > s.dist[b]
	}
	return s.order[a] < s.order[b]
}

// branch returns the arcs of the cheapest path from v to goal that avoids the
// blocked vertices and arcs of the current gen, starting at a cost of base. It
// gives up once the cheapest total left reaches bound, and returns false when
// there is no such path.
func (s *kShortestSearch[T]) branch(v int32, base, bound float64) ([]int32, float64, bool) {
	defer s.queue.clear()

	s.updates = 0
	s.seen[v] = s.gen
	s.dist[v] = base
	s.total[v] = base + s.toGoal[v]
	s.queue.push(v)

	for s.queue.len() > 0 {
		u := s.queue.pop()
		if s.total[u] >= bound {
			return nil, 0, false
		}
		if u == s.goal {
			return s.arcsTo(v), s.dist[u], true
		}

		for a := s.head[u]; a < s.head[u+1]; a++ {
			w := s.arcTo[a]
			if s.arcBlocked[a] == s.gen || s.blocked[w] == s.gen || math.IsInf(s.toGoal[w], 1) {
				continue
			}
			d := s.dist[u] + s.arcCost[a]
			if math.IsInf(d, 1) || (s.seen[w] == s.gen && d >= s.dist[w]) {
				continue
			}
			s.seen[w] = s.gen
			s.dist[w] = d
			s.total[w] = d + s.toGoal[w]
			s.pred[w] = u
			s.predArc[w] = a
			s.updates++
			s.order[w] = s.updates
			s.queue.update(w)
		}
	}
	return nil, 0, false
}

// arcsTo returns the arcs from v to goal along predArc.
func (s *kShortestSearch[T]) arcsTo(v int32) []int32 {
	size := 0
	for u := s.goal; u != v; u = s.pred[u] {
		size++
	}
	arcs := make([]int32, size)
	for u := s.goal; u != v; u = s.pred[u] {
		size--
		arcs[size] = s.predArc[u]
	}
	return arcs
}

// run finds up to k paths.
func (s *kShortestSearch[T]) run(k int) [][]*gograph.Vertex[T] {
	s.nextGen()
	arcs, total, ok := s.branch(0, 0, math.Inf(1))
	if !ok {
		return [][]*gograph.Vertex[T]{}
	}

	var cands candidateQueue
	seq := 0
	cands.add(&kCandidate{cost: total, arcs: arcs, seq: seq})

	paths := make([][]*gograph.Vertex[T], 0, min(k, 16))
	for len(paths) < k && cands.len() > 0 {
		c := cands.popMin()
		paths = append(paths, s.verticesOf(c.arcs))
		limit := k - len(paths)
		if limit == 0 {
			break
		}

		// Lawler's change: branch only from c.deviation on. Before it, c shares
		// its vertices with the path it came from, whose branches are already
		// candidates or found.
		prefix := 0.0
		u := int32(0)
		for i := 0; i < c.deviation; i++ {
			prefix += s.arcCost[c.arcs[i]]
			u = s.arcTo[c.arcs[i]]
		}
		for i := c.deviation; i < len(c.arcs); i++ {
			bound := math.Inf(1)
			if cands.len() >= limit {
				bound = cands.worst().cost
			}

			// The first i vertices of c are fixed, so the cheapest this branch
			// can be is prefix plus the cost from u to goal in the whole graph.
			// A later candidate loses ties, so equal is not enough.
			if prefix+s.toGoal[u] < bound {
				excluded := &excludedArc{arc: c.arcs[i]}
				if i == c.deviation {
					excluded.next = c.excluded
				}
				s.nextGen()
				s.blockPrefix(c.arcs[:i])
				for x := excluded; x != nil; x = x.next {
					s.arcBlocked[x.arc] = s.gen
				}

				if rest, total, ok := s.branch(u, prefix, bound); ok {
					arcs := make([]int32, i+len(rest))
					copy(arcs, c.arcs[:i])
					copy(arcs[i:], rest)
					seq++
					cands.add(&kCandidate{cost: total, arcs: arcs, deviation: i, excluded: excluded, seq: seq})
					if cands.len() > limit {
						cands.removeWorst()
					}
				}
			}

			prefix += s.arcCost[c.arcs[i]]
			u = s.arcTo[c.arcs[i]]
		}
	}
	return paths
}

// nextGen starts a new branch search, which makes every stamp stale. When the
// counter would wrap, it clears the stamps instead.
func (s *kShortestSearch[T]) nextGen() {
	if s.gen == math.MaxUint32 {
		clear(s.seen)
		clear(s.blocked)
		clear(s.arcBlocked)
		s.gen = 0
	}
	s.gen++
}

// blockPrefix blocks start and the vertices that arcs lead to, which includes
// the vertex the branch starts from.
func (s *kShortestSearch[T]) blockPrefix(arcs []int32) {
	s.blocked[0] = s.gen
	for _, a := range arcs {
		s.blocked[s.arcTo[a]] = s.gen
	}
}

func (s *kShortestSearch[T]) verticesOf(arcs []int32) []*gograph.Vertex[T] {
	path := make([]*gograph.Vertex[T], len(arcs)+1)
	path[0] = s.vertices[0]
	for i, a := range arcs {
		path[i+1] = s.vertices[s.arcTo[a]]
	}
	return path
}

// vertexHeap is an indexed binary min-heap of vertex indexes. pos[v] is the
// position of v in items, or -1 when v is not queued.
type vertexHeap struct {
	items []int32
	pos   []int32
	less  func(a, b int32) bool
}

func (q *vertexHeap) len() int { return len(q.items) }

func (q *vertexHeap) push(v int32) {
	q.pos[v] = int32(len(q.items)) //nolint:gosec // an index fits in int32
	q.items = append(q.items, v)
	q.up(len(q.items) - 1)
}

// update moves v up after its key went down, or pushes it if it's not queued.
func (q *vertexHeap) update(v int32) {
	if q.pos[v] < 0 {
		q.push(v)
		return
	}
	q.up(int(q.pos[v]))
}

func (q *vertexHeap) pop() int32 {
	v := q.items[0]
	last := len(q.items) - 1
	q.swap(0, last)
	q.items = q.items[:last]
	q.pos[v] = -1
	q.down(0)
	return v
}

// clear empties the heap in O(len).
func (q *vertexHeap) clear() {
	for _, v := range q.items {
		q.pos[v] = -1
	}
	q.items = q.items[:0]
}

func (q *vertexHeap) up(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !q.less(q.items[i], q.items[parent]) {
			return
		}
		q.swap(i, parent)
		i = parent
	}
}

func (q *vertexHeap) down(i int) {
	n := len(q.items)
	for {
		smallest, left := i, 2*i+1
		if left < n && q.less(q.items[left], q.items[smallest]) {
			smallest = left
		}
		if right := left + 1; right < n && q.less(q.items[right], q.items[smallest]) {
			smallest = right
		}
		if smallest == i {
			return
		}
		q.swap(i, smallest)
		i = smallest
	}
}

func (q *vertexHeap) swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
	q.pos[q.items[i]] = int32(i) //nolint:gosec // an index fits in int32
	q.pos[q.items[j]] = int32(j) //nolint:gosec // an index fits in int32
}

// kCandidate is a path that may be returned. Its first deviation arcs are the
// same as the path it branched from, and excluded lists the arcs it can't take
// from the vertex it branched at, which keeps every candidate distinct.
type kCandidate struct {
	cost      float64
	arcs      []int32
	deviation int
	excluded  *excludedArc
	seq       int
	minIndex  int
	maxIndex  int
}

// excludedArc is a list of arcs. Candidates that branch at the same vertex
// share its tail.
type excludedArc struct {
	arc  int32
	next *excludedArc
}

// candidateLess orders candidates by cost, then by when they were found.
func candidateLess(a, b *kCandidate) bool {
	if a.cost != b.cost {
		return a.cost < b.cost
	}
	return a.seq < b.seq
}

// candidateQueue keeps the candidates in a min-heap and a max-heap at once, so
// it can return the best one and drop the worst one in O(log k).
type candidateQueue struct {
	min candidateMinHeap
	max candidateMaxHeap
}

func (q *candidateQueue) len() int { return len(q.min) }

func (q *candidateQueue) add(c *kCandidate) {
	heap.Push(&q.min, c)
	heap.Push(&q.max, c)
}

func (q *candidateQueue) popMin() *kCandidate {
	c, _ := heap.Pop(&q.min).(*kCandidate)
	heap.Remove(&q.max, c.maxIndex)
	return c
}

func (q *candidateQueue) worst() *kCandidate { return q.max[0] }

func (q *candidateQueue) removeWorst() {
	c, _ := heap.Pop(&q.max).(*kCandidate)
	heap.Remove(&q.min, c.minIndex)
}

type candidateMinHeap []*kCandidate

func (h candidateMinHeap) Len() int           { return len(h) }
func (h candidateMinHeap) Less(i, j int) bool { return candidateLess(h[i], h[j]) }
func (h candidateMinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].minIndex = i
	h[j].minIndex = j
}

func (h *candidateMinHeap) Push(x any) {
	c, _ := x.(*kCandidate)
	c.minIndex = len(*h)
	*h = append(*h, c)
}

func (h *candidateMinHeap) Pop() any {
	old := *h
	c := old[len(old)-1]
	old[len(old)-1] = nil
	*h = old[:len(old)-1]
	return c
}

type candidateMaxHeap []*kCandidate

func (h candidateMaxHeap) Len() int           { return len(h) }
func (h candidateMaxHeap) Less(i, j int) bool { return candidateLess(h[j], h[i]) }
func (h candidateMaxHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].maxIndex = i
	h[j].maxIndex = j
}

func (h *candidateMaxHeap) Push(x any) {
	c, _ := x.(*kCandidate)
	c.maxIndex = len(*h)
	*h = append(*h, c)
}

func (h *candidateMaxHeap) Pop() any {
	old := *h
	c := old[len(old)-1]
	old[len(old)-1] = nil
	*h = old[:len(old)-1]
	return c
}
