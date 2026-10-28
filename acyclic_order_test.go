package gograph

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"testing"
)

// checkOrder fails the test unless g has a topological order that gives
// each vertex its own place and puts the source of every edge before its
// destination.
func checkOrder[T comparable](t testing.TB, g *baseGraph[T], step string) {
	t.Helper()

	if len(g.ord) != len(g.order) {
		t.Fatalf("%s: order has %d places for %d slots", step, len(g.ord), len(g.order))
	}
	owner := make(map[int]T, len(g.vertices))
	for _, v := range g.order {
		if v == nil {
			continue
		}
		place := g.ord[v.position]
		if other, ok := owner[place]; ok {
			t.Fatalf("%s: %v and %v share place %d", step, other, v.label, place)
		}
		if place < g.first || place > g.last {
			t.Fatalf("%s: %v has place %d outside [%d, %d]", step, v.label, place, g.first, g.last)
		}
		owner[place] = v.label
	}
	for _, e := range g.AllEdges() {
		if from, to := g.ord[e.source.position], g.ord[e.dest.position]; from >= to {
			t.Fatalf("%s: edge %v -> %v goes from place %d to %d", step, e.source.label, e.dest.label, from, to)
		}
	}
}

// checkWorkListsCleared fails the test if a reused work list of g still holds
// a vertex, which would keep a removed vertex alive.
func checkWorkListsCleared[T comparable](t testing.TB, g *baseGraph[T], step string) {
	t.Helper()

	for name, list := range map[string][]*Vertex[T]{"stack": g.stack, "behind": g.behind} {
		if slices.ContainsFunc(list[:cap(list)], func(v *Vertex[T]) bool { return v != nil }) {
			t.Fatalf("%s: the %s work list still holds vertices", step, name)
		}
	}
}

// reachesByScan reports whether "to" reaches "from" in g, following labels
// through GetVertexByID.
func reachesByScan(g Graph[int], to, from int) bool {
	seen := map[int]bool{to: true}
	queue := []int{to}
	for len(queue) > 0 {
		label := queue[0]
		queue = queue[1:]
		if label == from {
			return true
		}
		for _, n := range g.GetVertexByID(label).Neighbors() {
			if !seen[n.Label()] {
				seen[n.Label()] = true
				queue = append(queue, n.Label())
			}
		}
	}

	return false
}

func places(g *baseGraph[int]) map[int]int {
	if len(g.ord) != len(g.order) {
		return nil
	}
	out := make(map[int]int, len(g.vertices))
	for label, v := range g.vertices {
		out[label] = g.ord[v.position]
	}

	return out
}

// checkPlacesKept fails the test if adding the edge from -> to moved a vertex
// placed outside the endpoints that "to" doesn't reach in ref, or if a
// rejected edge moved any vertex.
func checkPlacesKept(
	t *testing.T, g *baseGraph[int], ref Graph[int], before map[int]int, from, to int, added bool, step string,
) {
	t.Helper()

	if before == nil {
		return
	}
	lower, upper := before[to], before[from]
	for label, place := range places(g) {
		moved := place != before[label]
		switch {
		case moved && !added:
			t.Fatalf("%s: rejected edge %d -> %d moved %d", step, from, to, label)
		case moved && label != from && label != to && (before[label] < lower || before[label] > upper) &&
			!reachesByScan(ref, to, label):
			t.Fatalf("%s: edge %d -> %d moved %d from place %d, outside [%d, %d] and unreachable from %d",
				step, from, to, label, before[label], lower, upper, to)
		}
	}
}

func TestAcyclicOrder_RandomOperations(t *testing.T) {
	for _, tc := range []struct{ n, seeds int }{{24, 6}, {200, 3}} {
		for seed := range int64(tc.seeds) {
			randomOperations(t, tc.n, seed)
		}
	}
}

// randomOperations compares an acyclic graph with a plain directed graph over
// random edge adds and removals and vertex replacements on n vertices.
func randomOperations(t *testing.T, n int, seed int64) {
	t.Helper()

	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // seeded so failures are reproducible
	g := newBaseGraph[int](newProperties(Acyclic()))
	ref := New[int](Directed())
	for _, label := range rng.Perm(n) {
		g.AddVertexByLabel(label)
		ref.AddVertexByLabel(label)
	}

	for step := range 3000 {
		name := fmt.Sprintf("n %d seed %d step %d", n, seed, step)
		switch r := rng.Intn(20); {
		case r < 16:
			from, to := rng.Intn(n), rng.Intn(n)
			var want error
			switch {
			case ref.ContainsEdge(ref.GetVertexByID(from), ref.GetVertexByID(to)):
				want = ErrEdgeAlreadyExists
			case reachesByScan(ref, to, from):
				want = ErrDAGCycle
			default:
				if _, err := ref.AddEdge(ref.GetVertexByID(from), ref.GetVertexByID(to)); err != nil {
					t.Fatal(err)
				}
			}
			before := places(g)
			// the errors come back unwrapped, so they compare equal
			if _, err := g.AddEdge(g.GetVertexByID(from), g.GetVertexByID(to)); err != want { //nolint:errorlint // checks for no wrapping
				t.Fatalf("%s: AddEdge(%d, %d) = %v, want %v", name, from, to, err, want)
			}
			checkPlacesKept(t, g, ref, before, from, to, want == nil, name)
		case r < 18:
			if edges := ref.AllEdges(); len(edges) > 0 {
				e := edges[rng.Intn(len(edges))]
				ref.RemoveEdges(e)
				g.RemoveEdges(g.GetEdge(g.GetVertexByID(e.Source().Label()), g.GetVertexByID(e.Destination().Label())))
			}
		default:
			label := rng.Intn(n)
			ref.RemoveVertices(ref.GetVertexByID(label))
			g.RemoveVertices(g.GetVertexByID(label))
			ref.AddVertexByLabel(label)
			g.AddVertexByLabel(label)
		}
		checkOrder(t, g, name)
		checkWorkListsCleared(t, g, name)
	}

	if g.Size() != ref.Size() {
		t.Fatalf("n %d seed %d: Size() = %d, want %d", n, seed, g.Size(), ref.Size())
	}
}

func TestAcyclicOrder_ReordersLargeRegion(t *testing.T) {
	const n = 50
	g := newBaseGraph[int](newProperties(Acyclic()))
	vertices := make([]*Vertex[int], 2*n)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	// two chains, 0 to n-1 placed before n to 2n-1
	for i := 0; i+1 < n; i++ {
		for _, offset := range []int{0, n} {
			if _, err := g.AddEdge(vertices[offset+i], vertices[offset+i+1]); err != nil {
				t.Fatal(err)
			}
		}
	}

	// the end of the second chain now leads to the start of the first, so
	// the whole second chain moves ahead of the first
	if _, err := g.AddEdge(vertices[2*n-1], vertices[0]); err != nil {
		t.Fatal(err)
	}
	checkOrder(t, g, "after the reorder")
	if g.ord[vertices[n].position] >= g.ord[vertices[n-1].position] {
		t.Fatalf("vertex %d is placed after vertex %d", n, n-1)
	}

	if _, err := g.AddEdge(vertices[n-1], vertices[n]); !errors.Is(err, ErrDAGCycle) {
		t.Fatalf("AddEdge(%d, %d) error = %v, want ErrDAGCycle", n-1, n, err)
	}
	checkOrder(t, g, "after the rejected edge")
}

func TestAcyclicOrder_SearchStaysBetweenEndpoints(t *testing.T) {
	g := newBaseGraph[string](newProperties(Acyclic()))
	p := g.AddVertexByLabel("p")
	y := g.AddVertexByLabel("y")
	x := g.AddVertexByLabel("x")
	q := g.AddVertexByLabel("q")
	// p reaches x from before y, and y reaches q after x
	for _, e := range [][2]*Vertex[string]{{p, x}, {y, q}} {
		if _, err := g.AddEdge(e[0], e[1]); err != nil {
			t.Fatal(err)
		}
	}

	last := g.search
	if _, err := g.AddEdge(x, y); err != nil {
		t.Fatal(err)
	}
	checkOrder(t, g, "after x -> y")
	if x.seen <= last || y.seen <= last {
		t.Fatalf("x -> y didn't search between its endpoints")
	}
	for _, v := range []*Vertex[string]{p, q} {
		if v.seen > last {
			t.Errorf("x -> y visited %s, placed outside its endpoints", v.label)
		}
	}
}

func TestAcyclicOrder_DerivedGraphs(t *testing.T) {
	g := New[int](Acyclic())
	vertices := make([]*Vertex[int], 4)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for i := 0; i+1 < len(vertices); i++ {
		if _, err := g.AddEdge(vertices[i], vertices[i+1]); err != nil {
			t.Fatal(err)
		}
	}
	sub, err := InducedSubgraph(g, 3, 2, 1, 0)
	if err != nil {
		t.Fatal(err)
	}

	derived := map[string]Graph[int]{
		"Clone":            Clone(g),
		"Clone of methods": Clone[int](methodsOnly[int]{g}),
		"InducedSubgraph":  sub,
	}
	for name, d := range derived {
		if _, err := d.AddEdge(d.GetVertexByID(3), d.GetVertexByID(1)); !errors.Is(err, ErrDAGCycle) {
			t.Fatalf("%s: AddEdge(3, 1) error = %v, want ErrDAGCycle", name, err)
		}
		if _, err := d.AddEdge(d.GetVertexByID(0), d.GetVertexByID(3)); err != nil {
			t.Fatalf("%s: AddEdge(0, 3) error = %v", name, err)
		}
		checkOrder(t, asBaseGraph(t, d), name)
	}
}

func TestAcyclicOrder_KeptThroughCompaction(t *testing.T) {
	g := newBaseGraph[int](newProperties(Acyclic()))
	vertices := make([]*Vertex[int], 10)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for i := 0; i+1 < len(vertices); i++ {
		if _, err := g.AddEdge(vertices[i], vertices[i+1]); err != nil {
			t.Fatal(err)
		}
	}

	g.RemoveVertices(vertices[1:7]...)
	if len(g.order) != 4 {
		t.Fatalf("order has %d slots, want 4 after compaction", len(g.order))
	}
	checkOrder(t, g, "after compaction")

	if _, err := g.AddEdge(vertices[9], vertices[7]); !errors.Is(err, ErrDAGCycle) {
		t.Fatalf("AddEdge(9, 7) error = %v, want ErrDAGCycle", err)
	}
	if _, err := g.AddEdge(vertices[9], vertices[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddEdge(vertices[0], vertices[7]); !errors.Is(err, ErrDAGCycle) {
		t.Fatalf("AddEdge(0, 7) error = %v, want ErrDAGCycle", err)
	}
	checkOrder(t, g, "after the new edges")
}

func TestAcyclicOrder_MovesReachedVerticesToEnd(t *testing.T) {
	const n = 100
	g := newBaseGraph[int](newProperties(Acyclic()))
	hubGraph(t, g, n, func(step string) { checkOrder(t, g, step) })

	// Each chain edge reaches only its destination and n, while all of the
	// chain above it reaches its source, so those two vertices move to the
	// end instead.
	if want := n + 2*(n-4); g.last < want {
		t.Fatalf("last place is %d, want at least %d after moving vertices to the end", g.last, want)
	}
	if _, err := g.AddEdge(g.GetVertexByID(0), g.GetVertexByID(n-1)); err != ErrDAGCycle { //nolint:errorlint // checks for no wrapping
		t.Fatalf("AddEdge(0, %d) error = %v, want ErrDAGCycle", n-1, err)
	}
}

func TestAcyclicOrder_PlacesNearLimits(t *testing.T) {
	tests := []struct {
		name string
		// before runs before d, e and f are added, and after runs after
		before, after func(g *baseGraph[int])
	}{
		{"last at the limit", func(g *baseGraph[int]) { g.last = math.MaxInt }, nil},
		{"last near the limit", nil, func(g *baseGraph[int]) { g.last = math.MaxInt - 1 }},
		{"first at the limit", nil, func(g *baseGraph[int]) { g.first = math.MinInt }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newBaseGraph[int](newProperties(Acyclic()))
			a, b, c := g.AddVertexByLabel(1), g.AddVertexByLabel(2), g.AddVertexByLabel(3)
			for _, e := range [][2]*Vertex[int]{{a, b}, {b, c}} {
				if _, err := g.AddEdge(e[0], e[1]); err != nil {
					t.Fatal(err)
				}
			}

			if tc.before != nil {
				tc.before(g)
			}
			d, e, f := g.AddVertexByLabel(4), g.AddVertexByLabel(5), g.AddVertexByLabel(6)
			if tc.after != nil {
				tc.after(g)
			}

			if _, err := g.AddEdge(c, a); err != ErrDAGCycle { //nolint:errorlint // checks for no wrapping
				t.Fatalf("AddEdge(3, 1) error = %v, want ErrDAGCycle", err)
			}
			// e -> c and c -> f move their destinations past the last place,
			// and d -> a moves d before the first place
			for _, edge := range [][2]*Vertex[int]{{d, e}, {e, c}, {c, f}, {d, a}} {
				if _, err := g.AddEdge(edge[0], edge[1]); err != nil {
					t.Fatal(err)
				}
			}
			checkOrder(t, g, tc.name)
		})
	}
}

func TestAcyclicOrder_DerivedGraphsKeepOrder(t *testing.T) {
	g := New[int](Acyclic())
	hubGraph(t, g, 20, nil)
	g.RemoveVertices(g.GetVertexByID(7))
	bg := asBaseGraph(t, g)
	checkOrder(t, bg, "original")
	want := places(bg)

	clone := asBaseGraph(t, Clone(g))
	sub, err := InducedSubgraph(g, 20, 3, 12, 0)
	if err != nil {
		t.Fatal(err)
	}
	for name, d := range map[string]*baseGraph[int]{"Clone": clone, "InducedSubgraph": asBaseGraph(t, sub)} {
		got := places(d)
		if got == nil {
			t.Fatalf("%s has no order", name)
		}
		for label, place := range got {
			if place != want[label] {
				t.Errorf("%s places %d at %d, want %d", name, label, place, want[label])
			}
		}
		if d.first != bg.first || d.last != bg.last {
			t.Errorf("%s has places [%d, %d], want [%d, %d]", name, d.first, d.last, bg.first, bg.last)
		}
		checkOrder(t, d, name)
	}
}
