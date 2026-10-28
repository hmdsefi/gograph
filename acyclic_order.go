package gograph

import (
	"cmp"
	"math"
	"slices"
)

// An acyclic graph keeps a topological order of its vertices while edges are
// added, with the dynamic topological sort of D. J. Pearce and P. H. J. Kelly,
// "A Dynamic Topological Sort Algorithm for Directed Acyclic Graphs", ACM
// Journal of Experimental Algorithmics 11 (2007). An edge that follows the
// order needs no search. For any other edge, a forward search from its
// destination among the vertices placed before its source finds any cycle.
// The order then changes in one of two ways: the vertices placed after the
// destination that reach the source swap places with the ones the forward
// search found, or everything the destination reaches moves to the end. The
// searches for the two run side by side and the first to finish is used, so
// an edge costs at most a constant factor more than one search over the
// vertices its destination reaches, plus sorting the vertices that move.
// Removing edges or vertices keeps the order valid.

// A walk is a breadth-first search that can stop after a number of edges and
// go on later. list holds the vertices found so far, i is the vertex whose
// edges come next, and j is its next edge.
type walk[T comparable] struct {
	list []*Vertex[T]
	i, j int
}

// placeEdge updates the order for a new edge from "from" to "to". It reports
// false, and changes nothing, when the edge would close a cycle.
func (g *baseGraph[T]) placeEdge(from, to *Vertex[T]) bool {
	if from == to {
		return false
	}
	// moving every vertex to the end at once must not overflow last
	if len(g.ord) != len(g.order) || g.first == math.MinInt || g.last > math.MaxInt-len(g.order) {
		g.buildOrder()
	}

	upper, lower := g.ord[from.position], g.ord[to.position]
	switch {
	case upper < lower:
		return true
	case len(from.incoming) == 0:
		// nothing reaches "from", so it can go first
		g.first--
		g.ord[from.position] = g.first
		return true
	case len(to.neighbors) == 0:
		// "to" reaches nothing, so it can go last
		g.last++
		g.ord[to.position] = g.last
		return true
	}

	mark := g.nextSearch()
	ahead, ok := g.forwardRegion(to, from, upper, mark)
	if !ok {
		return false
	}

	// Without a cycle, no vertex that reaches "from" is reachable from "to",
	// so the walks never meet and can share one search number. The walk that
	// reaches everything starts over the forward region, so its list keeps
	// that region as a prefix.
	from.seen = mark
	behind := walk[T]{list: append(g.behind[:0], from)}
	reach := walk[T]{list: ahead}
	for budget := 2 * len(ahead); ; budget *= 2 {
		if g.growBehind(&behind, mark, lower, budget) {
			clear(reach.list[len(ahead):])
			g.reorder(behind.list, reach.list[:len(ahead)])
			return true
		}
		if g.growReach(&reach, mark, budget) {
			g.behind = release(behind.list)
			g.moveToEnd(reach.list)
			return true
		}
	}
}

// forwardRegion returns the vertices that "to" reaches and that are placed
// before upper, the place of "from", marked with mark. It reports false if
// "from" is one of them.
func (g *baseGraph[T]) forwardRegion(to, from *Vertex[T], upper int, mark uint32) ([]*Vertex[T], bool) {
	to.seen = mark
	region := g.stack[:0]
	region = append(region, to)
	for i := 0; i < len(region); i++ {
		for _, n := range region[i].neighbors {
			if n == from {
				g.stack = release(region)
				return nil, false
			}
			if n.seen != mark && g.ord[n.position] < upper {
				n.seen = mark
				region = append(region, n)
			}
		}
	}

	return region, true
}

// growBehind goes on with the search for the vertices that reach "from" and
// are placed after lower, the place of "to", for up to budget edges. It
// reports whether the search is complete.
func (g *baseGraph[T]) growBehind(w *walk[T], mark uint32, lower, budget int) bool {
	for ; w.i < len(w.list); w.i, w.j = w.i+1, 0 {
		in := w.list[w.i].incoming
		for ; w.j < len(in); w.j++ {
			if budget == 0 {
				return false
			}
			budget--
			if s := in[w.j].source; s.seen != mark && g.ord[s.position] > lower {
				s.seen = mark
				w.list = append(w.list, s)
			}
		}
	}

	return true
}

// growReach goes on with the search for every vertex that "to" reaches, for
// up to budget edges. It reports whether the search is complete.
func (g *baseGraph[T]) growReach(w *walk[T], mark uint32, budget int) bool {
	for ; w.i < len(w.list); w.i, w.j = w.i+1, 0 {
		out := w.list[w.i].neighbors
		for ; w.j < len(out); w.j++ {
			if budget == 0 {
				return false
			}
			budget--
			if n := out[w.j]; n.seen != mark {
				n.seen = mark
				w.list = append(w.list, n)
			}
		}
	}

	return true
}

// reorder gives the vertices of behind and ahead the places they hold
// between them, behind first, and keeps the order within each set.
func (g *baseGraph[T]) reorder(behind, ahead []*Vertex[T]) {
	g.sortByPlace(behind)
	g.sortByPlace(ahead)

	places := g.places[:0]
	for i, j := 0, 0; i < len(behind) || j < len(ahead); {
		if j == len(ahead) || i < len(behind) && g.ord[behind[i].position] < g.ord[ahead[j].position] {
			places = append(places, g.ord[behind[i].position])
			i++
		} else {
			places = append(places, g.ord[ahead[j].position])
			j++
		}
	}

	for i, v := range behind {
		g.ord[v.position] = places[i]
	}
	for i, v := range ahead {
		g.ord[v.position] = places[len(behind)+i]
	}

	g.places = places[:0]
	g.behind = release(behind)
	g.stack = release(ahead)
}

// moveToEnd gives the vertices of reached new places after all others, and
// keeps their order. reached must hold every vertex that its vertices reach.
func (g *baseGraph[T]) moveToEnd(reached []*Vertex[T]) {
	g.sortByPlace(reached)
	for _, v := range reached {
		g.last++
		g.ord[v.position] = g.last
	}
	g.stack = release(reached)
}

func (g *baseGraph[T]) sortByPlace(list []*Vertex[T]) {
	slices.SortFunc(list, func(a, b *Vertex[T]) int {
		return cmp.Compare(g.ord[a.position], g.ord[b.position])
	})
}

// buildOrder sets ord to a topological order found with Kahn's algorithm.
// It runs when the graph has no order yet, as after Clone copies another
// implementation of Graph, and before first or last could overflow.
func (g *baseGraph[T]) buildOrder() {
	g.ord = make([]int, len(g.order))
	remaining := make([]int, len(g.order))
	queue := make([]*Vertex[T], 0, len(g.vertices))
	for _, v := range g.order {
		if v != nil {
			remaining[v.position] = len(v.incoming)
			if len(v.incoming) == 0 {
				queue = append(queue, v)
			}
		}
	}

	for i := 0; i < len(queue); i++ {
		v := queue[i]
		g.ord[v.position] = i
		for _, n := range v.neighbors {
			remaining[n.position]--
			if remaining[n.position] == 0 {
				queue = append(queue, n)
			}
		}
	}
	g.first, g.last = 0, len(queue)-1
}

// nextSearch returns a new number to mark the vertices of a search with.
// When the numbers wrap around, it clears the old marks first.
func (g *baseGraph[T]) nextSearch() uint32 {
	g.search++
	if g.search == 0 {
		for _, u := range g.vertices {
			u.seen = 0
		}
		g.search = 1
	}

	return g.search
}

// release clears a reused work list, so it doesn't keep removed vertices
// alive, and returns it empty.
func release[T comparable](list []*Vertex[T]) []*Vertex[T] {
	clear(list)
	return list[:0]
}
