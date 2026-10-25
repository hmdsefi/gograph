package gograph

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
)

// checkOrder fails the test unless g has a topological order that gives
// each vertex its own place and puts the source of every edge before its
// destination.
func checkOrder[T comparable](t *testing.T, g *baseGraph[T], step string) {
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
// placed outside the endpoints, or if a rejected edge moved any vertex.
func checkPlacesKept(t *testing.T, g *baseGraph[int], before map[int]int, from, to int, added bool, step string) {
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
		case moved && label != from && label != to && (before[label] < lower || before[label] > upper):
			t.Fatalf("%s: edge %d -> %d moved %d from place %d, outside [%d, %d]",
				step, from, to, label, before[label], lower, upper)
		}
	}
}

func TestAcyclicOrder_RandomOperations(t *testing.T) {
	const n = 24
	for seed := range int64(6) {
		rng := rand.New(rand.NewSource(seed)) //nolint:gosec // seeded so failures are reproducible
		g := newBaseGraph[int](newProperties(Acyclic()))
		ref := New[int](Directed())
		for _, label := range rng.Perm(n) {
			g.AddVertexByLabel(label)
			ref.AddVertexByLabel(label)
		}

		for step := range 3000 {
			name := fmt.Sprintf("seed %d step %d", seed, step)
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
				if _, err := g.AddEdge(g.GetVertexByID(from), g.GetVertexByID(to)); !errors.Is(err, want) || (want == nil) != (err == nil) {
					t.Fatalf("%s: AddEdge(%d, %d) = %v, want %v", name, from, to, err, want)
				}
				checkPlacesKept(t, g, before, from, to, want == nil, name)
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
		}

		if g.Size() != ref.Size() {
			t.Fatalf("seed %d: Size() = %d, want %d", seed, g.Size(), ref.Size())
		}
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
