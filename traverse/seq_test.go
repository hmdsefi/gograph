package traverse

import (
	"errors"
	"iter"
	"math"
	"math/rand"
	"slices"
	"testing"

	"github.com/hmdsefi/gograph"
)

// undirectedTestGraph returns an undirected graph with n vertices and m
// random edges, self-loops included.
func undirectedTestGraph(n int, m uint32) gograph.Graph[int] {
	rng := rand.New(rand.NewSource(3)) //nolint:gosec // seeded so failures are reproducible
	g := gograph.New[int]()
	vertices := make([]*gograph.Vertex[int], n)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for g.Size() < m {
		_, _ = g.AddEdge(vertices[rng.Intn(n)], vertices[rng.Intn(n)])
	}

	return g
}

// weightedUndirectedTestGraph returns an undirected weighted graph with n
// vertices and m random edges, where about one edge in ten weighs NaN and
// one in ten +Inf.
func weightedUndirectedTestGraph(n int, m uint32) gograph.Graph[int] {
	rng := rand.New(rand.NewSource(4)) //nolint:gosec // seeded so failures are reproducible
	g := gograph.New[int](gograph.Weighted())
	vertices := make([]*gograph.Vertex[int], n)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for g.Size() < m {
		weight := rng.Float64() * 10
		switch rng.Intn(10) {
		case 0:
			weight = math.NaN()
		case 1:
			weight = math.Inf(1)
		}
		_, _ = g.AddEdge(vertices[rng.Intn(n)], vertices[rng.Intn(n)], gograph.WithEdgeWeight(weight))
	}

	return g
}

func iteratorVertices[T comparable](t *testing.T, it Iterator[T], err error) []*gograph.Vertex[T] {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
	var out []*gograph.Vertex[T]
	for it.HasNext() {
		out = append(out, it.Next())
	}

	return out
}

func checkSame[T comparable](t *testing.T, name string, g gograph.Graph[T], got, want []*gograph.Vertex[T]) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Fatalf("%s: got %d vertices %v, want %d %v", name, len(got), labels(got), len(want), labels(want))
	}
	for _, v := range got {
		if g.GetVertexByID(v.Label()) != v {
			t.Fatalf("%s: vertex %v isn't the graph's own", name, v.Label())
		}
	}
}

func labels[T comparable](vertices []*gograph.Vertex[T]) []T {
	out := make([]T, len(vertices))
	for i, v := range vertices {
		out[i] = v.Label()
	}

	return out
}

func keys[K, V any](seq iter.Seq2[K, V]) []K {
	var out []K
	for k := range seq {
		out = append(out, k)
	}

	return out
}

func TestSequences_MatchIterators(t *testing.T) {
	graphs := map[string]gograph.Graph[int]{
		"directed":            benchGraph(300, 1200),
		"undirected":          undirectedTestGraph(300, 600),
		"weighted undirected": weightedUndirectedTestGraph(300, 600),
	}
	for name, g := range graphs {
		for start := 0; start < 300; start += 7 {
			it, err := NewBreadthFirstIterator(g, start)
			bfs := slices.Collect(BFS(g, start))
			checkSame(t, name+" BFS", g, bfs, iteratorVertices(t, it, err))
			checkSame(t, name+" BFSWithDepth", g, keys(BFSWithDepth(g, start)), bfs)

			it, err = NewDepthFirstIterator(g, start)
			checkSame(t, name+" DFS", g, slices.Collect(DFS(g, start)), iteratorVertices(t, it, err))

			it, err = NewClosestFirstIterator(g, start)
			if err != nil {
				t.Fatal(err)
			}
			closest, ok := it.(*closestFirstIterator[int])
			if !ok {
				t.Fatalf("iterator has type %T", it)
			}
			var want []*gograph.Vertex[int]
			var wantDist []float64
			for closest.HasNext() {
				want = append(want, closest.Next())
				wantDist = append(wantDist, closest.currDist)
			}
			var got []*gograph.Vertex[int]
			var gotDist []float64
			for v, d := range ClosestFirst(g, start) {
				got = append(got, v)
				gotDist = append(gotDist, d)
			}
			checkSame(t, name+" ClosestFirst", g, got, want)
			if !slices.Equal(gotDist, wantDist) {
				t.Fatalf("%s ClosestFirst from %d: distances %v, want %v", name, start, gotDist, wantDist)
			}
		}
	}
}

func TestTopological_MatchesIterator(t *testing.T) {
	rng := rand.New(rand.NewSource(5)) //nolint:gosec // seeded so failures are reproducible
	g := gograph.New[int](gograph.Directed())
	vertices := make([]*gograph.Vertex[int], 200)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for g.Size() < 800 {
		from, to := rng.Intn(200), rng.Intn(200)
		if from < to {
			_, _ = g.AddEdge(vertices[from], vertices[to])
		}
	}

	seq, err := Topological(g)
	if err != nil {
		t.Fatal(err)
	}
	it, err := NewTopologicalIterator(g)
	want := iteratorVertices(t, it, err)
	checkSame(t, "Topological", g, slices.Collect(seq), want)
	checkSame(t, "Topological again", g, slices.Collect(seq), want)
}

func TestTopological_Errors(t *testing.T) {
	cyclic := gograph.New[string](gograph.Directed())
	a, b := cyclic.AddVertexByLabel("a"), cyclic.AddVertexByLabel("b")
	_, _ = cyclic.AddEdge(a, b)
	_, _ = cyclic.AddEdge(b, a)

	undirected := gograph.New[string]()
	_, _ = undirected.AddEdge(gograph.NewVertex("a"), gograph.NewVertex("b"))

	for _, tc := range []struct {
		name string
		g    gograph.Graph[string]
		want error
	}{
		{"cycle", cyclic, gograph.ErrDAGHasCycle},
		{"undirected", undirected, gograph.ErrNotDirected},
	} {
		seq, err := Topological(tc.g)
		if !errors.Is(err, tc.want) {
			t.Fatalf("%s: error = %v, want %v", tc.name, err, tc.want)
		}
		if got := slices.Collect(seq); len(got) != 0 {
			t.Fatalf("%s: sequence yielded %v, want nothing", tc.name, labels(got))
		}
	}
}

// shortestHops finds the number of edges on a shortest path from start to
// each vertex it reaches, following labels through GetVertexByID.
func shortestHops(g gograph.Graph[int], start int) map[int]int {
	hops := map[int]int{start: 0}
	queue := []int{start}
	for len(queue) > 0 {
		label := queue[0]
		queue = queue[1:]
		for _, n := range g.GetVertexByID(label).Neighbors() {
			if _, ok := hops[n.Label()]; !ok {
				hops[n.Label()] = hops[label] + 1
				queue = append(queue, n.Label())
			}
		}
	}

	return hops
}

func TestBFSWithDepth_Depths(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	for _, e := range [][2]string{
		{"A", "B"}, {"A", "C"}, {"B", "D"}, {"C", "D"}, {"C", "E"}, {"D", "F"}, {"E", "F"}, {"F", "A"},
	} {
		_, _ = g.AddEdge(gograph.NewVertex(e[0]), gograph.NewVertex(e[1]))
	}

	got := map[string]int{}
	var order []string
	for v, depth := range BFSWithDepth(g, "A") {
		got[v.Label()] = depth
		order = append(order, v.Label())
	}
	want := map[string]int{"A": 0, "B": 1, "C": 1, "D": 2, "E": 2, "F": 3}
	if len(got) != len(want) {
		t.Fatalf("depths %v, want %v", got, want)
	}
	for label, depth := range want {
		if got[label] != depth {
			t.Fatalf("depth of %s = %d, want %d (all: %v)", label, got[label], depth, got)
		}
	}
	if !slices.Equal(order, []string{"A", "B", "C", "D", "E", "F"}) {
		t.Fatalf("order %v", order)
	}

	for name, rg := range map[string]gograph.Graph[int]{
		"directed":   benchGraph(300, 900),
		"undirected": undirectedTestGraph(300, 400),
	} {
		for start := 0; start < 300; start += 11 {
			hops := shortestHops(rg, start)
			n := 0
			for v, depth := range BFSWithDepth(rg, start) {
				if want, ok := hops[v.Label()]; !ok || depth != want {
					t.Fatalf("%s from %d: depth of %d = %d, want %d", name, start, v.Label(), depth, want)
				}
				n++
			}
			if n != len(hops) {
				t.Fatalf("%s from %d: yielded %d vertices, want %d", name, start, n, len(hops))
			}
		}
	}
}

func TestClosestFirst_Distances(t *testing.T) {
	g := initClosestFirstIteratorTestGraph()
	want := map[string]float64{"A": 0, "B": 2, "C": 3, "D": 5}
	n := 0
	for v, dist := range ClosestFirst(g, "A") {
		if dist != want[v.Label()] {
			t.Fatalf("distance of %s = %v, want %v", v.Label(), dist, want[v.Label()])
		}
		n++
	}
	if n != len(want) {
		t.Fatalf("yielded %d vertices, want %d", n, len(want))
	}
}

// edgeCounter counts the GetEdge calls a traversal makes.
type edgeCounter[T comparable] struct {
	gograph.Graph[T]
	calls int
}

func (c *edgeCounter[T]) GetEdge(from, to *gograph.Vertex[T]) *gograph.Edge[T] {
	c.calls++
	return c.Graph.GetEdge(from, to)
}

func TestSequences_StopEarly(t *testing.T) {
	g := benchGraph(300, 1200)
	topological, err := Topological(dagForStop())
	if err != nil {
		t.Fatal(err)
	}
	sequences := map[string]iter.Seq[*gograph.Vertex[int]]{
		"BFS": BFS(g, 0),
		"BFSWithDepth": func(yield func(*gograph.Vertex[int]) bool) {
			for v := range BFSWithDepth(g, 0) {
				if !yield(v) {
					return
				}
			}
		},
		"DFS": DFS(g, 0),
		"ClosestFirst": func(yield func(*gograph.Vertex[int]) bool) {
			for v := range ClosestFirst(g, 0) {
				if !yield(v) {
					return
				}
			}
		},
		"Topological": topological,
	}
	for name, seq := range sequences {
		for _, k := range []int{1, 2, 5} {
			n := 0
			for range seq {
				n++
				if n == k {
					break
				}
			}
			if n != k {
				t.Fatalf("%s: stopped after %d vertices, want %d", name, n, k)
			}
		}
	}

	// the first vertex comes before any edge is read
	counter := &edgeCounter[int]{Graph: g}
	for range ClosestFirst[int](counter, 0) {
		break
	}
	if counter.calls != 0 {
		t.Fatalf("ClosestFirst read %d edges before the first vertex", counter.calls)
	}
}

func dagForStop() gograph.Graph[int] {
	g := gograph.New[int](gograph.Directed())
	vertices := make([]*gograph.Vertex[int], 10)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for i := 0; i+1 < len(vertices); i++ {
		_, _ = g.AddEdge(vertices[i], vertices[i+1])
	}

	return g
}

func TestSequences_MissingStart(t *testing.T) {
	g := benchGraph(20, 40)
	if got := slices.Collect(BFS(g, 99)); len(got) != 0 {
		t.Fatalf("BFS yielded %v", labels(got))
	}
	if got := slices.Collect(DFS(g, 99)); len(got) != 0 {
		t.Fatalf("DFS yielded %v", labels(got))
	}
	if got := keys(BFSWithDepth(g, 99)); len(got) != 0 {
		t.Fatalf("BFSWithDepth yielded %v", labels(got))
	}
	if got := keys(ClosestFirst(g, 99)); len(got) != 0 {
		t.Fatalf("ClosestFirst yielded %v", labels(got))
	}
}

func TestSequences_RunAgain(t *testing.T) {
	g := benchGraph(100, 300)
	topological, err := Topological(dagForStop())
	if err != nil {
		t.Fatal(err)
	}
	sequences := map[string]iter.Seq[*gograph.Vertex[int]]{
		"BFS":          BFS(g, 0),
		"BFSWithDepth": func(yield func(*gograph.Vertex[int]) bool) { BFSWithDepth(g, 0)(withoutValue[int](yield)) },
		"DFS":          DFS(g, 0),
		"ClosestFirst": func(yield func(*gograph.Vertex[int]) bool) { ClosestFirst(g, 0)(withoutValue[float64](yield)) },
		"Topological":  topological,
	}
	for name, seq := range sequences {
		first := slices.Collect(seq)
		if len(first) < 2 {
			t.Fatalf("%s yielded %d vertices", name, len(first))
		}
		for v := range seq {
			if v != first[0] {
				t.Fatalf("%s: a broken loop started at %v, want %v", name, v.Label(), first[0].Label())
			}
			break
		}
		if second := slices.Collect(seq); !slices.Equal(second, first) {
			t.Fatalf("%s: the second range yielded %v, want %v", name, labels(second), labels(first))
		}
	}
}

// withoutValue adapts yield to a two-value sequence, dropping the value.
func withoutValue[V any](yield func(*gograph.Vertex[int]) bool) func(*gograph.Vertex[int], V) bool {
	return func(v *gograph.Vertex[int], _ V) bool { return yield(v) }
}

func TestSequences_SkipRemovedVertices(t *testing.T) {
	build := func() gograph.Graph[string] {
		g := gograph.New[string](gograph.Directed())
		for _, e := range [][2]string{{"A", "B"}, {"A", "C"}, {"A", "D"}, {"B", "E"}, {"C", "E"}, {"D", "E"}} {
			_, _ = g.AddEdge(gograph.NewVertex(e[0]), gograph.NewVertex(e[1]))
		}
		return g
	}
	// visit ranges over a traversal and, at its second vertex, removes the
	// neighbors of A it hasn't visited yet, which are already queued
	visit := func(g gograph.Graph[string], seq iter.Seq[*gograph.Vertex[string]]) []string {
		var out []string
		for v := range seq {
			out = append(out, v.Label())
			if len(out) == 2 {
				for _, label := range []string{"B", "C", "D"} {
					if !slices.Contains(out, label) {
						g.RemoveVertices(g.GetVertexByID(label))
					}
				}
			}
		}
		return out
	}
	iteratorSeq := func(it Iterator[string], err error) iter.Seq[*gograph.Vertex[string]] {
		if err != nil {
			t.Fatal(err)
		}
		return func(yield func(*gograph.Vertex[string]) bool) {
			for it.HasNext() && yield(it.Next()) {
			}
		}
	}

	for _, tc := range []struct {
		name string
		seq  func(g gograph.Graph[string]) iter.Seq[*gograph.Vertex[string]]
		it   func(g gograph.Graph[string]) (Iterator[string], error)
	}{
		{
			"BFS",
			func(g gograph.Graph[string]) iter.Seq[*gograph.Vertex[string]] { return BFS(g, "A") },
			func(g gograph.Graph[string]) (Iterator[string], error) { return NewBreadthFirstIterator(g, "A") },
		},
		{
			"DFS",
			func(g gograph.Graph[string]) iter.Seq[*gograph.Vertex[string]] { return DFS(g, "A") },
			func(g gograph.Graph[string]) (Iterator[string], error) { return NewDepthFirstIterator(g, "A") },
		},
		{
			"ClosestFirst",
			func(g gograph.Graph[string]) iter.Seq[*gograph.Vertex[string]] {
				return func(yield func(*gograph.Vertex[string]) bool) {
					for v := range ClosestFirst(g, "A") {
						if !yield(v) {
							return
						}
					}
				}
			},
			func(g gograph.Graph[string]) (Iterator[string], error) { return NewClosestFirstIterator(g, "A") },
		},
	} {
		g := build()
		got := visit(g, tc.seq(g))
		if len(got) != 3 || got[2] != "E" {
			t.Errorf("%s yielded %v, want two vertices and then E", tc.name, got)
		}
		g = build()
		if want := visit(g, iteratorSeq(tc.it(g))); !slices.Equal(got, want) {
			t.Errorf("%s yielded %v, and the iterator %v", tc.name, got, want)
		}
	}
}
