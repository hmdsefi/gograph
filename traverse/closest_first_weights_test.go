package traverse

import (
	"math"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestClosestFirstIterator_NonFiniteWeights(t *testing.T) {
	type edge struct {
		from, to string
		weight   float64
	}
	tests := []struct {
		name  string
		edges []edge
		want  []string
	}{
		{
			// N is skipped, and the NaN doesn't change the order of the others
			name: "NaN weight",
			edges: []edge{
				{"A", "B", 1}, {"A", "N", math.NaN()}, {"A", "C", 5},
				{"A", "E", 6}, {"B", "D", 0.5},
			},
			want: []string{"A", "B", "D", "C", "E"},
		},
		{
			name:  "NaN weight on the only edge",
			edges: []edge{{"A", "N", math.NaN()}, {"N", "B", 1}},
			want:  []string{"A"},
		},
		{
			// X is infinitely far, so it comes after every finite distance
			name:  "positive infinite weight",
			edges: []edge{{"A", "X", math.Inf(1)}, {"A", "B", 2}, {"B", "C", 3}},
			want:  []string{"A", "B", "C", "X"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gograph.New[string](gograph.Directed(), gograph.Weighted())
			for _, e := range tt.edges {
				_, _ = g.AddEdge(gograph.NewVertex(e.from), gograph.NewVertex(e.to), gograph.WithEdgeWeight(e.weight))
			}

			it, err := NewClosestFirstIterator(g, "A")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var got []string
			for it.HasNext() {
				got = append(got, it.Next().Label())
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
