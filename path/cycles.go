package path

import "github.com/hmdsefi/gograph"

// Cycles returns elementary cycles of a directed graph, up to limit.
//
// A cycle is a slice of vertices in walk order. The slice does not repeat the
// first vertex. The edge from the last vertex back to the first closes it, so
// [a b c] means a -> b -> c -> a. Each cycle starts at its earliest vertex in
// GetAllVertices order, so a rotation is not reported twice. A cycle never
// repeats a vertex, and a self-loop is a cycle of one vertex. The vertices are
// the graph's own pointers.
//
// When the graph has more cycles than limit, the slice has length limit and
// the error is nil. The cycles included are the ones Johnson's algorithm finds
// first under GetAllVertices order: all the cycles that start at the first
// vertex, then those that start at the second, and so on, each group in the
// order a depth-first search finds them following the order of Neighbors. The
// same graph and the same limit give the same cycles in the same order on
// every run. A graph with no cycle returns a nil slice and a nil error.
//
// A limit below 1 returns ErrInvalidLimit, and an undirected graph returns
// ErrNotDirected: each of its cycles can be walked in two directions, which is
// a separate question. A graph created with gograph.Acyclic returns nil without
// a search.
//
// A graph can have exponentially many cycles, and the running time grows with
// the number returned, so limit is how the caller bounds it. Time is
// O((V+E)·(c+1)) for the c cycles returned: Johnson's algorithm blocks each
// vertex that can't lead back to the start, so the search does O(V+E) work
// between two cycles. The strongly connected components are found once, and
// after the cycles through a start are listed, only the component of that
// start is split again, without it. Extra space is O(V+E), plus the cycles
// returned.
func Cycles[T comparable](g gograph.Graph[T], limit int) ([][]*gograph.Vertex[T], error) {
	if limit < 1 {
		return nil, ErrInvalidLimit
	}
	if !g.IsDirected() {
		return nil, ErrNotDirected
	}
	if g.IsAcyclic() {
		return nil, nil
	}

	s := newCycleSearch(g, limit)
	s.split(0, int32(len(s.vertices)), allVertices, -1) //nolint:gosec // a vertex number fits in int32
	for start := range s.vertices {
		v := int32(start) //nolint:gosec // a vertex number fits in int32
		// The vertices before v are gone, so the component of v is its
		// strongly connected component among the vertices from v on. It
		// starts at v when v is its smallest vertex.
		if s.comp[v] != v || !s.cyclic[v] {
			continue
		}
		lo, hi := s.rangeStart[v], s.rangeEnd[v]
		s.circuits(v, s.members[lo:hi])
		if len(s.cycles) == limit {
			break
		}
		s.comp[v] = removed
		s.split(lo, hi, v, v)
	}
	return s.cycles, nil
}

// cycleSearch holds the graph as arrays, with the vertices numbered in
// GetAllVertices order. The arcs that leave vertex v are head[v] up to
// head[v+1].
type cycleSearch[T comparable] struct {
	vertices []*gograph.Vertex[T]
	head     []int32
	arcTo    []int32
	arcFrom  []int32
	selfLoop []bool
	limit    int
	cycles   [][]*gograph.Vertex[T]

	// Strongly connected components of what is left of the graph. A
	// component is named by its smallest vertex, and comp[v] is the name of
	// the component of v, or removed. The members of component c are
	// members[rangeStart[c]:rangeEnd[c]], and cyclic[c] says whether it has a
	// cycle: more than one vertex, or a self-loop. current is the component
	// whose cycles are being listed.
	comp       []int32
	members    []int32
	rangeStart []int32
	rangeEnd   []int32
	cyclic     []bool
	current    int32

	// Tarjan's algorithm. visit[v] is 0 until split reaches v, and is reset
	// for the vertices of each component it splits, so it never exceeds the
	// size of one component.
	visit    []int32
	low      []int32
	onStack  []bool
	sccStack []int32
	found    []int32
	frames   []cycleFrame

	// Johnson's blocked vertices. The vertices that wait for v to be
	// unblocked are a linked list of arcs into v: blockedBy[v] is the first
	// arc and nextBlocked[a] the one after arc a. inList[a] says whether arc a
	// is in its target's list, so no arc is listed twice.
	blocked     []bool
	blockedBy   []int32
	nextBlocked []int32
	inList      []bool
	path        []int32
	unblocking  []int32
}

// cycleFrame is a vertex on the search path and the next arc to look at.
// found says whether a cycle was found through the vertex.
type cycleFrame struct {
	vertex int32
	arc    int32
	found  bool
}

func newCycleSearch[T comparable](g gograph.Graph[T], limit int) *cycleSearch[T] {
	index := make(map[*gograph.Vertex[T]]int32, g.Order())
	s := &cycleSearch[T]{limit: limit}
	for v := range gograph.Vertices(g) {
		index[v] = int32(len(s.vertices)) //nolint:gosec // a vertex number fits in int32
		s.vertices = append(s.vertices, v)
	}

	n := len(s.vertices)
	s.head = make([]int32, n+1)
	s.selfLoop = make([]bool, n)
	for u, v := range s.vertices {
		s.head[u] = int32(len(s.arcTo)) //nolint:gosec // an arc number fits in int32
		for next := range v.AllNeighbors() {
			w, ok := index[next]
			if !ok {
				continue
			}
			if int(w) == u {
				s.selfLoop[u] = true
			}
			s.arcTo = append(s.arcTo, w)
			s.arcFrom = append(s.arcFrom, int32(u)) //nolint:gosec // a vertex number fits in int32
		}
	}
	s.head[n] = int32(len(s.arcTo)) //nolint:gosec // an arc number fits in int32

	m := len(s.arcTo)
	s.comp = make([]int32, n)
	s.members = make([]int32, n)
	s.rangeStart = make([]int32, n)
	s.rangeEnd = make([]int32, n)
	s.cyclic = make([]bool, n)
	s.visit = make([]int32, n)
	s.low = make([]int32, n)
	s.onStack = make([]bool, n)
	s.blocked = make([]bool, n)
	s.blockedBy = make([]int32, n)
	s.nextBlocked = make([]int32, m)
	s.inList = make([]bool, m)
	for i := range s.comp {
		s.comp[i] = allVertices
		s.members[i] = int32(i) //nolint:gosec // a vertex number fits in int32
		s.blockedBy[i] = -1
	}
	return s
}

// Component names that aren't vertices.
const (
	allVertices int32 = -2 // the whole graph, before the first split
	removed     int32 = -1 // a start whose cycles are listed
)

// split finds the strongly connected components of the vertices of component
// old in members[lo:hi], leaving out skip, with Tarjan's algorithm and its own
// stack. It writes the members of each component back into that range and
// names it by its smallest vertex. It takes time linear in the vertices of the
// range and the arcs that leave them.
func (s *cycleSearch[T]) split(lo, hi, old, skip int32) {
	part := s.members[lo:hi]
	for _, v := range part {
		s.visit[v] = 0
	}
	s.found = s.found[:0]
	visits := int32(0)
	enter := func(v int32) {
		visits++
		s.visit[v] = visits
		s.low[v] = visits
		s.onStack[v] = true
		s.sccStack = append(s.sccStack, v)
		s.frames = append(s.frames, cycleFrame{vertex: v, arc: s.head[v]})
	}

	for _, root := range part {
		if root == skip || s.visit[root] != 0 || s.comp[root] != old {
			continue
		}
		enter(root)
		for len(s.frames) > 0 {
			f := &s.frames[len(s.frames)-1]
			v := f.vertex
			if f.arc < s.head[v+1] {
				w := s.arcTo[f.arc]
				f.arc++
				switch {
				case w == skip || (s.comp[w] != old && !s.onStack[w]):
				case s.visit[w] == 0:
					enter(w)
				case s.onStack[w]:
					s.low[v] = min(s.low[v], s.visit[w])
				}
				continue
			}

			s.frames = s.frames[:len(s.frames)-1]
			if len(s.frames) > 0 {
				parent := s.frames[len(s.frames)-1].vertex
				s.low[parent] = min(s.low[parent], s.low[v])
			}
			if s.low[v] == s.visit[v] {
				s.popComponent(v, lo)
			}
		}
	}
	copy(part, s.found)
}

// popComponent pops the component whose first vertex is v off Tarjan's stack,
// appends its members to found, and names it. lo is where found starts in
// members.
func (s *cycleSearch[T]) popComponent(v, lo int32) {
	first := len(s.found)
	smallest := v
	for {
		w := s.sccStack[len(s.sccStack)-1]
		s.sccStack = s.sccStack[:len(s.sccStack)-1]
		s.onStack[w] = false
		s.found = append(s.found, w)
		smallest = min(smallest, w)
		if w == v {
			break
		}
	}
	for _, w := range s.found[first:] {
		s.comp[w] = smallest
	}
	s.rangeStart[smallest] = lo + int32(first)      //nolint:gosec // a vertex number fits in int32
	s.rangeEnd[smallest] = lo + int32(len(s.found)) //nolint:gosec // a vertex number fits in int32
	s.cyclic[smallest] = len(s.found)-first > 1 || s.selfLoop[v]
}

// circuits finds the cycles through start in its component, whose vertices
// are members, with Johnson's CIRCUIT procedure without recursion, and stops
// at the limit.
func (s *cycleSearch[T]) circuits(start int32, members []int32) {
	s.current = start
	for _, v := range members {
		s.blocked[v] = false
		for a := s.blockedBy[v]; a >= 0; a = s.nextBlocked[a] {
			s.inList[a] = false
		}
		s.blockedBy[v] = -1
	}

	s.blocked[start] = true
	s.path = append(s.path[:0], start)
	s.frames = append(s.frames[:0], cycleFrame{vertex: start, arc: s.head[start]})
	for len(s.frames) > 0 {
		f := &s.frames[len(s.frames)-1]
		v := f.vertex
		if f.arc < s.head[v+1] {
			w := s.arcTo[f.arc]
			f.arc++
			switch {
			case s.comp[w] != s.current:
			case w == start:
				s.addCycle()
				f.found = true
				if len(s.cycles) == s.limit {
					s.frames = s.frames[:0]
					return
				}
			case !s.blocked[w]:
				s.blocked[w] = true
				s.path = append(s.path, w)
				s.frames = append(s.frames, cycleFrame{vertex: w, arc: s.head[w]})
			}
			continue
		}

		found := f.found
		if found {
			s.unblock(v)
		} else {
			// v stays blocked until one of its neighbors is unblocked.
			for a := s.head[v]; a < s.head[v+1]; a++ {
				if w := s.arcTo[a]; s.comp[w] == s.current && !s.inList[a] {
					s.inList[a] = true
					s.nextBlocked[a] = s.blockedBy[w]
					s.blockedBy[w] = a
				}
			}
		}
		s.path = s.path[:len(s.path)-1]
		s.frames = s.frames[:len(s.frames)-1]
		if found && len(s.frames) > 0 {
			s.frames[len(s.frames)-1].found = true
		}
	}
}

// unblock unblocks v, and each vertex that waits for a vertex it unblocks.
func (s *cycleSearch[T]) unblock(v int32) {
	s.blocked[v] = false
	s.unblocking = append(s.unblocking[:0], v)
	for len(s.unblocking) > 0 {
		u := s.unblocking[len(s.unblocking)-1]
		s.unblocking = s.unblocking[:len(s.unblocking)-1]
		for a := s.blockedBy[u]; a >= 0; a = s.nextBlocked[a] {
			s.inList[a] = false
			if w := s.arcFrom[a]; s.blocked[w] {
				s.blocked[w] = false
				s.unblocking = append(s.unblocking, w)
			}
		}
		s.blockedBy[u] = -1
	}
}

// addCycle copies the search path as a cycle.
func (s *cycleSearch[T]) addCycle() {
	cycle := make([]*gograph.Vertex[T], len(s.path))
	for i, v := range s.path {
		cycle[i] = s.vertices[v]
	}
	s.cycles = append(s.cycles, cycle)
}
