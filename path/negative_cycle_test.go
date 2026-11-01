package path

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

type weightedEdge struct {
	from, to string
	weight   float64
}

func negativeCycleGraph(t *testing.T, edges []weightedEdge) gograph.Graph[string] {
	t.Helper()
	g := gograph.New[string](gograph.Weighted(), gograph.Directed())
	vertex := func(label string) *gograph.Vertex[string] {
		if v := g.GetVertexByID(label); v != nil {
			return v
		}
		return g.AddVertexByLabel(label)
	}
	for _, e := range edges {
		if _, err := g.AddEdge(vertex(e.from), vertex(e.to), gograph.WithEdgeWeight(e.weight)); err != nil {
			t.Fatalf("AddEdge(%s, %s): %v", e.from, e.to, err)
		}
	}
	return g
}

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := negativeCycleGraph(t, tt.edges)

			cycle, err := NegativeCycle(g)
			if err != nil {
				t.Fatalf("NegativeCycle() error = %v", err)
			}
			if len(cycle) != tt.want {
				t.Fatalf("cycle = %v, want %d vertices", labels(cycle), tt.want)
			}
			if !math.IsInf(tt.edges[0].weight, -1) {
				assertNegativeCycle(t, g, cycle)
			}
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
		{name: "NaN weight", edges: []weightedEdge{{"A", "B", math.NaN()}, {"B", "A", 1}}},
		{name: "no edges", edges: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cycle, err := NegativeCycle(negativeCycleGraph(t, tt.edges))
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
	g := negativeCycleGraph(t, edges)

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

	fmt.Println(len(cycle))
	// Output: 3
}
