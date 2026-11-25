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

// beamDiamond is a diamond whose first edge toward A is cheaper, while the full
// path through B is cheaper.
func beamDiamond() gograph.Graph[string] {
	return directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "G", 10},
		weightedEdge{"S", "B", 2}, weightedEdge{"B", "G", 1},
	)
}

func TestBeam_WidthOneFollowsTheBetterFirstEdge(t *testing.T) {
	p, cost, err := Beam(beamDiamond(), "S", "G", 1, nil)
	if err != nil {
		t.Fatalf("Beam() error = %v", err)
	}
	if want := []string{"S", "A", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 11 {
		t.Errorf("path = %v, cost = %v, want %v and 11", pathLabels(p), cost, want)
	}
}

func TestBeam_LargerWidthFindsTheCheaperPath(t *testing.T) {
	g := beamDiamond()
	p, cost, err := Beam(g, "S", "G", 2, nil)
	if err != nil {
		t.Fatalf("Beam() error = %v", err)
	}
	if want := []string{"S", "B", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 3 {
		t.Errorf("path = %v, cost = %v, want %v and 3", pathLabels(p), cost, want)
	}
	assertPath(t, g, p, "S", "G", cost)
}

func TestBeam_Errors(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 1}, weightedEdge{"C", "D", 1})

	tests := []struct {
		name        string
		start, goal string
		width       int
		want        error
	}{
		{"width 0", "A", "B", 0, ErrInvalidWidth},
		{"negative width", "A", "B", -1, ErrInvalidWidth},
		{"missing start", "X", "B", 1, gograph.ErrVertexDoesNotExist},
		{"missing goal", "A", "X", 1, gograph.ErrVertexDoesNotExist},
		{"same missing vertex", "X", "X", 1, gograph.ErrVertexDoesNotExist},
		{"unreachable", "A", "D", 3, ErrUnreachable},
		{"no edge back", "B", "A", 3, ErrUnreachable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, cost, err := Beam(g, tt.start, tt.goal, tt.width, nil)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if p != nil || cost != 0 {
				t.Errorf("path = %v, cost = %v, want nil and 0", pathLabels(p), cost)
			}
		})
	}
}

func TestBeam_DroppedPathMakesTheGoalUnreachable(t *testing.T) {
	// The only way to G is through B, which ranks below A and is dropped.
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "C", 1},
		weightedEdge{"S", "B", 2}, weightedEdge{"B", "G", 1},
	)
	if _, _, err := Beam(g, "S", "G", 1, nil); !errors.Is(err, ErrUnreachable) {
		t.Errorf("width 1: error = %v, want %v", err, ErrUnreachable)
	}
	if _, cost, err := Beam(g, "S", "G", 2, nil); err != nil || cost != 3 {
		t.Errorf("width 2: cost = %v, error = %v, want 3 and nil", cost, err)
	}
}

func TestBeam_StartEqualsGoal(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "A", 3}, weightedEdge{"A", "B", 1})
	for _, width := range []int{1, 50} {
		p, cost, err := Beam(g, "A", "A", width, func(*gograph.Edge[string], float64) float64 {
			t.Error("score should not be called")
			return 0
		})
		if err != nil {
			t.Fatalf("Beam() error = %v", err)
		}
		if len(p) != 1 || p[0] != g.GetVertexByID("A") || cost != 0 {
			t.Errorf("width %d: path = %v, cost = %v, want the graph's own A and 0", width, pathLabels(p), cost)
		}
	}
}

func TestBeam_SamePathOnEveryRun(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "B", 1}, weightedEdge{"S", "A", 1}, weightedEdge{"S", "C", 1},
		weightedEdge{"A", "G", 1}, weightedEdge{"B", "G", 1}, weightedEdge{"C", "G", 1},
	)
	for range 20 {
		p, cost, err := Beam(g, "S", "G", 2, nil)
		if err != nil {
			t.Fatalf("Beam() error = %v", err)
		}
		// S to B was added first, so it ranks first among the ties.
		if want := []string{"S", "B", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 2 {
			t.Fatalf("path = %v, cost = %v, want %v and 2", pathLabels(p), cost, want)
		}
	}
}

func TestBeam_PathsDoNotRepeatAVertex(t *testing.T) {
	// A, B, A would keep the cheap edges, but A can't come twice.
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "B", 0}, weightedEdge{"B", "A", 0},
		weightedEdge{"A", "A", 0}, weightedEdge{"B", "C", 5}, weightedEdge{"C", "G", 1},
	)
	p, cost, err := Beam(g, "S", "G", 1, nil)
	if err != nil {
		t.Fatalf("Beam() error = %v", err)
	}
	if want := []string{"S", "A", "B", "C", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 7 {
		t.Errorf("path = %v, cost = %v, want %v and 7", pathLabels(p), cost, want)
	}
}

func TestBeam_HighDegreeVertexDoesNotRepeatAVertex(t *testing.T) {
	// H has more neighbors than beamWalkLimit, so its path goes in the map. The
	// cheap edges from H lead back to S and A, which are already on the path.
	edges := []weightedEdge{{"S", "A", 1}, {"A", "H", 1}, {"H", "S", 0}, {"H", "A", 0}}
	for i := range beamWalkLimit {
		edges = append(edges, weightedEdge{"H", "X" + strconv.Itoa(i), 5})
	}
	edges = append(edges, weightedEdge{"H", "G", 4}, weightedEdge{"X0", "S", 0})
	g := directedWeighted(edges...)

	p, cost, err := Beam(g, "S", "G", 1, nil)
	if err != nil {
		t.Fatalf("Beam() error = %v", err)
	}
	if want := []string{"S", "A", "H", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 6 {
		t.Errorf("path = %v, cost = %v, want %v and 6", pathLabels(p), cost, want)
	}
}

func TestBeam_StopsAtTheFirstStepThatKeepsTheGoal(t *testing.T) {
	// The direct edge reaches G in one step, before the cheaper two-step path.
	g := directedWeighted(
		weightedEdge{"S", "G", 9}, weightedEdge{"S", "A", 1}, weightedEdge{"A", "G", 1},
	)
	p, cost, err := Beam(g, "S", "G", 2, nil)
	if err != nil {
		t.Fatalf("Beam() error = %v", err)
	}
	if want := []string{"S", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 9 {
		t.Errorf("path = %v, cost = %v, want %v and 9", pathLabels(p), cost, want)
	}
}

func TestBeam_ReturnsTheBestGoalPathOfTheStep(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"S", "B", 2},
		weightedEdge{"A", "G", 9}, weightedEdge{"B", "G", 1},
	)
	p, cost, err := Beam(g, "S", "G", 5, nil)
	if err != nil {
		t.Fatalf("Beam() error = %v", err)
	}
	if want := []string{"S", "B", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 3 {
		t.Errorf("path = %v, cost = %v, want %v and 3", pathLabels(p), cost, want)
	}
}

func TestBeam_ScoreRanksButTheCostIsTheWeights(t *testing.T) {
	g := beamDiamond()
	// An estimate of the cost left after each edge guides width 1 to B.
	left := map[string]float64{"A": 10, "B": 1, "G": 0}
	var calls []float64
	p, cost, err := Beam(g, "S", "G", 1, func(e *gograph.Edge[string], costSoFar float64) float64 {
		calls = append(calls, costSoFar)
		return costSoFar + e.Weight() + left[e.Destination().Label()]
	})
	if err != nil {
		t.Fatalf("Beam() error = %v", err)
	}
	if want := []string{"S", "B", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 3 {
		t.Errorf("path = %v, cost = %v, want %v and 3", pathLabels(p), cost, want)
	}
	if want := []float64{0, 0, 2}; !reflect.DeepEqual(calls, want) {
		t.Errorf("score saw costs so far of %v, want %v", calls, want)
	}
}

func TestBeam_SkippedEdges(t *testing.T) {
	tests := []struct {
		name   string
		weight float64
		score  func(*gograph.Edge[string], float64) float64
	}{
		{"NaN weight", math.NaN(), nil},
		{"+Inf weight", math.Inf(1), nil},
		{"NaN score", 1, func(*gograph.Edge[string], float64) float64 { return math.NaN() }},
		{"+Inf score", 1, func(*gograph.Edge[string], float64) float64 { return math.Inf(1) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := directedWeighted(weightedEdge{"S", "G", tt.weight})
			if _, _, err := Beam(g, "S", "G", 3, tt.score); !errors.Is(err, ErrUnreachable) {
				t.Errorf("error = %v, want %v", err, ErrUnreachable)
			}
		})
	}
}

func TestBeam_OverflowingCostIsSkipped(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", math.MaxFloat64}, weightedEdge{"A", "G", math.MaxFloat64},
	)
	if _, _, err := Beam(g, "S", "G", 3, nil); !errors.Is(err, ErrUnreachable) {
		t.Errorf("error = %v, want %v", err, ErrUnreachable)
	}
}

func TestBeam_NegativeWeights(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 2}, weightedEdge{"A", "G", -5},
		weightedEdge{"S", "B", 1}, weightedEdge{"B", "G", 1},
	)
	p, cost, err := Beam(g, "S", "G", 2, nil)
	if err != nil {
		t.Fatalf("Beam() error = %v", err)
	}
	if want := []string{"S", "A", "G"}; !reflect.DeepEqual(pathLabels(p), want) || cost != -3 {
		t.Errorf("path = %v, cost = %v, want %v and -3", pathLabels(p), cost, want)
	}
}

func TestBeam_Undirected(t *testing.T) {
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
		p, cost, err := Beam(g, tt.start, tt.goal, 1, nil)
		if err != nil {
			t.Fatalf("%s to %s: error = %v", tt.start, tt.goal, err)
		}
		if !reflect.DeepEqual(pathLabels(p), tt.want) || cost != 5 {
			t.Errorf("%s to %s: path = %v, cost = %v, want %v and 5", tt.start, tt.goal, pathLabels(p), cost, tt.want)
		}
	}
}

func TestBeam_UnweightedGraphCostsZero(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	a, b := g.AddVertexByLabel("A"), g.AddVertexByLabel("B")
	_, _ = g.AddEdge(a, b)

	p, cost, err := Beam(g, "A", "B", 1, nil)
	if err != nil {
		t.Fatalf("Beam() error = %v", err)
	}
	if want := []string{"A", "B"}; !reflect.DeepEqual(pathLabels(p), want) || cost != 0 {
		t.Errorf("path = %v, cost = %v, want %v and 0", pathLabels(p), cost, want)
	}
}

// beamNoEdges is a graph that lists neighbors it has no edge for.
type beamNoEdges struct {
	gograph.Graph[string]
}

func (beamNoEdges) GetEdge(_, _ *gograph.Vertex[string]) *gograph.Edge[string] { return nil }

func TestBeam_NeighborWithoutAnEdgeIsSkipped(t *testing.T) {
	g := beamNoEdges{directedWeighted(weightedEdge{"A", "B", 1})}
	if _, _, err := Beam[string](g, "A", "B", 2, nil); !errors.Is(err, ErrUnreachable) {
		t.Errorf("error = %v, want %v", err, ErrUnreachable)
	}
}

// referenceBeam is the plain way to write Beam: make every candidate of a step,
// sort them all, and keep the first width.
func referenceBeam(g gograph.Graph[string], start, goal string, width int) ([]string, float64, bool) {
	type partial struct {
		labels []string
		cost   float64
	}
	frontier := []partial{{labels: []string{start}}}
	for len(frontier) > 0 {
		var next []partial
		for _, p := range frontier {
			last := g.GetVertexByID(p.labels[len(p.labels)-1])
			for v := range last.AllNeighbors() {
				if slices.Contains(p.labels, v.Label()) {
					continue
				}
				labels := append(slices.Clone(p.labels), v.Label())
				next = append(next, partial{labels: labels, cost: p.cost + g.GetEdge(last, v).Weight()})
			}
		}
		slices.SortStableFunc(next, func(a, b partial) int {
			switch {
			case a.cost < b.cost:
				return -1
			case a.cost > b.cost:
				return 1
			}
			return 0
		})
		frontier = next[:min(width, len(next))]
		for _, p := range frontier {
			if p.labels[len(p.labels)-1] == goal {
				return p.labels, p.cost, true
			}
		}
	}
	return nil, 0, false
}

// TestBeam_MatchesTheReference compares random graphs with referenceBeam. Small
// integer weights, zeros included, make many ties.
func TestBeam_MatchesTheReference(t *testing.T) {
	rng := rand.New(rand.NewPCG(20, 0)) //nolint:gosec // seeded so failures are reproducible

	for trial := range 3000 {
		var g gograph.Graph[string]
		if trial%2 == 0 {
			g = gograph.New[string](gograph.Directed(), gograph.Weighted())
		} else {
			g = gograph.New[string](gograph.Weighted())
		}
		n := 2 + rng.IntN(11)
		labels := make([]string, n)
		for i := range labels {
			labels[i] = strconv.Itoa(i)
			g.AddVertexByLabel(labels[i])
		}
		for range rng.IntN(3*n) + trial%3*n*n/2 {
			from, to := g.GetVertexByID(labels[rng.IntN(n)]), g.GetVertexByID(labels[rng.IntN(n)])
			_, _ = g.AddEdge(from, to, gograph.WithEdgeWeight(float64(rng.IntN(4))))
		}

		start, goal := labels[0], labels[n-1]
		for _, width := range []int{1, 2, 3, 1 + rng.IntN(20), 1 << 20} {
			want, wantCost, ok := referenceBeam(g, start, goal, width)
			p, cost, err := Beam(g, start, goal, width, nil)
			if !ok {
				if !errors.Is(err, ErrUnreachable) {
					t.Fatalf("trial %d, width %d: error = %v, want %v", trial, width, err, ErrUnreachable)
				}
				continue
			}
			if err != nil {
				t.Fatalf("trial %d, width %d: error = %v", trial, width, err)
			}
			if !reflect.DeepEqual(pathLabels(p), want) || cost != wantCost {
				t.Fatalf("trial %d, width %d: path = %v, cost = %v, want %v and %v",
					trial, width, pathLabels(p), cost, want, wantCost)
			}
			assertPath(t, g, p, start, goal, cost)
		}
	}
}

func ExampleBeam() {
	g := gograph.New[string](gograph.Weighted())
	home := g.AddVertexByLabel("home")
	park := g.AddVertexByLabel("park")
	shop := g.AddVertexByLabel("shop")
	work := g.AddVertexByLabel("work")

	_, _ = g.AddEdge(home, shop, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(shop, work, gograph.WithEdgeWeight(5))
	_, _ = g.AddEdge(home, park, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(park, work, gograph.WithEdgeWeight(2))

	for _, width := range []int{1, 2} {
		p, cost, err := Beam(g, "home", "work", width, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(width, pathLabels(p), cost)
	}
	// Output:
	// 1 [home shop work] 6
	// 2 [home park work] 4
}

// beamGrid returns a directed n by n grid with edges going right and down, and
// random weights from 1 to 9. The vertex at column x and row y has the label
// x*n+y.
func beamGrid(n int) gograph.Graph[int] {
	rng := rand.New(rand.NewPCG(4, 2)) //nolint:gosec // seeded so runs are comparable
	g := gograph.New[int](gograph.Directed(), gograph.Weighted())
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

func BenchmarkBeam(b *testing.B) {
	const n = 200
	g := beamGrid(n)
	for _, width := range []int{1, 10, 100} {
		b.Run("width="+strconv.Itoa(width), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, _, err := Beam(g, 0, n*n-1, width, nil); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
