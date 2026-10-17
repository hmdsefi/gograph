package dag

import (
	"errors"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

type tEdge struct{ from, to string }

func buildTree(order []string, edges ...tEdge) gograph.Graph[string] {
	g := gograph.New[string](gograph.Directed())
	for _, label := range order {
		g.AddVertexByLabel(label)
	}
	for _, e := range edges {
		_, _ = g.AddEdge(g.GetVertexByID(e.from), g.GetVertexByID(e.to))
	}
	return g
}

func weightOf(w map[string]float64) func(*gograph.Vertex[string]) float64 {
	return func(v *gograph.Vertex[string]) float64 { return w[v.Label()] }
}

func TestHeaviestBranch(t *testing.T) {
	tests := []struct {
		name   string
		order  []string
		edges  []tEdge
		root   string
		weight map[string]float64
		want   []string
	}{
		{
			// a's first block is heavier, but b's chain carries more in total.
			name:   "heavier chain wins over a heavier first edge",
			order:  []string{"g", "a1", "a2", "b1", "b2", "b3"},
			edges:  []tEdge{{"g", "a1"}, {"a1", "a2"}, {"g", "b1"}, {"b1", "b2"}, {"b2", "b3"}},
			root:   "g",
			weight: map[string]float64{"a1": 5, "a2": 1, "b1": 2, "b2": 2, "b3": 3},
			want:   []string{"g", "b1", "b2", "b3"},
		},
		{
			name:   "subtree weight includes every branch below a child",
			order:  []string{"r", "x", "y", "x1", "x2", "y1"},
			edges:  []tEdge{{"r", "x"}, {"r", "y"}, {"x", "x1"}, {"x", "x2"}, {"y", "y1"}},
			root:   "r",
			weight: map[string]float64{"x1": 3, "x2": 3, "y": 1, "y1": 4},
			want:   []string{"r", "x", "x1"},
		},
		{
			name:   "negative weights change the winner",
			order:  []string{"r", "a", "b"},
			edges:  []tEdge{{"r", "a"}, {"r", "b"}},
			root:   "r",
			weight: map[string]float64{"a": -1, "b": -2},
			want:   []string{"r", "a"},
		},
		{
			name:  "single vertex",
			order: []string{"r"},
			root:  "r",
			want:  []string{"r"},
		},
		{
			name:   "ties go to the earlier child in GetAllVertices",
			order:  []string{"r", "b", "a"},
			edges:  []tEdge{{"r", "a"}, {"r", "b"}},
			root:   "r",
			weight: map[string]float64{"a": 2, "b": 2},
			want:   []string{"r", "b"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := buildTree(tc.order, tc.edges...)
			path, err := HeaviestBranch(g, tc.root, weightOf(tc.weight))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := labels(path); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("HeaviestBranch = %v, want %v", got, tc.want)
			}
			for _, v := range path {
				if v != g.GetVertexByID(v.Label()) {
					t.Fatalf("vertex %q is not the graph's own pointer", v.Label())
				}
			}
		})
	}
}

func TestHeaviestBranchTiesAreStable(t *testing.T) {
	g := buildTree([]string{"r", "c", "a", "b", "c1", "a1"},
		tEdge{"r", "a"}, tEdge{"r", "b"}, tEdge{"r", "c"}, tEdge{"c", "c1"}, tEdge{"a", "a1"})
	for i := 0; i < 20; i++ {
		path, err := HeaviestBranch(g, "r", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got, want := labels(path), []string{"r", "c", "c1"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("run %d: HeaviestBranch = %v, want %v", i, got, want)
		}
	}
}

func TestHeaviestBranchCallsWeightOncePerVertex(t *testing.T) {
	g := buildTree([]string{"r", "a", "b", "a1", "a2"},
		tEdge{"r", "a"}, tEdge{"r", "b"}, tEdge{"a", "a1"}, tEdge{"a", "a2"})
	calls := map[string]int{}
	_, err := HeaviestBranch(g, "r", func(v *gograph.Vertex[string]) float64 {
		calls[v.Label()]++
		return 1
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, label := range []string{"r", "a", "b", "a1", "a2"} {
		if calls[label] != 1 {
			t.Fatalf("weight called %d times for %q, want 1", calls[label], label)
		}
	}
}

func TestHeaviestBranchErrors(t *testing.T) {
	undirected := gograph.New[string]()
	undirected.AddVertexByLabel("r")

	cycle := buildTree([]string{"r", "a", "b"}, tEdge{"r", "a"}, tEdge{"a", "b"}, tEdge{"b", "a"})
	selfLoop := buildTree([]string{"r", "a"}, tEdge{"r", "a"}, tEdge{"a", "a"})

	tests := []struct {
		name string
		g    gograph.Graph[string]
		root string
		want error
	}{
		{"undirected", undirected, "r", gograph.ErrNotDirected},
		{"missing root", buildTree([]string{"r"}), "x", gograph.ErrVertexDoesNotExist},
		{
			"vertex with two parents",
			buildTree([]string{"r", "a", "b", "c"}, tEdge{"r", "a"}, tEdge{"r", "b"}, tEdge{"a", "c"}, tEdge{"b", "c"}),
			"r", ErrNotTree,
		},
		{
			"vertex outside the root's reach",
			buildTree([]string{"r", "a", "x"}, tEdge{"r", "a"}),
			"r", ErrNotTree,
		},
		{
			"root has a parent",
			buildTree([]string{"p", "r", "a"}, tEdge{"p", "r"}, tEdge{"r", "a"}),
			"r", ErrNotTree,
		},
		{"cycle", cycle, "r", gograph.ErrDAGHasCycle},
		{
			"cycle the root can't reach",
			buildTree([]string{"r", "a", "b"}, tEdge{"a", "b"}, tEdge{"b", "a"}),
			"r", gograph.ErrDAGHasCycle,
		},
		{"self-loop", selfLoop, "r", gograph.ErrDAGHasCycle},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path, err := HeaviestBranch(tc.g, tc.root, nil)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if path != nil {
				t.Fatalf("path = %v, want nil", labels(path))
			}
		})
	}
}
