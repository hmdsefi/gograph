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

func TestIDAStar_MatchesAStarOnASmallGraph(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 4}, weightedEdge{"S", "B", 1}, weightedEdge{"B", "A", 2},
		weightedEdge{"A", "C", 1}, weightedEdge{"B", "C", 5}, weightedEdge{"C", "G", 3},
		weightedEdge{"A", "G", 6},
	)
	estimate := map[string]float64{"S": 5, "A": 3, "B": 4, "C": 2}
	h := func(label string) float64 { return estimate[label] }

	for name, heuristic := range map[string]func(string) float64{"zero": nil, "estimate": h} {
		t.Run(name, func(t *testing.T) {
			p, cost, err := IDAStar(g, "S", "G", nil, heuristic)
			if err != nil {
				t.Fatal(err)
			}
			wantPath, wantCost, _ := AStar(g, "S", "G", nil, heuristic)
			if !reflect.DeepEqual(p, wantPath) || cost != wantCost {
				t.Errorf("IDAStar = %v, %v; AStar = %v, %v", pathLabels(p), cost, pathLabels(wantPath), wantCost)
			}
			if want := []string{"S", "B", "A", "C", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 7 {
				t.Errorf("path = %v, cost = %v, want %v and 7", pathLabels(p), cost, want)
			}
		})
	}
}

func TestIDAStar_RaisesTheLimitToTheNextEstimatedTotal(t *testing.T) {
	// The limits are 0, then 1, 3 and 6. Each search follows the edges within
	// the limit and reads the cost of each edge it tries.
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "B", 2}, weightedEdge{"B", "G", 3},
		weightedEdge{"S", "G", 7},
	)
	calls := 0
	cost := func(e *gograph.Edge[string]) float64 {
		calls++
		return e.Weight()
	}
	p, total, err := IDAStar(g, "S", "G", cost, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"S", "A", "B", "G"}; !reflect.DeepEqual(pathLabels(p), want) || total != 6 {
		t.Fatalf("path = %v, cost = %v, want %v and 6", pathLabels(p), total, want)
	}
	// Search 1 tries S-A and S-G, search 2 adds A-B, search 3 B-G, and
	// search 4 reaches G: 2 + 3 + 4 + 3 edges.
	if calls != 12 {
		t.Errorf("cost was called %d times, want 12", calls)
	}
}

func TestIDAStar_CostFunctionAndUnweightedGraph(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 1}, weightedEdge{"B", "C", 1}, weightedEdge{"A", "C", 5})
	hops := func(*gograph.Edge[string]) float64 { return 1 }
	p, cost, err := IDAStar(g, "A", "C", hops, nil)
	if err != nil || !reflect.DeepEqual(pathLabels(p), []string{"A", "C"}) || cost != 1 {
		t.Errorf("hop cost: got %v, %v, %v; want [A C], 1", pathLabels(p), cost, err)
	}

	u := gograph.New[string](gograph.Directed())
	_, _ = u.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))
	_, _ = u.AddEdge(u.GetVertexByID("B"), gograph.NewVertex("C"))
	p, cost, err = IDAStar(u, "A", "C", nil, nil)
	if err != nil || !reflect.DeepEqual(pathLabels(p), []string{"A", "B", "C"}) || cost != 0 {
		t.Errorf("unweighted: got %v, %v, %v; want [A B C], 0", pathLabels(p), cost, err)
	}
}

func TestIDAStar_UndirectedEdgeWorksBothWays(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a, b, c := g.AddVertexByLabel("A"), g.AddVertexByLabel("B"), g.AddVertexByLabel("C")
	_, _ = g.AddEdge(b, a, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(c, b, gograph.WithEdgeWeight(3))
	p, cost, err := IDAStar(g, "A", "C", nil, nil)
	if err != nil || !reflect.DeepEqual(pathLabels(p), []string{"A", "B", "C"}) || cost != 5 {
		t.Errorf("got %v, %v, %v; want [A B C], 5", pathLabels(p), cost, err)
	}
}

func TestIDAStar_Errors(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 1}, weightedEdge{"B", "C", -1}, weightedEdge{"D", "A", 1})

	tests := []struct {
		name        string
		start, goal string
		want        error
	}{
		{"missing start", "X", "B", gograph.ErrVertexDoesNotExist},
		{"missing goal", "A", "X", gograph.ErrVertexDoesNotExist},
		{"unreachable", "C", "D", ErrUnreachable},
		{"negative cost", "A", "C", ErrNegativeWeight},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, cost, err := IDAStar(g, tc.start, tc.goal, nil, nil)
			if !errors.Is(err, tc.want) || p != nil || cost != 0 {
				t.Errorf("got %v, %v, %v; want nil, 0, %v", pathLabels(p), cost, err, tc.want)
			}
		})
	}
}

func TestIDAStar_StartEqualsGoal(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "A", -1}, weightedEdge{"A", "B", 1})
	p, cost, err := IDAStar(g, "A", "A", nil, func(string) float64 { panic("not called") })
	if err != nil || len(p) != 1 || p[0] != g.GetVertexByID("A") || cost != 0 {
		t.Errorf("got %v, %v, %v; want the vertex A and 0", pathLabels(p), cost, err)
	}
}

func TestIDAStar_NonFiniteCostsAndEstimates(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", math.NaN()}, weightedEdge{"A", "G", 1},
		weightedEdge{"S", "B", math.Inf(1)}, weightedEdge{"B", "G", 1},
		weightedEdge{"S", "C", math.MaxFloat64}, weightedEdge{"C", "G", math.MaxFloat64},
		weightedEdge{"S", "D", 4}, weightedEdge{"D", "G", 4},
	)
	p, cost, err := IDAStar(g, "S", "G", nil, nil)
	if err != nil || !reflect.DeepEqual(pathLabels(p), []string{"S", "D", "G"}) || cost != 8 {
		t.Errorf("got %v, %v, %v; want [S D G], 8", pathLabels(p), cost, err)
	}

	// A NaN estimate counts as 0.
	p, cost, err = IDAStar(g, "S", "G", nil, func(string) float64 { return math.NaN() })
	if err != nil || !reflect.DeepEqual(pathLabels(p), []string{"S", "D", "G"}) || cost != 8 {
		t.Errorf("NaN estimate: got %v, %v, %v; want [S D G], 8", pathLabels(p), cost, err)
	}

	// A +Inf estimate is searched last, but still searched.
	p, cost, err = IDAStar(g, "S", "G", nil, func(string) float64 { return math.Inf(1) })
	if err != nil || !reflect.DeepEqual(pathLabels(p), []string{"S", "D", "G"}) || cost != 8 {
		t.Errorf("+Inf estimate: got %v, %v, %v; want [S D G], 8", pathLabels(p), cost, err)
	}
	unreachable := directedWeighted(weightedEdge{"S", "A", 1}, weightedEdge{"G", "S", 1})
	if _, _, err := IDAStar(unreachable, "S", "G", nil, func(string) float64 { return math.Inf(1) }); !errors.Is(err, ErrUnreachable) {
		t.Errorf("+Inf estimate, unreachable: error = %v, want %v", err, ErrUnreachable)
	}
}

func TestIDAStar_InfiniteEstimatesStillGiveTheCheapestPath(t *testing.T) {
	inf := math.Inf(1)
	tests := []struct {
		name     string
		edges    []weightedEdge
		estimate map[string]float64
		want     []string
		cost     float64
	}{
		{
			// The estimate of start is never compared with a limit.
			name: "start",
			edges: []weightedEdge{
				{"S", "A", 1}, {"S", "B", 1}, {"A", "G", 10}, {"B", "G", 1},
			},
			estimate: map[string]float64{"S": inf},
			want:     []string{"S", "B", "G"},
			cost:     2,
		},
		{
			// Every path goes through X, so the searches start over with its
			// estimate counted as 0, and still pick the cheaper branch.
			name: "every path",
			edges: []weightedEdge{
				{"S", "X", 1}, {"X", "A", 10}, {"X", "B", 1}, {"A", "G", 1}, {"B", "G", 1},
			},
			estimate: map[string]float64{"X": inf},
			want:     []string{"S", "X", "B", "G"},
			cost:     3,
		},
		{
			// A total that overflows to +Inf counts the same way.
			name: "overflowing total",
			edges: []weightedEdge{
				{"S", "A", math.MaxFloat64 / 2}, {"A", "G", 1},
			},
			estimate: map[string]float64{"A": math.MaxFloat64},
			want:     []string{"S", "A", "G"},
			cost:     math.MaxFloat64/2 + 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := directedWeighted(tc.edges...)
			h := func(label string) float64 { return tc.estimate[label] }
			p, cost, err := IDAStar(g, "S", "G", nil, h)
			if err != nil || !reflect.DeepEqual(pathLabels(p), tc.want) || cost != tc.cost {
				t.Errorf("got %v, %v, %v; want %v, %v", pathLabels(p), cost, err, tc.want, tc.cost)
			}
			if _, want, _ := AStar(g, "S", "G", nil, h); cost != want {
				t.Errorf("cost = %v, AStar gives %v", cost, want)
			}
		})
	}
}

func TestIDAStar_NegativeCostBackIntoThePathIsReported(t *testing.T) {
	directed := directedWeighted(weightedEdge{"S", "A", 1}, weightedEdge{"A", "S", -1}, weightedEdge{"A", "G", 1})
	undirected := gograph.New[string](gograph.Weighted())
	s, g := undirected.AddVertexByLabel("S"), undirected.AddVertexByLabel("G")
	_, _ = undirected.AddEdge(s, s, gograph.WithEdgeWeight(-1))
	_, _ = undirected.AddEdge(s, g, gograph.WithEdgeWeight(1))

	for name, graph := range map[string]gograph.Graph[string]{"edge back to start": directed, "self-loop": undirected} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := IDAStar(graph, "S", "G", nil, nil); !errors.Is(err, ErrNegativeWeight) {
				t.Errorf("error = %v, want %v", err, ErrNegativeWeight)
			}
			if _, _, err := AStar(graph, "S", "G", nil, nil); !errors.Is(err, ErrNegativeWeight) {
				t.Errorf("AStar error = %v, want %v", err, ErrNegativeWeight)
			}
		})
	}
}

func TestIDAStar_HeuristicIsNeverCalledForTheGoal(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 1}, weightedEdge{"B", "G", 1})
	_, _, err := IDAStar(g, "A", "G", nil, func(label string) float64 {
		if label == "G" {
			t.Fatal("heuristic called for the goal")
		}
		return 0
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIDAStar_TiesFollowTheOrderOfNeighbors(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "B", 1}, weightedEdge{"S", "A", 1},
		weightedEdge{"A", "G", 1}, weightedEdge{"B", "G", 1},
	)
	for range 5 {
		p, _, err := IDAStar(g, "S", "G", nil, nil)
		if err != nil || !reflect.DeepEqual(pathLabels(p), []string{"S", "B", "G"}) {
			t.Fatalf("got %v, %v; want [S B G] on every run", pathLabels(p), err)
		}
	}
}

func TestIDAStar_ZeroCostCycleIsHarmless(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 0}, weightedEdge{"A", "S", 0}, weightedEdge{"A", "A", 0},
		weightedEdge{"A", "G", 2},
	)
	p, cost, err := IDAStar(g, "S", "G", nil, nil)
	if err != nil || !reflect.DeepEqual(pathLabels(p), []string{"S", "A", "G"}) || cost != 2 {
		t.Errorf("got %v, %v, %v; want [S A G], 2", pathLabels(p), cost, err)
	}
}

func TestIDAStar_NeighborWithoutAnEdgeIsSkipped(t *testing.T) {
	g := astarNoEdges{directedWeighted(weightedEdge{"A", "B", 1})}
	if _, _, err := IDAStar[string](g, "A", "B", nil, nil); !errors.Is(err, ErrUnreachable) {
		t.Errorf("error = %v, want %v", err, ErrUnreachable)
	}
}

// TestIDAStar_MatchesAStar compares random graphs with AStar, under
// heuristics of 0, of the exact remaining cost, and of a random share of it.
func TestIDAStar_MatchesAStar(t *testing.T) {
	rng := rand.New(rand.NewPCG(23, 7)) //nolint:gosec // seeded so failures are reproducible

	for trial := range 3000 {
		directed := trial%2 == 0
		var g gograph.Graph[string]
		if directed {
			g = gograph.New[string](gograph.Directed(), gograph.Weighted())
		} else {
			g = gograph.New[string](gograph.Weighted())
		}

		n := 2 + rng.IntN(8)
		labels := make([]string, n)
		for i := range labels {
			labels[i] = strconv.Itoa(i)
			g.AddVertexByLabel(labels[i])
		}
		for range rng.IntN(3 * n) {
			from, to := g.GetVertexByID(labels[rng.IntN(n)]), g.GetVertexByID(labels[rng.IntN(n)])
			_, _ = g.AddEdge(from, to, gograph.WithEdgeWeight(float64(rng.IntN(10))))
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

		for name, h := range map[string]func(string) float64{
			"zero":  nil,
			"exact": func(label string) float64 { return remaining[label] },
			"share": func(label string) float64 { return share[label] },
		} {
			_, want, wantErr := AStar(g, start, goal, nil, h)
			p, cost, err := IDAStar(g, start, goal, nil, h)
			if !errors.Is(err, wantErr) {
				t.Fatalf("trial %d, %s: error = %v, want %v", trial, name, err, wantErr)
			}
			if err != nil {
				continue
			}
			if cost != want {
				t.Fatalf("trial %d, %s: cost = %v, want %v", trial, name, cost, want)
			}
			assertPath(t, g, p, start, goal, cost)
		}
	}
}

func ExampleIDAStar() {
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

	p, cost, err := IDAStar(g, "home", "work", nil, func(label string) float64 {
		return straightLine[label]
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(pathLabels(p), cost)
	// Output:
	// [home park work] 4
}

func BenchmarkIDAStar(b *testing.B) {
	const n = 300
	g := astarGrid(n)
	goal := n*n - 1
	b.Run("manhattan", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, _, err := IDAStar(g, 0, goal, nil, astarManhattan(n)); err != nil {
				b.Fatal(err)
			}
		}
	})

	// A small grid without a heuristic is the bad case: every path within
	// each limit is followed again for each new limit.
	small := astarGrid(7)
	b.Run("grid 7x7 zero", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, _, err := IDAStar(small, 0, 48, nil, nil); err != nil {
				b.Fatal(err)
			}
		}
	})

	tree := idaTree(4, 8)
	b.Run("tree", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, _, err := IDAStar(tree, 0, tree.Order()-1, nil, nil); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// idaTree returns a directed tree where each vertex has `branching` children
// and the leaves are `depth` edges from the root, vertex 0. An edge to the
// first child of a vertex costs 2 and an edge to any other child 1, so the
// cheapest leaf is the last one, which has the largest label.
func idaTree(branching, depth int) gograph.Graph[uint32] {
	g := gograph.New[uint32](gograph.Directed(), gograph.Weighted())
	level := []*gograph.Vertex[uint32]{g.AddVertexByLabel(0)}
	next := uint32(1)
	for range depth {
		var children []*gograph.Vertex[uint32]
		for _, parent := range level {
			for c := range branching {
				child := g.AddVertexByLabel(next)
				next++
				w := 1.0
				if c == 0 {
					w = 2
				}
				_, _ = g.AddEdge(parent, child, gograph.WithEdgeWeight(w))
				children = append(children, child)
			}
		}
		level = children
	}
	return g
}
