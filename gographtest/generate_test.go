package gographtest

import (
	"math"
	"math/rand/v2"
	"reflect"
	"slices"
	"testing"

	"github.com/hmdsefi/gograph"
)

// edgeList returns the edges of g as label pairs, each undirected edge once
// with the lower label first, sorted.
func edgeList(g gograph.Graph[int]) [][2]int {
	var out [][2]int
	for e := range gograph.Edges(g) {
		from, to := e.Source().Label(), e.Destination().Label()
		if !g.IsDirected() {
			if from > to {
				continue
			}
		}
		out = append(out, [2]int{from, to})
	}
	slices.SortFunc(out, func(a, b [2]int) int {
		if a[0] != b[0] {
			return a[0] - b[0]
		}
		return a[1] - b[1]
	})
	return out
}

// checkLabels fails t unless the vertices of g are 0 to n-1, in order.
func checkLabels(t *testing.T, g gograph.Graph[int], n int) {
	t.Helper()
	var got []int
	for v := range gograph.Vertices(g) {
		got = append(got, v.Label())
	}
	for i, label := range got {
		if label != i {
			t.Fatalf("vertices %v, want 0 to %d in order", got, n-1)
		}
	}
	if len(got) != n {
		t.Fatalf("got %d vertices, want %d", len(got), n)
	}
}

func TestPath(t *testing.T) {
	checkLabels(t, Path(0), 0)
	checkLabels(t, Path(-3), 0)

	g := Path(4, gograph.Directed())
	checkLabels(t, g, 4)
	if got, want := edgeList(g), [][2]int{{0, 1}, {1, 2}, {2, 3}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("edges %v, want %v", got, want)
	}

	g = Path(4)
	if g.IsDirected() || len(edgeList(g)) != 3 {
		t.Fatalf("undirected path: directed %v, edges %v", g.IsDirected(), edgeList(g))
	}
	if g := Path(1); len(edgeList(g)) != 0 {
		t.Fatalf("one vertex: edges %v", edgeList(g))
	}
}

func TestCycle(t *testing.T) {
	tests := []struct {
		name string
		g    gograph.Graph[int]
		n    int
		want [][2]int
	}{
		{name: "empty", g: Cycle(0, gograph.Directed())},
		{name: "self-loop", g: Cycle(1, gograph.Directed()), n: 1, want: [][2]int{{0, 0}}},
		{name: "two vertices, directed", g: Cycle(2, gograph.Directed()), n: 2, want: [][2]int{{0, 1}, {1, 0}}},
		{name: "two vertices, undirected", g: Cycle(2), n: 2, want: [][2]int{{0, 1}}},
		{name: "directed", g: Cycle(4, gograph.Directed()), n: 4, want: [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}}},
		{name: "undirected", g: Cycle(4), n: 4, want: [][2]int{{0, 1}, {0, 3}, {1, 2}, {2, 3}}},
		{name: "acyclic", g: Cycle(3, gograph.Acyclic()), n: 3, want: [][2]int{{0, 1}, {1, 2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkLabels(t, tt.g, tt.n)
			if got := edgeList(tt.g); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("edges %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComplete(t *testing.T) {
	tests := []struct {
		name string
		g    gograph.Graph[int]
		n    int
		want [][2]int
	}{
		{name: "empty", g: Complete(0)},
		{name: "one vertex", g: Complete(1, gograph.Directed()), n: 1},
		{name: "directed", g: Complete(3, gograph.Directed()), n: 3, want: [][2]int{{0, 1}, {0, 2}, {1, 0}, {1, 2}, {2, 0}, {2, 1}}},
		{name: "undirected", g: Complete(3), n: 3, want: [][2]int{{0, 1}, {0, 2}, {1, 2}}},
		{name: "acyclic", g: Complete(3, gograph.Acyclic()), n: 3, want: [][2]int{{0, 1}, {0, 2}, {1, 2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkLabels(t, tt.g, tt.n)
			if got := edgeList(tt.g); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("edges %v, want %v", got, tt.want)
			}
		})
	}

	if got := len(edgeList(Complete(20, gograph.Directed()))); got != 20*19 {
		t.Fatalf("directed: %d edges, want %d", got, 20*19)
	}
	if got := len(edgeList(Complete(20))); got != 20*19/2 {
		t.Fatalf("undirected: %d edges, want %d", got, 20*19/2)
	}
}

func TestGrid(t *testing.T) {
	g := Grid(2, 3, gograph.Directed())
	checkLabels(t, g, 6)
	// 0 1 2
	// 3 4 5
	want := [][2]int{{0, 1}, {0, 3}, {1, 2}, {1, 4}, {2, 5}, {3, 4}, {4, 5}}
	if got := edgeList(g); !reflect.DeepEqual(got, want) {
		t.Fatalf("edges %v, want %v", got, want)
	}

	if got := edgeList(Grid(2, 3)); !reflect.DeepEqual(got, want) {
		t.Fatalf("undirected edges %v, want %v", got, want)
	}

	for _, size := range [][2]int{{0, 3}, {3, 0}, {-1, 2}, {-1, -2}} {
		if g := Grid(size[0], size[1]); g.Order() != 0 {
			t.Fatalf("Grid(%d, %d) has %d vertices", size[0], size[1], g.Order())
		}
	}

	rows, cols := 7, 4
	if got, want := len(edgeList(Grid(rows, cols))), rows*(cols-1)+(rows-1)*cols; got != want {
		t.Fatalf("%dx%d grid: %d edges, want %d", rows, cols, got, want)
	}
}

func TestGenerators_PassOptions(t *testing.T) {
	graphs := map[string]gograph.Graph[int]{
		"Path":     Path(3, gograph.Directed(), gograph.Weighted()),
		"Cycle":    Cycle(3, gograph.Directed(), gograph.Weighted()),
		"Complete": Complete(3, gograph.Directed(), gograph.Weighted()),
		"Grid":     Grid(2, 2, gograph.Directed(), gograph.Weighted()),
		"Random":   Random(3, 1, nil, gograph.Directed(), gograph.Weighted()),
	}
	for name, g := range graphs {
		if !g.IsDirected() || !g.IsWeighted() {
			t.Errorf("%s: directed %v, weighted %v", name, g.IsDirected(), g.IsWeighted())
		}
	}
}

func TestRandom_Bounds(t *testing.T) {
	for _, p := range []float64{0, -1, math.NaN(), math.Inf(-1)} {
		for _, directed := range []bool{true, false} {
			if g := Random(10, p, nil, directedOption(directed)...); len(edgeList(g)) != 0 || g.Order() != 10 {
				t.Fatalf("p %v, directed %v: %d vertices and edges %v", p, directed, g.Order(), edgeList(g))
			}
		}
	}

	for _, p := range []float64{1, 2, math.Inf(1)} {
		if got, want := edgeList(Random(6, p, nil, gograph.Directed())), edgeList(Complete(6, gograph.Directed())); !reflect.DeepEqual(got, want) {
			t.Fatalf("p %v, directed: edges %v, want %v", p, got, want)
		}
		if got, want := edgeList(Random(6, p, nil)), edgeList(Complete(6)); !reflect.DeepEqual(got, want) {
			t.Fatalf("p %v, undirected: edges %v, want %v", p, got, want)
		}
		if got, want := edgeList(RandomDAG(6, p, nil)), edgeList(Complete(6, gograph.Acyclic())); !reflect.DeepEqual(got, want) {
			t.Fatalf("p %v, DAG: edges %v, want %v", p, got, want)
		}
		if got, want := edgeList(Random(6, p, nil, gograph.Acyclic())), edgeList(Complete(6, gograph.Acyclic())); !reflect.DeepEqual(got, want) {
			t.Fatalf("p %v, acyclic: edges %v, want %v", p, got, want)
		}
	}

	checkLabels(t, Random(0, 0.5, rand.New(rand.NewPCG(1, 1))), 0)     //nolint:gosec // seeded so failures are reproducible
	checkLabels(t, RandomDAG(1, 0.5, rand.New(rand.NewPCG(1, 1))), 1)  //nolint:gosec // seeded so failures are reproducible
	checkLabels(t, Random(-2, 0.5, rand.New(rand.NewPCG(1, 1))), 0)    //nolint:gosec // seeded so failures are reproducible
	checkLabels(t, RandomDAG(-2, 0.5, rand.New(rand.NewPCG(1, 1))), 0) //nolint:gosec // seeded so failures are reproducible
}

func directedOption(directed bool) []gograph.GraphOptionFunc {
	if directed {
		return []gograph.GraphOptionFunc{gograph.Directed()}
	}
	return nil
}

func TestRandom_SameSeedSameGraph(t *testing.T) {
	build := map[string]func(rng *rand.Rand) gograph.Graph[int]{
		"directed":   func(rng *rand.Rand) gograph.Graph[int] { return Random(40, 0.1, rng, gograph.Directed()) },
		"undirected": func(rng *rand.Rand) gograph.Graph[int] { return Random(40, 0.1, rng) },
		"DAG":        func(rng *rand.Rand) gograph.Graph[int] { return RandomDAG(40, 0.1, rng) },
	}
	for name, f := range build {
		a := edgeList(f(rand.New(rand.NewPCG(7, 126)))) //nolint:gosec // seeded so failures are reproducible
		b := edgeList(f(rand.New(rand.NewPCG(7, 126)))) //nolint:gosec // seeded so failures are reproducible
		c := edgeList(f(rand.New(rand.NewPCG(8, 126)))) //nolint:gosec // seeded so failures are reproducible
		if !reflect.DeepEqual(a, b) {
			t.Errorf("%s: the same seed gave different graphs", name)
		}
		if reflect.DeepEqual(a, c) {
			t.Errorf("%s: two seeds gave the same graph", name)
		}
	}
}

func TestRandom_Distribution(t *testing.T) {
	// each pair is picked with probability p, so the number of edges is
	// binomial, and each pair is picked about as often as any other
	const n, p, runs = 30, 0.2, 400
	rng := rand.New(rand.NewPCG(3, 126)) //nolint:gosec // seeded so failures are reproducible

	tests := []struct {
		name  string
		pairs int
		build func() gograph.Graph[int]
		valid func(from, to int) bool
	}{
		{
			name:  "directed",
			pairs: n * (n - 1),
			build: func() gograph.Graph[int] { return Random(n, p, rng, gograph.Directed()) },
			valid: func(from, to int) bool { return from != to },
		},
		{
			name:  "undirected",
			pairs: n * (n - 1) / 2,
			build: func() gograph.Graph[int] { return Random(n, p, rng) },
			valid: func(from, to int) bool { return from < to },
		},
		{
			name:  "DAG",
			pairs: n * (n - 1) / 2,
			build: func() gograph.Graph[int] { return RandomDAG(n, p, rng) },
			valid: func(from, to int) bool { return from < to },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hits := map[[2]int]int{}
			total := 0
			for range runs {
				g := tt.build()
				checkLabels(t, g, n)
				for _, e := range edgeList(g) {
					if !tt.valid(e[0], e[1]) {
						t.Fatalf("edge %v", e)
					}
					hits[e]++
					total++
				}
			}

			mean := float64(runs*tt.pairs) * p
			sd := math.Sqrt(mean * (1 - p))
			if math.Abs(float64(total)-mean) > 5*sd {
				t.Fatalf("%d edges in %d runs, want about %.0f", total, runs, mean)
			}
			if len(hits) != tt.pairs {
				t.Fatalf("%d of %d pairs got an edge", len(hits), tt.pairs)
			}
			// a pair picked in a run is binomial(runs, p)
			pairMean, pairSD := runs*p, math.Sqrt(runs*p*(1-p))
			for pair, count := range hits {
				if math.Abs(float64(count)-pairMean) > 6*pairSD {
					t.Fatalf("pair %v picked %d times in %d runs, want about %.0f", pair, count, runs, pairMean)
				}
			}
		})
	}
}

func TestRandomDAG_Acyclic(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 126)) //nolint:gosec // seeded so failures are reproducible
	for range 50 {
		g := RandomDAG(25, 0.3, rng)
		if !g.IsDirected() || g.IsAcyclic() {
			t.Fatalf("directed %v, acyclic option %v", g.IsDirected(), g.IsAcyclic())
		}
		if gograph.HasCycle(g) {
			t.Fatal("RandomDAG has a cycle")
		}
		order, err := gograph.TopologySort(g)
		if err != nil {
			t.Fatal(err)
		}
		r := &recorder{}
		IsTopologicalOrder(r, g, order)
		r.check(t, "")
	}
}

func BenchmarkRandom(b *testing.B) {
	// 100,000 vertices have about 5*10^9 ordered pairs, and p picks about
	// 50,000 of them
	rng := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // seeded so runs are comparable
	b.ReportAllocs()
	for b.Loop() {
		Random(100_000, 1e-5, rng, gograph.Directed())
	}
}
