package path

import (
	"math"

	"github.com/hmdsefi/gograph"
)

// ShortestPaths holds the shortest path from one start vertex to each vertex
// it reaches, as returned by DijkstraPaths, BellmanFordPaths and BFSPaths. It
// keeps its own record of the search, so later changes to the graph don't
// change its answers.
type ShortestPaths[T comparable] struct {
	vertices []*gograph.Vertex[T]
	index    map[T]int
	dist     []float64
	pred     []int
}

// Distance returns the distance from the start vertex to the vertex with the
// given label, and false when the start doesn't reach it or it wasn't in the
// graph. The start is at distance 0 from itself.
func (s *ShortestPaths[T]) Distance(to T) (float64, bool) {
	i, ok := s.reached(to)
	if !ok {
		return 0, false
	}
	return s.dist[i], true
}

// PathTo returns the shortest path from the start vertex to the vertex with
// the given label, start first, and false when the start doesn't reach it or
// it wasn't in the graph. The path to the start is the start alone. The
// vertices are the graph's own pointers. It takes O(L) time for a path of L
// vertices.
func (s *ShortestPaths[T]) PathTo(to T) ([]*gograph.Vertex[T], bool) {
	i, ok := s.reached(to)
	if !ok {
		return nil, false
	}

	size := 0
	for x := i; x >= 0; x = s.pred[x] {
		size++
	}
	path := make([]*gograph.Vertex[T], size)
	for x := i; x >= 0; x = s.pred[x] {
		size--
		path[size] = s.vertices[x]
	}
	return path, true
}

func (s *ShortestPaths[T]) reached(to T) (int, bool) {
	i, ok := s.index[to]
	if !ok || math.IsInf(s.dist[i], 1) {
		return 0, false
	}
	return i, true
}

// newShortestPaths returns a result that holds only the start vertex.
func newShortestPaths[T comparable](start *gograph.Vertex[T]) *ShortestPaths[T] {
	return &ShortestPaths[T]{
		vertices: []*gograph.Vertex[T]{start},
		index:    map[T]int{start.Label(): 0},
		dist:     []float64{0},
		pred:     []int{-1},
	}
}

// add records a vertex the search just found, and returns its index.
func (s *ShortestPaths[T]) add(v *gograph.Vertex[T], dist float64, pred int) int {
	i := len(s.vertices)
	s.index[v.Label()] = i
	s.vertices = append(s.vertices, v)
	s.dist = append(s.dist, dist)
	s.pred = append(s.pred, pred)
	return i
}

// DijkstraPaths finds the shortest path from start to every vertex it reaches,
// by edge weight. Distance and PathTo on the result return each distance and
// path, and report false for a vertex start doesn't reach, so there's no
// math.MaxFloat64 or +Inf to compare with.
//
// Edge weights are read the way Dijkstra reads them: the graph does not have to
// be created with Weighted, an edge without a weight counts as 0, an edge with
// a NaN weight is skipped, and a path through a +Inf weight is as long as no
// path. Directed graphs follow outgoing edges, and an undirected edge works
// both ways.
//
// Weights must be zero or positive. A negative weight on an edge that leaves a
// vertex start reaches returns ErrNegativeWeight. BellmanFordPaths handles
// negative weights. A start that is not in the graph returns
// gograph.ErrVertexDoesNotExist.
//
// Equal paths keep the one found first. Vertices are settled in order of
// distance, then the insertion index of the edge that reached them, so the same
// graph gives the same paths on every run. For a single source, these are the
// paths of DijkstraMultiSource.
//
// The search only reads the vertices start reaches and the edges that leave
// them. With V and E counting those, time is O((V+E) log V) for a graph from
// gograph.New, and extra space is O(V): a vertex has one entry in the queue,
// which moves when a shorter path turns up.
func DijkstraPaths[T comparable](g gograph.Graph[T], start T) (*ShortestPaths[T], error) {
	from := g.GetVertexByID(start)
	if from == nil {
		return nil, gograph.ErrVertexDoesNotExist
	}

	s := newShortestPaths(from)
	queue := pathHeap{pos: []int{-1}}
	queue.update(0, 0, -1)

	for queue.len() > 0 {
		u := queue.pop()
		v := s.vertices[u]
		for next := range v.AllNeighbors() {
			e := g.GetEdge(v, next)
			if e == nil {
				continue
			}
			w := e.Weight()
			if w < 0 {
				return nil, ErrNegativeWeight
			}
			d := s.dist[u] + w
			if math.IsNaN(d) || math.IsInf(d, 1) {
				continue
			}

			i, ok := s.index[next.Label()]
			switch {
			case !ok:
				i = s.add(next, d, u)
				queue.pos = append(queue.pos, -1)
			case d >= s.dist[i]:
				continue
			default:
				s.dist[i] = d
				s.pred[i] = u
			}
			queue.update(i, d, e.InsertionIndex())
		}
	}
	return s, nil
}

// BFSPaths finds the path with the fewest edges from start to every vertex it
// reaches. Edge weights are ignored, and the distance of a vertex is the number
// of edges on its path. Distance and PathTo on the result return each distance
// and path, and report false for a vertex start doesn't reach.
//
// Directed graphs follow outgoing edges, and an undirected edge works both
// ways. A start that is not in the graph returns gograph.ErrVertexDoesNotExist.
// Vertices are found in the order of traverse.BFS, and a vertex keeps the first
// path that finds it, so the same graph gives the same paths on every run.
//
// The search only reads the vertices start reaches and the edges that leave
// them. With V and E counting those, time is O(V+E) and extra space is O(V):
// the result is also the queue of the search.
func BFSPaths[T comparable](g gograph.Graph[T], start T) (*ShortestPaths[T], error) {
	from := g.GetVertexByID(start)
	if from == nil {
		return nil, gograph.ErrVertexDoesNotExist
	}

	s := newShortestPaths(from)
	for u := 0; u < len(s.vertices); u++ {
		for next := range s.vertices[u].AllNeighbors() {
			if _, ok := s.index[next.Label()]; !ok {
				s.add(next, s.dist[u]+1, u)
			}
		}
	}
	return s, nil
}

// BellmanFordPaths finds the shortest path from start to every vertex it
// reaches, by edge weight, on a graph whose weights can be negative. Distance
// and PathTo on the result return each distance and path, and report false for
// a vertex start doesn't reach.
//
// It accepts the graphs BellmanFord accepts, returns the same errors, and gives
// the same distances: ErrNotWeighted, ErrNotDirected,
// gograph.ErrVertexDoesNotExist for a start that is not in the graph, and
// ErrNegativeWeightCycle for a negative cycle start reaches. An edge with a NaN
// weight is skipped, and a path through a +Inf weight is as long as no path. A
// vertex that a path through a -Inf weight reaches has a distance of -Inf, and
// PathTo returns such a path.
//
// Among the shortest paths to a vertex with a finite distance, PathTo returns
// one with the fewest edges, and after that the one whose edges come first in
// the order of g.AllEdges. The path to a vertex at -Inf follows the shortest
// path to the -Inf weight it goes through, so another path at -Inf can have
// fewer edges. The same graph gives the same paths on every run. The weights
// along a path add up to its distance. Rounding is the one exception: with
// weights such as 0.1, a cycle that sums to zero can round to slightly below
// zero, and the path of a vertex on it can then add up to slightly more than
// the distance.
//
// Time is O(V·E), or O(k·E) when every shortest path has at most k edges, as
// for BellmanFord. Building the paths from the distances adds O(V+E), or
// O((V+E) log V) when rounding leaves a vertex that no edge reaches exactly.
// Extra space is O(V+E).
func BellmanFordPaths[T comparable](g gograph.Graph[T], start T) (*ShortestPaths[T], error) {
	b, err := bellmanFordDistances(g, start)
	if err != nil {
		return nil, err
	}
	return &ShortestPaths[T]{
		vertices: b.vertices,
		index:    b.index,
		dist:     b.dist,
		pred:     shortestPathTree(b.dist, b.start, b.arcs, b.negInf),
	}, nil
}

// shortestPathTree returns, for each vertex that start reaches, the vertex
// before it on a shortest path from start, and -1 for start and for the
// vertices it doesn't reach. dist holds the final distances of Bellman-Ford,
// and arcs the edges with each -Inf weight read as 0 and listed in negInf.
//
// An arc from u to v is tight when dist[u] plus its weight equals dist[v]. A
// breadth-first search from start over the tight arcs gives every vertex with
// a finite distance a path with the fewest edges. Reading the links from the
// distances, not recording them during the passes, keeps the links a tree:
// links recorded during the passes can close a loop when rounding makes a
// zero cycle look negative.
//
// A vertex at -Inf then hangs off the tree below a -Inf arc. No -Inf arc is on
// a cycle here, so every such vertex is reachable from a -Inf arc whose tail
// has a finite distance.
func shortestPathTree(dist []float64, start int, arcs, negInf []arc) []int {
	n := len(dist)
	pred := make([]int, n)
	for i := range pred {
		pred[i] = -1
	}
	offset, order := outgoingArcs(n, arcs)
	attached := make([]bool, n)
	attached[start] = true
	queue := make([]int, 1, n)
	queue[0] = start

	head := 0
	for ; head < len(queue); head++ {
		u := queue[head]
		for _, i := range order[offset[u]:offset[u+1]] {
			a := arcs[i]
			if !attached[a.to] && !math.IsInf(dist[a.to], 0) && dist[u]+a.weight == dist[a.to] {
				attached[a.to] = true
				pred[a.to] = u
				queue = append(queue, a.to)
			}
		}
	}

	finite := 0
	for _, d := range dist {
		if !math.IsInf(d, 0) {
			finite++
		}
	}
	if len(queue) < finite {
		queue = attachBySlack(dist, arcs, offset, order, attached, pred, queue)
		head = len(queue)
	}

	negInfinity := math.Inf(-1)
	for _, a := range negInf {
		if attached[a.from] && dist[a.from] != negInfinity && !attached[a.to] {
			attached[a.to] = true
			pred[a.to] = a.from
			queue = append(queue, a.to)
		}
	}
	for ; head < len(queue); head++ {
		u := queue[head]
		for _, i := range order[offset[u]:offset[u+1]] {
			if v := arcs[i].to; !attached[v] && dist[v] == negInfinity {
				attached[v] = true
				pred[v] = u
				queue = append(queue, v)
			}
		}
	}
	return pred
}

// attachBySlack links the vertices with a finite distance that no tight arc
// reaches, which only happens when rounding made a zero cycle look negative.
// The slack of an arc from u to v is dist[u] plus its weight, minus dist[v].
// It's never below zero once the passes have settled. Each vertex left gets
// the path from the tree with the least total slack, by Dijkstra's algorithm
// over the slacks, so its weights add up as close to its distance as a path
// can. It returns queue with the vertices it attached added. It takes
// O((V+E) log V) time and O(V) space.
func attachBySlack(dist []float64, arcs []arc, offset, order []int, attached []bool, pred, queue []int) []int {
	n := len(dist)
	slack := make([]float64, n)
	h := pathHeap{pos: make([]int, n)}
	for v := range n {
		slack[v] = math.Inf(1)
		h.pos[v] = -1
	}
	for i, u := range queue {
		slack[u] = 0
		h.update(u, 0, i)
	}

	for h.len() > 0 {
		u := h.pop()
		if !attached[u] {
			attached[u] = true
			queue = append(queue, u)
		}
		for _, i := range order[offset[u]:offset[u+1]] {
			a := arcs[i]
			if attached[a.to] || math.IsInf(dist[a.to], 0) {
				continue
			}
			if s := slack[u] + (dist[u] + a.weight - dist[a.to]); s < slack[a.to] {
				slack[a.to] = s
				pred[a.to] = u
				h.update(a.to, s, n+i)
			}
		}
	}
	return queue
}

// outgoingArcs groups the arcs by the vertex they leave: the arcs from vertex
// v are arcs[i] for each i in order[offset[v]:offset[v+1]], in the order of
// arcs.
func outgoingArcs(n int, arcs []arc) (offset, order []int) {
	offset = make([]int, n+1)
	for _, a := range arcs {
		offset[a.from+1]++
	}
	for v := range n {
		offset[v+1] += offset[v]
	}

	order = make([]int, len(arcs))
	next := make([]int, n)
	copy(next, offset)
	for i, a := range arcs {
		order[next[a.from]] = i
		next[a.from]++
	}
	return offset, order
}

// pathEntry is a queued vertex with its key. A smaller key comes first, and
// then a smaller tie.
type pathEntry struct {
	key float64
	tie int
	v   int
}

// pathHeap is a binary min-heap with one entry per vertex. pos[v] is the index
// of v's entry, or -1 when v is not queued. The entries carry their keys, so a
// comparison reads two neighbors in one slice.
type pathHeap struct {
	entries []pathEntry
	pos     []int
}

func (h *pathHeap) len() int { return len(h.entries) }

// update queues v with the given key, or moves its entry up when v is already
// queued. The key of a queued vertex only goes down.
func (h *pathHeap) update(v int, key float64, tie int) {
	i := h.pos[v]
	if i < 0 {
		i = len(h.entries)
		h.entries = append(h.entries, pathEntry{})
	}

	e := pathEntry{key: key, tie: tie, v: v}
	for i > 0 {
		parent := (i - 1) / 2
		if !e.before(h.entries[parent]) {
			break
		}
		h.entries[i] = h.entries[parent]
		h.pos[h.entries[i].v] = i
		i = parent
	}
	h.entries[i] = e
	h.pos[v] = i
}

// pop removes and returns the vertex with the smallest key.
func (h *pathHeap) pop() int {
	top := h.entries[0].v
	h.pos[top] = -1

	last := len(h.entries) - 1
	e := h.entries[last]
	h.entries = h.entries[:last]
	if last == 0 {
		return top
	}

	i := 0
	for {
		child := 2*i + 1
		if child >= last {
			break
		}
		if right := child + 1; right < last && h.entries[right].before(h.entries[child]) {
			child = right
		}
		if !h.entries[child].before(e) {
			break
		}
		h.entries[i] = h.entries[child]
		h.pos[h.entries[i].v] = i
		i = child
	}
	h.entries[i] = e
	h.pos[e.v] = i
	return top
}

func (e pathEntry) before(o pathEntry) bool {
	if e.key != o.key {
		return e.key < o.key
	}
	return e.tie < o.tie
}
