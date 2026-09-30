package traverse

import (
	"math"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestRandomWalkIterator_FractionalWeight(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(0.5))

	it, err := NewRandomWalkIterator(g, "A", 3)
	if err != nil {
		t.Fatalf("Expect NewRandomWalkIterator doesn't return error, but got %s", err)
	}

	visited := make([]string, 0)
	for it.HasNext() {
		visited = append(visited, it.Next().Label())
	}

	if expected := []string{"A", "B", "A"}; !reflect.DeepEqual(visited, expected) {
		t.Errorf("Expect walk %v, but got %v", expected, visited)
	}
}

func TestRandomWalkIterator_RandomVertexWithoutEdges(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a := g.AddVertexByLabel("A")
	r := &randomWalkIterator[string]{graph: g}

	if v := r.randomVertex(nil); v != nil {
		t.Errorf("Expect nil for a nil vertex, but got %v", v.Label())
	}

	if v := r.randomVertex(a); v != nil {
		t.Errorf("Expect nil for a vertex without edges, but got %v", v.Label())
	}
}

func TestRandomWalkIterator_WeightOdds(t *testing.T) {
	const walks = 10000
	const tolerance = 0.03

	tests := []struct {
		name    string
		weights [2]float64
		shares  [2]float64
	}{
		{name: "equal fractional weights", weights: [2]float64{1.5, 1.5}, shares: [2]float64{0.5, 0.5}},
		{name: "one to three", weights: [2]float64{1, 3}, shares: [2]float64{0.25, 0.75}},
		{name: "weights below one", weights: [2]float64{0.2, 0.6}, shares: [2]float64{0.25, 0.75}},
		{name: "zero weight", weights: [2]float64{0, 2}, shares: [2]float64{0, 1}},
		{name: "negative weight", weights: [2]float64{-1, 1}, shares: [2]float64{0, 1}},
		{name: "only zero weights", weights: [2]float64{0, 0}, shares: [2]float64{0.5, 0.5}},
		{name: "only negative weights", weights: [2]float64{-1, -2}, shares: [2]float64{0.5, 0.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gograph.New[string](gograph.Directed(), gograph.Weighted())
			a := g.AddVertexByLabel("A")
			labels := [2]string{"B", "C"}
			for i, label := range labels {
				_, _ = g.AddEdge(a, g.AddVertexByLabel(label), gograph.WithEdgeWeight(tt.weights[i]))
			}

			it, err := NewRandomWalkIterator(g, "A", 2)
			if err != nil {
				t.Fatalf("Expect NewRandomWalkIterator doesn't return error, but got %s", err)
			}

			counts := make(map[string]int)
			for range walks {
				it.Reset()
				it.Next()
				v := it.Next()
				if v == nil {
					t.Fatal("Expect a second vertex, but got nil")
				}
				counts[v.Label()]++
			}

			for i, label := range labels {
				got := float64(counts[label]) / walks
				want := tt.shares[i]
				if math.Abs(got-want) > tolerance || (want == 0 && got != 0) {
					t.Errorf("Expect %s to be picked in %.2f of the walks, but got %.3f", label, want, got)
				}
			}
		})
	}
}
