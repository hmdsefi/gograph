package gograph

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"testing"
)

func TestVertex_AllNeighbors(t *testing.T) {
	g := New[string](Directed())
	a := g.AddVertexByLabel("A")
	for _, label := range []string{"C", "B", "D"} {
		if _, err := g.AddEdge(a, g.AddVertexByLabel(label)); err != nil {
			t.Fatal(err)
		}
	}

	var got []string
	for n := range a.AllNeighbors() {
		if n != g.GetVertexByID(n.Label()) {
			t.Fatalf("AllNeighbors yielded a copy of %s", n.Label())
		}
		got = append(got, n.Label())
	}
	if want := vertexLabels(a.Neighbors()); !slices.Equal(got, want) {
		t.Fatalf("AllNeighbors() = %v, want the order of Neighbors() %v", got, want)
	}

	count := 0
	for range a.AllNeighbors() {
		count++
		break
	}
	if count != 1 {
		t.Fatalf("break after the first neighbor visited %d", count)
	}

	g.RemoveVertices(a)
	for n := range a.AllNeighbors() {
		t.Fatalf("removed vertex yielded %s", n.Label())
	}
}

func TestVertex_AllNeighborsUndirectedSelfLoop(t *testing.T) {
	g := New[string]()
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	if _, err := g.AddEdge(a, a); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddEdge(a, b); err != nil {
		t.Fatal(err)
	}

	got := vertexLabels(slices.Collect(a.AllNeighbors()))
	if want := []string{"A", "B"}; !slices.Equal(got, want) {
		t.Fatalf("AllNeighbors() = %v, want %v", got, want)
	}
	if got := vertexLabels(slices.Collect(b.AllNeighbors())); !slices.Equal(got, []string{"A"}) {
		t.Fatalf("AllNeighbors() of B = %v, want [A]", got)
	}
}

func TestVertex_AllNeighborsRemovedDuringLoop(t *testing.T) {
	g := New[int](Directed())
	a := g.AddVertexByLabel(0)
	for i := 1; i <= 3; i++ {
		if _, err := g.AddEdge(a, g.AddVertexByLabel(i)); err != nil {
			t.Fatal(err)
		}
	}

	for n := range a.AllNeighbors() {
		if n == nil {
			t.Fatal("AllNeighbors yielded nil")
		}
		g.RemoveVertices(n)
	}
}

func TestAddEdge_AcyclicSkipsSearchWithoutPath(t *testing.T) {
	g := newBaseGraph[int](newProperties(Acyclic()))
	a, b, c, d := g.AddVertexByLabel(1), g.AddVertexByLabel(2), g.AddVertexByLabel(3), g.AddVertexByLabel(4)
	if _, err := g.AddEdge(a, b); err != nil {
		t.Fatal(err)
	}

	// c has no outgoing edges, and d has no incoming edges
	for _, e := range [][2]*Vertex[int]{{b, c}, {d, a}} {
		search := g.search
		if _, err := g.AddEdge(e[0], e[1]); err != nil {
			t.Fatal(err)
		}
		if g.search != search {
			t.Fatalf("AddEdge(%d, %d) ran a search", e[0].label, e[1].label)
		}
	}
}

// edgesOfByScan is EdgesOf written against the Graph methods: the outgoing
// edges in AllEdges order, then the incoming ones in GetAllVertices order.
func edgesOfByScan[T comparable](g Graph[T], v *Vertex[T]) []*Edge[T] {
	var edges []*Edge[T]
	for _, e := range g.AllEdges() {
		if e.Source() == v {
			edges = append(edges, e)
		}
	}
	for _, u := range g.GetAllVertices() {
		if u == v {
			continue
		}
		if e := g.GetEdge(u, v); e != nil {
			edges = append(edges, e)
		}
	}

	return edges
}

// checkIncoming fails the test unless every vertex of g lists exactly the
// stored edges that end at it, and EdgesOf matches edgesOfByScan.
func checkIncoming[T comparable](t *testing.T, g *baseGraph[T], step string) {
	t.Helper()

	want := make(map[*Vertex[T]][]*Edge[T])
	for _, dests := range g.edges {
		for _, e := range dests {
			want[e.dest] = append(want[e.dest], e)
		}
	}

	for _, v := range g.GetAllVertices() {
		w := want[v]
		listed := make(map[*Edge[T]]int, len(v.incoming))
		for _, e := range v.incoming {
			listed[e]++
		}
		for _, e := range w {
			listed[e]--
		}
		for e, count := range listed {
			if count != 0 {
				t.Fatalf("%s: vertex %v lists edge %v -> %v %+d times too often",
					step, v.label, e.source.label, e.dest.label, count)
			}
		}
		if v.InDegree() != len(w) {
			t.Fatalf("%s: vertex %v has InDegree() %d, want %d", step, v.label, v.InDegree(), len(w))
		}
		if edges, scan := g.EdgesOf(v), edgesOfByScan[T](g, v); !slices.Equal(edges, scan) {
			t.Fatalf("%s: EdgesOf(%v) = %v, want %v", step, v.label, edges, scan)
		}
	}
}

// methodsOnly hides the graph type, so Clone copies it through its methods.
type methodsOnly[T comparable] struct{ Graph[T] }

func asBaseGraph[T comparable](t *testing.T, g Graph[T]) *baseGraph[T] {
	t.Helper()

	bg, ok := g.(*baseGraph[T])
	if !ok {
		t.Fatalf("got %T, want a graph from New", g)
	}

	return bg
}

func TestEdgesOf_RandomOperations(t *testing.T) {
	for _, options := range [][]GraphOptionFunc{nil, {Directed()}, {Acyclic()}} {
		rng := rand.New(rand.NewSource(int64(len(options)))) //nolint:gosec // seeded so failures are reproducible
		g := newBaseGraph[int](newProperties(options...))
		const n = 12
		for i := range n {
			g.AddVertexByLabel(i)
		}

		for step := range 3000 {
			name := fmt.Sprintf("directed=%v step %d", g.IsDirected(), step)
			switch r := rng.Intn(10); {
			case r < 6:
				from, to := NewVertex(rng.Intn(n)), NewVertex(rng.Intn(n))
				_, _ = g.AddEdge(from, to)
			case r < 8:
				if edges := g.AllEdges(); len(edges) > 0 {
					g.RemoveEdges(edges[rng.Intn(len(edges))])
				}
			case r < 9:
				g.RemoveVertices(NewVertex(rng.Intn(n)))
			default:
				g.AddVertexByLabel(rng.Intn(n))
			}
			checkIncoming(t, g, name)
		}
	}
}

func TestAddEdge_AcyclicMatchesCycleCheck(t *testing.T) {
	rng := rand.New(rand.NewSource(7)) //nolint:gosec // seeded so failures are reproducible
	const n = 20
	g := New[int](Acyclic())
	ref := New[int](Directed())
	for i := range n {
		g.AddVertexByLabel(i)
		ref.AddVertexByLabel(i)
	}

	for step := range 4000 {
		switch r := rng.Intn(10); {
		case r < 8:
			from, to := rng.Intn(n), rng.Intn(n)
			_, refErr := ref.AddEdge(NewVertex(from), NewVertex(to))
			wantCycle := false
			if refErr == nil {
				if _, err := TopologySort(ref); err != nil {
					wantCycle = true
					ref.RemoveEdges(ref.GetEdge(ref.GetVertexByID(from), ref.GetVertexByID(to)))
				}
			}

			_, err := g.AddEdge(NewVertex(from), NewVertex(to))
			switch {
			case wantCycle && !errors.Is(err, ErrDAGCycle):
				t.Fatalf("step %d: AddEdge(%d, %d) = %v, want ErrDAGCycle", step, from, to, err)
			case !wantCycle && (err == nil) != (refErr == nil):
				t.Fatalf("step %d: AddEdge(%d, %d) = %v, want %v", step, from, to, err, refErr)
			}
		case r < 9:
			if edges := ref.AllEdges(); len(edges) > 0 {
				e := edges[rng.Intn(len(edges))]
				ref.RemoveEdges(e)
				g.RemoveEdges(g.GetEdge(g.GetVertexByID(e.Source().Label()), g.GetVertexByID(e.Destination().Label())))
			}
		default:
			label := rng.Intn(n)
			ref.RemoveVertices(NewVertex(label))
			g.RemoveVertices(NewVertex(label))
			ref.AddVertexByLabel(label)
			g.AddVertexByLabel(label)
		}
	}

	if g.Size() != ref.Size() {
		t.Fatalf("Size() = %d, want %d", g.Size(), ref.Size())
	}
	checkIncoming(t, asBaseGraph(t, g), "end")
}

func TestAddEdge_AcyclicSearchWraparound(t *testing.T) {
	g := newBaseGraph[int](newProperties(Acyclic()))
	a, b, c := g.AddVertexByLabel(1), g.AddVertexByLabel(2), g.AddVertexByLabel(3)
	for _, e := range [][2]*Vertex[int]{{a, b}, {b, c}} {
		if _, err := g.AddEdge(e[0], e[1]); err != nil {
			t.Fatal(err)
		}
	}

	// the next search number wraps around, and b still holds the mark of search 1
	g.search = ^uint32(0)
	b.seen = 1
	if _, err := g.AddEdge(c, a); !errors.Is(err, ErrDAGCycle) {
		t.Fatalf("AddEdge(3, 1) error = %v, want ErrDAGCycle", err)
	}
}

func TestClone_KeepsIncomingEdges(t *testing.T) {
	for _, options := range [][]GraphOptionFunc{nil, {Directed()}} {
		g, _ := benchGraph(40, benchEdges(40, 120, false), options...)
		for _, c := range []Graph[int]{Clone(g), Clone[int](methodsOnly[int]{g})} {
			bc := asBaseGraph(t, c)
			checkIncoming(t, bc, "clone")
			bc.RemoveVertices(bc.GetVertexByID(3), bc.GetVertexByID(17))
			checkIncoming(t, bc, "clone after removal")
		}
	}
}

func TestInducedSubgraph_KeepsIncomingEdges(t *testing.T) {
	for _, options := range [][]GraphOptionFunc{nil, {Directed()}} {
		g, _ := benchGraph(40, benchEdges(40, 200, false), options...)
		labels := make([]int, 0, 20)
		for i := 0; i < 40; i += 2 {
			labels = append(labels, i)
		}
		sub, err := InducedSubgraph(g, labels...)
		if err != nil {
			t.Fatal(err)
		}
		bs := asBaseGraph(t, sub)
		checkIncoming(t, bs, "induced")
		bs.RemoveVertices(bs.GetVertexByID(4))
		checkIncoming(t, bs, "induced after removal")
	}
}

func TestRemoveItem(t *testing.T) {
	a, b, c, d := NewVertex(1), NewVertex(2), NewVertex(3), NewVertex(4)
	tests := []struct {
		name string
		x    *Vertex[int]
		want []*Vertex[int]
	}{
		{"first", a, []*Vertex[int]{b, c, d}},
		{"last", d, []*Vertex[int]{a, b, c}},
		{"middle", b, []*Vertex[int]{a, c, d}},
		{"missing", NewVertex(5), []*Vertex[int]{a, b, c, d}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			backing := []*Vertex[int]{a, b, c, d}
			got := removeItem(backing, tc.x)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			if kept := slices.DeleteFunc(slices.Clone(backing), func(v *Vertex[int]) bool { return v == nil }); len(kept) != len(tc.want) {
				t.Errorf("the backing array still holds %d vertices, want %d", len(kept), len(tc.want))
			}
		})
	}

	if got := removeItem[*Vertex[int]](nil, a); got != nil {
		t.Errorf("got %v from an empty list", got)
	}
}
