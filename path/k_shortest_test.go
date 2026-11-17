package path

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/hmdsefi/gograph"
)

// assertKShortest checks that paths are distinct simple paths from start to
// goal over edges of g, cheapest first, and that their costs are want.
func assertKShortest(
	t *testing.T,
	g gograph.Graph[string],
	paths [][]*gograph.Vertex[string],
	start, goal string,
	want []float64,
) {
	t.Helper()
	if len(paths) != len(want) {
		t.Fatalf("got %d paths %v, want %d with costs %v", len(paths), allLabels(paths), len(want), want)
	}
	seen := map[string]bool{}
	for i, cost := range want {
		p := paths[i]
		if len(p) == 0 || p[0].Label() != start || p[len(p)-1].Label() != goal {
			t.Fatalf("path %d = %v, want it to run from %s to %s", i, pathLabels(p), start, goal)
		}
		visited := map[string]bool{}
		var sum float64
		for j, v := range p {
			if g.GetVertexByID(v.Label()) != v {
				t.Errorf("vertex %s is not the graph's own pointer", v.Label())
			}
			if visited[v.Label()] {
				t.Fatalf("path %d = %v repeats %s", i, pathLabels(p), v.Label())
			}
			visited[v.Label()] = true
			if j > 0 {
				e := g.GetEdge(p[j-1], v)
				if e == nil {
					t.Fatalf("no edge from %s to %s in %v", p[j-1].Label(), v.Label(), pathLabels(p))
				}
				sum += e.Weight()
			}
		}
		if sum != cost {
			t.Errorf("path %d = %v costs %v, want %v", i, pathLabels(p), sum, cost)
		}
		key := fmt.Sprint(pathLabels(p))
		if seen[key] {
			t.Errorf("path %v is returned twice", pathLabels(p))
		}
		seen[key] = true
	}
}

func allLabels(paths [][]*gograph.Vertex[string]) [][]string {
	out := make([][]string, len(paths))
	for i, p := range paths {
		out[i] = pathLabels(p)
	}
	return out
}

// simplePathCosts returns the cost of every simple path from start to goal,
// cheapest first, by trying every path.
func simplePathCosts(g gograph.Graph[string], start, goal string) []float64 {
	var costs []float64
	visited := map[string]bool{}
	var walk func(v *gograph.Vertex[string], sum float64)
	walk = func(v *gograph.Vertex[string], sum float64) {
		if v.Label() == goal {
			costs = append(costs, sum)
			return
		}
		visited[v.Label()] = true
		for next := range v.AllNeighbors() {
			if !visited[next.Label()] {
				walk(next, sum+g.GetEdge(v, next).Weight())
			}
		}
		visited[v.Label()] = false
	}
	walk(g.GetVertexByID(start), 0)
	slices.Sort(costs)
	return costs
}

func TestKShortest_Diamond(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "G", 1},
		weightedEdge{"S", "B", 1}, weightedEdge{"B", "G", 3},
	)
	paths, err := KShortest(g, "S", "G", 2, nil)
	if err != nil {
		t.Fatalf("KShortest() error = %v", err)
	}
	want := [][]string{{"S", "A", "G"}, {"S", "B", "G"}}
	if !reflect.DeepEqual(allLabels(paths), want) {
		t.Errorf("paths = %v, want %v", allLabels(paths), want)
	}
	assertKShortest(t, g, paths, "S", "G", []float64{2, 4})
}

func TestKShortest_KAboveThePathCountReturnsAll(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "G", 1},
		weightedEdge{"S", "B", 2}, weightedEdge{"B", "G", 2},
		weightedEdge{"A", "B", 1}, weightedEdge{"S", "G", 7},
	)
	paths, err := KShortest(g, "S", "G", 100, nil)
	if err != nil {
		t.Fatalf("KShortest() error = %v", err)
	}
	assertKShortest(t, g, paths, "S", "G", []float64{2, 4, 4, 7})
}

func TestKShortest_FirstPathMatchesDijkstra(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "G", 10},
		weightedEdge{"S", "B", 4}, weightedEdge{"B", "C", 1}, weightedEdge{"C", "G", 1},
	)
	paths, err := KShortest(g, "S", "G", 1, nil)
	if err != nil {
		t.Fatalf("KShortest() error = %v", err)
	}
	if want := []string{"S", "B", "C", "G"}; len(paths) != 1 || !reflect.DeepEqual(pathLabels(paths[0]), want) {
		t.Fatalf("paths = %v, want [%v]", allLabels(paths), want)
	}
	assertKShortest(t, g, paths, "S", "G", []float64{Dijkstra(g, "S")["G"]})
}

func TestKShortest_Errors(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 1})
	negative := func(*gograph.Edge[string]) float64 { return -1 }

	tests := []struct {
		name        string
		start, goal string
		k           int
		cost        func(*gograph.Edge[string]) float64
		want        error
	}{
		{"k is 0", "A", "B", 0, nil, ErrInvalidK},
		{"k is negative", "A", "B", -3, nil, ErrInvalidK},
		{"missing start", "X", "B", 1, nil, gograph.ErrVertexDoesNotExist},
		{"missing goal", "A", "X", 1, nil, gograph.ErrVertexDoesNotExist},
		{"same missing vertex", "X", "X", 1, nil, gograph.ErrVertexDoesNotExist},
		{"negative cost", "A", "B", 1, negative, ErrNegativeWeight},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := KShortest(g, tt.start, tt.goal, tt.k, tt.cost)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if paths != nil {
				t.Errorf("paths = %v, want nil", allLabels(paths))
			}
		})
	}
}

func TestKShortest_NoPathIsAnEmptySlice(t *testing.T) {
	tests := []struct {
		name  string
		edges []weightedEdge
	}{
		{"goal not reached", []weightedEdge{{"S", "A", 1}, {"G", "S", 1}}},
		{"NaN weight is skipped", []weightedEdge{{"S", "G", math.NaN()}}},
		{"+Inf weight is no path", []weightedEdge{{"S", "G", math.Inf(1)}}},
		{
			"a sum past the float64 range is no path",
			[]weightedEdge{{"S", "A", math.MaxFloat64}, {"A", "G", math.MaxFloat64}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := directedWeighted(tt.edges...)
			for _, k := range []int{1, 3} {
				paths, err := KShortest(g, "S", "G", k, nil)
				if err != nil || paths == nil || len(paths) != 0 {
					t.Errorf("k = %d: paths = %v, error = %v, want an empty slice and nil", k, allLabels(paths), err)
				}
			}
		})
	}
}

func TestKShortest_OverflowOnlyDropsThatPath(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", math.MaxFloat64}, weightedEdge{"A", "G", math.MaxFloat64},
		weightedEdge{"S", "B", math.MaxFloat64}, weightedEdge{"B", "G", 1},
	)
	paths, err := KShortest(g, "S", "G", 5, nil)
	if err != nil {
		t.Fatalf("KShortest() error = %v", err)
	}
	assertKShortest(t, g, paths, "S", "G", []float64{math.MaxFloat64})
}

func TestKShortest_StartEqualsGoal(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 1}, weightedEdge{"B", "A", 1}, weightedEdge{"A", "A", 1})
	for _, k := range []int{1, 5} {
		paths, err := KShortest(g, "A", "A", k, func(*gograph.Edge[string]) float64 {
			t.Error("cost should not be called")
			return 0
		})
		if err != nil {
			t.Fatalf("KShortest() error = %v", err)
		}
		if len(paths) != 1 || len(paths[0]) != 1 || paths[0][0] != g.GetVertexByID("A") {
			t.Errorf("k = %d: paths = %v, want one path of the graph's own A", k, allLabels(paths))
		}
	}
}

func TestKShortest_PathsAreSimple(t *testing.T) {
	// S, A, B, A, G would cost 3, below the only simple path's 10.
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "B", 0}, weightedEdge{"B", "A", 0},
		weightedEdge{"A", "A", 0}, weightedEdge{"A", "G", 10}, weightedEdge{"B", "C", 1},
	)
	paths, err := KShortest(g, "S", "G", 5, nil)
	if err != nil {
		t.Fatalf("KShortest() error = %v", err)
	}
	assertKShortest(t, g, paths, "S", "G", []float64{11})
}

func TestKShortest_Undirected(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a, b, c, d := g.AddVertexByLabel("A"), g.AddVertexByLabel("B"), g.AddVertexByLabel("C"), g.AddVertexByLabel("D")
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(b, d, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(a, c, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(c, d, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(1))

	for _, tt := range []struct{ start, goal string }{{"A", "D"}, {"D", "A"}} {
		paths, err := KShortest(g, tt.start, tt.goal, 10, nil)
		if err != nil {
			t.Fatalf("KShortest() error = %v", err)
		}
		assertKShortest(t, g, paths, tt.start, tt.goal, []float64{2, 4, 4, 4})
	}
}

func TestKShortest_CostFunctionReplacesTheWeight(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "G", 1},
		weightedEdge{"S", "G", 5},
	)
	calls := 0
	paths, err := KShortest(g, "S", "G", 2, func(e *gograph.Edge[string]) float64 {
		calls++
		if e.Source().Label() == "S" && e.Destination().Label() == "G" {
			return 0.5
		}
		return e.Weight()
	})
	if err != nil {
		t.Fatalf("KShortest() error = %v", err)
	}
	want := [][]string{{"S", "G"}, {"S", "A", "G"}}
	if !reflect.DeepEqual(allLabels(paths), want) {
		t.Errorf("paths = %v, want %v", allLabels(paths), want)
	}
	if calls != 3 {
		t.Errorf("cost called %d times, want once per edge", calls)
	}
}

func TestKShortest_UnweightedGraphCostsZero(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	s, a, goal := g.AddVertexByLabel("S"), g.AddVertexByLabel("A"), g.AddVertexByLabel("G")
	_, _ = g.AddEdge(s, a)
	_, _ = g.AddEdge(a, goal)
	_, _ = g.AddEdge(s, goal)

	paths, err := KShortest(g, "S", "G", 3, nil)
	if err != nil {
		t.Fatalf("KShortest() error = %v", err)
	}
	// Both cost 0, so they come in the order the search finds them.
	want := [][]string{{"S", "G"}, {"S", "A", "G"}}
	if !reflect.DeepEqual(allLabels(paths), want) {
		t.Errorf("paths = %v, want %v", allLabels(paths), want)
	}
}

func TestKShortest_SamePathsOnEveryRun(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "B", 1}, weightedEdge{"S", "A", 1}, weightedEdge{"S", "C", 1},
		weightedEdge{"A", "G", 1}, weightedEdge{"B", "G", 1}, weightedEdge{"C", "G", 1},
		weightedEdge{"A", "B", 0}, weightedEdge{"B", "C", 0},
	)
	first, err := KShortest(g, "S", "G", 10, nil)
	if err != nil {
		t.Fatalf("KShortest() error = %v", err)
	}
	assertKShortest(t, g, first, "S", "G", simplePathCosts(g, "S", "G"))
	for range 20 {
		paths, err := KShortest(g, "S", "G", 10, nil)
		if err != nil {
			t.Fatalf("KShortest() error = %v", err)
		}
		if !reflect.DeepEqual(allLabels(paths), allLabels(first)) {
			t.Fatalf("paths = %v, then %v", allLabels(first), allLabels(paths))
		}
	}
}

// kNoEdges is a graph that lists neighbors it has no edge for.
type kNoEdges struct {
	gograph.Graph[string]
}

func (kNoEdges) GetEdge(_, _ *gograph.Vertex[string]) *gograph.Edge[string] { return nil }

func TestKShortest_NeighborWithoutAnEdgeIsSkipped(t *testing.T) {
	g := kNoEdges{directedWeighted(weightedEdge{"A", "B", 1})}
	paths, err := KShortest[string](g, "A", "B", 2, nil)
	if err != nil || len(paths) != 0 {
		t.Errorf("paths = %v, error = %v, want none and nil", allLabels(paths), err)
	}
}

func TestKShortest_StampsSurviveWrapping(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "G", 1},
		weightedEdge{"S", "B", 1}, weightedEdge{"B", "G", 2},
		weightedEdge{"A", "B", 1}, weightedEdge{"B", "A", 1},
	)
	s, err := newKShortestSearch(g, g.GetVertexByID("S"), g.GetVertexByID("G"), (*gograph.Edge[string]).Weight)
	if err != nil {
		t.Fatalf("newKShortestSearch() error = %v", err)
	}
	s.gen = math.MaxUint32 - 2
	assertKShortest(t, g, s.run(10), "S", "G", simplePathCosts(g, "S", "G"))
}

// TestKShortest_MatchesEverySimplePath compares random graphs with a search
// over every simple path. Small integer weights, zeros included, make many
// ties, which is where the bound on candidates matters.
func TestKShortest_MatchesEverySimplePath(t *testing.T) {
	rng := rand.New(rand.NewPCG(18, 9)) //nolint:gosec // seeded so failures are reproducible

	for trial := range 2000 {
		var g gograph.Graph[string]
		if trial%2 == 0 {
			g = gograph.New[string](gograph.Directed(), gograph.Weighted())
		} else {
			g = gograph.New[string](gograph.Weighted())
		}
		n := 2 + rng.IntN(7)
		labels := make([]string, n)
		for i := range labels {
			labels[i] = strconv.Itoa(i)
			g.AddVertexByLabel(labels[i])
		}
		for range rng.IntN(3 * n) {
			from, to := g.GetVertexByID(labels[rng.IntN(n)]), g.GetVertexByID(labels[rng.IntN(n)])
			_, _ = g.AddEdge(from, to, gograph.WithEdgeWeight(float64(rng.IntN(4))))
		}

		start, goal := labels[0], labels[n-1]
		all := simplePathCosts(g, start, goal)
		for _, k := range []int{1, 2, 3, 1 + rng.IntN(10), len(all) + 1} {
			if k < 1 {
				continue
			}
			paths, err := KShortest(g, start, goal, k, nil)
			if err != nil {
				t.Fatalf("trial %d, k = %d: error = %v", trial, k, err)
			}
			want := all[:min(k, len(all))]
			if len(want) == 0 {
				want = nil
			}
			if t.Run("", func(t *testing.T) { assertKShortest(t, g, paths, start, goal, want) }); t.Failed() {
				t.Fatalf("trial %d, k = %d", trial, k)
			}
		}
	}
}

func ExampleKShortest() {
	g := gograph.New[string](gograph.Weighted())
	home := g.AddVertexByLabel("home")
	park := g.AddVertexByLabel("park")
	shop := g.AddVertexByLabel("shop")
	work := g.AddVertexByLabel("work")

	_, _ = g.AddEdge(home, park, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(park, work, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(home, shop, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(shop, work, gograph.WithEdgeWeight(5))
	_, _ = g.AddEdge(park, shop, gograph.WithEdgeWeight(1))

	paths, err := KShortest(g, "home", "work", 3, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, p := range paths {
		var cost float64
		for i := 1; i < len(p); i++ {
			cost += g.GetEdge(p[i-1], p[i]).Weight()
		}
		fmt.Println(pathLabels(p), cost)
	}
	// Output:
	// [home park work] 4
	// [home shop park work] 4
	// [home shop work] 6
}

// kShortestGrid returns an undirected n by n grid with random weights from 1
// to 9. The vertex at column x and row y has the label x*n+y.
func kShortestGrid(n int) gograph.Graph[int] {
	rng := rand.New(rand.NewPCG(3, 7)) //nolint:gosec // seeded so runs are comparable
	g := gograph.New[int](gograph.Weighted())
	vertices := make([]*gograph.Vertex[int], n*n)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for x := range n {
		for y := range n {
			if x+1 < n {
				_, _ = g.AddEdge(vertices[x*n+y], vertices[(x+1)*n+y], gograph.WithEdgeWeight(float64(1+rng.IntN(9))))
			}
			if y+1 < n {
				_, _ = g.AddEdge(vertices[x*n+y], vertices[x*n+y+1], gograph.WithEdgeWeight(float64(1+rng.IntN(9))))
			}
		}
	}
	return g
}

func BenchmarkKShortest(b *testing.B) {
	const n = 100
	g := kShortestGrid(n)
	for _, k := range []int{1, 10, 100} {
		b.Run("k="+strconv.Itoa(k), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := KShortest(g, 0, n*n-1, k, nil); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
