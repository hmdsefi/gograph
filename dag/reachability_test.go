package dag

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/hmdsefi/gograph"
)

func newGraph(t *testing.T, edges ...[2]string) gograph.Graph[string] {
	t.Helper()
	g := gograph.New[string](gograph.Directed())
	for _, e := range edges {
		if _, err := g.AddEdge(gograph.NewVertex(e[0]), gograph.NewVertex(e[1])); err != nil {
			t.Fatalf("AddEdge(%s, %s): %v", e[0], e[1], err)
		}
	}
	return g
}

func labels[T comparable](vertices []*gograph.Vertex[T]) []T {
	out := make([]T, len(vertices))
	for i, v := range vertices {
		out[i] = v.Label()
	}
	return out
}

func mustLabels(t *testing.T, vertices []*gograph.Vertex[string], err error) []string {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return labels(vertices)
}

func assertLabels(t *testing.T, name string, got, want []string) {
	t.Helper()
	if len(got) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}

// assertLevels checks that got holds the labels of each level in order,
// in any order within a level.
func assertLevels(t *testing.T, name string, got []string, levels ...[]string) {
	t.Helper()
	i := 0
	for _, level := range levels {
		if i+len(level) > len(got) {
			t.Fatalf("%s = %v, want levels %v", name, got, levels)
		}
		part := append([]string(nil), got[i:i+len(level)]...)
		want := append([]string(nil), level...)
		sort.Strings(part)
		sort.Strings(want)
		if !reflect.DeepEqual(part, want) {
			t.Fatalf("%s = %v, want levels %v", name, got, levels)
		}
		i += len(level)
	}
	if i != len(got) {
		t.Fatalf("%s = %v, want levels %v", name, got, levels)
	}
}

func TestChain(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"}, [2]string{"b", "c"}, [2]string{"c", "d"})

	got, err := Descendants(g, "a")
	assertLabels(t, "Descendants(a)", mustLabels(t, got, err), []string{"b", "c", "d"})
	got, err = Descendants(g, "d")
	assertLabels(t, "Descendants(d)", mustLabels(t, got, err), nil)

	got, err = Ancestors(g, "d")
	assertLabels(t, "Ancestors(d)", mustLabels(t, got, err), []string{"c", "b", "a"})
	got, err = Ancestors(g, "a")
	assertLabels(t, "Ancestors(a)", mustLabels(t, got, err), nil)

	got, err = Affected(g, "c")
	assertLabels(t, "Affected(c)", mustLabels(t, got, err), []string{"c", "d"})
}

func TestDiamond(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"}, [2]string{"a", "c"}, [2]string{"b", "d"}, [2]string{"c", "d"})

	got, err := Descendants(g, "a")
	assertLabels(t, "Descendants(a)", mustLabels(t, got, err), []string{"b", "c", "d"})

	got, err = Ancestors(g, "d")
	assertLevels(t, "Ancestors(d)", mustLabels(t, got, err), []string{"b", "c"}, []string{"a"})

	got, err = Affected(g, "b")
	assertLabels(t, "Affected(b)", mustLabels(t, got, err), []string{"b", "d"})
}

func TestTwoRoots(t *testing.T) {
	g := newGraph(t, [2]string{"r1", "x"}, [2]string{"r2", "x"}, [2]string{"x", "y"})

	got, err := Descendants(g, "r1")
	assertLabels(t, "Descendants(r1)", mustLabels(t, got, err), []string{"x", "y"})
	got, err = Descendants(g, "r2")
	assertLabels(t, "Descendants(r2)", mustLabels(t, got, err), []string{"x", "y"})

	got, err = Ancestors(g, "y")
	assertLevels(t, "Ancestors(y)", mustLabels(t, got, err), []string{"x"}, []string{"r1", "r2"})

	got, err = Affected(g, "r1", "r2")
	assertLabels(t, "Affected(r1, r2)", mustLabels(t, got, err), []string{"r1", "r2", "x", "y"})
}

func TestCycle(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"}, [2]string{"b", "c"}, [2]string{"c", "a"}, [2]string{"c", "d"})

	// a is on the cycle, so it can reach itself
	got, err := Descendants(g, "a")
	assertLabels(t, "Descendants(a)", mustLabels(t, got, err), []string{"b", "c", "a", "d"})
	got, err = Ancestors(g, "a")
	assertLabels(t, "Ancestors(a)", mustLabels(t, got, err), []string{"c", "b", "a"})

	// d is not on the cycle
	got, err = Descendants(g, "d")
	assertLabels(t, "Descendants(d)", mustLabels(t, got, err), nil)
	got, err = Ancestors(g, "d")
	assertLabels(t, "Ancestors(d)", mustLabels(t, got, err), []string{"c", "b", "a"})

	got, err = Affected(g, "b")
	assertLabels(t, "Affected(b)", mustLabels(t, got, err), []string{"b", "c", "a", "d"})
}

func TestSelfLoop(t *testing.T) {
	g := newGraph(t, [2]string{"a", "a"}, [2]string{"a", "b"})

	got, err := Descendants(g, "a")
	assertLabels(t, "Descendants(a)", mustLabels(t, got, err), []string{"a", "b"})
	got, err = Ancestors(g, "a")
	assertLabels(t, "Ancestors(a)", mustLabels(t, got, err), []string{"a"})
	got, err = Ancestors(g, "b")
	assertLabels(t, "Ancestors(b)", mustLabels(t, got, err), []string{"a"})
}

func TestAffected_Overlapping(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"}, [2]string{"b", "c"}, [2]string{"x", "c"})

	// changed vertices come first, in the order given, each once
	got, err := Affected(g, "b", "a", "b", "x")
	assertLabels(t, "Affected(b, a, b, x)", mustLabels(t, got, err), []string{"b", "a", "x", "c"})

	got, err = Affected(g)
	assertLabels(t, "Affected()", mustLabels(t, got, err), nil)
}

func TestGraphPointers(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"}, [2]string{"b", "c"})

	results := map[string]func() ([]*gograph.Vertex[string], error){
		"Descendants": func() ([]*gograph.Vertex[string], error) { return Descendants(g, "a") },
		"Ancestors":   func() ([]*gograph.Vertex[string], error) { return Ancestors(g, "c") },
		"Affected":    func() ([]*gograph.Vertex[string], error) { return Affected(g, "a") },
	}
	for name, run := range results {
		vertices, err := run()
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", name, err)
		}
		for _, v := range vertices {
			if v != g.GetVertexByID(v.Label()) {
				t.Fatalf("%s returned a copy of %s instead of the graph's vertex", name, v.Label())
			}
		}
	}
}

func TestUnknownLabel(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"})

	calls := map[string]func() ([]*gograph.Vertex[string], error){
		"Descendants": func() ([]*gograph.Vertex[string], error) { return Descendants(g, "missing") },
		"Ancestors":   func() ([]*gograph.Vertex[string], error) { return Ancestors(g, "missing") },
		"Affected":    func() ([]*gograph.Vertex[string], error) { return Affected(g, "a", "missing") },
	}
	for name, call := range calls {
		vertices, err := call()
		if !errors.Is(err, gograph.ErrVertexDoesNotExist) {
			t.Fatalf("%s: expected %v, got %v", name, gograph.ErrVertexDoesNotExist, err)
		}
		if !strings.Contains(err.Error(), "missing") {
			t.Fatalf("%s: expected the label in the error, got %q", name, err)
		}
		if vertices != nil {
			t.Fatalf("%s: expected nil result, got %v", name, labels(vertices))
		}
	}
}

func TestUndirected(t *testing.T) {
	g := gograph.New[string]()
	_, _ = g.AddEdge(gograph.NewVertex("a"), gograph.NewVertex("b"))

	calls := map[string]func() ([]*gograph.Vertex[string], error){
		"Descendants": func() ([]*gograph.Vertex[string], error) { return Descendants(g, "a") },
		"Ancestors":   func() ([]*gograph.Vertex[string], error) { return Ancestors(g, "a") },
		"Affected":    func() ([]*gograph.Vertex[string], error) { return Affected(g, "a") },
	}
	for name, call := range calls {
		vertices, err := call()
		if !errors.Is(err, gograph.ErrNotDirected) {
			t.Fatalf("%s: expected %v, got %v", name, gograph.ErrNotDirected, err)
		}
		if vertices != nil {
			t.Fatalf("%s: expected nil result, got %v", name, labels(vertices))
		}
	}
}

func TestLongChain(t *testing.T) {
	const n = 100_000
	g := gograph.New[int](gograph.Directed())
	prev := g.AddVertexByLabel(0)
	for i := 1; i < n; i++ {
		curr := g.AddVertexByLabel(i)
		if _, err := g.AddEdge(prev, curr); err != nil {
			t.Fatalf("AddEdge: %v", err)
		}
		prev = curr
	}

	descendants, err := Descendants(g, 0)
	if err != nil || len(descendants) != n-1 || descendants[0].Label() != 1 || descendants[n-2].Label() != n-1 {
		t.Fatalf("Descendants(0): got %d vertices, err %v", len(descendants), err)
	}

	ancestors, err := Ancestors(g, n-1)
	if err != nil || len(ancestors) != n-1 || ancestors[0].Label() != n-2 || ancestors[n-2].Label() != 0 {
		t.Fatalf("Ancestors(%d): got %d vertices, err %v", n-1, len(ancestors), err)
	}

	affected, err := Affected(g, 0)
	if err != nil || len(affected) != n {
		t.Fatalf("Affected(0): got %d vertices, err %v", len(affected), err)
	}
}
