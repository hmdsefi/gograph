package path

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

// astarGrid returns an undirected n by n grid with a weight of 1 on each edge.
// The vertex at column x and row y has the label x*n+y.
func astarGrid(n int) gograph.Graph[int] {
	g := gograph.New[int](gograph.Weighted())
	vertices := make([]*gograph.Vertex[int], n*n)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for x := range n {
		for y := range n {
			if x+1 < n {
				_, _ = g.AddEdge(vertices[x*n+y], vertices[(x+1)*n+y], gograph.WithEdgeWeight(1))
			}
			if y+1 < n {
				_, _ = g.AddEdge(vertices[x*n+y], vertices[x*n+y+1], gograph.WithEdgeWeight(1))
			}
		}
	}
	return g
}

// astarManhattan returns the grid distance to the far corner of an n by n grid.
func astarManhattan(n int) func(int) float64 {
	return func(label int) float64 {
		return float64(n-1-label/n) + float64(n-1-label%n)
	}
}

func TestAStar_GridHeuristicExpandsFewerVertices(t *testing.T) {
	const n = 40
	g := astarGrid(n)
	goal := n*n - 1

	var calls int
	counted := func(e *gograph.Edge[int]) float64 {
		calls++
		return e.Weight()
	}

	zeroPath, zeroCost, err := AStar(g, 0, goal, counted, nil)
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}
	zeroCalls := calls

	calls = 0
	guidedPath, guidedCost, err := AStar(g, 0, goal, counted, astarManhattan(n))
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}

	want := float64(2 * (n - 1))
	if zeroCost != want || guidedCost != want {
		t.Errorf("cost = %v without a heuristic and %v with one, want %v", zeroCost, guidedCost, want)
	}
	if len(zeroPath) != 2*n-1 || len(guidedPath) != 2*n-1 {
		t.Errorf("paths have %d and %d vertices, want %d", len(zeroPath), len(guidedPath), 2*n-1)
	}
	// Every vertex on a straight route ties, and the larger cost goes first, so
	// the guided search walks one route to the corner.
	if calls >= zeroCalls/10 {
		t.Errorf("the heuristic read %d edges, a zero heuristic %d, want far fewer", calls, zeroCalls)
	}
}

func TestAStar_PicksTheCheaperPath(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "G", 10},
		weightedEdge{"S", "B", 4}, weightedEdge{"B", "C", 1}, weightedEdge{"C", "G", 1},
	)

	p, cost, err := AStar(g, "S", "G", nil, nil)
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}
	if want := []string{"S", "B", "C", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 6 {
		t.Fatalf("path = %v, cost = %v, want %v and 6", pathLabels(p), cost, want)
	}
	assertPath(t, g, p, "S", "G", cost)
}

func TestAStar_CostFunctionReplacesTheWeight(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "G", 1},
		weightedEdge{"S", "G", 5},
	)
	closed := func(e *gograph.Edge[string]) float64 {
		if e.Source().Label() == "A" {
			return math.Inf(1)
		}
		return e.Weight()
	}

	p, cost, err := AStar(g, "S", "G", closed, nil)
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}
	if want := []string{"S", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 5 {
		t.Errorf("path = %v, cost = %v, want %v and 5", pathLabels(p), cost, want)
	}
}

func TestAStar_UnweightedGraphCostsZero(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	a, b := g.AddVertexByLabel("A"), g.AddVertexByLabel("B")
	_, _ = g.AddEdge(a, b)

	p, cost, err := AStar(g, "A", "B", nil, nil)
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}
	if want := []string{"A", "B"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 0 {
		t.Errorf("path = %v, cost = %v, want %v and 0", pathLabels(p), cost, want)
	}
}

func TestAStar_DirectedEdgesGoOneWay(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 1})

	if _, _, err := AStar(g, "A", "B", nil, nil); err != nil {
		t.Errorf("A to B: error = %v", err)
	}
	if _, _, err := AStar(g, "B", "A", nil, nil); !errors.Is(err, ErrUnreachable) {
		t.Errorf("B to A: error = %v, want %v", err, ErrUnreachable)
	}
}

func TestAStar_UndirectedEdgeWorksBothWaysAndCountsOnce(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a, b, c := g.AddVertexByLabel("A"), g.AddVertexByLabel("B"), g.AddVertexByLabel("C")
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(3))

	for _, tt := range []struct {
		start, goal string
		want        []string
	}{
		{"A", "C", []string{"A", "B", "C"}},
		{"C", "A", []string{"C", "B", "A"}},
	} {
		p, cost, err := AStar(g, tt.start, tt.goal, nil, nil)
		if err != nil {
			t.Fatalf("%s to %s: error = %v", tt.start, tt.goal, err)
		}
		if !reflect.DeepEqual(pathLabels(p), tt.want) || cost != 5 {
			t.Errorf("%s to %s: path = %v, cost = %v, want %v and 5", tt.start, tt.goal, pathLabels(p), cost, tt.want)
		}
	}
}

func TestAStar_Errors(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 1}, weightedEdge{"C", "D", 1})
	negative := func(*gograph.Edge[string]) float64 { return -1 }

	tests := []struct {
		name        string
		start, goal string
		cost        func(*gograph.Edge[string]) float64
		want        error
	}{
		{"missing start", "X", "A", nil, gograph.ErrVertexDoesNotExist},
		{"missing goal", "A", "X", nil, gograph.ErrVertexDoesNotExist},
		{"missing both", "X", "Y", nil, gograph.ErrVertexDoesNotExist},
		{"same missing vertex", "X", "X", nil, gograph.ErrVertexDoesNotExist},
		{"unreachable", "A", "D", nil, ErrUnreachable},
		{"negative cost", "A", "B", negative, ErrNegativeWeight},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, cost, err := AStar(g, tt.start, tt.goal, tt.cost, nil)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if p != nil || cost != 0 {
				t.Errorf("path = %v, cost = %v, want nil and 0", pathLabels(p), cost)
			}
		})
	}
}

func TestAStar_NegativeWeightsAndInfinities(t *testing.T) {
	tests := []struct {
		name  string
		edges []weightedEdge
		want  error
	}{
		{"negative weight", []weightedEdge{{"S", "G", -1}}, ErrNegativeWeight},
		{"-Inf weight", []weightedEdge{{"S", "G", math.Inf(-1)}}, ErrNegativeWeight},
		{"NaN weight is skipped", []weightedEdge{{"S", "G", math.NaN()}}, ErrUnreachable},
		{"+Inf weight is no path", []weightedEdge{{"S", "G", math.Inf(1)}}, ErrUnreachable},
		{
			"a sum past the float64 range is no path",
			[]weightedEdge{{"S", "A", math.MaxFloat64}, {"A", "G", math.MaxFloat64}},
			ErrUnreachable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := directedWeighted(tt.edges...)
			if _, _, err := AStar(g, "S", "G", nil, nil); !errors.Is(err, tt.want) {
				t.Errorf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestAStar_NegativeCostOnAnEdgeNeverReachedIsNotReported(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "G", 1},
		weightedEdge{"S", "A", 5}, weightedEdge{"A", "B", -1},
	)
	if _, cost, err := AStar(g, "S", "G", nil, nil); err != nil || cost != 1 {
		t.Errorf("cost = %v, error = %v, want 1 and nil", cost, err)
	}
}

func TestAStar_StartEqualsGoal(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "A", 3}, weightedEdge{"A", "B", 1})
	p, cost, err := AStar(g, "A", "A", func(*gograph.Edge[string]) float64 {
		t.Error("cost should not be called")
		return 0
	}, func(string) float64 {
		t.Error("heuristic should not be called")
		return 0
	})
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}
	if len(p) != 1 || p[0] != g.GetVertexByID("A") || cost != 0 {
		t.Errorf("path = %v, cost = %v, want the graph's own A and 0", pathLabels(p), cost)
	}
}

func TestAStar_SelfLoopAndZeroCycleAreHarmless(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "S", 0}, weightedEdge{"S", "A", 0},
		weightedEdge{"A", "S", 0}, weightedEdge{"A", "G", 2},
	)
	p, cost, err := AStar(g, "S", "G", nil, nil)
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}
	if want := []string{"S", "A", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 2 {
		t.Errorf("path = %v, cost = %v, want %v and 2", pathLabels(p), cost, want)
	}
}

func TestAStar_SamePathOnEveryRun(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "B", 1}, weightedEdge{"S", "A", 1},
		weightedEdge{"A", "G", 1}, weightedEdge{"B", "G", 1},
	)
	for range 20 {
		p, cost, err := AStar(g, "S", "G", nil, nil)
		if err != nil {
			t.Fatalf("AStar() error = %v", err)
		}
		// B's edge from S was added first.
		if want := []string{"S", "B", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 2 {
			t.Fatalf("path = %v, cost = %v, want %v and 2", pathLabels(p), cost, want)
		}
	}
}

func TestAStar_LargerCostGoesFirstOnEqualTotals(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"S", "B", 2},
		weightedEdge{"A", "X", 1}, weightedEdge{"B", "G", 1},
	)
	heuristic := map[string]float64{"S": 3, "A": 2, "B": 1, "X": 1}

	var expanded []string
	_, cost, err := AStar(g, "S", "G", func(e *gograph.Edge[string]) float64 {
		expanded = append(expanded, e.Source().Label())
		return e.Weight()
	}, func(label string) float64 { return heuristic[label] })
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}
	if want := []string{"S", "S", "B"}; !reflect.DeepEqual(expanded, want) || cost != 3 {
		t.Errorf("read the edges of %v, cost = %v, want %v and 3", expanded, cost, want)
	}
}

func TestAStar_StopsAtTheGoal(t *testing.T) {
	const n = 10_001
	g := gograph.New[int](gograph.Weighted())
	prev := g.AddVertexByLabel(0)
	for i := 1; i < n; i++ {
		v := g.AddVertexByLabel(i)
		_, _ = g.AddEdge(prev, v, gograph.WithEdgeWeight(1))
		prev = v
	}

	var costs, estimates int
	_, cost, err := AStar(g, 0, 2, func(e *gograph.Edge[int]) float64 {
		costs++
		return e.Weight()
	}, func(label int) float64 {
		estimates++
		return math.Max(float64(2-label), 0)
	})
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}
	if cost != 2 {
		t.Errorf("cost = %v, want 2", cost)
	}
	if costs > 4 || estimates > 3 {
		t.Errorf("read %d costs and %d estimates, want at most 4 and 3", costs, estimates)
	}
}

func TestAStar_Heuristic(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "G", 1},
		weightedEdge{"S", "B", 1}, weightedEdge{"B", "G", 5},
	)
	tests := []struct {
		name string
		h    func(string) float64
	}{
		{"NaN counts as 0", func(string) float64 { return math.NaN() }},
		{"negative everywhere", func(string) float64 { return -100 }},
		{"-Inf everywhere", func(string) float64 { return math.Inf(-1) }},
		{"+Inf is searched last", func(string) float64 { return math.Inf(1) }},
		{"a wrong estimate for goal is ignored", func(label string) float64 {
			if label == "G" {
				return -50
			}
			return 0
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, cost, err := AStar(g, "S", "G", nil, tt.h)
			if err != nil {
				t.Fatalf("AStar() error = %v", err)
			}
			if want := []string{"S", "A", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 2 {
				t.Errorf("path = %v, cost = %v, want %v and 2", pathLabels(p), cost, want)
			}
		})
	}
}

func TestAStar_HeuristicCalledOncePerVertex(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 5}, weightedEdge{"S", "B", 1},
		weightedEdge{"B", "A", 1}, weightedEdge{"A", "G", 1},
	)
	calls := map[string]int{}
	_, cost, err := AStar(g, "S", "G", nil, func(label string) float64 {
		calls[label]++
		return 0
	})
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}
	if cost != 3 {
		t.Errorf("cost = %v, want 3", cost)
	}
	for label, n := range calls {
		if n != 1 {
			t.Errorf("heuristic called %d times for %s", n, label)
		}
	}
	if calls["G"] != 0 {
		t.Errorf("heuristic called for the goal")
	}
}

// astarHardGraph returns a graph with an estimate that stays within the true
// remaining cost, and that makes plain A* take about 2^levels expansions.
//
// Each level i runs from V(i-1) to V(i) directly at a cost of 2^(levels-i), or
// for free through X(i), whose high estimate delays it until everything past
// it has been expanded. The deepest X goes first, so each cheaper way to a
// vertex repeats all the levels below it, like a binary counter. The final edge
// to G is costly enough that all the estimates stay within the remaining cost.
func astarHardGraph(levels int) (g gograph.Graph[string], heuristic func(string) float64, want float64) {
	delay := math.Ldexp(1, levels+2)
	want = float64(levels+2) * delay

	estimates := map[string]float64{}
	var edges []weightedEdge
	v := func(i int) string { return "V" + strconv.Itoa(i) }
	for i := 1; i <= levels; i++ {
		x := "X" + strconv.Itoa(i)
		estimates[x] = float64(levels-i+1) * delay
		edges = append(edges,
			weightedEdge{v(i - 1), v(i), math.Ldexp(1, levels-i)},
			weightedEdge{v(i - 1), x, 0},
			weightedEdge{x, v(i), 0},
		)
	}
	edges = append(edges, weightedEdge{v(levels), "G", want})

	return directedWeighted(edges...), func(label string) float64 { return estimates[label] }, want
}

func TestAStar_InconsistentHeuristicStaysWithinQuadraticExpansions(t *testing.T) {
	const levels = 18
	g, heuristic, want := astarHardGraph(levels)
	vertices := int(g.Order())

	var calls int
	p, cost, err := AStar(g, "V0", "G", func(e *gograph.Edge[string]) float64 {
		calls++
		return e.Weight()
	}, heuristic)
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}
	if cost != want {
		t.Errorf("cost = %v, want %v", cost, want)
	}
	assertPath(t, g, p, "V0", "G", cost)

	// Each vertex has at most 2 outgoing edges. Plain A* reads over 1,000,000
	// edges here.
	if limit := 2 * vertices * vertices; calls > limit {
		t.Errorf("read %d edges, want at most %d", calls, limit)
	}
}

// TestAStar_NodeFallsBelowTheFloor uses a heuristic that drops by more than an
// edge costs, from C to D. Expanding C raises the floor to 10. D, queued at a
// total of 20, then improves to 5 and moves below the floor. E joins it at 8,
// and improves to 6 through D while it's still queued.
func TestAStar_NodeFallsBelowTheFloor(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "C", 0}, weightedEdge{"S", "D", 20},
		weightedEdge{"C", "D", 5}, weightedEdge{"C", "E", 8},
		weightedEdge{"D", "E", 1}, weightedEdge{"E", "G", 20},
	)
	heuristic := map[string]float64{"C": 10}

	p, cost, err := AStar(g, "S", "G", nil, func(label string) float64 { return heuristic[label] })
	if err != nil {
		t.Fatalf("AStar() error = %v", err)
	}
	if want := []string{"S", "C", "D", "E", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 26 {
		t.Errorf("path = %v, cost = %v, want %v and 26", pathLabels(p), cost, want)
	}
}

// astarNoEdges is a graph that lists neighbors it has no edge for.
type astarNoEdges struct {
	gograph.Graph[string]
}

func (astarNoEdges) GetEdge(_, _ *gograph.Vertex[string]) *gograph.Edge[string] { return nil }

func TestAStar_NeighborWithoutAnEdgeIsSkipped(t *testing.T) {
	g := astarNoEdges{directedWeighted(weightedEdge{"A", "B", 1})}
	if _, _, err := AStar[string](g, "A", "B", nil, nil); !errors.Is(err, ErrUnreachable) {
		t.Errorf("error = %v, want %v", err, ErrUnreachable)
	}
}

// TestAStar_MatchesDijkstra compares random graphs with Dijkstra, under
// heuristics of 0, of the exact remaining cost, and of a random share of it,
// which is admissible but inconsistent.
func TestAStar_MatchesDijkstra(t *testing.T) {
	rng := rand.New(rand.NewPCG(18, 5)) //nolint:gosec // seeded so failures are reproducible

	for trial := range 3000 {
		directed := trial%2 == 0
		var g gograph.Graph[string]
		if directed {
			g = gograph.New[string](gograph.Directed(), gograph.Weighted())
		} else {
			g = gograph.New[string](gograph.Weighted())
		}

		n := 2 + rng.IntN(12)
		labels := make([]string, n)
		for i := range labels {
			labels[i] = strconv.Itoa(i)
			g.AddVertexByLabel(labels[i])
		}
		for range rng.IntN(4 * n) {
			from, to := g.GetVertexByID(labels[rng.IntN(n)]), g.GetVertexByID(labels[rng.IntN(n)])
			_, _ = g.AddEdge(from, to, gograph.WithEdgeWeight(float64(rng.IntN(20))))
		}

		start, goal := labels[0], labels[n-1]
		remaining := make(map[string]float64, n)
		share := make(map[string]float64, n)
		for _, label := range labels {
			d := Dijkstra(g, label)[goal]
			if d == math.MaxFloat64 {
				d = math.Inf(1)
			}
			remaining[label] = d
			share[label] = d * rng.Float64()
		}

		want := remaining[start]
		for name, h := range map[string]func(string) float64{
			"zero":  nil,
			"exact": func(label string) float64 { return remaining[label] },
			"share": func(label string) float64 { return share[label] },
		} {
			p, cost, err := AStar(g, start, goal, nil, h)
			if math.IsInf(want, 1) {
				if !errors.Is(err, ErrUnreachable) {
					t.Fatalf("trial %d, %s: error = %v, want %v", trial, name, err, ErrUnreachable)
				}
				continue
			}
			if err != nil {
				t.Fatalf("trial %d, %s: error = %v", trial, name, err)
			}
			if cost != want {
				t.Fatalf("trial %d, %s: cost = %v, want %v", trial, name, cost, want)
			}
			assertPath(t, g, p, start, goal, cost)
		}
	}
}

func ExampleAStar() {
	g := gograph.New[string](gograph.Weighted())
	home := g.AddVertexByLabel("home")
	park := g.AddVertexByLabel("park")
	shop := g.AddVertexByLabel("shop")
	work := g.AddVertexByLabel("work")

	_, _ = g.AddEdge(home, park, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(park, work, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(home, shop, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(shop, work, gograph.WithEdgeWeight(5))

	// The straight-line distance to work, which no road can beat.
	straightLine := map[string]float64{"home": 3, "park": 2, "shop": 4}

	p, cost, err := AStar(g, "home", "work", nil, func(label string) float64 {
		return straightLine[label]
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, v := range p {
		fmt.Println(v.Label())
	}
	fmt.Println(cost)
	// Output:
	// home
	// park
	// work
	// 4
}

func BenchmarkAStar(b *testing.B) {
	const n = 300
	g := astarGrid(n)
	goal := n*n - 1

	for _, bm := range []struct {
		name      string
		heuristic func(int) float64
	}{
		{"zero", nil},
		{"manhattan", astarManhattan(n)},
	} {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, _, err := AStar(g, 0, goal, nil, bm.heuristic); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
