package gographtest

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/hmdsefi/gograph"
)

// recorder is a testing.TB that keeps the failures instead of failing the
// test. Calling a method it doesn't override panics, since the embedded TB
// is nil.
type recorder struct {
	testing.TB
	helpers  int
	failures []string
}

func (r *recorder) Helper() { r.helpers++ }

func (r *recorder) Errorf(format string, args ...any) {
	r.failures = append(r.failures, fmt.Sprintf(format, args...))
}

// check fails t unless r recorded no failure when want is empty, or exactly
// the failure want.
func (r *recorder) check(t *testing.T, want string) {
	t.Helper()
	if r.helpers == 0 {
		t.Error("the assertion didn't call Helper")
	}
	var got string
	switch len(r.failures) {
	case 0:
	case 1:
		got = r.failures[0]
	default:
		t.Fatalf("expected at most one failure, got %q", r.failures)
	}
	if got != want {
		t.Fatalf("failure:\n%s\nwant:\n%s", got, want)
	}
}

const (
	notTopological = "gographtest: not a topological order of the graph:"
	groupsDiffer   = "gographtest: groups differ:"
	graphsDiffer   = "gographtest: graphs differ:"
)

// dag returns a -> b, a -> c, b -> d, c -> d.
func dag() gograph.Graph[string] {
	g := gograph.New[string](gograph.Directed())
	for _, e := range [][2]string{{"a", "b"}, {"a", "c"}, {"b", "d"}, {"c", "d"}} {
		_, _ = g.AddEdge(gograph.NewVertex(e[0]), gograph.NewVertex(e[1]))
	}
	return g
}

// vertices returns new vertices with the labels, and nil for each empty
// label.
func vertices(labels ...string) []*gograph.Vertex[string] {
	out := make([]*gograph.Vertex[string], len(labels))
	for i, label := range labels {
		if label != "" {
			out[i] = gograph.NewVertex(label)
		}
	}
	return out
}

func TestIsTopologicalOrder(t *testing.T) {
	tests := []struct {
		name  string
		order []*gograph.Vertex[string]
		want  string
	}{
		{name: "valid", order: vertices("a", "b", "c", "d")},
		{name: "another valid order", order: vertices("a", "c", "b", "d")},
		{
			name:  "edge backward",
			order: vertices("a", "d", "b", "c"),
			want: notTopological +
				"\n\tedge b -> d goes from position 2 to 1" +
				"\n\tedge c -> d goes from position 3 to 1",
		},
		{name: "missing vertex", order: vertices("a", "b", "c"), want: notTopological + "\n\tvertex d is missing"},
		{
			name:  "repeated vertex",
			order: vertices("a", "b", "c", "d", "b"),
			want:  notTopological + "\n\tvertex b comes more than once",
		},
		{
			name:  "vertex not in the graph",
			order: vertices("a", "b", "c", "d", "e"),
			want:  notTopological + "\n\tvertex e isn't in the graph",
		},
		{name: "nil vertex", order: vertices("a", "b", "", "c", "d"), want: notTopological + "\n\tposition 2 holds nil"},
		{
			name:  "every kind of problem",
			order: vertices("d", "", "x", "d", "a"),
			want: notTopological +
				"\n\tposition 1 holds nil" +
				"\n\tvertex x isn't in the graph" +
				"\n\tvertex d comes more than once" +
				"\n\tvertex b is missing" +
				"\n\tvertex c is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &recorder{}
			IsTopologicalOrder(r, dag(), tt.order)
			r.check(t, tt.want)
		})
	}
}

func TestIsTopologicalOrder_GraphOrder(t *testing.T) {
	g := Grid(5, 5, gograph.Directed())
	order, err := gograph.TopologySort(g)
	if err != nil {
		t.Fatal(err)
	}
	r := &recorder{}
	IsTopologicalOrder(r, g, order)
	r.check(t, "")
}

func TestIsTopologicalOrder_SelfLoop(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	a := g.AddVertexByLabel("a")
	_, _ = g.AddEdge(a, a)

	r := &recorder{}
	IsTopologicalOrder(r, g, vertices("a"))
	r.check(t, notTopological+"\n\tedge a -> a goes from position 0 to 0")
}

func TestIsTopologicalOrder_Undirected(t *testing.T) {
	g := gograph.New[string]()
	g.AddVertexByLabel("a")

	r := &recorder{}
	IsTopologicalOrder(r, g, vertices("a"))
	r.check(t, "gographtest: IsTopologicalOrder needs a directed graph")
}

func TestIsTopologicalOrder_ListsTenProblems(t *testing.T) {
	// the reversed path n-1, ..., 0 has each of its n-1 edges backward
	for _, n := range []int{11, 15} {
		g := Path(n, gograph.Directed())
		order, _ := gograph.TopologySort(g)
		slices.Reverse(order)

		want := notTopological
		for i := range min(n-1, 10) {
			want += fmt.Sprintf("\n\tedge %d -> %d goes from position %d to %d", i, i+1, n-1-i, n-2-i)
		}
		if n-1 > 10 {
			want += fmt.Sprintf("\n\tand %d more", n-11)
		}

		r := &recorder{}
		IsTopologicalOrder(r, g, order)
		r.check(t, want)
	}
}

func TestSameGroups(t *testing.T) {
	a, b, c, d := gograph.NewVertex("a"), gograph.NewVertex("b"), gograph.NewVertex("c"), gograph.NewVertex("d")
	tests := []struct {
		name string
		want [][]string
		got  [][]*gograph.Vertex[string]
		fail string
	}{
		{name: "same order", want: [][]string{{"a", "b"}, {"c"}}, got: [][]*gograph.Vertex[string]{{a, b}, {c}}},
		{
			name: "groups and labels in another order",
			want: [][]string{{"a", "b"}, {"c", "d"}},
			got:  [][]*gograph.Vertex[string]{{d, c}, {b, a}},
		},
		{
			name: "groups that share labels",
			want: [][]string{{"a", "b", "c"}, {"b", "c", "d"}},
			got:  [][]*gograph.Vertex[string]{{c, d, b}, {c, b, a}},
		},
		{name: "both empty"},
		{name: "empty group", want: [][]string{{}}, got: [][]*gograph.Vertex[string]{{}}},
		{
			name: "missing group",
			want: [][]string{{"a", "b"}, {"c"}},
			got:  [][]*gograph.Vertex[string]{{b, a}},
			fail: groupsDiffer + "\n\tmissing [c]",
		},
		{
			name: "unexpected group",
			want: [][]string{{"a", "b"}},
			got:  [][]*gograph.Vertex[string]{{b, a}, {d}},
			fail: groupsDiffer + "\n\tunexpected [d]",
		},
		{
			name: "group split in two",
			want: [][]string{{"a", "b", "c"}},
			got:  [][]*gograph.Vertex[string]{{a}, {c, b}},
			fail: groupsDiffer + "\n\tmissing [a b c]\n\tunexpected [a]\n\tunexpected [c b]",
		},
		{
			name: "repeated label",
			want: [][]string{{"a", "b"}},
			got:  [][]*gograph.Vertex[string]{{a, b, a}},
			fail: groupsDiffer + "\n\tmissing [a b]\n\tunexpected [a b a]",
		},
		{
			name: "repeated group",
			want: [][]string{{"a"}, {"a"}},
			got:  [][]*gograph.Vertex[string]{{a}},
			fail: groupsDiffer + "\n\tmissing [a]",
		},
		{
			name: "nil vertex",
			want: [][]string{{"a", "b"}},
			got:  [][]*gograph.Vertex[string]{{a, nil}},
			fail: groupsDiffer + "\n\tgroup 0 of got holds nil\n\tmissing [a b]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &recorder{}
			SameGroups(r, tt.want, tt.got)
			r.check(t, tt.fail)
		})
	}
}

func TestSameGroups_ListsTenProblemsInOrder(t *testing.T) {
	want := make([][]int, 12)
	for i := range want {
		want[i] = []int{i}
	}

	r := &recorder{}
	SameGroups(r, want, [][]*gograph.Vertex[int]{{gograph.NewVertex(12)}})

	var fail strings.Builder
	fail.WriteString(groupsDiffer)
	for i := range 10 {
		fmt.Fprintf(&fail, "\n\tmissing [%d]", i)
	}
	fail.WriteString("\n\tand 3 more")
	r.check(t, fail.String())
}

func TestSameGroups_ListsMissingBeforeUnexpected(t *testing.T) {
	r := &recorder{}
	SameGroups(r, [][]int{{1}, {2}}, [][]*gograph.Vertex[int]{{gograph.NewVertex(3)}})
	r.check(t, groupsDiffer+"\n\tmissing [1]\n\tmissing [2]\n\tunexpected [3]")
}

func TestEqual(t *testing.T) {
	// reversed builds dag with its vertices and edges in reverse order
	reversed := func() gograph.Graph[string] {
		g := gograph.New[string](gograph.Directed())
		for _, label := range []string{"d", "c", "b", "a"} {
			g.AddVertexByLabel(label)
		}
		for _, e := range [][2]string{{"c", "d"}, {"b", "d"}, {"a", "c"}, {"a", "b"}} {
			_, _ = g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1]))
		}
		return g
	}
	weighted := func(w float64) gograph.Graph[string] {
		g := gograph.New[string](gograph.Directed())
		_, _ = g.AddEdge(g.AddVertexByLabel("a", gograph.WithVertexWeight(w)), gograph.NewVertex("b"))
		return g
	}

	tests := []struct {
		name      string
		want, got gograph.Graph[string]
		opts      []gograph.DiffOption
		fail      string
	}{
		{name: "same graph", want: dag(), got: dag()},
		{name: "another order", want: dag(), got: reversed()},
		{name: "both nil"},
		{name: "nil want", got: dag(), fail: graphsDiffer + "\n\twant is nil"},
		{name: "nil got", want: dag(), fail: graphsDiffer + "\n\tgot is nil"},
		{name: "different weight", want: weighted(1), got: weighted(1.5), fail: graphsDiffer + "\n\t~ vertex a weight 1 -> 1.5"},
		{name: "weight within the tolerance", want: weighted(1), got: weighted(1.5), opts: []gograph.DiffOption{gograph.WithWeightTolerance(0.5)}},
		{name: "missing vertex", want: dag(), got: gograph.New[string](gograph.Directed()), fail: graphsDiffer +
			"\n\t- vertex a\n\t- vertex b\n\t- vertex c\n\t- vertex d" +
			"\n\t- edge a -> b\n\t- edge a -> c\n\t- edge b -> d\n\t- edge c -> d"},
		{
			name: "directed and undirected",
			want: dag(),
			got:  gograph.New[string](),
			fail: "gographtest: graphs are not both directed or both undirected: a is directed: true, b is directed: false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &recorder{}
			Equal(r, tt.want, tt.got, tt.opts...)
			r.check(t, tt.fail)
		})
	}
}

func TestEqual_ListsTenProblems(t *testing.T) {
	r := &recorder{}
	Equal(r, gograph.New[int](), Path(12))

	var fail strings.Builder
	fail.WriteString(graphsDiffer)
	for i := range 10 {
		fmt.Fprintf(&fail, "\n\t+ vertex %d", i)
	}
	fail.WriteString("\n\tand 13 more")
	r.check(t, fail.String())
}
