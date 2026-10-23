package path

import (
	"errors"
	"math"
	"testing"

	"github.com/hmdsefi/gograph"
)

type weightedEdge struct {
	from, to string
	weight   float64
}

// directedWeighted returns a directed, weighted graph with the given edges.
// The vertices are added in the order they first appear.
func directedWeighted(edges ...weightedEdge) gograph.Graph[string] {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	for _, e := range edges {
		_, _ = g.AddEdge(gograph.NewVertex(e.from), gograph.NewVertex(e.to), gograph.WithEdgeWeight(e.weight))
	}
	return g
}

// distanceFunc returns the distance from the start vertex to v, and false
// when the function reports v as unreachable.
type distanceFunc func(v string) (float64, bool)

// shortestPaths runs every single-source shortest-path function on the
// same graph. Each one reports an unreachable vertex in its own way.
var shortestPaths = []struct {
	name string
	run  func(g gograph.Graph[string], start string) (distanceFunc, error)
}{
	{
		name: "FloydWarshall",
		run: func(g gograph.Graph[string], start string) (distanceFunc, error) {
			dist, err := FloydWarshall(g)
			if err != nil {
				return nil, err
			}
			return func(v string) (float64, bool) {
				return dist[start][v], !math.IsInf(dist[start][v], 1)
			}, nil
		},
	},
	{
		name: "BellmanFord",
		run: func(g gograph.Graph[string], start string) (distanceFunc, error) {
			dist, err := BellmanFord(g, start)
			if err != nil {
				return nil, err
			}
			return func(v string) (float64, bool) {
				return dist[v], !math.IsInf(dist[v], 1)
			}, nil
		},
	},
	{
		name: "Dijkstra",
		run: func(g gograph.Graph[string], start string) (distanceFunc, error) {
			dist := Dijkstra(g, start)
			return func(v string) (float64, bool) {
				return dist[v], dist[v] != math.MaxFloat64
			}, nil
		},
	},
	{
		name: "DijkstraSimple",
		run: func(g gograph.Graph[string], start string) (distanceFunc, error) {
			dist := DijkstraSimple(g, start)
			return func(v string) (float64, bool) {
				return dist[v], dist[v] != math.MaxFloat64
			}, nil
		},
	},
	{
		name: "DijkstraMultiSource",
		run: func(g gograph.Graph[string], start string) (distanceFunc, error) {
			nearest, err := DijkstraMultiSource(g, start)
			if err != nil {
				return nil, err
			}
			return nearest.Distance, nil
		},
	},
}

func TestShortestPaths_NaNWeight(t *testing.T) {
	// the NaN edge to b is skipped, so b is reached through c, and d,
	// which only a NaN edge leads to, is unreachable
	g := directedWeighted(
		weightedEdge{"a", "b", math.NaN()},
		weightedEdge{"a", "c", 1},
		weightedEdge{"c", "b", 1},
		weightedEdge{"a", "d", math.NaN()},
	)

	for _, tt := range shortestPaths {
		t.Run(tt.name, func(t *testing.T) {
			distance, err := tt.run(g, "a")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d, ok := distance("b"); !ok || d != 2 {
				t.Errorf("b: expected distance 2, got %v (reachable %v)", d, ok)
			}
			if d, ok := distance("d"); ok {
				t.Errorf("d: expected it to be unreachable, got distance %v", d)
			}
		})
	}
}

func TestShortestPaths_PositiveInfinityWeight(t *testing.T) {
	// a path through a +Inf edge is as long as no path
	g := directedWeighted(
		weightedEdge{"a", "b", math.Inf(1)},
		weightedEdge{"a", "c", 1},
		weightedEdge{"b", "d", 1},
	)

	for _, tt := range shortestPaths {
		t.Run(tt.name, func(t *testing.T) {
			distance, err := tt.run(g, "a")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d, ok := distance("c"); !ok || d != 1 {
				t.Errorf("c: expected distance 1, got %v (reachable %v)", d, ok)
			}
			for _, v := range []string{"b", "d"} {
				if d, ok := distance(v); ok {
					t.Errorf("%s: expected it to be unreachable, got distance %v", v, d)
				}
			}
		})
	}
}

func TestShortestPaths_NegativeInfinityDistance(t *testing.T) {
	tests := []struct {
		name  string
		edges []weightedEdge
	}{
		{
			name:  "negative infinite weight in a cycle",
			edges: []weightedEdge{{"a", "b", math.Inf(-1)}, {"b", "a", 1}},
		},
		{
			name:  "negative infinite weight",
			edges: []weightedEdge{{"a", "b", math.Inf(-1)}, {"b", "c", 1}},
		},
		{
			name:  "negative cycle that sums below the float64 range",
			edges: []weightedEdge{{"a", "b", -1e308}, {"b", "a", -1e308}, {"c", "c", 0}},
		},
		{
			name:  "path that sums below the float64 range",
			edges: []weightedEdge{{"a", "b", -1e308}, {"b", "c", -1e308}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := directedWeighted(tt.edges...)
			if dist, err := BellmanFord(g, "a"); !errors.Is(err, ErrNegativeWeightCycle) {
				t.Errorf("BellmanFord: expected ErrNegativeWeightCycle, got %v and %v", dist, err)
			}
			if dist, err := FloydWarshall(g); !errors.Is(err, ErrNegativeWeightCycle) {
				t.Errorf("FloydWarshall: expected ErrNegativeWeightCycle, got %v and %v", dist, err)
			}
			if _, err := DijkstraMultiSource(g, "a"); !errors.Is(err, ErrNegativeWeight) {
				t.Errorf("DijkstraMultiSource: expected ErrNegativeWeight, got %v", err)
			}
		})
	}
}

func TestKCenter_NonFiniteWeights(t *testing.T) {
	tests := []struct {
		name   string
		edges  []weightedEdge
		radius float64
	}{
		{
			// b is reached through c, past the NaN edge
			name:   "NaN weight",
			edges:  []weightedEdge{{"a", "b", math.NaN()}, {"a", "c", 1}, {"c", "b", 1}},
			radius: 2,
		},
		{
			// b is only behind the +Inf edge, so a can't cover it
			name:   "positive infinite weight",
			edges:  []weightedEdge{{"a", "b", math.Inf(1)}, {"a", "c", 1}},
			radius: math.Inf(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			centers, radius, err := KCenter(directedWeighted(tt.edges...), 1, "a")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(centers) != 1 || centers[0].Label() != "a" {
				t.Fatalf("expected the fixed center a, got %v", centers)
			}
			if radius != tt.radius {
				t.Fatalf("expected radius %v, got %v", tt.radius, radius)
			}
		})
	}
}
