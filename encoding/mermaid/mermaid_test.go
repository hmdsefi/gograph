package mermaid

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hmdsefi/gograph"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata")

func assertGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name+".mmd")
	if *update {
		if err := os.WriteFile(path, got, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	want, err := os.ReadFile(path) //nolint:gosec // path is built from a test name
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("output doesn't match %s\ngot:\n%s\nwant:\n%s", path, got, want)
	}
}

func addEdges(t *testing.T, g gograph.Graph[string], edges ...[2]string) {
	t.Helper()
	for _, e := range edges {
		if _, err := g.AddEdge(gograph.NewVertex(e[0]), gograph.NewVertex(e[1])); err != nil {
			t.Fatalf("AddEdge(%s, %s): %v", e[0], e[1], err)
		}
	}
}

func mustMarshal[T comparable](t *testing.T, g gograph.Graph[T], opts ...Option[T]) []byte {
	t.Helper()
	out, err := Marshal(g, opts...)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	return out
}

func TestMarshal_Directed(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	addEdges(t, g, [2]string{"A", "B"}, [2]string{"A", "C"}, [2]string{"B", "C"})

	assertGolden(t, "directed", mustMarshal(t, g))
}

func TestMarshal_Undirected(t *testing.T) {
	g := gograph.New[string]()
	addEdges(t, g, [2]string{"A", "B"}, [2]string{"B", "C"}, [2]string{"C", "A"})

	assertGolden(t, "undirected", mustMarshal(t, g))
}

func TestMarshal_Weighted(t *testing.T) {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	a := g.AddVertexByLabel("A")
	_, _ = g.AddEdge(a, gograph.NewVertex("B"), gograph.WithEdgeWeight(4))
	_, _ = g.AddEdge(a, gograph.NewVertex("C"), gograph.WithEdgeWeight(3))
	_, _ = g.AddEdge(g.GetVertexByID("B"), g.GetVertexByID("C"), gograph.WithEdgeWeight(0.25))

	assertGolden(t, "weighted", mustMarshal(t, g, WithDirection[string]("LR")))
}

func TestMarshal_EdgePropertyLabels(t *testing.T) {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	_, _ = g.AddEdge(a, b,
		gograph.WithEdgeWeight(1),
		gograph.WithEdgeLabel("calls \"service\"\nnow"),
	)
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(2))

	assertGolden(t, "edge_property_labels", mustMarshal(t, g))
	assertGolden(t, "edge_property_labels_custom", mustMarshal(t, g,
		WithEdgeLabel(func(*gograph.Edge[string]) string { return "custom" }),
	))
}

func TestMarshal_IsolatedVertex(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	addEdges(t, g, [2]string{"A", "B"})
	g.AddVertexByLabel("C")

	assertGolden(t, "isolated", mustMarshal(t, g))
}

func TestMarshal_SelfLoop(t *testing.T) {
	directed := gograph.New[string](gograph.Directed())
	addEdges(t, directed, [2]string{"A", "A"}, [2]string{"A", "B"})
	assertGolden(t, "self_loop_directed", mustMarshal(t, directed))

	undirected := gograph.New[string]()
	addEdges(t, undirected, [2]string{"A", "A"}, [2]string{"A", "B"})
	assertGolden(t, "self_loop_undirected", mustMarshal(t, undirected))
}

func TestMarshal_SpecialLabels(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	addEdges(t, g,
		[2]string{`say "hi"`, "two words"},
		[2]string{"two words", "a#b;"},
		[2]string{"a#b;", "<b>bold</b> & more"},
		[2]string{"<b>bold</b> & more", "line one\nline two"},
		[2]string{"line one\nline two", "`code`"},
	)

	edgeLabel := func(e *gograph.Edge[string]) string {
		if e.Source().Label() == `say "hi"` {
			return `pipe | and "quotes"`
		}
		return ""
	}
	assertGolden(t, "special_labels", mustMarshal(t, g, WithEdgeLabel(edgeLabel)))
}

func TestMarshal_Keywords(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	addEdges(t, g, [2]string{"graph", "end"}, [2]string{"end", "subgraph"}, [2]string{"subgraph", "class"})

	assertGolden(t, "keywords", mustMarshal(t, g))
}

func TestMarshal_Classes(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	addEdges(t, g, [2]string{"A", "B"}, [2]string{"B", "C"}, [2]string{"C", "D"})

	critical := map[string]bool{"A": true, "C": true}
	class := func(v *gograph.Vertex[string]) string {
		switch {
		case critical[v.Label()]:
			return "critical"
		case v.Label() == "D":
			return "done"
		}
		return ""
	}

	assertGolden(t, "classes", mustMarshal(t, g,
		WithVertexClass(class),
		WithClassDef[string]("critical", "fill:#f96,stroke:#333"),
		WithClassDef[string]("done", "fill:#9f6"),
	))
}

func TestMarshal_CustomLabels(t *testing.T) {
	g := gograph.New[int](gograph.Directed(), gograph.Weighted())
	for _, label := range []int{10, 2, 1} {
		g.AddVertexByLabel(label)
	}
	_, _ = g.AddEdge(g.GetVertexByID(10), g.GetVertexByID(2), gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(g.GetVertexByID(2), g.GetVertexByID(1), gograph.WithEdgeWeight(2))

	out := mustMarshal(t, g,
		WithVertexLabel(func(v *gograph.Vertex[int]) string {
			return "task " + string(rune('A'+v.Label()%26))
		}),
		WithEdgeLabel(func(e *gograph.Edge[int]) string {
			if e.Weight() > 1 {
				return "slow"
			}
			return ""
		}),
	)
	assertGolden(t, "custom_labels", out)
}

func TestMarshal_EmptyLabels(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	addEdges(t, g, [2]string{"", "a"})
	assertGolden(t, "empty_label", mustMarshal(t, g))

	blank := WithVertexLabel(func(*gograph.Vertex[string]) string { return "" })
	assertGolden(t, "empty_label_func", mustMarshal(t, g, blank))
}

func TestMarshal_SamePrintedLabels(t *testing.T) {
	// all three labels print as "1", so the order comes from their types
	build := func() gograph.Graph[any] {
		g := gograph.New[any](gograph.Directed())
		_, _ = g.AddEdge(gograph.NewVertex[any]("1"), gograph.NewVertex[any](1))
		_, _ = g.AddEdge(gograph.NewVertex[any](int64(1)), gograph.NewVertex[any]("1"))
		return g
	}

	want := mustMarshal(t, build())
	assertGolden(t, "same_printed_labels", want)
	for range 50 {
		if got := mustMarshal(t, build()); !bytes.Equal(got, want) {
			t.Fatalf("output changed between runs\ngot:\n%s\nwant:\n%s", got, want)
		}
	}
}

func TestMarshal_Empty(t *testing.T) {
	assertGolden(t, "empty", mustMarshal(t, gograph.New[string]()))
}

func TestMarshal_StableOutput(t *testing.T) {
	build := func() gograph.Graph[int] {
		g := gograph.New[int]()
		for i := range 30 {
			_, _ = g.AddEdge(gograph.NewVertex(i), gograph.NewVertex((i*7+3)%30))
		}
		return g
	}

	want := mustMarshal(t, build())
	for range 50 {
		if got := mustMarshal(t, build()); !bytes.Equal(got, want) {
			t.Fatalf("output changed between runs\ngot:\n%s\nwant:\n%s", got, want)
		}
	}
}

func TestMarshal_Directions(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	addEdges(t, g, [2]string{"A", "B"})

	for _, dir := range []string{"TD", "TB", "LR", "BT", "RL"} {
		out := mustMarshal(t, g, WithDirection[string](dir))
		if first, _, _ := strings.Cut(string(out), "\n"); first != "flowchart "+dir {
			t.Fatalf("WithDirection(%q): first line is %q", dir, first)
		}
	}
}

func TestMarshal_InvalidOptions(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	addEdges(t, g, [2]string{"A", "B"})

	tests := map[string][]Option[string]{
		"direction":        {WithDirection[string]("up")},
		"lowercase dir":    {WithDirection[string]("lr")},
		"class name":       {WithVertexClass(func(*gograph.Vertex[string]) string { return "two words" })},
		"class def name":   {WithClassDef[string]("a,b", "fill:#f96")},
		"empty class def":  {WithClassDef[string]("", "fill:#f96")},
		"class def style":  {WithClassDef[string]("hot", "fill:#f96;\nA --> B")},
		"empty class body": {WithClassDef[string]("hot", "")},
		"hyphen in name":   {WithClassDef[string]("is-hot", "fill:#f96")},
		"digit first":      {WithClassDef[string]("1hot", "fill:#f96")},
	}
	for _, keyword := range []string{"end", "graph", "class", "classDef", "style", "subgraph", "_self"} {
		tests["keyword "+keyword+" as class def"] = []Option[string]{WithClassDef[string](keyword, "fill:#f96")}
		tests["keyword "+keyword+" as vertex class"] = []Option[string]{
			WithVertexClass(func(*gograph.Vertex[string]) string { return keyword }),
		}
	}
	for name, opts := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := Write(&buf, g, opts...); err == nil {
				t.Fatal("expected an error")
			}
			if buf.Len() != 0 {
				t.Fatalf("expected no output, got %q", buf.String())
			}
			if out, err := Marshal(g, opts...); err == nil || out != nil {
				t.Fatalf("Marshal: expected an error and no output, got %q, %v", out, err)
			}
		})
	}
}

func TestMarshal_NilGraph(t *testing.T) {
	if _, err := Marshal[string](nil); err == nil {
		t.Fatal("expected an error for a nil graph")
	}
}

type failingWriter struct{}

var errWrite = errors.New("write failed")

func (failingWriter) Write([]byte) (int, error) { return 0, errWrite }

func TestWrite(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	addEdges(t, g, [2]string{"A", "B"})

	var buf bytes.Buffer
	if err := Write(&buf, g); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), mustMarshal(t, g)) {
		t.Fatal("Write and Marshal returned different output")
	}

	if err := Write(failingWriter{}, g); !errors.Is(err, errWrite) {
		t.Fatalf("expected the writer's error, got %v", err)
	}
}
