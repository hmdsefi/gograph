package path

import (
	"errors"
	"math"

	"github.com/hmdsefi/gograph"
)

// ErrNegativeWeight is returned when an edge weight is negative.
// The searches that use it refuse the graph before they start.
var ErrNegativeWeight = errors.New("edge weight is negative")

// NearestSources holds, for every vertex, the nearest of a set of sources
// and the distance from it.
type NearestSources[T comparable] struct {
	vertices []*gograph.Vertex[T]
	index    map[T]int
	dist     []float64
	origin   []int
	pred     []int
}

// Distance returns the distance from v's nearest source to v, and false
// when no source reaches v.
func (n *NearestSources[T]) Distance(v T) (float64, bool) {
	i, ok := n.reached(v)
	if !ok {
		return 0, false
	}
	return n.dist[i], true
}

// Source returns v's nearest source, and false when no source reaches v.
func (n *NearestSources[T]) Source(v T) (*gograph.Vertex[T], bool) {
	i, ok := n.reached(v)
	if !ok {
		return nil, false
	}
	return n.vertices[n.origin[i]], true
}

// PathTo returns the shortest path from v's nearest source to v, source
// first, and false when no source reaches v.
func (n *NearestSources[T]) PathTo(v T) ([]*gograph.Vertex[T], bool) {
	i, ok := n.reached(v)
	if !ok {
		return nil, false
	}

	rev := make([]int, 0, 8)
	for x := i; x >= 0; x = n.pred[x] {
		rev = append(rev, x)
	}
	path := make([]*gograph.Vertex[T], len(rev))
	for j, x := range rev {
		path[len(rev)-1-j] = n.vertices[x]
	}
	return path, true
}

func (n *NearestSources[T]) reached(v T) (int, bool) {
	i, ok := n.index[v]
	if !ok || n.origin[i] < 0 {
		return 0, false
	}
	return i, true
}

// DijkstraMultiSource finds, for every vertex, the source it is closest to
// and the distance from that source. It runs one search for all sources,
// so it costs the same as Dijkstra from a single start.
//
// Distances run from the sources along the edges. On a directed graph that
// is the distance from a source to the vertex. Undirected edges work both
// ways. Edge weights are read the way Dijkstra reads them. The graph does
// not have to be created with Weighted, and an edge without a weight counts
// as 0.
//
// A negative edge weight returns ErrNegativeWeight. The check covers every
// edge, including edges no source reaches. A source that is not in the graph
// returns gograph.ErrVertexDoesNotExist. A source listed twice counts once,
// at its first position. No sources returns a result where no vertex is
// reachable, and no error.
//
// When two sources are equally close, the one earlier in sources wins. A
// source is at distance 0 from itself, so it is its own nearest source,
// unless a zero-weight path from an earlier source reaches it. Equal paths
// from the same source follow the edge added earlier. Edges from a vertex
// are relaxed in the order they were added, and a vertex reached by an
// earlier edge is taken before one reached by a later edge, so the same
// graph and the same sources give the same paths on every run. Returned
// vertices are the graph's own pointers.
//
// Time is O((V+E) log V) however many sources there are. Extra space is
// O(V+E). Each vertex is extracted from the heap at most once.
func DijkstraMultiSource[T comparable](g gograph.Graph[T], sources ...T) (*NearestSources[T], error) {
	adj, vertices, index, err := buildSourceAdj(g)
	if err != nil {
		return nil, err
	}
	chosen, err := bindSources(index, sources)
	if err != nil {
		return nil, err
	}

	dist, rank, origin, pred := newSourceState(len(vertices), len(sources))
	starts := make([]int, 0, len(chosen))
	for pos, label := range chosen {
		i := index[label]
		dist[i] = 0
		rank[i] = pos
		origin[i] = i
		starts = append(starts, i)
	}
	if len(starts) > 0 {
		continueNearest(adj, dist, rank, origin, pred, starts)
	}

	return &NearestSources[T]{
		vertices: vertices,
		index:    index,
		dist:     dist,
		origin:   origin,
		pred:     pred,
	}, nil
}

// sourceAdj lists every outgoing edge in one pair of arrays. off[v] is the
// first edge of vertex v, and off[v+1] is the one after its last edge.
type sourceAdj struct {
	off    []int
	to     []int
	weight []float64
}

// buildSourceAdj copies the edges into arrays the search can walk without
// calling Neighbors, which copies every vertex. Edges stay in the order
// they were added for each source.
func buildSourceAdj[T comparable](g gograph.Graph[T]) (sourceAdj, []*gograph.Vertex[T], map[T]int, error) {
	vertices := g.GetAllVertices()
	index := make(map[T]int, len(vertices))
	for i, v := range vertices {
		index[v.Label()] = i
	}

	edges := g.AllEdges()
	n := len(vertices)
	off := make([]int, n+1)
	for _, e := range edges {
		if e.Weight() < 0 {
			return sourceAdj{}, nil, nil, ErrNegativeWeight
		}
		off[index[e.Source().Label()]+1]++
	}
	for i := 1; i <= n; i++ {
		off[i] += off[i-1]
	}

	adj := sourceAdj{
		off:    off,
		to:     make([]int, len(edges)),
		weight: make([]float64, len(edges)),
	}
	cursor := make([]int, n)
	copy(cursor, off)
	for _, e := range edges {
		from := index[e.Source().Label()]
		at := cursor[from]
		adj.to[at] = index[e.Destination().Label()]
		adj.weight[at] = e.Weight()
		cursor[from]++
	}
	return adj, vertices, index, nil
}

// bindSources keeps the first occurrence of each source. A missing label
// returns gograph.ErrVertexDoesNotExist.
func bindSources[T comparable](index map[T]int, sources []T) ([]T, error) {
	seen := make(map[T]struct{}, len(sources))
	chosen := make([]T, 0, len(sources))
	for _, label := range sources {
		if _, ok := index[label]; !ok {
			return nil, gograph.ErrVertexDoesNotExist
		}
		if _, ok := seen[label]; ok {
			continue
		}
		seen[label] = struct{}{}
		chosen = append(chosen, label)
	}
	return chosen, nil
}

// newSourceState returns distances of +Inf, a rank worse than any source,
// and no origin or predecessor. worseRank must be greater than every source
// position the caller will write.
func newSourceState(n, worseRank int) ([]float64, []int, []int, []int) {
	dist := make([]float64, n)
	rank := make([]int, n)
	origin := make([]int, n)
	pred := make([]int, n)
	for i := range dist {
		dist[i] = math.Inf(1)
		rank[i] = worseRank
		origin[i] = -1
		pred[i] = -1
	}
	return dist, rank, origin, pred
}

// continueNearest runs the search from starts. The caller sets dist, rank,
// origin and pred for each start first. A vertex is updated only when the
// new pair (distance, source position) is strictly better, so a later call
// can add one source and leave anything already closer untouched.
//
// The heap is ordered by that pair, then by the edge that reached the
// vertex, earlier first. A vertex already in the heap is moved rather than
// pushed again, so the heap holds at most one entry per vertex.
func continueNearest(adj sourceAdj, dist []float64, rank, origin, pred []int, starts []int) {
	edge := make([]int, len(dist))
	for i := range edge {
		edge[i] = math.MaxInt
	}
	h := newSourceHeap(dist, rank, edge)
	for _, s := range starts {
		edge[s] = -1
		h.push(s)
	}
	for h.len() > 0 {
		v := h.pop()
		for e := adj.off[v]; e < adj.off[v+1]; e++ {
			to := adj.to[e]
			next := dist[v] + adj.weight[e]
			if !closerSource(next, rank[v], dist[to], rank[to]) {
				continue
			}
			dist[to] = next
			rank[to] = rank[v]
			origin[to] = origin[v]
			pred[to] = v
			edge[to] = e
			h.improve(to)
		}
	}
}

// closerSource reports whether (next, nextRank) is strictly better than
// (dist, rank). A smaller distance wins. An equal distance keeps the
// earlier source.
func closerSource(next float64, nextRank int, dist float64, rank int) bool {
	if next < dist {
		return true
	}
	return next == dist && nextRank < rank
}

// sourceHeap is a binary heap of vertex indexes. The key lives in dist,
// rank and edge, so improving a vertex updates those slices and then moves
// its one heap entry. loc[v] is the entry's index, or -1 when v is not in
// the heap. edge[v] is the index of the edge that reached v, or -1 for a
// source.
type sourceHeap struct {
	v    []int
	loc  []int
	dist []float64
	rank []int
	edge []int
}

func newSourceHeap(dist []float64, rank, edge []int) *sourceHeap {
	loc := make([]int, len(dist))
	for i := range loc {
		loc[i] = -1
	}
	return &sourceHeap{
		v:    make([]int, 0, len(dist)),
		loc:  loc,
		dist: dist,
		rank: rank,
		edge: edge,
	}
}

func (h *sourceHeap) len() int { return len(h.v) }

func (h *sourceHeap) before(i, j int) bool {
	vi, vj := h.v[i], h.v[j]
	if h.dist[vi] != h.dist[vj] {
		return h.dist[vi] < h.dist[vj]
	}
	if h.rank[vi] != h.rank[vj] {
		return h.rank[vi] < h.rank[vj]
	}
	return h.edge[vi] < h.edge[vj]
}

func (h *sourceHeap) push(v int) {
	h.loc[v] = len(h.v)
	h.v = append(h.v, v)
	h.up(h.loc[v])
}

func (h *sourceHeap) improve(v int) {
	if h.loc[v] >= 0 {
		h.up(h.loc[v])
		return
	}
	h.push(v)
}

func (h *sourceHeap) pop() int {
	n := len(h.v) - 1
	h.swap(0, n)
	v := h.v[n]
	h.v = h.v[:n]
	h.loc[v] = -1
	if n > 0 {
		h.down(0)
	}
	return v
}

func (h *sourceHeap) up(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !h.before(i, parent) {
			return
		}
		h.swap(i, parent)
		i = parent
	}
}

func (h *sourceHeap) down(i int) {
	n := len(h.v)
	for {
		left := 2*i + 1
		if left >= n {
			return
		}
		small := left
		if right := left + 1; right < n && h.before(right, left) {
			small = right
		}
		if !h.before(small, i) {
			return
		}
		h.swap(i, small)
		i = small
	}
}

func (h *sourceHeap) swap(i, j int) {
	h.v[i], h.v[j] = h.v[j], h.v[i]
	h.loc[h.v[i]] = i
	h.loc[h.v[j]] = j
}
