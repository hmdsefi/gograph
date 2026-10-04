package dag

import (
	"errors"
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

type wEdge struct {
	from, to string
	weight   float64
}

func buildWeighted(order []string, vw map[string]float64, edges ...wEdge) gograph.Graph[string] {
	g := gograph.New[string](gograph.Directed())
	for _, label := range order {
		g.AddVertexByLabel(label, gograph.WithVertexWeight(vw[label]))
	}
	for _, e := range edges {
		_, _ = g.AddEdge(g.GetVertexByID(e.from), g.GetVertexByID(e.to), gograph.WithEdgeWeight(e.weight))
	}
	return g
}

func assertCriticalPath(t *testing.T, g gograph.Graph[string], wantPath []string, wantCost float64) {
	t.Helper()
	path, cost, err := CriticalPath(g)
	if err != nil {
		t.Fatalf("CriticalPath: unexpected error: %v", err)
	}
	if got := labels(path); !reflect.DeepEqual(got, wantPath) {
		t.Fatalf("CriticalPath = %v, want %v", got, wantPath)
	}
	if cost != wantCost {
		t.Fatalf("CriticalPath cost = %v, want %v", cost, wantCost)
	}
}

func TestCriticalPath_Chain(t *testing.T) {
	g := buildWeighted([]string{"a", "b", "c"},
		map[string]float64{"a": 1, "b": 2, "c": 3},
		wEdge{"a", "b", 0}, wEdge{"b", "c", 0})
	assertCriticalPath(t, g, []string{"a", "b", "c"}, 6)
}

func TestCriticalPath_DiamondUnequalBranches(t *testing.T) {
	g := buildWeighted([]string{"start", "fast", "slow", "end"},
		map[string]float64{"start": 1, "fast": 2, "slow": 10, "end": 1},
		wEdge{"start", "fast", 0}, wEdge{"start", "slow", 0},
		wEdge{"fast", "end", 0}, wEdge{"slow", "end", 0})
	assertCriticalPath(t, g, []string{"start", "slow", "end"}, 12)
}

func TestCriticalPath_VertexWeightsOnly(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	a := g.AddVertexByLabel("a", gograph.WithVertexWeight(2))
	b := g.AddVertexByLabel("b", gograph.WithVertexWeight(5))
	c := g.AddVertexByLabel("c", gograph.WithVertexWeight(1))
	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(a, c)
	assertCriticalPath(t, g, []string{"a", "b"}, 7)
}

func TestCriticalPath_EdgeWeightsOnly(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	a := g.AddVertexByLabel("a")
	b := g.AddVertexByLabel("b")
	c := g.AddVertexByLabel("c")
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(3))
	_, _ = g.AddEdge(a, c, gograph.WithEdgeWeight(8))
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(1))
	assertCriticalPath(t, g, []string{"a", "c"}, 8)
}

func TestCriticalPath_VertexAndEdgeWeights(t *testing.T) {
	// a -> b costs 1+4+1 = 6, a -> c costs 1+0+9 = 10
	g := buildWeighted([]string{"a", "b", "c"},
		map[string]float64{"a": 1, "b": 1, "c": 9},
		wEdge{"a", "b", 4}, wEdge{"a", "c", 0})
	assertCriticalPath(t, g, []string{"a", "c"}, 10)
}

func TestCriticalPath_NegativeWeights(t *testing.T) {
	// the path may start and end anywhere, so the costly tail of the chain wins
	g := buildWeighted([]string{"a", "b", "c"},
		map[string]float64{"a": -5, "b": -1, "c": 4},
		wEdge{"a", "b", -2}, wEdge{"b", "c", 0})
	assertCriticalPath(t, g, []string{"c"}, 4)

	// all weights negative: the best path is the cheapest single vertex
	g = buildWeighted([]string{"a", "b"},
		map[string]float64{"a": -3, "b": -7},
		wEdge{"a", "b", -1})
	assertCriticalPath(t, g, []string{"a"}, -3)
}

func TestCriticalPath_NoEdges(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	g.AddVertexByLabel("a", gograph.WithVertexWeight(2))
	g.AddVertexByLabel("b", gograph.WithVertexWeight(7))
	g.AddVertexByLabel("c", gograph.WithVertexWeight(3))
	assertCriticalPath(t, g, []string{"b"}, 7)
}

func TestCriticalPath_Empty(t *testing.T) {
	path, cost, err := CriticalPath(gograph.New[string](gograph.Directed()))
	if err != nil || path != nil || cost != 0 {
		t.Fatalf("CriticalPath(empty) = %v, %v, %v; want nil, 0, nil", path, cost, err)
	}
}

func TestCriticalPath_UnweightedGraphReadsWeights(t *testing.T) {
	// the graph isn't created with Weighted(), weights are still read
	g := gograph.New[string](gograph.Directed())
	if g.IsWeighted() {
		t.Fatal("test graph should not be weighted")
	}
	a := g.AddVertexByLabel("a", gograph.WithVertexWeight(1))
	b := g.AddVertexByLabel("b", gograph.WithVertexWeight(1))
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(5))
	assertCriticalPath(t, g, []string{"a", "b"}, 7)
}

func TestCriticalPath_TiesAreStable(t *testing.T) {
	build := func() gograph.Graph[string] {
		return buildWeighted([]string{"a", "b", "c", "d"},
			map[string]float64{"a": 1, "b": 1, "c": 1, "d": 1},
			wEdge{"a", "b", 0}, wEdge{"a", "c", 0},
			wEdge{"b", "d", 0}, wEdge{"c", "d", 0})
	}

	first, _, err := CriticalPath(build())
	if err != nil {
		t.Fatalf("CriticalPath: %v", err)
	}
	// b is reached first in topological order, so the path goes through b
	if got := labels(first); !reflect.DeepEqual(got, []string{"a", "b", "d"}) {
		t.Fatalf("CriticalPath = %v, want [a b d]", got)
	}

	for i := 0; i < 20; i++ {
		again, _, _ := CriticalPath(build())
		if !reflect.DeepEqual(labels(again), labels(first)) {
			t.Fatalf("run %d returned %v, first run returned %v", i, labels(again), labels(first))
		}
	}
}

func TestCriticalPath_ReturnsGraphVertices(t *testing.T) {
	g := buildWeighted([]string{"a", "b"},
		map[string]float64{"a": 1, "b": 1}, wEdge{"a", "b", 1})
	path, _, err := CriticalPath(g)
	if err != nil {
		t.Fatalf("CriticalPath: %v", err)
	}
	for _, v := range path {
		if v != g.GetVertexByID(v.Label()) {
			t.Fatalf("vertex %q is not the graph's own pointer", v.Label())
		}
	}
}

func TestCriticalPath_Errors(t *testing.T) {
	undirected := gograph.New[string]()
	undirected.AddVertexByLabel("a")
	if _, _, err := CriticalPath(undirected); !errors.Is(err, gograph.ErrNotDirected) {
		t.Fatalf("undirected: err = %v, want ErrNotDirected", err)
	}

	cyclic := newGraph(t, [2]string{"a", "b"}, [2]string{"b", "c"}, [2]string{"c", "a"})
	if _, _, err := CriticalPath(cyclic); !errors.Is(err, gograph.ErrDAGHasCycle) {
		t.Fatalf("cycle: err = %v, want ErrDAGHasCycle", err)
	}

	selfLoop := gograph.New[string](gograph.Directed())
	a := selfLoop.AddVertexByLabel("a")
	_, _ = selfLoop.AddEdge(a, a)
	if _, _, err := CriticalPath(selfLoop); !errors.Is(err, gograph.ErrDAGHasCycle) {
		t.Fatalf("self-loop: err = %v, want ErrDAGHasCycle", err)
	}

	// a cycle that sits next to an acyclic part still fails
	partial := newGraph(t, [2]string{"x", "y"}, [2]string{"a", "b"}, [2]string{"b", "a"})
	if _, _, err := CriticalPath(partial); !errors.Is(err, gograph.ErrDAGHasCycle) {
		t.Fatalf("partial cycle: err = %v, want ErrDAGHasCycle", err)
	}
}

func TestCriticalPathFunc(t *testing.T) {
	// durations kept outside the graph
	g := newGraph(t, [2]string{"a", "b"}, [2]string{"a", "c"}, [2]string{"b", "d"}, [2]string{"c", "d"})
	duration := map[string]float64{"a": 1, "b": 3, "c": 5, "d": 2}
	delay := map[[2]string]float64{{"a", "b"}: 10, {"a", "c"}: 0}

	path, cost, err := CriticalPathFunc(g,
		func(v *gograph.Vertex[string]) float64 { return duration[v.Label()] },
		func(e *gograph.Edge[string]) float64 {
			return delay[[2]string{e.Source().Label(), e.Destination().Label()}]
		})
	if err != nil {
		t.Fatalf("CriticalPathFunc: %v", err)
	}
	// a -> b -> d is 1+10+3+0+2 = 16, a -> c -> d is 1+0+5+0+2 = 8
	if got := labels(path); !reflect.DeepEqual(got, []string{"a", "b", "d"}) || cost != 16 {
		t.Fatalf("CriticalPathFunc = %v, %v; want [a b d], 16", got, cost)
	}
}

func TestCriticalPathFunc_NilFunctionsCostZero(t *testing.T) {
	g := buildWeighted([]string{"a", "b"},
		map[string]float64{"a": 4, "b": 4}, wEdge{"a", "b", 4})
	path, cost, err := CriticalPathFunc[string](g, nil, nil)
	if err != nil {
		t.Fatalf("CriticalPathFunc: %v", err)
	}
	if cost != 0 || len(path) == 0 {
		t.Fatalf("CriticalPathFunc(nil, nil) = %v, %v; want a path with cost 0", labels(path), cost)
	}

	// only the vertex cost is set
	_, cost, _ = CriticalPathFunc(g, func(v *gograph.Vertex[string]) float64 { return v.Weight() }, nil)
	if cost != 8 {
		t.Fatalf("vertex cost only = %v, want 8", cost)
	}

	// only the edge cost is set
	_, cost, _ = CriticalPathFunc(g, nil, func(e *gograph.Edge[string]) float64 { return e.Weight() })
	if cost != 4 {
		t.Fatalf("edge cost only = %v, want 4", cost)
	}
}

func TestCriticalPathFunc_Errors(t *testing.T) {
	undirected := gograph.New[string]()
	if _, _, err := CriticalPathFunc[string](undirected, nil, nil); !errors.Is(err, gograph.ErrNotDirected) {
		t.Fatalf("err = %v, want ErrNotDirected", err)
	}
}

// bruteForceCriticalPath returns the largest cost over every path in a DAG.
func bruteForceCriticalPath(g gograph.Graph[int]) float64 {
	next := map[int][]*gograph.Edge[int]{}
	for _, e := range g.AllEdges() {
		next[e.Source().Label()] = append(next[e.Source().Label()], e)
	}

	best := math.Inf(-1)
	var walk func(v *gograph.Vertex[int], cost float64)
	walk = func(v *gograph.Vertex[int], cost float64) {
		cost += v.Weight()
		best = math.Max(best, cost)
		for _, e := range next[v.Label()] {
			walk(g.GetVertexByID(e.Destination().Label()), cost+e.Weight())
		}
	}
	for _, v := range g.GetAllVertices() {
		walk(v, 0)
	}
	return best
}

func TestCriticalPath_MatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(121)) //nolint:gosec // seeded so failures are reproducible
	for i := 0; i < 300; i++ {
		n := 1 + rng.Intn(9)
		g := gograph.New[int](gograph.Directed())
		for _, label := range rng.Perm(n) {
			g.AddVertexByLabel(label, gograph.WithVertexWeight(float64(rng.Intn(21)-10)))
		}
		for a := 0; a < n; a++ {
			for b := a + 1; b < n; b++ {
				if rng.Intn(3) == 0 {
					_, _ = g.AddEdge(g.GetVertexByID(a), g.GetVertexByID(b),
						gograph.WithEdgeWeight(float64(rng.Intn(21)-10)))
				}
			}
		}

		path, cost, err := CriticalPath(g)
		if err != nil {
			t.Fatalf("graph %d: CriticalPath: %v", i, err)
		}
		if want := bruteForceCriticalPath(g); cost != want {
			t.Fatalf("graph %d: cost = %v, brute force = %v", i, cost, want)
		}

		// the returned path must be a real path and add up to the cost
		sum := path[0].Weight()
		for j := 1; j < len(path); j++ {
			edges := g.GetAllEdges(path[j-1], path[j])
			if len(edges) != 1 {
				t.Fatalf("graph %d: no edge between %v and %v", i, path[j-1].Label(), path[j].Label())
			}
			sum += edges[0].Weight() + path[j].Weight()
		}
		if sum != cost {
			t.Fatalf("graph %d: path adds up to %v, reported cost is %v", i, sum, cost)
		}
	}
}

func TestCriticalPath_LongChain(t *testing.T) {
	const n = 100000
	g := gograph.New[int](gograph.Directed())
	var prev *gograph.Vertex[int]
	for i := 0; i < n; i++ {
		v := g.AddVertexByLabel(i, gograph.WithVertexWeight(1))
		if prev != nil {
			_, _ = g.AddEdge(prev, v)
		}
		prev = v
	}

	path, cost, err := CriticalPath(g)
	if err != nil {
		t.Fatalf("CriticalPath: %v", err)
	}
	if len(path) != n || cost != n {
		t.Fatalf("len(path) = %d, cost = %v; want %d, %d", len(path), cost, n, n)
	}
}
