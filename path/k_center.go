package path

import (
	"errors"
	"math"

	"github.com/hmdsefi/gograph"
)

var (
	// ErrNegativeWeight is returned when an edge weight is negative.
	// The searches that use it refuse the graph before they start.
	ErrNegativeWeight = errors.New("edge weight is negative")

	// ErrInvalidK is returned when k is below 1, or below the number of
	// distinct fixed centers.
	ErrInvalidK = errors.New("k is below 1 or below the number of fixed centers")
)

// KCenter chooses k centers so that the vertex farthest from its nearest
// center is as close to one as it can make it. It returns the centers in
// the order it chose them, and the radius: the largest distance from a
// vertex to its nearest center.
//
// fixed are centers that are already open, such as existing warehouses.
// They come first, in the order given, and count toward k. A fixed center
// listed twice counts once. Without fixed centers, the first center is the
// first vertex added to the graph.
//
// Each next center is the vertex farthest from the centers chosen so far.
// Ties take the vertex added earlier, so the same graph gives the same
// centers on every run. On an undirected graph, the radius is at most twice
// the smallest radius any k centers can have. On a directed graph the
// distances are not symmetric, and that factor of two does not hold.
//
// A vertex no center reaches is infinitely far and is chosen next. If some
// vertex is still unreachable after k centers, the radius is +Inf. When k
// is at least the number of vertices, every vertex is a center, fewer than
// k are returned, and the radius is 0. An empty graph returns nil centers
// and radius 0.
//
// Distances follow the edges. The graph does not have to be weighted, and
// an edge without a weight counts as 0. A negative weight returns
// ErrNegativeWeight. k below 1, or below the number of fixed centers,
// returns ErrInvalidK. A fixed center that is not in the graph returns
// gograph.ErrVertexDoesNotExist.
//
// Time is O(k (V + E) log V) and extra space is O(V + E). The distances
// from one round are kept. Opening a center searches only the vertices it
// gets strictly closer to, which is safe because a path that is no closer
// at a vertex cannot be closer past it when every weight is at least 0.
func KCenter[T comparable](g gograph.Graph[T], k int, fixed ...T) ([]*gograph.Vertex[T], float64, error) {
	if k < 1 {
		return nil, 0, ErrInvalidK
	}

	vertices := g.GetAllVertices()
	if len(vertices) == 0 {
		if len(fixed) > 0 {
			return nil, 0, gograph.ErrVertexDoesNotExist
		}
		return nil, 0, nil
	}

	index := make(map[T]int, len(vertices))
	for i, v := range vertices {
		index[v.Label()] = i
	}

	pinned, err := fixedCenterIndexes(g, index, fixed)
	if err != nil {
		return nil, 0, err
	}
	if k < len(pinned) {
		return nil, 0, ErrInvalidK
	}

	adj, err := buildCenterAdj(g, index)
	if err != nil {
		return nil, 0, err
	}

	dist := make([]float64, len(vertices))
	for i := range dist {
		dist[i] = math.Inf(1)
	}
	isCenter := make([]bool, len(vertices))
	centers := make([]*gograph.Vertex[T], 0, min(k, len(vertices)))
	h := &distHeap{}

	open := func(i int) {
		isCenter[i] = true
		centers = append(centers, vertices[i])
		closerFrom(h, adj, dist, i)
	}
	for _, i := range pinned {
		open(i)
	}
	for len(centers) < k && len(centers) < len(vertices) {
		open(farthestUncovered(dist, isCenter))
	}

	return centers, farthestDist(dist), nil
}

type centerEdge struct {
	to     int
	weight float64
}

// centerAdj holds every outgoing edge in one array. off[v] is the first
// edge of vertex v, and off[v+1] is the one after its last edge.
type centerAdj struct {
	off  []int
	edge []centerEdge
}

func (a centerAdj) from(v int) []centerEdge {
	return a.edge[a.off[v]:a.off[v+1]]
}

// buildCenterAdj lists each edge once per direction. AllEdges is used
// instead of Neighbors because Neighbors copies every vertex. The edges
// sit in one array so the search does not allocate a slice per vertex.
func buildCenterAdj[T comparable](g gograph.Graph[T], index map[T]int) (centerAdj, error) {
	edges := g.AllEdges()
	count := make([]int, len(index))
	for _, e := range edges {
		if e.Weight() < 0 {
			return centerAdj{}, ErrNegativeWeight
		}
		count[index[e.Source().Label()]]++
	}

	off := make([]int, len(index)+1)
	for i, n := range count {
		off[i+1] = off[i] + n
	}
	adj := centerAdj{off: off, edge: make([]centerEdge, off[len(index)])}
	cursor := count
	for i := range cursor {
		cursor[i] = off[i]
	}
	for _, e := range edges {
		from := index[e.Source().Label()]
		adj.edge[cursor[from]] = centerEdge{
			to:     index[e.Destination().Label()],
			weight: e.Weight(),
		}
		cursor[from]++
	}
	return adj, nil
}

func fixedCenterIndexes[T comparable](g gograph.Graph[T], index map[T]int, fixed []T) ([]int, error) {
	if len(fixed) == 0 {
		return nil, nil
	}
	seen := make(map[T]struct{}, len(fixed))
	out := make([]int, 0, len(fixed))
	for _, label := range fixed {
		if g.GetVertexByID(label) == nil {
			return nil, gograph.ErrVertexDoesNotExist
		}
		if _, ok := seen[label]; ok {
			continue
		}
		seen[label] = struct{}{}
		out = append(out, index[label])
	}
	return out, nil
}

// closerFrom sets start's distance to 0 and updates a vertex only when the
// path from start is strictly shorter than the distance it already has.
func closerFrom(h *distHeap, adj centerAdj, dist []float64, start int) {
	h.clear()
	dist[start] = 0
	h.push(start, 0)
	for h.len() > 0 {
		v, d := h.pop()
		if d > dist[v] {
			continue
		}
		for _, e := range adj.from(v) {
			next := d + e.weight
			if next < dist[e.to] {
				dist[e.to] = next
				h.push(e.to, next)
			}
		}
	}
}

// farthestUncovered returns the unchosen vertex with the largest distance.
// An earlier vertex wins a tie.
func farthestUncovered(dist []float64, isCenter []bool) int {
	best := -1
	bestDist := math.Inf(-1)
	for i, d := range dist {
		if isCenter[i] {
			continue
		}
		if d > bestDist {
			best = i
			bestDist = d
		}
	}
	return best
}

func farthestDist(dist []float64) float64 {
	radius := 0.0
	for _, d := range dist {
		if d > radius {
			radius = d
		}
	}
	return radius
}

// distHeap is a binary min-heap of vertex indexes keyed by distance.
// Entries are values in two slices, so a push does not allocate its own object.
// A vertex may be pushed again when a shorter path is found. The pop that
// carries an older distance is ignored.
type distHeap struct {
	v []int
	d []float64
}

func (h *distHeap) len() int { return len(h.v) }

func (h *distHeap) clear() {
	h.v = h.v[:0]
	h.d = h.d[:0]
}

func (h *distHeap) push(v int, d float64) {
	h.v = append(h.v, v)
	h.d = append(h.d, d)
	i := len(h.v) - 1
	for i > 0 {
		parent := (i - 1) / 2
		if h.d[parent] <= h.d[i] {
			return
		}
		h.swap(parent, i)
		i = parent
	}
}

func (h *distHeap) pop() (int, float64) {
	n := len(h.v) - 1
	h.swap(0, n)
	h.down(0, n)
	v, d := h.v[n], h.d[n]
	h.v = h.v[:n]
	h.d = h.d[:n]
	return v, d
}

func (h *distHeap) down(i, n int) {
	for {
		left := 2*i + 1
		if left >= n {
			return
		}
		small := left
		if right := left + 1; right < n && h.d[right] < h.d[left] {
			small = right
		}
		if h.d[small] >= h.d[i] {
			return
		}
		h.swap(i, small)
		i = small
	}
}

func (h *distHeap) swap(i, j int) {
	h.v[i], h.v[j] = h.v[j], h.v[i]
	h.d[i], h.d[j] = h.d[j], h.d[i]
}
