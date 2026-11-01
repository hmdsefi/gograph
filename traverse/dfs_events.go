package traverse

import (
	"iter"
	"slices"
	"strconv"

	"github.com/hmdsefi/gograph"
)

// DFSEventKind tells which step of a depth-first search a DFSEvent reports.
type DFSEventKind int

const (
	// DiscoverVertex is the first time the search reaches a vertex.
	DiscoverVertex DFSEventKind = iota
	// FinishVertex comes once every vertex reachable from the vertex has been
	// handled.
	FinishVertex
	// TreeEdge is an edge to an undiscovered vertex. The tree edges form the
	// depth-first tree, and the source of each one is the parent of its
	// destination.
	TreeEdge
	// BackEdge is an edge to a vertex still on the current path, which closes
	// a cycle. A self-loop is a back edge.
	BackEdge
	// ForwardEdge is an edge to a finished descendant. It occurs in directed
	// graphs only.
	ForwardEdge
	// CrossEdge is an edge to a finished vertex that isn't a descendant. It
	// occurs in directed graphs only.
	CrossEdge
)

var dfsEventKindNames = [...]string{
	DiscoverVertex: "DiscoverVertex",
	FinishVertex:   "FinishVertex",
	TreeEdge:       "TreeEdge",
	BackEdge:       "BackEdge",
	ForwardEdge:    "ForwardEdge",
	CrossEdge:      "CrossEdge",
}

// String returns the name of the kind, such as "TreeEdge", or
// "DFSEventKind(n)" for a value outside the constants.
func (k DFSEventKind) String() string {
	if k < 0 || int(k) >= len(dfsEventKindNames) {
		return "DFSEventKind(" + strconv.Itoa(int(k)) + ")"
	}
	return dfsEventKindNames[k]
}

// DFSEvent is one step of a depth-first search.
type DFSEvent[T comparable] struct {
	Kind DFSEventKind
	// Vertex is set for DiscoverVertex and FinishVertex.
	Vertex *gograph.Vertex[T]
	// Edge is set for the edge kinds. Its source is the vertex the search is
	// at, and its destination the neighbor the search looks at.
	Edge *gograph.Edge[T]
	// Depth is the depth of Vertex, or of the source of Edge, in the
	// depth-first tree. The vertex a search starts from has depth 0.
	Depth int
}

// DFSEvents returns the events of a depth-first search from start. It
// discovers a vertex when it first enters it, then looks at its neighbors in
// the order of Neighbors, one edge event each, and finishes the vertex once
// all of them are handled. A tree edge discovers its destination right away,
// so the events of each vertex nest inside the events of its parent, like
// parentheses. A missing start yields nothing.
//
// In a directed graph, each edge from a vertex the search reaches comes
// once, as one of the four edge kinds. A forward edge goes to a finished
// vertex discovered after the edge's source, and a cross edge to one
// discovered before it. In an undirected graph, each edge comes once too:
// as a tree edge when it discovers a vertex, or as a back edge from the
// deeper of its two vertices. The search skips the edge back to the parent
// and the second direction of an edge it already reported, so only tree and
// back edges occur. A self-loop is a back edge in both kinds of graph.
//
// The search runs as the loop asks for the next event, so breaking out of
// the loop ends it there, and ranging over the sequence again searches
// again. It keeps its own stack, so the goroutine stack stays small on a
// long path. A vertex or an edge removed from g during the loop is skipped
// when the search gets to it.
//
// Ranging over all of it takes O(V+E) time for the V vertices and E edges
// that start reaches. It keeps O(V) state, plus the neighbors still to look
// at of each vertex on the current path, which is at most O(V+E).
func DFSEvents[T comparable](g gograph.Graph[T], start T) iter.Seq[DFSEvent[T]] {
	return func(yield func(DFSEvent[T]) bool) {
		if root := g.GetVertexByID(start); root != nil {
			newDFSSearch(g, 0).run(root, yield)
		}
	}
}

// DFSEventsAll returns the events of a depth-first search over all of g. It
// searches from each vertex not discovered yet, in the order of
// GetAllVertices, and each of those vertices has depth 0. The events follow
// the rules of DFSEvents, and a cross edge can go to a vertex of an earlier
// search.
//
// Ranging over all of it takes O(V+E) time. It keeps O(V) state, plus the
// neighbors still to look at of each vertex on the current path, which is at
// most O(V+E).
func DFSEventsAll[T comparable](g gograph.Graph[T]) iter.Seq[DFSEvent[T]] {
	return func(yield func(DFSEvent[T]) bool) {
		s := newDFSSearch(g, int(g.Order()))
		for root := range gograph.Vertices(g) {
			if _, found := s.order[root]; found || (!hasEdges(root) && !stored(g, root)) {
				continue
			}
			if !s.run(root, yield) {
				return
			}
		}
	}
}

// dfsSearch is the state of one depth-first search, kept across the roots
// of DFSEventsAll.
type dfsSearch[T comparable] struct {
	g        gograph.Graph[T]
	directed bool
	// order is the discovery number of each discovered vertex, from 1,
	// negated once the vertex is finished.
	order  map[*gograph.Vertex[T]]int
	frames []dfsFrame[T]
	// pending holds the neighbors still to look at of the vertices in
	// frames. Each frame's neighbors are in reverse, so the next one is last.
	pending []*gograph.Vertex[T]
}

// dfsFrame is a vertex on the current path.
type dfsFrame[T comparable] struct {
	vertex *gograph.Vertex[T]
	parent *gograph.Vertex[T] // nil for the vertex the search started from
	start  int                // the vertex's neighbors begin at pending[start]
}

// newDFSSearch returns a search with room for size vertices.
func newDFSSearch[T comparable](g gograph.Graph[T], size int) *dfsSearch[T] {
	return &dfsSearch[T]{
		g:        g,
		directed: g.IsDirected(),
		order:    make(map[*gograph.Vertex[T]]int, size),
	}
}

// run searches from root, and reports false once yield does.
func (s *dfsSearch[T]) run(root *gograph.Vertex[T], yield func(DFSEvent[T]) bool) bool {
	if !s.discover(root, nil, yield) {
		return false
	}

	for len(s.frames) > 0 {
		depth := len(s.frames) - 1
		f := s.frames[depth]
		if len(s.pending) == f.start {
			s.frames = s.frames[:depth]
			s.order[f.vertex] = -s.order[f.vertex]
			if !yield(DFSEvent[T]{Kind: FinishVertex, Vertex: f.vertex, Depth: depth}) {
				return false
			}
			continue
		}

		v := s.pending[len(s.pending)-1]
		s.pending = s.pending[:len(s.pending)-1]

		kind := TreeEdge
		order, found := s.order[v]
		switch {
		case !found:
		case order > 0:
			if !s.directed && v == f.parent {
				continue
			}
			kind = BackEdge
		case !s.directed:
			// v is a finished descendant, which reported this edge as a back
			// edge
			continue
		case -order > s.order[f.vertex]:
			kind = ForwardEdge
		default:
			kind = CrossEdge
		}

		edge := s.g.GetEdge(f.vertex, v)
		if edge == nil {
			continue
		}
		if !yield(DFSEvent[T]{Kind: kind, Edge: edge, Depth: depth}) {
			return false
		}
		if kind == TreeEdge && !s.discover(v, f.vertex, yield) {
			return false
		}
	}

	return true
}

// discover enters v, pushes its frame and neighbors, and yields the event.
func (s *dfsSearch[T]) discover(v, parent *gograph.Vertex[T], yield func(DFSEvent[T]) bool) bool {
	s.order[v] = len(s.order) + 1
	start := len(s.pending)
	s.frames = append(s.frames, dfsFrame[T]{vertex: v, parent: parent, start: start})
	for n := range v.AllNeighbors() {
		s.pending = append(s.pending, n)
	}
	slices.Reverse(s.pending[start:])

	return yield(DFSEvent[T]{Kind: DiscoverVertex, Vertex: v, Depth: len(s.frames) - 1})
}
