package gograph

import (
	"cmp"
	"slices"
)

// An acyclic graph keeps a topological order of its vertices while edges are
// added, with the dynamic topological sort of D. J. Pearce and P. H. J. Kelly,
// "A Dynamic Topological Sort Algorithm for Directed Acyclic Graphs", ACM
// Journal of Experimental Algorithmics 11 (2007). An edge that follows the
// order needs no search, and any other edge only visits the vertices placed
// between its endpoints. Removing edges or vertices keeps the order valid.

// placeEdge updates the order for a new edge from "from" to "to". It reports
// false, and changes nothing, when the edge would close a cycle.
func (g *baseGraph[T]) placeEdge(from, to *Vertex[T]) bool {
	if from == to {
		return false
	}
	if len(g.ord) != len(g.order) {
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

	ahead, ok := g.forwardRegion(to, from, upper)
	if !ok {
		return false
	}
	g.reorder(g.backwardRegion(from, lower), ahead)

	return true
}

// forwardRegion returns the vertices that "to" reaches and that are placed
// before upper, the place of "from". It reports false if "from" is one of
// them.
func (g *baseGraph[T]) forwardRegion(to, from *Vertex[T], upper int) ([]*Vertex[T], bool) {
	mark := g.nextSearch()
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

// backwardRegion returns the vertices that reach "from" and that are placed
// after lower, the place of "to".
func (g *baseGraph[T]) backwardRegion(from *Vertex[T], lower int) []*Vertex[T] {
	mark := g.nextSearch()
	from.seen = mark
	region := g.behind[:0]
	region = append(region, from)
	for i := 0; i < len(region); i++ {
		for _, edge := range region[i].incoming {
			if s := edge.source; s.seen != mark && g.ord[s.position] > lower {
				s.seen = mark
				region = append(region, s)
			}
		}
	}

	return region
}

// reorder gives the vertices of behind and ahead the places they hold
// between them, behind first, and keeps the order within each set.
func (g *baseGraph[T]) reorder(behind, ahead []*Vertex[T]) {
	byPlace := func(a, b *Vertex[T]) int {
		return cmp.Compare(g.ord[a.position], g.ord[b.position])
	}
	slices.SortFunc(behind, byPlace)
	slices.SortFunc(ahead, byPlace)

	places := g.places[:0]
	for _, v := range behind {
		places = append(places, g.ord[v.position])
	}
	for _, v := range ahead {
		places = append(places, g.ord[v.position])
	}
	slices.Sort(places)

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

// buildOrder sets ord to a topological order found with Kahn's algorithm.
// It runs when the graph has no order yet, as after Clone or
// InducedSubgraph.
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
