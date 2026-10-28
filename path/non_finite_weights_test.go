package path

import (
	"errors"
	"maps"
	"math"
	"math/rand/v2"
	"reflect"
	"strconv"
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

func TestShortestPaths_NegativeCycleOrOverflow(t *testing.T) {
	negInf := math.Inf(-1)
	tests := []struct {
		name  string
		edges []weightedEdge
	}{
		{
			name:  "negative infinite weight in a cycle",
			edges: []weightedEdge{{"a", "b", negInf}, {"b", "a", 1}},
		},
		{
			name:  "negative infinite self-loop",
			edges: []weightedEdge{{"a", "b", 1}, {"b", "b", negInf}},
		},
		{
			name:  "cycle of negative infinite weights behind one",
			edges: []weightedEdge{{"a", "b", negInf}, {"b", "c", negInf}, {"c", "b", negInf}},
		},
		{
			name:  "negative cycle that a negative infinite weight leads into",
			edges: []weightedEdge{{"a", "b", negInf}, {"b", "c", -1}, {"c", "b", -1}},
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

func TestShortestPaths_NegativeInfinityPath(t *testing.T) {
	inf := math.Inf(1)
	tests := []struct {
		name  string
		edges []weightedEdge
		want  map[string]float64 // distances from a
	}{
		{
			name:  "single negative infinite weight",
			edges: []weightedEdge{{"a", "b", -inf}},
			want:  map[string]float64{"a": 0, "b": -inf},
		},
		{
			name:  "path through a negative infinite weight",
			edges: []weightedEdge{{"a", "b", -inf}, {"b", "c", 1}},
			want:  map[string]float64{"a": 0, "b": -inf, "c": -inf},
		},
		{
			name:  "two negative infinite weights in a row",
			edges: []weightedEdge{{"a", "b", -inf}, {"b", "c", -inf}},
			want:  map[string]float64{"a": 0, "b": -inf, "c": -inf},
		},
		{
			name:  "positive cycle that a negative infinite weight leads into",
			edges: []weightedEdge{{"a", "b", -inf}, {"b", "c", 1}, {"c", "b", 1}},
			want:  map[string]float64{"a": 0, "b": -inf, "c": -inf},
		},
		{
			name:  "shorter finite path to the same vertex",
			edges: []weightedEdge{{"a", "c", -5}, {"a", "b", -inf}, {"b", "c", 1}, {"c", "d", 1}},
			want:  map[string]float64{"a": 0, "b": -inf, "c": -inf, "d": -inf},
		},
		{
			name:  "positive infinite weight after it",
			edges: []weightedEdge{{"a", "b", -inf}, {"b", "c", inf}},
			want:  map[string]float64{"a": 0, "b": -inf, "c": inf},
		},
		{
			name:  "positive infinite weight back to the start",
			edges: []weightedEdge{{"a", "b", -inf}, {"b", "a", inf}},
			want:  map[string]float64{"a": 0, "b": -inf},
		},
		{
			name:  "NaN weight back to the start",
			edges: []weightedEdge{{"a", "b", -inf}, {"b", "a", math.NaN()}},
			want:  map[string]float64{"a": 0, "b": -inf},
		},
		{
			name:  "negative infinite weight that the start doesn't reach",
			edges: []weightedEdge{{"a", "b", 2}, {"c", "b", -inf}, {"b", "d", 1}},
			want:  map[string]float64{"a": 0, "b": 2, "c": inf, "d": 3},
		},
		{
			name:  "two negative infinite weights into one vertex",
			edges: []weightedEdge{{"a", "b", 1}, {"a", "c", 1}, {"b", "d", -inf}, {"c", "d", -inf}, {"d", "e", 1}},
			want:  map[string]float64{"a": 0, "b": 1, "c": 1, "d": -inf, "e": -inf},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := directedWeighted(tt.edges...)
			dist, err := BellmanFord(g, "a")
			if err != nil {
				t.Fatalf("BellmanFord: unexpected error: %v", err)
			}
			if !reflect.DeepEqual(dist, tt.want) {
				t.Errorf("BellmanFord: expected %v, got %v", tt.want, dist)
			}

			all, err := FloydWarshall(g)
			if err != nil {
				t.Fatalf("FloydWarshall: unexpected error: %v", err)
			}
			if !reflect.DeepEqual(all["a"], tt.want) {
				t.Errorf("FloydWarshall: expected %v from a, got %v", tt.want, all["a"])
			}
		})
	}
}

func TestShortestPaths_UnreachableNegativeInfinityCycle(t *testing.T) {
	// a negative cycle counts for BellmanFord only when the start reaches it,
	// and for FloydWarshall anywhere in the graph
	g := directedWeighted(
		weightedEdge{"a", "b", 1},
		weightedEdge{"c", "d", math.Inf(-1)},
		weightedEdge{"d", "c", 1},
	)

	dist, err := BellmanFord(g, "a")
	if err != nil {
		t.Fatalf("BellmanFord: unexpected error: %v", err)
	}
	want := map[string]float64{"a": 0, "b": 1, "c": math.Inf(1), "d": math.Inf(1)}
	if !reflect.DeepEqual(dist, want) {
		t.Errorf("BellmanFord: expected %v, got %v", want, dist)
	}

	if _, err = BellmanFord(g, "d"); !errors.Is(err, ErrNegativeWeightCycle) {
		t.Errorf("BellmanFord from d: expected ErrNegativeWeightCycle, got %v", err)
	}
	if _, err = FloydWarshall(g); !errors.Is(err, ErrNegativeWeightCycle) {
		t.Errorf("FloydWarshall: expected ErrNegativeWeightCycle, got %v", err)
	}
}

// referenceBellmanFord follows the rules in the BellmanFord doc one by one,
// without the shortcuts BellmanFord takes. It reports true for a negative
// cycle that start reaches.
func referenceBellmanFord(g gograph.Graph[string], start string) (map[string]float64, bool) {
	usable := func(w float64) bool { return !math.IsNaN(w) && !math.IsInf(w, 1) }
	reaches := func(from, to string) bool {
		seen := map[string]bool{from: true}
		queue := []string{from}
		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			if v == to {
				return true
			}
			for _, e := range g.AllEdges() {
				if e.Source().Label() == v && usable(e.Weight()) && !seen[e.Destination().Label()] {
					seen[e.Destination().Label()] = true
					queue = append(queue, e.Destination().Label())
				}
			}
		}
		return false
	}
	weight := func(e *gograph.Edge[string]) float64 {
		if math.IsInf(e.Weight(), -1) {
			return 0
		}
		return e.Weight()
	}

	inf := math.Inf(1)
	dist := make(map[string]float64)
	for _, v := range g.GetAllVertices() {
		dist[v.Label()] = inf
	}
	dist[start] = 0
	for range g.Order() {
		for _, e := range g.AllEdges() {
			if d := dist[e.Source().Label()]; d != inf && d+weight(e) < dist[e.Destination().Label()] {
				dist[e.Destination().Label()] = d + weight(e)
			}
		}
	}
	for _, e := range g.AllEdges() {
		if d := dist[e.Source().Label()]; d != inf && d+weight(e) < dist[e.Destination().Label()] {
			return nil, true
		}
	}

	out := maps.Clone(dist)
	for _, e := range g.AllEdges() {
		if !math.IsInf(e.Weight(), -1) || dist[e.Source().Label()] == inf {
			continue
		}
		if reaches(e.Destination().Label(), e.Source().Label()) {
			return nil, true
		}
		for _, v := range g.GetAllVertices() {
			if reaches(e.Destination().Label(), v.Label()) {
				out[v.Label()] = math.Inf(-1)
			}
		}
	}

	return out, false
}

func TestShortestPaths_MatchReference(t *testing.T) {
	// integer weights add up exactly, so every correct order of additions
	// gives the same distances
	weights := []float64{-3, -1, 0, 1, 2, 5, math.Inf(-1), math.Inf(1), math.NaN()}
	for seed := range uint64(500) {
		rng := rand.New(rand.NewPCG(seed, 291)) //nolint:gosec // seeded so failures are reproducible
		n := 1 + rng.IntN(7)
		g := gograph.New[string](gograph.Directed(), gograph.Weighted())
		for i := range n {
			g.AddVertexByLabel(strconv.Itoa(i))
		}
		for range rng.IntN(2*n*n/3 + 1) {
			from, to := g.GetVertexByID(strconv.Itoa(rng.IntN(n))), g.GetVertexByID(strconv.Itoa(rng.IntN(n)))
			_, _ = g.AddEdge(from, to, gograph.WithEdgeWeight(weights[rng.IntN(len(weights))]))
		}

		all, allErr := FloydWarshall(g)
		anyCycle := false
		for _, v := range g.GetAllVertices() {
			want, cycle := referenceBellmanFord(g, v.Label())
			anyCycle = anyCycle || cycle

			got, err := BellmanFord(g, v.Label())
			if cycle != errors.Is(err, ErrNegativeWeightCycle) || !reflect.DeepEqual(got, want) {
				t.Fatalf("seed %d, BellmanFord from %s: expected %v (cycle %v), got %v and %v\nedges: %v",
					seed, v.Label(), want, cycle, got, err, g.AllEdges())
			}
			if allErr == nil && !reflect.DeepEqual(all[v.Label()], want) {
				t.Fatalf("seed %d, FloydWarshall from %s: expected %v, got %v\nedges: %v",
					seed, v.Label(), want, all[v.Label()], g.AllEdges())
			}
		}

		if anyCycle != errors.Is(allErr, ErrNegativeWeightCycle) {
			t.Fatalf("seed %d, FloydWarshall: expected a negative cycle %v, got %v\nedges: %v",
				seed, anyCycle, allErr, g.AllEdges())
		}
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
