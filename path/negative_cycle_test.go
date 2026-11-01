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

// assertNegativeCycle walks the cycle through the graph's edges, including the
// one that closes it, and checks that the weights sum to less than zero.
func assertNegativeCycle(t *testing.T, g gograph.Graph[string], cycle []*gograph.Vertex[string]) {
	t.Helper()
	if len(cycle) == 0 {
		t.Fatal("expected a cycle, got none")
	}
	seen := make(map[string]bool, len(cycle))
	var sum float64
	for i, v := range cycle {
		if g.GetVertexByID(v.Label()) != v {
			t.Errorf("vertex %s is not the graph's own pointer", v.Label())
		}
		if seen[v.Label()] {
			t.Errorf("vertex %s appears twice in %v", v.Label(), labels(cycle))
		}
		seen[v.Label()] = true

		next := cycle[(i+1)%len(cycle)]
		edge := g.GetEdge(v, next)
		if edge == nil {
			t.Fatalf("no edge from %s to %s in %v", v.Label(), next.Label(), labels(cycle))
		}
		sum += edge.Weight()
	}
	if sum >= 0 {
		t.Errorf("cycle %v sums to %v, want less than 0", labels(cycle), sum)
	}
}

func labels(vertices []*gograph.Vertex[string]) []string {
	out := make([]string, len(vertices))
	for i, v := range vertices {
		out[i] = v.Label()
	}
	return out
}

func TestNegativeCycle(t *testing.T) {
	tests := []struct {
		name  string
		edges []weightedEdge
		want  int
	}{
		{
			name:  "triangle",
			edges: []weightedEdge{{"A", "B", 1}, {"B", "C", -3}, {"C", "A", 1}},
			want:  3,
		},
		{
			name:  "self-loop",
			edges: []weightedEdge{{"A", "B", 1}, {"B", "B", -1}},
			want:  1,
		},
		{
			name:  "infinite self-loop",
			edges: []weightedEdge{{"A", "A", math.Inf(-1)}},
			want:  1,
		},
		{
			name: "not reachable from the first vertex",
			edges: []weightedEdge{
				{"A", "B", 1},
				{"X", "Y", 1}, {"Y", "Z", -3}, {"Z", "X", 1},
			},
			want: 3,
		},
		{
			name: "tail leading into the cycle",
			edges: []weightedEdge{
				{"A", "B", 2}, {"B", "C", 2},
				{"C", "D", 1}, {"D", "C", -2},
			},
			want: 2,
		},
		{
			name: "edges in the order that relaxes slowest",
			edges: []weightedEdge{
				{"D", "A", 1}, {"C", "D", 1}, {"B", "C", 1}, {"A", "B", -4},
			},
			want: 4,
		},
		{
			name:  "infinite tail leading into the cycle",
			edges: []weightedEdge{{"X", "A", math.Inf(-1)}, {"A", "B", 1}, {"B", "A", -3}},
			want:  2,
		},
		{
			name:  "infinite tail with the edges in the other order",
			edges: []weightedEdge{{"B", "A", -2}, {"A", "B", 1}, {"S", "A", math.Inf(-1)}},
			want:  2,
		},
		{
			name:  "tail large enough to swallow the cycle's weights",
			edges: []weightedEdge{{"X", "A", -1e20}, {"A", "B", 1}, {"B", "A", -3}},
			want:  2,
		},
		{
			name:  "cycle through an infinite weight that another one reaches first",
			edges: []weightedEdge{{"S", "B", math.Inf(-1)}, {"A", "B", math.Inf(-1)}, {"B", "A", 6}},
			want:  2,
		},
		{
			name: "tail that overflows",
			edges: []weightedEdge{
				{"S", "T", -math.MaxFloat64}, {"T", "A", -math.MaxFloat64},
				{"A", "B", 1}, {"B", "A", -2},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := directedWeighted(tt.edges...)

			cycle, err := NegativeCycle(g)
			if err != nil {
				t.Fatalf("NegativeCycle() error = %v", err)
			}
			if len(cycle) != tt.want {
				t.Fatalf("cycle = %v, want %d vertices", labels(cycle), tt.want)
			}
			assertNegativeCycle(t, g, cycle)
		})
	}
}

func TestNegativeCycle_NoCycle(t *testing.T) {
	tests := []struct {
		name  string
		edges []weightedEdge
	}{
		{name: "positive cycle", edges: []weightedEdge{{"A", "B", 1}, {"B", "C", 1}, {"C", "A", 1}}},
		{name: "cycle summing to zero", edges: []weightedEdge{{"A", "B", 2}, {"B", "A", -2}}},
		{name: "negative edge without a cycle", edges: []weightedEdge{{"A", "B", -5}, {"B", "C", -5}}},
		{
			name:  "cycle summing to exactly zero with rounded weights",
			edges: []weightedEdge{{"A", "B", 0.1}, {"B", "C", 0.1}, {"C", "D", 0.2}, {"D", "A", -0.4}},
		},
		{name: "NaN weight", edges: []weightedEdge{{"A", "B", math.NaN()}, {"B", "A", 1}}},
		{name: "cycle through a +Inf weight", edges: []weightedEdge{{"A", "B", math.Inf(1)}, {"B", "A", math.Inf(-1)}}},
		{name: "no edges", edges: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cycle, err := NegativeCycle(directedWeighted(tt.edges...))
			if err != nil {
				t.Fatalf("NegativeCycle() error = %v", err)
			}
			if cycle != nil {
				t.Errorf("cycle = %v, want nil", labels(cycle))
			}
		})
	}
}

func TestNegativeCycle_GraphRequirements(t *testing.T) {
	tests := []struct {
		name string
		g    gograph.Graph[string]
		want error
	}{
		{name: "undirected", g: gograph.New[string](gograph.Weighted()), want: ErrNotDirected},
		{name: "unweighted", g: gograph.New[string](gograph.Directed()), want: ErrNotWeighted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cycle, err := NegativeCycle(tt.g)
			if !errors.Is(err, tt.want) {
				t.Errorf("NegativeCycle() error = %v, want %v", err, tt.want)
			}
			if cycle != nil {
				t.Errorf("cycle = %v, want nil", labels(cycle))
			}
		})
	}
}

func TestNegativeCycle_SameCycleOnEveryRun(t *testing.T) {
	edges := []weightedEdge{
		{"A", "B", -1}, {"B", "A", -1},
		{"C", "D", -1}, {"D", "C", -1},
		{"E", "F", -1}, {"F", "G", -1}, {"G", "E", -1},
	}
	g := directedWeighted(edges...)

	first, err := NegativeCycle(g)
	if err != nil {
		t.Fatalf("NegativeCycle() error = %v", err)
	}
	for range 20 {
		next, err := NegativeCycle(g)
		if err != nil {
			t.Fatalf("NegativeCycle() error = %v", err)
		}
		if !reflect.DeepEqual(labels(next), labels(first)) {
			t.Fatalf("cycle changed from %v to %v", labels(first), labels(next))
		}
	}
}

// exactNegativeCycle runs Floyd-Warshall on int64 weights, so nothing is
// rounded. weights[i][j] is the edge from i to j, and absent means no edge.
func exactNegativeCycle(weights [][]int64, absent int64) bool {
	dist := make([][]int64, len(weights))
	for i := range weights {
		dist[i] = slices.Clone(weights[i])
	}
	for k := range dist {
		for i := range dist {
			for j := range dist {
				if dist[i][k] != absent && dist[k][j] != absent &&
					(dist[i][j] == absent || dist[i][k]+dist[k][j] < dist[i][j]) {
					dist[i][j] = dist[i][k] + dist[k][j]
				}
			}
			if dist[i][i] != absent && dist[i][i] < 0 {
				return true
			}
		}
	}
	return false
}

func TestNegativeCycle_MatchesExactCheck(t *testing.T) {
	const (
		absent      = math.MaxInt64
		negativeInf = -1 << 40
	)
	rng := rand.New(rand.NewPCG(1, 2))

	for range 5000 {
		n := 1 + rng.IntN(7)
		weights := make([][]int64, n)
		for i := range weights {
			weights[i] = make([]int64, n)
			for j := range weights[i] {
				weights[i][j] = absent
			}
		}

		g := gograph.New[string](gograph.Directed(), gograph.Weighted())
		for range rng.IntN(n*n + 1) {
			i, j := rng.IntN(n), rng.IntN(n)
			from, to := gograph.NewVertex(strconv.Itoa(i)), gograph.NewVertex(strconv.Itoa(j))
			if g.ContainsEdge(from, to) {
				continue
			}

			w := float64(rng.IntN(16) - 5)
			switch rng.IntN(24) {
			case 0:
				w = math.NaN()
			case 1:
				w = math.Inf(1)
			case 2:
				w = math.Inf(-1)
				weights[i][j] = negativeInf
			default:
				weights[i][j] = int64(w)
			}
			if _, err := g.AddEdge(from, to, gograph.WithEdgeWeight(w)); err != nil {
				t.Fatalf("AddEdge(%d, %d): %v", i, j, err)
			}
		}

		cycle, err := NegativeCycle(g)
		if err != nil {
			t.Fatalf("NegativeCycle() error = %v", err)
		}
		if want := exactNegativeCycle(weights, absent); (cycle != nil) != want {
			t.Fatalf("cycle = %v, want a cycle: %v, weights %v", labels(cycle), want, weights)
		}
		if cycle != nil {
			assertNegativeCycle(t, g, cycle)
		}
	}
}

func ExampleNegativeCycle() {
	g := gograph.New[string](gograph.Weighted(), gograph.Directed())
	usd := g.AddVertexByLabel("USD")
	eur := g.AddVertexByLabel("EUR")
	gbp := g.AddVertexByLabel("GBP")

	_, _ = g.AddEdge(usd, eur, gograph.WithEdgeWeight(0.1))
	_, _ = g.AddEdge(eur, gbp, gograph.WithEdgeWeight(0.2))
	_, _ = g.AddEdge(gbp, usd, gograph.WithEdgeWeight(-0.4))

	cycle, err := NegativeCycle(g)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, v := range cycle {
		fmt.Println(v.Label())
	}
	// Output:
	// EUR
	// GBP
	// USD
}
