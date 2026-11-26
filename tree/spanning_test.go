package tree

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"strconv"
	"testing"

	"github.com/hmdsefi/gograph"
)

type spanningFunc func(gograph.Graph[string]) ([]*gograph.Edge[string], float64, error)

var spanningFuncs = []struct {
	name string
	fn   spanningFunc
}{
	{"Kruskal", Kruskal[string]},
	{"Prim", Prim[string]},
}

// weightedGraph returns an undirected weighted graph with the given edges,
// written as "A-B:4". The vertices are added in the order they first appear.
func weightedGraph(t testing.TB, edges ...string) gograph.Graph[string] {
	t.Helper()
	g := gograph.New[string](gograph.Weighted())
	for _, s := range edges {
		var a, b string
		var w float64
		if _, err := fmt.Sscanf(s, "%1s-%1s:%g", &a, &b, &w); err != nil {
			t.Fatalf("bad edge %q: %v", s, err)
		}
		if g.GetVertexByID(a) == nil {
			g.AddVertexByLabel(a)
		}
		if g.GetVertexByID(b) == nil {
			g.AddVertexByLabel(b)
		}
		va, vb := g.GetVertexByID(a), g.GetVertexByID(b)
		if _, err := g.AddEdge(va, vb, gograph.WithEdgeWeight(w)); err != nil {
			t.Fatal(err)
		}
	}
	return g
}

// edgeNames writes each edge as "A-B:4", from source to destination.
func edgeNames[T comparable](edges []*gograph.Edge[T]) []string {
	names := make([]string, len(edges))
	for i, e := range edges {
		names[i] = fmt.Sprintf("%v-%v:%g", e.Source().Label(), e.Destination().Label(), e.Weight())
	}
	return names
}

func TestSpanning_ConnectedGraph(t *testing.T) {
	g := weightedGraph(t,
		"A-B:7", "A-D:5", "B-C:8", "B-D:9", "B-E:7", "C-E:5",
		"D-E:15", "D-F:6", "E-F:8", "E-G:9", "F-G:11",
	)
	want := []string{"A-D:5", "C-E:5", "D-F:6", "A-B:7", "B-E:7", "E-G:9"}

	for _, f := range spanningFuncs {
		t.Run(f.name, func(t *testing.T) {
			edges, total, err := f.fn(g)
			if err != nil {
				t.Fatal(err)
			}
			if got := edgeNames(edges); !reflect.DeepEqual(got, want) {
				t.Errorf("edges = %v, want %v", got, want)
			}
			if total != 39 {
				t.Errorf("total = %v, want 39", total)
			}
		})
	}
}

func TestSpanning_DisconnectedGraphGivesAForest(t *testing.T) {
	g := weightedGraph(t, "A-B:3", "B-C:1", "A-C:2", "D-E:4", "E-F:6", "D-F:5")
	g.AddVertexByLabel("G")
	want := []string{"B-C:1", "A-C:2", "D-E:4", "D-F:5"}

	for _, f := range spanningFuncs {
		t.Run(f.name, func(t *testing.T) {
			edges, total, err := f.fn(g)
			if err != nil {
				t.Fatal(err)
			}
			if got := edgeNames(edges); !reflect.DeepEqual(got, want) {
				t.Errorf("edges = %v, want %v", got, want)
			}
			if total != 12 {
				t.Errorf("total = %v, want 12", total)
			}
		})
	}
}

func TestSpanning_EdgesAreTheGraphsOwnFromTheEarlierEndpoint(t *testing.T) {
	// The edges are added from the later vertex, so the stored edge that
	// AddEdge returns starts at the later one.
	g := gograph.New[string](gograph.Weighted())
	a, b, c := g.AddVertexByLabel("A"), g.AddVertexByLabel("B"), g.AddVertexByLabel("C")
	_, _ = g.AddEdge(b, a, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(c, b, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(c, a, gograph.WithEdgeWeight(3))

	for _, f := range spanningFuncs {
		t.Run(f.name, func(t *testing.T) {
			edges, _, err := f.fn(g)
			if err != nil {
				t.Fatal(err)
			}
			want := []*gograph.Edge[string]{g.GetEdge(a, b), g.GetEdge(b, c)}
			if len(edges) != len(want) {
				t.Fatalf("edges = %v, want A-B and B-C", edgeNames(edges))
			}
			for i := range want {
				if edges[i] != want[i] {
					t.Errorf("edge %d = %v, want the graph's edge %v", i, edgeNames(edges[i:i+1]), edgeNames(want[i:i+1]))
				}
			}
		})
	}
}

func TestSpanning_TiesBreakByEndpointOrder(t *testing.T) {
	// A square with equal weights. The vertices are added in the order D, A,
	// C, B, so the edges compare as D-A, D-C, A-B and C-B, and C-B is left
	// out. Each edge starts at its endpoint added first.
	g := weightedGraph(t, "D-A:1", "C-D:1", "B-C:1", "A-B:1")
	want := []string{"D-A:1", "D-C:1", "A-B:1"}

	for _, f := range spanningFuncs {
		t.Run(f.name, func(t *testing.T) {
			edges, _, err := f.fn(g)
			if err != nil {
				t.Fatal(err)
			}
			if got := edgeNames(edges); !reflect.DeepEqual(got, want) {
				t.Errorf("edges = %v, want %v", got, want)
			}
		})
	}
}

func TestSpanning_SelfLoopsAndNaNAreLeftOut(t *testing.T) {
	g := weightedGraph(t, "A-A:-5", "A-B:2", "B-C:-1")
	b, c := g.GetVertexByID("B"), g.GetVertexByID("C")
	d := g.AddVertexByLabel("D")
	_, _ = g.AddEdge(c, d, gograph.WithEdgeWeight(math.NaN()))
	_, _ = g.AddEdge(b, d, gograph.WithEdgeWeight(math.Inf(1)))
	want := []string{"B-C:-1", "A-B:2", "B-D:+Inf"}

	for _, f := range spanningFuncs {
		t.Run(f.name, func(t *testing.T) {
			edges, total, err := f.fn(g)
			if err != nil {
				t.Fatal(err)
			}
			if got := edgeNames(edges); !reflect.DeepEqual(got, want) {
				t.Errorf("edges = %v, want %v", got, want)
			}
			if !math.IsInf(total, 1) {
				t.Errorf("total = %v, want +Inf", total)
			}
		})
	}
}

func TestSpanning_UnweightedGraph(t *testing.T) {
	g := gograph.New[string]()
	a, b, c := g.AddVertexByLabel("A"), g.AddVertexByLabel("B"), g.AddVertexByLabel("C")
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(a, c)
	_, _ = g.AddEdge(a, b)
	want := []string{"A-B:0", "A-C:0"}

	for _, f := range spanningFuncs {
		t.Run(f.name, func(t *testing.T) {
			edges, total, err := f.fn(g)
			if err != nil {
				t.Fatal(err)
			}
			if got := edgeNames(edges); !reflect.DeepEqual(got, want) {
				t.Errorf("edges = %v, want %v", got, want)
			}
			if total != 0 {
				t.Errorf("total = %v, want 0", total)
			}
		})
	}
}

func TestSpanning_EmptyAndEdgelessGraphs(t *testing.T) {
	empty := gograph.New[string](gograph.Weighted())
	single := gograph.New[string](gograph.Weighted())
	single.AddVertexByLabel("A")
	isolated := gograph.New[string](gograph.Weighted())
	isolated.AddVertexByLabel("A")
	isolated.AddVertexByLabel("B")

	for _, f := range spanningFuncs {
		for name, g := range map[string]gograph.Graph[string]{"empty": empty, "single": single, "isolated": isolated} {
			t.Run(f.name+"/"+name, func(t *testing.T) {
				edges, total, err := f.fn(g)
				if err != nil {
					t.Fatal(err)
				}
				if edges == nil || len(edges) != 0 || total != 0 {
					t.Errorf("got %v, %v; want an empty slice and 0", edges, total)
				}
			})
		}
	}
}

func TestSpanning_DirectedGraphIsAnError(t *testing.T) {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"), gograph.WithEdgeWeight(1))

	for _, f := range spanningFuncs {
		t.Run(f.name, func(t *testing.T) {
			edges, total, err := f.fn(g)
			if !errors.Is(err, gograph.ErrNotUndirected) {
				t.Errorf("err = %v, want ErrNotUndirected", err)
			}
			if edges != nil || total != 0 {
				t.Errorf("got %v, %v; want nil and 0", edges, total)
			}
		})
	}
}

func TestSpanning_SameEdgesOnEveryRun(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 9)) //nolint:gosec // seeded so failures are reproducible
	g := randomWeighted(rng, 60, 300, 3)
	first, _, _ := Kruskal(g)
	for _, f := range []func(gograph.Graph[int]) ([]*gograph.Edge[int], float64, error){Kruskal[int], Prim[int]} {
		for range 5 {
			edges, _, _ := f(g)
			if !reflect.DeepEqual(edges, first) {
				t.Fatalf("edges differ between runs")
			}
		}
	}
}

// randomWeighted returns an undirected graph with n vertices, about m random
// edges and self-loops, and integer weights from 0 to maxWeight-1, so many
// edges tie.
func randomWeighted(rng *rand.Rand, n, m, maxWeight int) gograph.Graph[int] {
	g := gograph.New[int](gograph.Weighted())
	vertices := make([]*gograph.Vertex[int], n)
	// Add the labels out of order so vertex order and label order differ.
	for i, label := range rng.Perm(n) {
		vertices[i] = g.AddVertexByLabel(label)
	}
	for range m {
		a, b := vertices[rng.IntN(n)], vertices[rng.IntN(n)]
		_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(float64(rng.IntN(maxWeight))))
	}
	return g
}

// bruteForceWeight returns the smallest total weight of a spanning forest of
// g, trying every subset of its edges.
func bruteForceWeight(g gograph.Graph[int]) float64 {
	var edges []*gograph.Edge[int]
	for e := range gograph.Edges(g) {
		if e.Source().Label() < e.Destination().Label() {
			edges = append(edges, e)
		}
	}
	want := len(g.GetAllVertices()) - components(g, edges)

	best := math.Inf(1)
	for mask := range 1 << len(edges) {
		var chosen []*gograph.Edge[int]
		total := 0.0
		for i, e := range edges {
			if mask&(1<<i) != 0 {
				chosen = append(chosen, e)
				total += e.Weight()
			}
		}
		if len(chosen) == want && components(g, chosen) == len(g.GetAllVertices())-want {
			best = min(best, total)
		}
	}
	return best
}

// components returns the number of connected components of the vertices of
// g with only the given edges.
func components(g gograph.Graph[int], edges []*gograph.Edge[int]) int {
	parent := map[int]int{}
	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	count := 0
	for _, v := range g.GetAllVertices() {
		parent[v.Label()] = v.Label()
		count++
	}
	for _, e := range edges {
		a, b := find(e.Source().Label()), find(e.Destination().Label())
		if a != b {
			parent[a] = b
			count--
		}
	}
	return count
}

func TestSpanning_MatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewPCG(11, 2)) //nolint:gosec // seeded so failures are reproducible
	for i := range 400 {
		n := 1 + rng.IntN(7)
		g := randomWeighted(rng, n, rng.IntN(13), 4)
		name := strconv.Itoa(i)

		kEdges, kTotal, err := Kruskal(g)
		if err != nil {
			t.Fatal(err)
		}
		pEdges, pTotal, err := Prim(g)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(kEdges, pEdges) || kTotal != pTotal {
			t.Fatalf("graph %s: Kruskal %v (%v) and Prim %v (%v) differ",
				name, edgeNames(kEdges), kTotal, edgeNames(pEdges), pTotal)
		}

		sum := 0.0
		for _, e := range kEdges {
			sum += e.Weight()
			if g.GetEdge(e.Source(), e.Destination()) != e {
				t.Fatalf("graph %s: %v is not the graph's own edge", name, edgeNames([]*gograph.Edge[int]{e}))
			}
		}
		if sum != kTotal {
			t.Fatalf("graph %s: total %v, but the edges add up to %v", name, kTotal, sum)
		}
		var all []*gograph.Edge[int]
		for e := range gograph.Edges(g) {
			all = append(all, e)
		}
		if c := components(g, kEdges); c != components(g, all) || len(kEdges) != n-c {
			t.Fatalf("graph %s: %v is not a spanning forest", name, edgeNames(kEdges))
		}
		if want := bruteForceWeight(g); kTotal != want {
			t.Fatalf("graph %s: total %v, want %v", name, kTotal, want)
		}
	}
}

func TestSpanning_PrimMatchesKruskalOnLargerGraphs(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 8)) //nolint:gosec // seeded so failures are reproducible
	for range 50 {
		n := 1 + rng.IntN(200)
		g := randomWeighted(rng, n, rng.IntN(n*n/2+1), 1+rng.IntN(10))
		kEdges, kTotal, _ := Kruskal(g)
		pEdges, pTotal, _ := Prim(g)
		if !reflect.DeepEqual(kEdges, pEdges) || kTotal != pTotal {
			t.Fatalf("Kruskal and Prim differ on a graph of %d vertices", n)
		}
	}
}

// randomSparse returns an undirected graph with n vertices and m distinct
// edges with random weights.
func randomSparse(n, m int) gograph.Graph[int] {
	rng := rand.New(rand.NewPCG(1, 1)) //nolint:gosec // seeded so runs are comparable
	g := gograph.New[int](gograph.Weighted())
	vertices := make([]*gograph.Vertex[int], n)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for added := 0; added < m; {
		a, b := vertices[rng.IntN(n)], vertices[rng.IntN(n)]
		if a == b || g.GetEdge(a, b) != nil {
			continue
		}
		_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(rng.Float64()))
		added++
	}
	return g
}

// complete returns the complete undirected graph on n vertices with random
// weights.
func complete(n int) gograph.Graph[int] {
	rng := rand.New(rand.NewPCG(2, 2)) //nolint:gosec // seeded so runs are comparable
	g := gograph.New[int](gograph.Weighted())
	vertices := make([]*gograph.Vertex[int], n)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for i := range n {
		for j := i + 1; j < n; j++ {
			_, _ = g.AddEdge(vertices[i], vertices[j], gograph.WithEdgeWeight(rng.Float64()))
		}
	}
	return g
}

func BenchmarkSpanning(b *testing.B) {
	graphs := []struct {
		name string
		g    gograph.Graph[int]
	}{
		{"sparse", randomSparse(10_000, 50_000)},
		{"complete", complete(1_000)},
	}
	for _, graph := range graphs {
		for _, f := range []struct {
			name string
			fn   func(gograph.Graph[int]) ([]*gograph.Edge[int], float64, error)
		}{
			{"Kruskal", Kruskal[int]},
			{"Prim", Prim[int]},
		} {
			b.Run(graph.name+"/"+f.name, func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, _, err := f.fn(graph.g); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// noEdges lists neighbors but has no edge for any of them.
type noEdges struct{ gograph.Graph[string] }

func (noEdges) GetEdge(_, _ *gograph.Vertex[string]) *gograph.Edge[string] { return nil }

// hiddenVertex leaves one vertex out of GetAllVertices.
type hiddenVertex struct {
	gograph.Graph[string]
	hidden string
}

func (h hiddenVertex) GetAllVertices() []*gograph.Vertex[string] {
	var out []*gograph.Vertex[string]
	for _, v := range h.Graph.GetAllVertices() {
		if v.Label() != h.hidden {
			out = append(out, v)
		}
	}
	return out
}

// oneWay answers GetEdge only from the later vertex to the earlier one, and
// lists fewer edges in AllEdges than GetEdge finds.
type oneWay struct {
	gograph.Graph[string]
	order map[string]int
}

func (o oneWay) GetEdge(from, to *gograph.Vertex[string]) *gograph.Edge[string] {
	if o.order[from.Label()] < o.order[to.Label()] {
		return nil
	}
	return o.Graph.GetEdge(from, to)
}

func (oneWay) AllEdges() []*gograph.Edge[string] { return nil }

func TestSpanning_OtherGraphTypesGiveTheSameResultFromBoth(t *testing.T) {
	base := weightedGraph(t, "A-B:5", "B-C:5", "C-D:1", "A-C:9", "B-D:2")
	order := map[string]int{"A": 0, "B": 1, "C": 2, "D": 3}
	graphs := []struct {
		name string
		g    gograph.Graph[string]
		want []string
	}{
		{"no edges", noEdges{base}, []string{}},
		{"hidden vertex", hiddenVertex{base, "D"}, []string{"A-B:5", "B-C:5"}},
		{"one-way GetEdge", oneWay{base, order}, []string{}},
	}
	for _, tc := range graphs {
		t.Run(tc.name, func(t *testing.T) {
			kEdges, kTotal, err := Kruskal(tc.g)
			if err != nil {
				t.Fatal(err)
			}
			pEdges, pTotal, err := Prim(tc.g)
			if err != nil {
				t.Fatal(err)
			}
			if got := edgeNames(kEdges); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Kruskal edges = %v, want %v", got, tc.want)
			}
			if !reflect.DeepEqual(kEdges, pEdges) || kTotal != pTotal {
				t.Errorf("Kruskal %v (%v) and Prim %v (%v) differ", edgeNames(kEdges), kTotal, edgeNames(pEdges), pTotal)
			}
		})
	}
}

func TestSpanning_InfiniteWeights(t *testing.T) {
	g := weightedGraph(t, "A-B:1", "B-C:2")
	a, c := g.GetVertexByID("A"), g.GetVertexByID("C")
	d, e := g.AddVertexByLabel("D"), g.AddVertexByLabel("E")
	_, _ = g.AddEdge(a, d, gograph.WithEdgeWeight(math.Inf(-1)))
	_, _ = g.AddEdge(c, d, gograph.WithEdgeWeight(math.Inf(-1)))

	for _, f := range spanningFuncs {
		t.Run(f.name, func(t *testing.T) {
			edges, total, err := f.fn(g)
			if err != nil {
				t.Fatal(err)
			}
			if want := []string{"A-D:-Inf", "C-D:-Inf", "A-B:1"}; !reflect.DeepEqual(edgeNames(edges), want) {
				t.Errorf("edges = %v, want %v", edgeNames(edges), want)
			}
			if !math.IsInf(total, -1) {
				t.Errorf("total = %v, want -Inf", total)
			}
		})
	}

	_, _ = g.AddEdge(a, e, gograph.WithEdgeWeight(math.Inf(1)))
	for _, f := range spanningFuncs {
		t.Run(f.name+" with +Inf", func(t *testing.T) {
			_, total, _ := f.fn(g)
			if !math.IsNaN(total) {
				t.Errorf("total = %v, want NaN for a forest with -Inf and +Inf edges", total)
			}
		})
	}
}

// noEdgeFrom returns no edge from one vertex to another, and answers every
// other GetEdge.
type noEdgeFrom struct {
	gograph.Graph[string]
	from, to string
}

func (n noEdgeFrom) GetEdge(from, to *gograph.Vertex[string]) *gograph.Edge[string] {
	if from.Label() == n.from && to.Label() == n.to {
		return nil
	}
	return n.Graph.GetEdge(from, to)
}

func TestPrim_KeepsTheEdgeItReadWhenTheOtherDirectionIsMissing(t *testing.T) {
	g := noEdgeFrom{weightedGraph(t, "A-C:1", "B-C:2"), "B", "C"}

	edges, total, err := Prim[string](g)
	if err != nil {
		t.Fatal(err)
	}
	// Vertices are added in the order A, C, B, so C is the earlier endpoint
	// of C-B and GetEdge(C, B) answers.
	if want := []string{"A-C:1", "C-B:2"}; !reflect.DeepEqual(edgeNames(edges), want) || total != 3 {
		t.Errorf("Prim = %v, %v; want %v, 3", edgeNames(edges), total, want)
	}

	// With B added before C, the earlier endpoint is B, and GetEdge(B, C)
	// returns nil, so Prim keeps C-B, the edge it read from C.
	h := gograph.New[string](gograph.Weighted())
	a, b, c := h.AddVertexByLabel("A"), h.AddVertexByLabel("B"), h.AddVertexByLabel("C")
	_, _ = h.AddEdge(a, c, gograph.WithEdgeWeight(1))
	_, _ = h.AddEdge(b, c, gograph.WithEdgeWeight(2))
	edges, total, err = Prim[string](noEdgeFrom{h, "B", "C"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"A-C:1", "C-B:2"}; !reflect.DeepEqual(edgeNames(edges), want) || total != 3 {
		t.Errorf("Prim = %v, %v; want %v, 3", edgeNames(edges), total, want)
	}
}

func TestSpanning_OverflowingTotal(t *testing.T) {
	g := weightedGraph(t, "A-B:1", "B-C:1")
	a, b, c := g.GetVertexByID("A"), g.GetVertexByID("B"), g.GetVertexByID("C")
	g.RemoveEdges(g.GetEdge(a, b), g.GetEdge(b, c))
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(-math.MaxFloat64))
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(-math.MaxFloat64))
	d := g.AddVertexByLabel("D")
	_, _ = g.AddEdge(c, d, gograph.WithEdgeWeight(math.Inf(1)))

	for _, f := range spanningFuncs {
		t.Run(f.name, func(t *testing.T) {
			edges, total, err := f.fn(g)
			if err != nil {
				t.Fatal(err)
			}
			if len(edges) != 3 || !math.IsNaN(total) {
				t.Errorf("got %d edges and total %v; want 3 edges and NaN, since the finite weights overflow to -Inf", len(edges), total)
			}
		})
	}
}
