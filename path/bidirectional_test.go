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

// assertPath checks that p runs from start to goal over edges of g and that
// its edge weights add up to cost.
func assertPath(t *testing.T, g gograph.Graph[string], p []*gograph.Vertex[string], start, goal string, cost float64) {
	t.Helper()
	if len(p) == 0 || p[0].Label() != start || p[len(p)-1].Label() != goal {
		t.Fatalf("path = %v, want it to run from %s to %s", pathLabels(p), start, goal)
	}
	var sum float64
	for i, v := range p {
		if g.GetVertexByID(v.Label()) != v {
			t.Errorf("vertex %s is not the graph's own pointer", v.Label())
		}
		if i == 0 {
			continue
		}
		e := g.GetEdge(p[i-1], v)
		if e == nil {
			t.Fatalf("no edge from %s to %s in %v", p[i-1].Label(), v.Label(), pathLabels(p))
		}
		sum += e.Weight()
	}
	if sum != cost {
		t.Errorf("path %v weighs %v, but the cost is %v", pathLabels(p), sum, cost)
	}
}

func TestBidirectional_PathGraph(t *testing.T) {
	var edges []weightedEdge
	for i := range 9 {
		edges = append(edges, weightedEdge{strconv.Itoa(i), strconv.Itoa(i + 1), float64(i + 1)})
	}
	g := directedWeighted(edges...)

	p, cost, err := Bidirectional(g, "0", "9", nil)
	if err != nil {
		t.Fatalf("Bidirectional() error = %v", err)
	}
	if cost != 45 || len(p) != 10 {
		t.Fatalf("path = %v, cost = %v, want 10 vertices and cost 45", pathLabels(p), cost)
	}
	assertPath(t, g, p, "0", "9", cost)

	if dist := Dijkstra(g, "0")["9"]; dist != cost {
		t.Errorf("Dijkstra gives %v, Bidirectional %v", dist, cost)
	}
}

func TestBidirectional_PicksTheCheaperPath(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "G", 10},
		weightedEdge{"S", "B", 4}, weightedEdge{"B", "C", 1}, weightedEdge{"C", "G", 1},
	)

	p, cost, err := Bidirectional(g, "S", "G", nil)
	if err != nil {
		t.Fatalf("Bidirectional() error = %v", err)
	}
	if want := []string{"S", "B", "C", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 6 {
		t.Fatalf("path = %v, cost = %v, want %v and 6", pathLabels(p), cost, want)
	}
}

func TestBidirectional_DirectedBackwardUsesIncomingEdges(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 1}, weightedEdge{"C", "B", 1})

	if _, _, err := Bidirectional(g, "B", "A", nil); !errors.Is(err, ErrUnreachable) {
		t.Errorf("B to A: error = %v, want %v", err, ErrUnreachable)
	}
	if _, _, err := Bidirectional(g, "A", "C", nil); !errors.Is(err, ErrUnreachable) {
		t.Errorf("A to C: error = %v, want %v", err, ErrUnreachable)
	}
}

func TestBidirectional_Undirected(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a, b, c := g.AddVertexByLabel("A"), g.AddVertexByLabel("B"), g.AddVertexByLabel("C")
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(3))

	for _, tt := range []struct{ start, goal string }{{"A", "C"}, {"C", "A"}, {"B", "A"}} {
		p, cost, err := Bidirectional(g, tt.start, tt.goal, nil)
		if err != nil {
			t.Fatalf("%s to %s: error = %v", tt.start, tt.goal, err)
		}
		assertPath(t, g, p, tt.start, tt.goal, cost)
	}

	_, cost, _ := Bidirectional(g, "C", "A", nil)
	if cost != 5 {
		t.Errorf("C to A: cost = %v, want 5", cost)
	}
}

func TestBidirectional_Cost(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "B", 1}, weightedEdge{"B", "G", 1},
		weightedEdge{"S", "G", 10},
	)
	hops := func(*gograph.Edge[string]) float64 { return 1 }

	p, cost, err := Bidirectional(g, "S", "G", hops)
	if err != nil {
		t.Fatalf("Bidirectional() error = %v", err)
	}
	if want := []string{"S", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 1 {
		t.Errorf("path = %v, cost = %v, want %v and 1", pathLabels(p), cost, want)
	}
}

func TestBidirectional_StartIsGoal(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 1})

	p, cost, err := Bidirectional(g, "A", "A", nil)
	if err != nil {
		t.Fatalf("Bidirectional() error = %v", err)
	}
	if len(p) != 1 || p[0] != g.GetVertexByID("A") || cost != 0 {
		t.Errorf("path = %v, cost = %v, want [A] and 0", pathLabels(p), cost)
	}
}

func TestBidirectional_Errors(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 1}, weightedEdge{"D", "E", 1}, weightedEdge{"F", "G", 1}, weightedEdge{"G", "H", -1})

	tests := []struct {
		name        string
		start, goal string
		cost        func(*gograph.Edge[string]) float64
		want        error
	}{
		{name: "missing start", start: "X", goal: "B", want: gograph.ErrVertexDoesNotExist},
		{name: "missing goal", start: "A", goal: "X", want: gograph.ErrVertexDoesNotExist},
		{name: "unreachable", start: "A", goal: "E", want: ErrUnreachable},
		{name: "negative weight", start: "F", goal: "H", want: ErrNegativeWeight},
		{
			name: "negative cost", start: "D", goal: "E",
			cost: func(*gograph.Edge[string]) float64 { return -1 },
			want: ErrNegativeWeight,
		},
		{
			name: "negative infinite cost", start: "D", goal: "E",
			cost: func(*gograph.Edge[string]) float64 { return math.Inf(-1) },
			want: ErrNegativeWeight,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _, err := Bidirectional(g, tt.start, tt.goal, tt.cost)
			if !errors.Is(err, tt.want) {
				t.Errorf("error = %v, want %v", err, tt.want)
			}
			if p != nil {
				t.Errorf("path = %v, want nil", pathLabels(p))
			}
		})
	}
}

func TestBidirectional_NonFiniteWeights(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "G", math.NaN()},
		weightedEdge{"S", "A", math.Inf(1)}, weightedEdge{"A", "G", 1},
		weightedEdge{"S", "B", 3}, weightedEdge{"B", "G", 4},
	)

	p, cost, err := Bidirectional(g, "S", "G", nil)
	if err != nil {
		t.Fatalf("Bidirectional() error = %v", err)
	}
	if want := []string{"S", "B", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 7 {
		t.Errorf("path = %v, cost = %v, want %v and 7", pathLabels(p), cost, want)
	}

	only := directedWeighted(weightedEdge{"S", "G", math.Inf(1)})
	if _, _, err := Bidirectional(only, "S", "G", nil); !errors.Is(err, ErrUnreachable) {
		t.Errorf("+Inf only: error = %v, want %v", err, ErrUnreachable)
	}
}

func TestBidirectional_SamePathOnEveryRun(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"A", "B", 1}, weightedEdge{"B", "D", 1},
		weightedEdge{"A", "C", 1}, weightedEdge{"C", "D", 1},
	)

	first, _, err := Bidirectional(g, "A", "D", nil)
	if err != nil {
		t.Fatalf("Bidirectional() error = %v", err)
	}
	for range 20 {
		next, _, _ := Bidirectional(g, "A", "D", nil)
		if !reflect.DeepEqual(pathLabels(next), pathLabels(first)) {
			t.Fatalf("path changed from %v to %v", pathLabels(first), pathLabels(next))
		}
	}
}

func TestBidirectional_StopsEarly(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	for i := range 10_000 {
		_, _ = g.AddEdge(gograph.NewVertex(strconv.Itoa(i)), gograph.NewVertex(strconv.Itoa(i+1)), gograph.WithEdgeWeight(1))
	}

	calls := 0
	count := func(e *gograph.Edge[string]) float64 {
		calls++
		return e.Weight()
	}
	if _, _, err := Bidirectional(g, "5000", "5002", count); err != nil {
		t.Fatalf("Bidirectional() error = %v", err)
	}
	if calls > 10 {
		t.Errorf("cost was called %d times for a goal two edges away", calls)
	}
}

func TestBidirectional_MatchesDijkstra(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 8)) //nolint:gosec // seeded so failures are reproducible

	for run := range 3000 {
		directed := run%2 == 0
		var g gograph.Graph[string]
		if directed {
			g = gograph.New[string](gograph.Directed(), gograph.Weighted())
		} else {
			g = gograph.New[string](gograph.Weighted())
		}
		n := 1 + rng.IntN(9)
		for i := range n {
			g.AddVertexByLabel(strconv.Itoa(i))
		}
		for range rng.IntN(2*n + 1) {
			from, to := g.GetVertexByID(strconv.Itoa(rng.IntN(n))), g.GetVertexByID(strconv.Itoa(rng.IntN(n)))
			if g.ContainsEdge(from, to) {
				continue
			}
			_, _ = g.AddEdge(from, to, gograph.WithEdgeWeight(float64(rng.IntN(6))))
		}

		start, goal := strconv.Itoa(rng.IntN(n)), strconv.Itoa(rng.IntN(n))
		want := Dijkstra(g, start)[goal]
		p, cost, err := Bidirectional(g, start, goal, nil)
		if want == math.MaxFloat64 {
			if !errors.Is(err, ErrUnreachable) {
				t.Fatalf("run %d: %s to %s: error = %v, want %v", run, start, goal, err, ErrUnreachable)
			}
			continue
		}
		if err != nil {
			t.Fatalf("run %d: %s to %s: error = %v", run, start, goal, err)
		}
		if cost != want {
			t.Fatalf("run %d: %s to %s: cost = %v, Dijkstra gives %v", run, start, goal, cost, want)
		}
		assertPath(t, g, p, start, goal, cost)
	}
}

func ExampleBidirectional() {
	g := gograph.New[string](gograph.Weighted())
	home := g.AddVertexByLabel("home")
	park := g.AddVertexByLabel("park")
	shop := g.AddVertexByLabel("shop")
	work := g.AddVertexByLabel("work")

	_, _ = g.AddEdge(home, park, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(park, work, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(home, shop, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(shop, work, gograph.WithEdgeWeight(5))

	p, cost, err := Bidirectional(g, "home", "work", nil)
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

func pathLabels(vertices []*gograph.Vertex[string]) []string {
	out := make([]string, len(vertices))
	for i, v := range vertices {
		out[i] = v.Label()
	}
	return out
}
