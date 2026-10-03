package dag

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

func levelLabels[T comparable](levels [][]*gograph.Vertex[T]) [][]T {
	var out [][]T
	for _, level := range levels {
		out = append(out, labels(level))
	}
	return out
}

func assertLevelLabels(t *testing.T, g gograph.Graph[string], want ...[]string) {
	t.Helper()
	levels, err := Levels(g)
	if err != nil {
		t.Fatalf("Levels: unexpected error: %v", err)
	}
	if got := levelLabels(levels); !reflect.DeepEqual(got, want) {
		t.Fatalf("Levels = %v, want %v", got, want)
	}
}

// randomDAG returns a DAG on the labels 0..n-1 where every edge goes from a
// lower label to a higher one. The vertices and edges are added in random
// order, so GetAllVertices isn't a topological order.
func randomDAG(rng *rand.Rand) gograph.Graph[int] {
	n := 1 + rng.Intn(20)
	g := gograph.New[int](gograph.Directed())
	for _, label := range rng.Perm(n) {
		g.AddVertexByLabel(label)
	}

	var edges [][2]int
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if rng.Intn(4) == 0 {
				edges = append(edges, [2]int{i, j})
			}
		}
	}
	rng.Shuffle(len(edges), func(i, j int) { edges[i], edges[j] = edges[j], edges[i] })
	for _, e := range edges {
		_, _ = g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1]))
	}

	return g
}

func TestLevels_Chain(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"}, [2]string{"b", "c"}, [2]string{"c", "d"})
	assertLevelLabels(t, g, []string{"a"}, []string{"b"}, []string{"c"}, []string{"d"})
}

func TestLevels_Diamond(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"}, [2]string{"a", "c"}, [2]string{"b", "d"}, [2]string{"c", "d"})
	assertLevelLabels(t, g, []string{"a"}, []string{"b", "c"}, []string{"d"})
}

func TestLevels_LongestPath(t *testing.T) {
	// c is one edge away from a, but has to wait for b
	g := newGraph(t, [2]string{"a", "c"}, [2]string{"a", "b"}, [2]string{"b", "c"})
	assertLevelLabels(t, g, []string{"a"}, []string{"b"}, []string{"c"})
}

func TestLevels_FanOut(t *testing.T) {
	g := gograph.New[int](gograph.Directed())
	root := g.AddVertexByLabel(-1)
	want := make([]int, 100)
	for i := range want {
		want[i] = i
		_, _ = g.AddEdge(root, gograph.NewVertex(i))
	}

	levels, err := Levels(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := levelLabels(levels); !reflect.DeepEqual(got, [][]int{{-1}, want}) {
		t.Fatalf("Levels = %v", got)
	}
}

func TestLevels_Disconnected(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"}, [2]string{"y", "z"})
	g.AddVertexByLabel("x")
	assertLevelLabels(t, g, []string{"a", "y", "x"}, []string{"b", "z"})
}

func TestLevels_FollowsVertexOrder(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	for _, label := range []string{"d", "c", "b", "a"} {
		g.AddVertexByLabel(label)
	}
	_, _ = g.AddEdge(g.GetVertexByID("a"), g.GetVertexByID("b"))
	_, _ = g.AddEdge(g.GetVertexByID("a"), g.GetVertexByID("d"))
	_, _ = g.AddEdge(g.GetVertexByID("c"), g.GetVertexByID("d"))

	assertLevelLabels(t, g, []string{"c", "a"}, []string{"d", "b"})
}

func TestLevels_Empty(t *testing.T) {
	levels, err := Levels(gograph.New[string](gograph.Directed()))
	if err != nil || levels != nil {
		t.Fatalf("Levels = %v, %v, want nil, nil", levels, err)
	}
}

func TestLevels_GraphPointers(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"}, [2]string{"b", "c"})
	levels, err := Levels(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, level := range levels {
		for _, v := range level {
			if v != g.GetVertexByID(v.Label()) {
				t.Fatalf("Levels returned a copy of %s", v.Label())
			}
		}
	}
}

func TestLevels_InvalidGraphs(t *testing.T) {
	tests := []struct {
		name string
		g    gograph.Graph[string]
		want error
	}{
		{"cycle", newGraph(t, [2]string{"a", "b"}, [2]string{"b", "c"}, [2]string{"c", "a"}), gograph.ErrDAGHasCycle},
		{"cycle after a DAG part", newGraph(t, [2]string{"x", "a"}, [2]string{"a", "b"}, [2]string{"b", "a"}), gograph.ErrDAGHasCycle},
		{"self-loop", newGraph(t, [2]string{"a", "b"}, [2]string{"b", "b"}), gograph.ErrDAGHasCycle},
		{"undirected", gograph.New[string](), gograph.ErrNotDirected},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if levels, err := Levels(tt.g); !errors.Is(err, tt.want) || levels != nil {
				t.Fatalf("Levels = %v, %v, want nil, %v", levels, err, tt.want)
			}
		})
	}
}

func TestLevels_RandomDAGs(t *testing.T) {
	rng := rand.New(rand.NewSource(1)) //nolint:gosec // seeded so failures are reproducible
	for run := 0; run < 300; run++ {
		g := randomDAG(rng)
		levels, err := Levels(g)
		if err != nil {
			t.Fatalf("run %d: unexpected error: %v", run, err)
		}

		position := make(map[int]int)
		for i, v := range g.GetAllVertices() {
			position[v.Label()] = i
		}
		level := make(map[int]int)
		for l, vertices := range levels {
			if len(vertices) == 0 {
				t.Fatalf("run %d: level %d is empty", run, l)
			}
			for i, v := range vertices {
				if _, ok := level[v.Label()]; ok {
					t.Fatalf("run %d: %d is in more than one level", run, v.Label())
				}
				if i > 0 && position[vertices[i-1].Label()] > position[v.Label()] {
					t.Fatalf("run %d: level %d doesn't follow GetAllVertices: %v", run, l, labels(vertices))
				}
				level[v.Label()] = l
			}
		}
		if len(level) != int(g.Order()) {
			t.Fatalf("run %d: levels have %d vertices, want %d", run, len(level), g.Order())
		}

		// every edge goes up a level, and every vertex above level 0 has a
		// predecessor right below it, so its level is its longest path
		hasPredBelow := make(map[int]bool)
		for _, e := range g.AllEdges() {
			from, to := level[e.Source().Label()], level[e.Destination().Label()]
			if from >= to {
				t.Fatalf("run %d: edge %v -> %v goes from level %d to %d",
					run, e.Source().Label(), e.Destination().Label(), from, to)
			}
			if from == to-1 {
				hasPredBelow[e.Destination().Label()] = true
			}
		}
		for label, l := range level {
			if l > 0 && !hasPredBelow[label] {
				t.Fatalf("run %d: %d is at level %d with no predecessor at level %d", run, label, l, l-1)
			}
		}
	}
}

func TestLevels_LongChain(t *testing.T) {
	const n = 200_000
	g := gograph.New[int](gograph.Directed())
	prev := g.AddVertexByLabel(0)
	for i := 1; i < n; i++ {
		v := g.AddVertexByLabel(i)
		_, _ = g.AddEdge(prev, v)
		prev = v
	}

	levels, err := Levels(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(levels) != n || levels[n-1][0].Label() != n-1 {
		t.Fatalf("got %d levels", len(levels))
	}
}
