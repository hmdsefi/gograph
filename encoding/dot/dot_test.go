package dot

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hmdsefi/gograph"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata")

func assertGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name+".gv")
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
	_, _ = g.AddEdge(a, gograph.NewVertex("C"), gograph.WithEdgeWeight(-3))
	_, _ = g.AddEdge(g.GetVertexByID("B"), g.GetVertexByID("C"), gograph.WithEdgeWeight(0.25))

	assertGolden(t, "weighted", mustMarshal(t, g))
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
		WithEdgeAttributes(func(*gograph.Edge[string]) map[string]string {
			return map[string]string{"label": "custom"}
		}),
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
	for _, label := range []string{
		`say "hi"`,
		`C:\path\`,
		"two\nlines",
		"windows\r\nline",
		"a b; c=d [e] {f} -> g",
		"graph",
		"Node",
		"<nil>",
		"",
		"日本",
	} {
		g.AddVertexByLabel(label)
	}

	assertGolden(t, "special_labels", mustMarshal(t, g))
}

func TestMarshal_HTMLLabels(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	addEdges(t, g, [2]string{"A", "B"})

	out := mustMarshal(t, g,
		WithVertexAttributes(func(v *gograph.Vertex[string]) map[string]string {
			if v.Label() == "A" {
				return map[string]string{"label": "<<b>A</b><br/>bold>", "shape": "box"}
			}
			// not HTML-like, since it doesn't end with >
			return map[string]string{"label": "<B"}
		}),
		WithEdgeAttributes(func(*gograph.Edge[string]) map[string]string {
			return map[string]string{"label": "<<i>x</i>>"}
		}),
	)
	assertGolden(t, "html_labels", out)
}

func TestMarshal_GraphAttributes(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	addEdges(t, g, [2]string{"A", "B"})

	out := mustMarshal(t, g,
		WithName[string]("deps"),
		WithGraphAttributes[string](map[string]string{
			"rankdir":  "LR",
			"label":    "Build\ngraph",
			"fontsize": "10",
		}),
	)
	assertGolden(t, "graph_attributes", out)
}

func TestMarshal_QuotedIDs(t *testing.T) {
	g := gograph.New[string]()
	addEdges(t, g, [2]string{"A", "B"})

	out := mustMarshal(t, g,
		WithName[string]("my graph"),
		WithVertexAttributes(func(*gograph.Vertex[string]) map[string]string {
			return map[string]string{"data key": "1", "Edge": "x", "_ok1": "y", "2x": "z"}
		}),
	)
	assertGolden(t, "quoted_ids", out)

	for _, name := range []string{"Graph", "1abc", "a-b", "é"} {
		if out := mustMarshal(t, g, WithName[string](name)); !bytes.HasPrefix(out, []byte(`graph "`)) {
			t.Fatalf("WithName(%q) wrote an unquoted name:\n%s", name, out)
		}
	}
}

func TestMarshal_OverriddenDefaults(t *testing.T) {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"), gograph.WithEdgeWeight(4))
	_, _ = g.AddEdge(g.GetVertexByID("B"), gograph.NewVertex("C"), gograph.WithEdgeWeight(2))

	out := mustMarshal(t, g,
		WithVertexAttributes(func(v *gograph.Vertex[string]) map[string]string {
			if v.Label() == "B" {
				return map[string]string{"label": "step B", "color": "red"}
			}
			return nil
		}),
		WithEdgeAttributes(func(e *gograph.Edge[string]) map[string]string {
			if e.Source().Label() == "A" {
				return map[string]string{"label": "", "style": "dashed"}
			}
			return map[string]string{"weight": "3"}
		}),
	)
	assertGolden(t, "overridden_defaults", out)
}

func TestMarshal_UndirectedEdgeAttributes(t *testing.T) {
	g := gograph.New[string]()
	addEdges(t, g, [2]string{"B", "A"})

	var got []string
	_ = mustMarshal(t, g, WithEdgeAttributes(func(e *gograph.Edge[string]) map[string]string {
		got = append(got, e.Source().Label()+e.Destination().Label())
		return nil
	}))
	if len(got) != 1 || got[0] != "AB" {
		t.Fatalf("WithEdgeAttributes got edges %v, want [AB]", got)
	}
}

func TestMarshal_SamePrintedLabels(t *testing.T) {
	g := gograph.New[any](gograph.Directed())
	_, _ = g.AddEdge(gograph.NewVertex[any](1), gograph.NewVertex[any]("1"))

	assertGolden(t, "same_printed_labels", mustMarshal(t, g))
}

func TestMarshal_Empty(t *testing.T) {
	assertGolden(t, "empty", mustMarshal(t, gograph.New[string](gograph.Directed())))
}

func TestMarshal_StableOutput(t *testing.T) {
	first := gograph.New[string](gograph.Directed())
	addEdges(t, first, [2]string{"A", "B"}, [2]string{"A", "C"}, [2]string{"C", "D"})
	second := gograph.New[string](gograph.Directed())
	for _, label := range []string{"D", "C", "B", "A"} {
		second.AddVertexByLabel(label)
	}
	addEdges(t, second, [2]string{"C", "D"}, [2]string{"A", "B"}, [2]string{"A", "C"})

	attrs := WithVertexAttributes(func(*gograph.Vertex[string]) map[string]string {
		return map[string]string{"z": "1", "a": "2", "m": "3", "label": "x"}
	})
	want := mustMarshal(t, first, attrs)
	for i := 0; i < 20; i++ {
		if got := mustMarshal(t, second, attrs); !bytes.Equal(got, want) {
			t.Fatalf("output depends on the order the graph was built in\ngot:\n%s\nwant:\n%s", got, want)
		}
	}
}

func TestWithGraphAttributes_CopiesMap(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	attrs := map[string]string{"rankdir": "LR"}
	opt := WithGraphAttributes[string](attrs)
	attrs["rankdir"] = "TB"

	if out := mustMarshal(t, g, opt); !bytes.Contains(out, []byte(`rankdir="LR"`)) {
		t.Fatalf("changing the map after WithGraphAttributes changed the output:\n%s", out)
	}
}

func TestMarshal_NilGraph(t *testing.T) {
	if _, err := Marshal[string](nil); err == nil {
		t.Fatal("Marshal(nil) returned no error")
	}
}

var errWrite = errors.New("write failed")

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errWrite }

func TestWrite(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	addEdges(t, g, [2]string{"A", "B"})

	var buf bytes.Buffer
	if err := Write(&buf, g); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), mustMarshal(t, g)) {
		t.Fatalf("Write and Marshal differ:\n%s", buf.String())
	}

	if err := Write(failingWriter{}, g); !errors.Is(err, errWrite) {
		t.Fatalf("Write to a failing writer = %v, want %v", err, errWrite)
	}
}

// TestGolden_Graphviz checks that Graphviz reads every golden file without
// errors or warnings. It runs when dot is on the PATH, which CI installs.
func TestGolden_Graphviz(t *testing.T) {
	dot, err := exec.LookPath("dot")
	if err != nil {
		t.Skip("dot is not installed")
	}

	files, err := filepath.Glob(filepath.Join("testdata", "*.gv"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no golden files: %v", err)
	}
	for _, file := range files {
		var stderr strings.Builder
		cmd := exec.Command(dot, "-Tsvg", "-o", os.DevNull, file) //nolint:gosec // runs dot on the test's own files
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil || stderr.Len() > 0 {
			t.Errorf("dot %s: %v\n%s", file, err, stderr.String())
		}
	}
}
