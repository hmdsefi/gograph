package path

import (
	"cmp"
	"errors"
	"math"
	"slices"

	"github.com/hmdsefi/gograph"
)

// ErrInvalidWidth is returned when a beam width is below 1.
var ErrInvalidWidth = errors.New("beam width is below 1")

// Beam returns a path from start to goal and its cost, keeping at most width
// partial paths at each step, ranked by score. The cost is the sum of the edge
// weights on the path.
//
// score(edge, costSoFar) returns the rank of the path that would use edge after
// spending costSoFar. Lower is better. A nil score uses costSoFar plus the edge
// weight. An edge whose score is NaN or +Inf is not used, and neither is an
// edge that would make the cost NaN or +Inf. Weights may be negative, since the
// search doesn't rely on them being positive.
//
// The path is approximate once width is smaller than the frontier: a small
// width can drop the partial path that leads to the cheapest path. Each step
// extends every kept path by one outgoing edge, and a path never repeats a
// vertex. The search stops at the first step that keeps a path ending at goal,
// and returns the best of those. Paths with the same score rank in the order
// they were made: by the rank of the path they extend, then by the order of the
// edges, so the same graph, width and score give the same path on every run.
//
// width below 1 returns ErrInvalidWidth, a start or goal that is not in the
// graph returns gograph.ErrVertexDoesNotExist, and a search whose kept paths
// never reach goal returns ErrUnreachable. When start equals goal, the path is
// that one vertex and the cost is 0. The path vertices are the graph's own
// pointers, start first and goal last.
//
// The search takes at most L steps, where L is the number of edges on the path
// it returns, or at most V-1 when it returns ErrUnreachable. With d the largest
// out-degree, time is O(L·width·(L + d log width)), and only the edges out of
// kept paths are read. Extra space is O(L·width): the kept paths share their
// common starts, and the candidates for a step never number more than width.
func Beam[T comparable](
	g gograph.Graph[T],
	start, goal T,
	width int,
	score func(edge *gograph.Edge[T], costSoFar float64) float64,
) ([]*gograph.Vertex[T], float64, error) {
	if width < 1 {
		return nil, 0, ErrInvalidWidth
	}
	from, to := g.GetVertexByID(start), g.GetVertexByID(goal)
	if from == nil || to == nil {
		return nil, 0, gograph.ErrVertexDoesNotExist
	}
	if from == to {
		return []*gograph.Vertex[T]{from}, 0, nil
	}

	frontier := []*beamNode[T]{{vertex: from}}
	next := beamHeap[T]{width: width}
	// onPath[v] == stamp when v is on the path being extended. It's only filled
	// for a vertex with many neighbors. For a few, walking the path back for
	// each neighbor is cheaper than writing the whole path to a map.
	var onPath map[*gograph.Vertex[T]]int
	stamp := 0

	for len(frontier) > 0 {
		order := 0
		for _, p := range frontier {
			useMap := p.vertex.OutDegree() > beamWalkLimit
			if useMap {
				if onPath == nil {
					onPath = map[*gograph.Vertex[T]]int{}
				}
				stamp++
				for x := p; x != nil; x = x.parent {
					onPath[x.vertex] = stamp
				}
			}

			for neighbor := range p.vertex.AllNeighbors() {
				if (useMap && onPath[neighbor] == stamp) || (!useMap && p.contains(neighbor)) {
					continue
				}
				e := g.GetEdge(p.vertex, neighbor)
				if e == nil {
					continue
				}
				cost := p.cost + e.Weight()
				rank := cost
				if score != nil {
					rank = score(e, p.cost)
				}
				if math.IsNaN(cost) || math.IsInf(cost, 1) || math.IsNaN(rank) || math.IsInf(rank, 1) {
					continue
				}
				next.offer(neighbor, p, cost, rank, order)
				order++
			}
		}

		frontier = next.take(frontier[:0])
		for _, p := range frontier {
			if p.vertex == to {
				return p.path(), p.cost, nil
			}
		}
	}
	return nil, 0, ErrUnreachable
}

// beamWalkLimit is the most neighbors a vertex can have for Beam to check each
// one by walking the path back.
const beamWalkLimit = 8

// beamNode is one partial path: its last vertex, the path it extends, its cost
// and score, and the order it was made in within its step.
type beamNode[T comparable] struct {
	vertex *gograph.Vertex[T]
	parent *beamNode[T]
	cost   float64
	score  float64
	order  int
}

// beamBefore reports whether a ranks before b.
func beamBefore[T comparable](a, b *beamNode[T]) bool {
	if a.score != b.score {
		return a.score < b.score
	}
	return a.order < b.order
}

// contains reports whether v is on the path that ends at n.
func (n *beamNode[T]) contains(v *gograph.Vertex[T]) bool {
	for x := n; x != nil; x = x.parent {
		if x.vertex == v {
			return true
		}
	}
	return false
}

// path returns the vertices from start to n, start first.
func (n *beamNode[T]) path() []*gograph.Vertex[T] {
	size := 0
	for x := n; x != nil; x = x.parent {
		size++
	}
	path := make([]*gograph.Vertex[T], size)
	for x := n; x != nil; x = x.parent {
		size--
		path[size] = x.vertex
	}
	return path
}

// beamHeap keeps the best width candidates of a step in a binary heap with
// the worst at the top, so a better candidate replaces it in O(log width).
type beamHeap[T comparable] struct {
	nodes []*beamNode[T]
	width int
}

// offer adds a candidate if it ranks among the best width so far. A candidate
// that doesn't is never allocated, and one that does reuses the node it pushes
// out, since nothing refers to a candidate before take returns it.
func (h *beamHeap[T]) offer(v *gograph.Vertex[T], parent *beamNode[T], cost, score float64, order int) {
	if len(h.nodes) < h.width {
		h.nodes = append(h.nodes, &beamNode[T]{vertex: v, parent: parent, cost: cost, score: score, order: order})
		h.up(len(h.nodes) - 1)
		return
	}
	worst := h.nodes[0]
	if score > worst.score || (score == worst.score && order > worst.order) {
		return
	}
	*worst = beamNode[T]{vertex: v, parent: parent, cost: cost, score: score, order: order}
	h.down(0)
}

// take returns the candidates best first, and keeps buf, an empty slice whose
// array is free, for the next step's candidates.
func (h *beamHeap[T]) take(buf []*beamNode[T]) []*beamNode[T] {
	nodes := h.nodes
	h.nodes = buf
	slices.SortFunc(nodes, func(a, b *beamNode[T]) int {
		if c := cmp.Compare(a.score, b.score); c != 0 {
			return c
		}
		return cmp.Compare(a.order, b.order)
	})
	return nodes
}

// worse reports whether the node at i ranks after the node at j.
func (h *beamHeap[T]) worse(i, j int) bool { return beamBefore(h.nodes[j], h.nodes[i]) }

func (h *beamHeap[T]) up(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !h.worse(i, parent) {
			return
		}
		h.nodes[i], h.nodes[parent] = h.nodes[parent], h.nodes[i]
		i = parent
	}
}

func (h *beamHeap[T]) down(i int) {
	n := len(h.nodes)
	for {
		top, left := i, 2*i+1
		if left < n && h.worse(left, top) {
			top = left
		}
		if right := left + 1; right < n && h.worse(right, top) {
			top = right
		}
		if top == i {
			return
		}
		h.nodes[i], h.nodes[top] = h.nodes[top], h.nodes[i]
		i = top
	}
}
