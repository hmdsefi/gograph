package path

import (
	"iter"
	"math"

	"github.com/hmdsefi/gograph"
)

// AllPathsOption sets an option of AllSimplePaths and AllSimplePathsSeq.
type AllPathsOption func(*allPathsOptions)

type allPathsOptions struct {
	maxLength int // 0 when there is no limit
	limit     int // 0 when there is no limit
	invalid   bool
}

// WithMaxLength limits paths to at most n edges. An n below 1 makes
// AllSimplePaths and AllSimplePathsSeq return ErrInvalidLimit.
func WithMaxLength(n int) AllPathsOption {
	return func(o *allPathsOptions) {
		o.maxLength = n
		o.invalid = o.invalid || n < 1
	}
}

// WithLimit stops after n paths. An n below 1 makes AllSimplePaths and
// AllSimplePathsSeq return ErrInvalidLimit.
func WithLimit(n int) AllPathsOption {
	return func(o *allPathsOptions) {
		o.limit = n
		o.invalid = o.invalid || n < 1
	}
}

// AllSimplePaths returns every path from `from` to `to` that doesn't repeat a
// vertex. Each path starts with from and ends with to, and each slice is newly
// allocated, so the caller can keep it. The vertices are the graph's own
// pointers.
//
// A simple path here has at least one edge, so from == to returns no paths.
// Other libraries differ on this. Self-loops never appear in a simple path.
// Directed graphs follow outgoing edges. In an undirected graph an edge is
// usable in either direction, and A - B is one path even though the edge is
// stored in both directions.
//
// The paths come in depth-first order, following the order of Neighbors, so
// the same graph gives the same paths in the same order on every run. An
// unknown from or to returns gograph.ErrVertexDoesNotExist, and a to that from
// can't reach returns an empty slice and no error. A WithLimit or
// WithMaxLength below 1 returns ErrInvalidLimit.
//
// The number of simple paths can grow exponentially with the size of the
// graph: between opposite corners, a 6 by 6 grid has 1,262,816 of them and a
// 10 by 10 grid about 4·10^19. There is no default limit. WithLimit stops after n paths, WithMaxLength skips
// paths of more than n edges, and AllSimplePathsSeq lets the caller stop at
// any point.
//
// The search blocks each vertex that can't lead to `to` without going through
// the current path, the way Johnson's algorithm does for cycles, and unblocks
// it once a path through one of its neighbors turns up. It does O(V+E) work
// between two paths, so time is O((V+E)·(N+1)) for the N paths returned. With
// WithMaxLength(L), a vertex that finds no path is only blocked at the depth
// it was tried at and deeper, since it can still reach to from closer to from.
// Between two paths, each of the up to L vertices on the last path unblocks
// the vertices that wait for it, and a vertex can then be tried at up to L
// depths before the next unblocking, so time is O(L²·(V+E)·(N+1)). Extra space
// is O(V+E), plus the paths returned. V and E count only what from reaches.
func AllSimplePaths[T comparable](
	g gograph.Graph[T],
	from, to T,
	opts ...AllPathsOption,
) ([][]*gograph.Vertex[T], error) {
	paths, err := AllSimplePathsSeq(g, from, to, opts...)
	if err != nil {
		return nil, err
	}

	out := [][]*gograph.Vertex[T]{}
	for p := range paths {
		out = append(out, p)
	}
	return out, nil
}

// AllSimplePathsSeq yields the same paths as AllSimplePaths, in the same
// order, lazily, so callers can stop early. The errors are returned before
// the search starts, and from and to are looked up then. The search runs each
// time the sequence is ranged over, and it reads the edges as they are then,
// starting from the vertices looked up at the call. Stopping the loop stops
// the search.
func AllSimplePathsSeq[T comparable](
	g gograph.Graph[T],
	from, to T,
	opts ...AllPathsOption,
) (iter.Seq[[]*gograph.Vertex[T]], error) {
	var o allPathsOptions
	for _, opt := range opts {
		opt(&o)
	}
	if o.invalid {
		return nil, ErrInvalidLimit
	}
	src, dst := g.GetVertexByID(from), g.GetVertexByID(to)
	if src == nil || dst == nil {
		return nil, gograph.ErrVertexDoesNotExist
	}

	return func(yield func([]*gograph.Vertex[T]) bool) {
		if src == dst {
			return
		}
		s := newSimplePathSearch(src, dst, o)
		if s != nil {
			s.run(yield)
		}
	}, nil
}

// noLock is the lock of a vertex that can be entered at any depth.
const noLock = math.MaxInt32

// simplePathSearch holds the part of the graph from reaches as arrays. The
// arcs that leave vertex v are head[v] up to head[v+1]. Vertex 0 is from. The
// arcs into from, the arcs out of to and self-loops are left out, since a
// simple path from `from` to `to` never uses them.
type simplePathSearch[T comparable] struct {
	vertices []*gograph.Vertex[T]
	head     []int32
	arcTo    []int32
	arcFrom  []int32
	goal     int32

	// maxLength is the most edges a path can have, noLock when there is no
	// limit. limit is the most paths to yield, 0 when there is no limit.
	maxLength int32
	limit     int

	// lock[v] is the depth v can be entered below. Entering v sets it to the
	// depth of v, which also keeps v out of the search while it's on the path.
	// When v finds no path to goal, it stays locked until a neighbor of v
	// finds one. waiting[w] is the first arc u -> w whose source u waits for
	// w, nextWaiting[a] the arc after a in that list, and listed[a] whether a
	// is in it.
	lock        []int32
	onPath      []bool
	waiting     []int32
	nextWaiting []int32
	listed      []bool

	path      []int32
	frames    []pathFrame
	unlocking []int32
}

// pathFrame is a vertex on the path, the next arc to look at, and whether a
// path to goal was found through the vertex.
type pathFrame struct {
	vertex int32
	arc    int32
	found  bool
}

// newSimplePathSearch reads the part of the graph that from reaches, and
// returns nil when to is not in it.
func newSimplePathSearch[T comparable](from, to *gograph.Vertex[T], o allPathsOptions) *simplePathSearch[T] {
	s := &simplePathSearch[T]{vertices: []*gograph.Vertex[T]{from}, goal: -1, maxLength: noLock, limit: o.limit}
	if o.maxLength > 0 && o.maxLength < noLock {
		s.maxLength = int32(o.maxLength) //nolint:gosec // checked against noLock
	}
	index := map[*gograph.Vertex[T]]int32{from: 0}

	for u := 0; u < len(s.vertices); u++ {
		v := s.vertices[u]
		s.head = append(s.head, int32(len(s.arcTo))) //nolint:gosec // an arc number fits in int32
		if v == to {
			continue
		}
		for next := range v.AllNeighbors() {
			if next == v || next == from {
				continue
			}
			i, ok := index[next]
			if !ok {
				i = int32(len(s.vertices)) //nolint:gosec // a vertex number fits in int32
				index[next] = i
				s.vertices = append(s.vertices, next)
			}
			s.arcTo = append(s.arcTo, i)
			s.arcFrom = append(s.arcFrom, int32(u)) //nolint:gosec // a vertex number fits in int32
		}
	}
	s.head = append(s.head, int32(len(s.arcTo))) //nolint:gosec // an arc number fits in int32

	goal, ok := index[to]
	if !ok {
		return nil
	}
	s.goal = goal

	n, m := len(s.vertices), len(s.arcTo)
	s.lock = make([]int32, n)
	s.onPath = make([]bool, n)
	s.waiting = make([]int32, n)
	s.nextWaiting = make([]int32, m)
	s.listed = make([]bool, m)
	for i := range s.lock {
		s.lock[i] = noLock
		s.waiting[i] = -1
	}
	return s
}

// run yields the paths in depth-first order until there are none left, the
// limit is reached or yield returns false.
func (s *simplePathSearch[T]) run(yield func([]*gograph.Vertex[T]) bool) {
	found := 0
	s.enter(0)
	for len(s.frames) > 0 {
		f := &s.frames[len(s.frames)-1]
		v := f.vertex
		if f.arc < s.head[v+1] {
			w := s.arcTo[f.arc]
			f.arc++
			depth := int32(len(s.path)) //nolint:gosec // a depth fits in int32
			switch {
			case w == s.goal:
				f.found = true
				found++
				if !yield(s.copyPath()) || found == s.limit {
					return
				}
			case depth < s.lock[w] && depth < s.maxLength:
				s.enter(w)
			}
			continue
		}

		ok := f.found
		s.frames = s.frames[:len(s.frames)-1]
		s.path = s.path[:len(s.path)-1]
		s.onPath[v] = false
		if !ok {
			s.wait(v)
			continue
		}
		s.unlock(v)
		if len(s.frames) > 0 {
			s.frames[len(s.frames)-1].found = true
		}
	}
}

func (s *simplePathSearch[T]) enter(v int32) {
	s.lock[v] = int32(len(s.path)) //nolint:gosec // a depth fits in int32
	s.onPath[v] = true
	s.path = append(s.path, v)
	s.frames = append(s.frames, pathFrame{vertex: v, arc: s.head[v]})
}

// wait records that v found no path to goal. Without a length limit, every
// path from v to goal went through the path, so v stays locked at any depth
// until a neighbor of v finds a path. With one, v keeps the lock it got when
// it was entered, so it can still be entered closer to from.
func (s *simplePathSearch[T]) wait(v int32) {
	if s.maxLength == noLock {
		s.lock[v] = 0
	}
	for a := s.head[v]; a < s.head[v+1]; a++ {
		if w := s.arcTo[a]; w != s.goal && !s.listed[a] {
			s.listed[a] = true
			s.nextWaiting[a] = s.waiting[w]
			s.waiting[w] = a
		}
	}
}

// unlock records that v found a path to goal, so v can be entered at any
// depth again, and so can each vertex that waits for it, directly or through
// other vertices. A vertex on the path stays locked: it's an ancestor of v, so
// it finds a path too and is unlocked when it leaves the path.
func (s *simplePathSearch[T]) unlock(v int32) {
	s.lock[v] = noLock
	s.unlocking = append(s.unlocking[:0], v)
	for len(s.unlocking) > 0 {
		w := s.unlocking[len(s.unlocking)-1]
		s.unlocking = s.unlocking[:len(s.unlocking)-1]
		for a := s.waiting[w]; a >= 0; a = s.nextWaiting[a] {
			s.listed[a] = false
			if u := s.arcFrom[a]; !s.onPath[u] && s.lock[u] != noLock {
				s.lock[u] = noLock
				s.unlocking = append(s.unlocking, u)
			}
		}
		s.waiting[w] = -1
	}
}

// copyPath returns the vertices of the path, then goal, in a new slice.
func (s *simplePathSearch[T]) copyPath() []*gograph.Vertex[T] {
	p := make([]*gograph.Vertex[T], len(s.path)+1)
	for i, v := range s.path {
		p[i] = s.vertices[v]
	}
	p[len(s.path)] = s.vertices[s.goal]
	return p
}
