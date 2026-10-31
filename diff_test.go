package gograph

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"
)

// graphSpec builds a graph with string labels. An edge whose vertices aren't
// listed adds them with weight 0.
type graphSpec struct {
	opts     []GraphOptionFunc
	vertices []VertexEntry[string]
	edges    []EdgeEntry[string]
}

func (s graphSpec) build(t *testing.T) Graph[string] {
	t.Helper()
	g := New[string](s.opts...)
	for _, v := range s.vertices {
		g.AddVertexByLabel(v.Label, WithVertexWeight(v.Weight))
	}
	for _, e := range s.edges {
		if _, err := g.AddEdge(NewVertex(e.From), NewVertex(e.To), WithEdgeWeight(e.Weight), WithEdgeLabel(e.Label)); err != nil {
			t.Fatalf("edge %v: %v", e, err)
		}
	}
	return g
}

func vx(label string, weight float64) VertexEntry[string] {
	return VertexEntry[string]{Label: label, Weight: weight}
}

func ed(from, to string, weight float64, label string) EdgeEntry[string] {
	return EdgeEntry[string]{From: from, To: to, Weight: weight, Label: label}
}

// fullState describes everything Diff compares, in the order of g.
func fullState[T comparable](g Graph[T]) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%+v;", summarize(g))
	for v := range Vertices(g) {
		fmt.Fprintf(&b, " %v:%g", v.Label(), v.Weight())
	}
	b.WriteString(";")
	for e := range Edges(g) {
		fmt.Fprintf(&b, " %v>%v:%g:%q", e.Source().Label(), e.Destination().Label(), e.Weight(), e.Label())
	}
	return b.String()
}

func TestDiff(t *testing.T) {
	directed := []GraphOptionFunc{Directed()}
	tests := []struct {
		name string
		a, b graphSpec
		want string
	}{
		{
			name: "identical",
			a:    graphSpec{opts: directed, vertices: []VertexEntry[string]{vx("A", 1)}, edges: []EdgeEntry[string]{ed("A", "B", 2, "x")}},
			b:    graphSpec{opts: directed, vertices: []VertexEntry[string]{vx("A", 1)}, edges: []EdgeEntry[string]{ed("A", "B", 2, "x")}},
		},
		{name: "both empty", a: graphSpec{opts: directed}, b: graphSpec{opts: directed}},
		{
			name: "added vertex",
			a:    graphSpec{opts: directed, vertices: []VertexEntry[string]{vx("A", 0)}},
			b:    graphSpec{opts: directed, vertices: []VertexEntry[string]{vx("A", 0), vx("B", 2)}},
			want: "+ vertex B weight 2",
		},
		{
			name: "removed vertex and its edges",
			a:    graphSpec{opts: directed, edges: []EdgeEntry[string]{ed("A", "B", 0, ""), ed("B", "A", 1, "back"), ed("A", "C", 0, "")}},
			b:    graphSpec{opts: directed, edges: []EdgeEntry[string]{ed("A", "C", 0, "")}},
			want: "- vertex B\n- edge A -> B\n- edge B -> A weight 1 label \"back\"",
		},
		{
			name: "changed vertex weight",
			a:    graphSpec{opts: directed, vertices: []VertexEntry[string]{vx("A", 1)}},
			b:    graphSpec{opts: directed, vertices: []VertexEntry[string]{vx("A", 2)}},
			want: "~ vertex A weight 1 -> 2",
		},
		{
			name: "added edge",
			a:    graphSpec{opts: directed, vertices: []VertexEntry[string]{vx("A", 0), vx("B", 0)}},
			b:    graphSpec{opts: directed, edges: []EdgeEntry[string]{ed("A", "B", 3, "uses")}},
			want: "+ edge A -> B weight 3 label \"uses\"",
		},
		{
			name: "removed edge",
			a:    graphSpec{opts: directed, edges: []EdgeEntry[string]{ed("A", "B", 0, "")}},
			b:    graphSpec{opts: directed, vertices: []VertexEntry[string]{vx("A", 0), vx("B", 0)}},
			want: "- edge A -> B",
		},
		{
			name: "changed edge weight",
			a:    graphSpec{opts: directed, edges: []EdgeEntry[string]{ed("A", "B", 3, "")}},
			b:    graphSpec{opts: directed, edges: []EdgeEntry[string]{ed("A", "B", 5, "")}},
			want: "~ edge A -> B weight 3 -> 5",
		},
		{
			name: "changed edge label",
			a:    graphSpec{opts: directed, edges: []EdgeEntry[string]{ed("A", "B", 0, "a")}},
			b:    graphSpec{opts: directed, edges: []EdgeEntry[string]{ed("A", "B", 0, "b")}},
			want: "~ edge A -> B label \"a\" -> \"b\"",
		},
		{
			name: "changed edge weight and label",
			a:    graphSpec{opts: directed, edges: []EdgeEntry[string]{ed("A", "B", 1, "")}},
			b:    graphSpec{opts: directed, edges: []EdgeEntry[string]{ed("A", "B", 2, "b")}},
			want: "~ edge A -> B weight 1 -> 2, label \"\" -> \"b\"",
		},
		{
			name: "weighted property",
			a:    graphSpec{opts: directed},
			b:    graphSpec{opts: []GraphOptionFunc{Directed(), Weighted()}},
			want: "~ weighted false -> true",
		},
		{
			name: "acyclic property",
			a:    graphSpec{opts: []GraphOptionFunc{Acyclic()}},
			b:    graphSpec{opts: directed},
			want: "~ acyclic true -> false",
		},
		{
			name: "self-loop",
			a:    graphSpec{opts: directed, vertices: []VertexEntry[string]{vx("A", 0)}},
			b:    graphSpec{opts: directed, edges: []EdgeEntry[string]{ed("A", "A", 0, "")}},
			want: "+ edge A -> A",
		},
		{
			name: "every kind of change",
			a: graphSpec{
				opts:     directed,
				vertices: []VertexEntry[string]{vx("A", 0), vx("B", 0), vx("C", 1)},
				edges:    []EdgeEntry[string]{ed("A", "B", 0, ""), ed("A", "C", 3, "")},
			},
			b: graphSpec{
				opts:     []GraphOptionFunc{Directed(), Weighted()},
				vertices: []VertexEntry[string]{vx("A", 0), vx("C", 2), vx("D", 0)},
				edges:    []EdgeEntry[string]{ed("A", "C", 5, ""), ed("A", "D", 0, "")},
			},
			want: "~ weighted false -> true\n" +
				"+ vertex D\n" +
				"- vertex B\n" +
				"~ vertex C weight 1 -> 2\n" +
				"+ edge A -> D\n" +
				"- edge A -> B\n" +
				"~ edge A -> C weight 3 -> 5",
		},
		{
			name: "undirected edge in either orientation",
			a:    graphSpec{edges: []EdgeEntry[string]{ed("A", "B", 1, "x")}},
			b:    graphSpec{vertices: []VertexEntry[string]{vx("B", 0)}, edges: []EdgeEntry[string]{ed("B", "A", 1, "x")}},
		},
		{
			name: "undirected edge in the orientation found first",
			a:    graphSpec{vertices: []VertexEntry[string]{vx("B", 0), vx("A", 0)}, edges: []EdgeEntry[string]{ed("A", "B", 0, "")}},
			b:    graphSpec{vertices: []VertexEntry[string]{vx("A", 0), vx("B", 0)}},
			want: "- edge B -> A",
		},
		{
			name: "undirected change in the orientation of b",
			a:    graphSpec{edges: []EdgeEntry[string]{ed("A", "B", 1, "")}},
			b:    graphSpec{vertices: []VertexEntry[string]{vx("B", 0)}, edges: []EdgeEntry[string]{ed("A", "B", 2, "")}},
			want: "~ edge B -> A weight 1 -> 2",
		},
		{
			name: "undirected self-loop",
			a:    graphSpec{edges: []EdgeEntry[string]{ed("A", "A", 0, ""), ed("A", "B", 0, "")}},
			b:    graphSpec{edges: []EdgeEntry[string]{ed("A", "B", 0, "")}},
			want: "- edge A -> A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b := tt.a.build(t), tt.b.build(t)
			for _, wrapped := range []bool{false, true} {
				x, y := a, b
				if wrapped {
					x, y = cloneWrapper[string]{a}, cloneWrapper[string]{b}
				}
				d, err := Diff(x, y)
				if err != nil {
					t.Fatal(err)
				}
				if got := d.String(); got != tt.want {
					t.Fatalf("wrapped %v:\n%s\nwant:\n%s", wrapped, got, tt.want)
				}
				if d.IsEmpty() != (tt.want == "") {
					t.Fatalf("wrapped %v: IsEmpty %v for %q", wrapped, d.IsEmpty(), tt.want)
				}
			}
		})
	}
}

func TestDiff_Entries(t *testing.T) {
	a := graphSpec{
		opts:     []GraphOptionFunc{Directed()},
		vertices: []VertexEntry[string]{vx("A", 1), vx("B", 2)},
		edges:    []EdgeEntry[string]{ed("A", "B", 3, "ab"), ed("B", "A", 4, "ba")},
	}.build(t)
	b := graphSpec{
		opts:     []GraphOptionFunc{Directed(), Weighted()},
		vertices: []VertexEntry[string]{vx("A", 5), vx("C", 6)},
		edges:    []EdgeEntry[string]{ed("A", "C", 7, "ac"), ed("A", "A", 8, "")},
	}.build(t)

	d, err := Diff(a, b)
	if err != nil {
		t.Fatal(err)
	}
	want := &GraphDiff[string]{
		AddedVertices:   []VertexEntry[string]{vx("C", 6)},
		RemovedVertices: []VertexEntry[string]{vx("B", 2)},
		ChangedVertices: []VertexChange[string]{{Label: "A", OldWeight: 1, NewWeight: 5}},
		AddedEdges:      []EdgeEntry[string]{ed("A", "C", 7, "ac"), ed("A", "A", 8, "")},
		RemovedEdges:    []EdgeEntry[string]{ed("A", "B", 3, "ab"), ed("B", "A", 4, "ba")},
		OldProperties:   &PropertiesSummary{Directed: true},
		NewProperties:   &PropertiesSummary{Directed: true, Weighted: true},
	}
	if !reflect.DeepEqual(d, want) {
		t.Fatalf("got %+v\nwant %+v", d, want)
	}
}

func TestDiff_Order(t *testing.T) {
	a := graphSpec{
		opts:     []GraphOptionFunc{Directed()},
		vertices: []VertexEntry[string]{vx("C", 1), vx("A", 1), vx("B", 1)},
		edges:    []EdgeEntry[string]{ed("B", "C", 0, ""), ed("C", "B", 0, ""), ed("C", "A", 0, ""), ed("A", "B", 1, "")},
	}.build(t)
	b := graphSpec{
		opts:     []GraphOptionFunc{Directed()},
		vertices: []VertexEntry[string]{vx("Z", 0), vx("B", 2), vx("X", 0), vx("A", 2)},
		edges:    []EdgeEntry[string]{ed("Z", "X", 0, ""), ed("B", "A", 0, ""), ed("A", "B", 2, ""), ed("X", "Z", 0, "")},
	}.build(t)

	d, err := Diff(a, b)
	if err != nil {
		t.Fatal(err)
	}
	want := "+ vertex Z\n+ vertex X\n- vertex C weight 1\n~ vertex B weight 1 -> 2\n~ vertex A weight 1 -> 2\n" +
		"+ edge Z -> X\n+ edge B -> A\n+ edge X -> Z\n" +
		"- edge C -> B\n- edge C -> A\n- edge B -> C\n" +
		"~ edge A -> B weight 1 -> 2"
	if got := d.String(); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestDiff_Weights(t *testing.T) {
	nan, inf := math.NaN(), math.Inf(1)
	tests := []struct {
		name string
		x, y float64
		opts []DiffOption
		same bool
	}{
		{name: "equal", x: 1, y: 1, same: true},
		{name: "different", x: 1, y: 1.05},
		{name: "both NaN", x: nan, y: nan, same: true},
		{name: "NaN and a number", x: nan, y: 1},
		{name: "both +Inf", x: inf, y: inf, same: true},
		{name: "+Inf and -Inf", x: inf, y: -inf},
		{name: "within the tolerance", x: 1, y: 1.05, opts: []DiffOption{WithWeightTolerance(0.1)}, same: true},
		{name: "beyond the tolerance", x: 1, y: 1.2, opts: []DiffOption{WithWeightTolerance(0.1)}},
		{name: "both +Inf with a tolerance", x: inf, y: inf, opts: []DiffOption{WithWeightTolerance(0.1)}, same: true},
		{name: "NaN and a number with a tolerance", x: nan, y: 1, opts: []DiffOption{WithWeightTolerance(0.1)}},
		{name: "negative tolerance", x: 1, y: 1.05, opts: []DiffOption{WithWeightTolerance(-1)}},
		{name: "NaN tolerance", x: 1, y: 1.05, opts: []DiffOption{WithWeightTolerance(nan)}},
		{name: "last tolerance wins", x: 1, y: 1.05, opts: []DiffOption{WithWeightTolerance(0.1), WithWeightTolerance(-1)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := graphSpec{opts: []GraphOptionFunc{Directed()}, vertices: []VertexEntry[string]{vx("A", tt.x)}, edges: []EdgeEntry[string]{ed("A", "B", tt.x, "")}}.build(t)
			b := graphSpec{opts: []GraphOptionFunc{Directed()}, vertices: []VertexEntry[string]{vx("A", tt.y)}, edges: []EdgeEntry[string]{ed("A", "B", tt.y, "")}}.build(t)
			d, err := Diff(a, b, tt.opts...)
			if err != nil {
				t.Fatal(err)
			}
			if got := len(d.ChangedVertices) == 0 && len(d.ChangedEdges) == 0; got != tt.same {
				t.Fatalf("same %v, want %v: %s", got, tt.same, d)
			}
			if !tt.same && (len(d.ChangedVertices) != 1 || len(d.ChangedEdges) != 1) {
				t.Fatalf("want one vertex and one edge change: %s", d)
			}
		})
	}

	a := graphSpec{opts: []GraphOptionFunc{Directed()}, vertices: []VertexEntry[string]{vx("A", nan)}, edges: []EdgeEntry[string]{ed("A", "B", nan, "")}}.build(t)
	if d, err := Diff(a, a); err != nil || !d.IsEmpty() {
		t.Fatalf("a graph with NaN weights differs from itself: %v, %s", err, d)
	}
}

func TestDiff_IncompatibleGraphs(t *testing.T) {
	for _, opts := range [][]GraphOptionFunc{{Directed()}, {Acyclic()}} {
		directed, undirected := New[string](opts...), New[string]()
		for _, pair := range [][2]Graph[string]{{directed, undirected}, {undirected, directed}} {
			d, err := Diff(pair[0], pair[1])
			if !errors.Is(err, ErrIncompatibleGraphs) || d != nil {
				t.Fatalf("got %v, %v", d, err)
			}
		}
	}
}

func TestDiff_NilGraph(t *testing.T) {
	var typedNil *baseGraph[string]
	for _, pair := range [][2]Graph[string]{{nil, New[string]()}, {New[string](), nil}, {typedNil, New[string]()}} {
		func() {
			defer func() {
				if r := recover(); r != "gograph: Diff requires non-nil graphs" {
					t.Fatalf("got panic %v", r)
				}
			}()
			_, _ = Diff(pair[0], pair[1])
		}()
	}
}

func TestGraphDiff_IsEmpty(t *testing.T) {
	var nilDiff *GraphDiff[string]
	if !nilDiff.IsEmpty() || nilDiff.String() != "" || !(&GraphDiff[string]{}).IsEmpty() {
		t.Fatal("a nil or zero diff isn't empty")
	}

	diffs := []*GraphDiff[string]{
		{AddedVertices: []VertexEntry[string]{vx("A", 0)}},
		{RemovedVertices: []VertexEntry[string]{vx("A", 0)}},
		{ChangedVertices: []VertexChange[string]{{Label: "A"}}},
		{AddedEdges: []EdgeEntry[string]{ed("A", "B", 0, "")}},
		{RemovedEdges: []EdgeEntry[string]{ed("A", "B", 0, "")}},
		{ChangedEdges: []EdgeChange[string]{{From: "A", To: "B"}}},
		{OldProperties: &PropertiesSummary{}},
		{NewProperties: &PropertiesSummary{}},
	}
	for i, d := range diffs {
		if d.IsEmpty() {
			t.Fatalf("diff %d is empty", i)
		}
	}
}

func TestGraphDiff_String(t *testing.T) {
	tests := []struct {
		name string
		d    GraphDiff[string]
		want string
	}{
		{
			name: "special weights and quoted labels",
			d: GraphDiff[string]{
				AddedVertices: []VertexEntry[string]{vx("A", math.NaN()), vx("B", math.Inf(-1)), vx("C", 1e21)},
				AddedEdges:    []EdgeEntry[string]{ed("A", "B", 0.5, "say \"hi\"\n")},
			},
			want: "+ vertex A weight NaN\n+ vertex B weight -Inf\n+ vertex C weight 1e+21\n+ edge A -> B weight 0.5 label \"say \\\"hi\\\"\\n\"",
		},
		{
			name: "edge change with nothing changed",
			d:    GraphDiff[string]{ChangedEdges: []EdgeChange[string]{{From: "A", To: "B", OldWeight: 1, NewWeight: 1, OldLabel: "x", NewLabel: "x"}}},
			want: "~ edge A -> B",
		},
		{
			name: "edge change with NaN weights",
			d:    GraphDiff[string]{ChangedEdges: []EdgeChange[string]{{From: "A", To: "B", OldWeight: math.NaN(), NewWeight: math.NaN(), NewLabel: "x"}}},
			want: "~ edge A -> B label \"\" -> \"x\"",
		},
		{
			name: "every property",
			d: GraphDiff[string]{
				OldProperties: &PropertiesSummary{Directed: true, Acyclic: true},
				NewProperties: &PropertiesSummary{Weighted: true},
			},
			want: "~ directed true -> false\n~ weighted false -> true\n~ acyclic true -> false",
		},
		{
			name: "only the new properties",
			d:    GraphDiff[string]{NewProperties: &PropertiesSummary{Weighted: true}},
			want: "~ weighted false -> true",
		},
		{
			name: "only the old properties",
			d:    GraphDiff[string]{OldProperties: &PropertiesSummary{Directed: true}},
			want: "~ directed true -> false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.String(); got != tt.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

// randomPair returns a random graph with int labels and a random change of
// it. The edges of an acyclic graph go from a lower label to a higher one.
func randomPair(rng *rand.Rand) (a, b Graph[int]) {
	weights := []float64{0, 1, 2.5, math.NaN()}
	labels := []string{"", "x", "y"}
	pick := func() float64 { return weights[rng.IntN(len(weights))] }

	var opts []GraphOptionFunc
	switch rng.IntN(3) {
	case 0:
		opts = append(opts, Directed())
	case 1:
		opts = append(opts, Acyclic())
	}
	// a directed graph with every edge from a lower label to a higher one
	// can become acyclic
	lowToHigh := len(opts) > 0 && rng.IntN(2) == 0
	if rng.IntN(2) == 0 {
		opts = append(opts, Weighted())
	}

	n := rng.IntN(8)
	g := New[int](opts...)
	for i := range n {
		g.AddVertexByLabel(i, WithVertexWeight(pick()))
	}
	addEdges := func(g Graph[int], count, labelsUpTo int, ordered bool) {
		for range count {
			if labelsUpTo == 0 {
				return
			}
			from, to := rng.IntN(labelsUpTo), rng.IntN(labelsUpTo)
			if (ordered || g.IsAcyclic()) && from >= to {
				continue
			}
			_, _ = g.AddEdge(NewVertex(from), NewVertex(to), WithEdgeWeight(pick()), WithEdgeLabel(labels[rng.IntN(len(labels))]))
		}
	}
	addEdges(g, rng.IntN(3*n+1), n, lowToHigh)

	c, _ := Clone(g).(*baseGraph[int])
	switch rng.IntN(4) {
	case 0:
		c.properties.isWeighted = !c.properties.isWeighted
	case 1:
		if c.properties.isAcyclic {
			c.properties.isAcyclic, c.ord = false, nil
			lowToHigh = false
		} else if c.properties.isDirected && (lowToHigh || c.Size() == 0) {
			c.properties.isAcyclic, c.ord = true, nil
		}
	}

	for _, v := range c.GetAllVertices() {
		switch rng.IntN(6) {
		case 0:
			c.RemoveVertices(v)
		case 1:
			v.properties.weight = pick()
		}
	}
	for _, e := range c.AllEdges() {
		switch rng.IntN(6) {
		case 0:
			c.RemoveEdges(e)
		case 1:
			if stored := c.edges[e.source.label][e.dest.label]; stored != nil {
				props := EdgeProperties{weight: pick(), label: labels[rng.IntN(len(labels))]}
				stored.properties = props
				if !c.IsDirected() {
					c.edges[e.dest.label][e.source.label].properties = props
				}
			}
		}
	}
	extra := rng.IntN(4)
	for i := n; i < n+extra; i++ {
		c.AddVertexByLabel(i, WithVertexWeight(pick()))
	}
	addEdges(c, rng.IntN(2*(n+extra)+1), n+extra, lowToHigh)

	return g, c
}

func TestGraphDiff_ApplyRandomGraphs(t *testing.T) {
	for seed := range uint64(600) {
		rng := rand.New(rand.NewPCG(seed, 125)) //nolint:gosec // seeded so failures are reproducible
		a, b := randomPair(rng)
		before := fullState(a)

		d, err := Diff(a, b)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		wrapped, err := Diff[int](cloneWrapper[int]{a}, cloneWrapper[int]{b})
		if err != nil || d.String() != wrapped.String() {
			t.Fatalf("seed %d: the diff of wrapped graphs differs:\n%s\n%s", seed, d, wrapped)
		}

		c, err := d.Apply(a)
		if err != nil {
			t.Fatalf("seed %d: %v\n%s", seed, err, d)
		}
		if rest, err := Diff(c, b); err != nil || !rest.IsEmpty() {
			t.Fatalf("seed %d: the result differs from b: %v\n%s", seed, err, rest)
		}
		if fullState(a) != before {
			t.Fatalf("seed %d: Apply changed a", seed)
		}
		for v := range Vertices(b) {
			w := c.GetVertexByID(v.Label())
			if w.InDegree() != v.InDegree() || w.OutDegree() != v.OutDegree() {
				t.Fatalf("seed %d: vertex %d has in and out degrees %d, %d, want %d, %d",
					seed, v.Label(), w.InDegree(), w.OutDegree(), v.InDegree(), v.OutDegree())
			}
		}
		if c.IsAcyclic() {
			if _, err := TopologySort(c); err != nil {
				t.Fatalf("seed %d: %v", seed, err)
			}
		}

		back, err := Diff(b, a)
		if err != nil {
			t.Fatal(err)
		}
		if len(back.AddedVertices) != len(d.RemovedVertices) || len(back.RemovedVertices) != len(d.AddedVertices) ||
			len(back.AddedEdges) != len(d.RemovedEdges) || len(back.RemovedEdges) != len(d.AddedEdges) ||
			len(back.ChangedVertices) != len(d.ChangedVertices) || len(back.ChangedEdges) != len(d.ChangedEdges) {
			t.Fatalf("seed %d: the diff back doesn't mirror the diff:\n%s\n%s", seed, d, back)
		}
		restored, err := back.Apply(b)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if rest, _ := Diff(restored, a); !rest.IsEmpty() {
			t.Fatalf("seed %d: the diff back doesn't restore a:\n%s", seed, rest)
		}
	}
}

func TestGraphDiff_ApplyKeepsAcyclicOrder(t *testing.T) {
	b := graphSpec{opts: []GraphOptionFunc{Acyclic()}, edges: []EdgeEntry[string]{ed("A", "B", 0, ""), ed("B", "C", 0, "")}}.build(t)
	for _, opts := range [][]GraphOptionFunc{{Directed()}, {Acyclic()}} {
		a := graphSpec{opts: opts, edges: []EdgeEntry[string]{ed("A", "B", 0, "")}}.build(t)
		d, err := Diff(a, b)
		if err != nil {
			t.Fatal(err)
		}
		c, err := d.Apply(a)
		if err != nil {
			t.Fatal(err)
		}
		if !c.IsAcyclic() {
			t.Fatal("the result isn't acyclic")
		}
		if _, err := c.AddEdge(c.GetVertexByID("C"), c.GetVertexByID("A")); !errors.Is(err, ErrDAGCycle) {
			t.Fatalf("an edge that closes a cycle gave %v", err)
		}
		if _, err := c.AddEdge(c.GetVertexByID("A"), c.GetVertexByID("C")); err != nil {
			t.Fatal(err)
		}
	}

	a := graphSpec{opts: []GraphOptionFunc{Directed()}, edges: []EdgeEntry[string]{ed("A", "B", 0, "")}}.build(t)
	back, err := Diff(b, a)
	if err != nil {
		t.Fatal(err)
	}
	c, err := back.Apply(b)
	if err != nil {
		t.Fatal(err)
	}
	if c.IsAcyclic() {
		t.Fatal("the result is still acyclic")
	}
	if _, err := c.AddEdge(c.GetVertexByID("B"), c.GetVertexByID("A")); err != nil {
		t.Fatalf("an edge that closes a cycle gave %v", err)
	}
}

func TestGraphDiff_ApplyUndirectedChange(t *testing.T) {
	a := graphSpec{edges: []EdgeEntry[string]{ed("A", "B", 1, "")}}.build(t)
	d := &GraphDiff[string]{ChangedEdges: []EdgeChange[string]{{From: "B", To: "A", OldWeight: 1, NewWeight: 2, NewLabel: "x"}}}
	c, err := d.Apply(a)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range []*Edge[string]{c.GetEdge(c.GetVertexByID("A"), c.GetVertexByID("B")), c.GetEdge(c.GetVertexByID("B"), c.GetVertexByID("A"))} {
		if e.Weight() != 2 || e.Label() != "x" {
			t.Fatalf("edge %v -> %v: weight %v, label %q", e.Source().Label(), e.Destination().Label(), e.Weight(), e.Label())
		}
	}
}

func TestGraphDiff_ApplyErrors(t *testing.T) {
	directed := graphSpec{
		opts:     []GraphOptionFunc{Directed()},
		vertices: []VertexEntry[string]{vx("C", 0)},
		edges:    []EdgeEntry[string]{ed("A", "B", 0, "")},
	}
	acyclic := graphSpec{opts: []GraphOptionFunc{Acyclic()}, edges: []EdgeEntry[string]{ed("A", "B", 0, "")}}
	cycle := graphSpec{opts: []GraphOptionFunc{Directed()}, edges: []EdgeEntry[string]{ed("A", "B", 0, ""), ed("B", "A", 0, "")}}

	tests := []struct {
		name string
		base graphSpec
		d    GraphDiff[string]
		err  error
		text string
	}{
		{name: "removed edge missing", base: directed, d: GraphDiff[string]{RemovedEdges: []EdgeEntry[string]{ed("B", "A", 0, "")}}, err: ErrDiffMismatch, text: "edge B -> A isn't in the graph"},
		{name: "removed vertex missing", base: directed, d: GraphDiff[string]{RemovedVertices: []VertexEntry[string]{vx("D", 0)}}, err: ErrDiffMismatch, text: "vertex D isn't in the graph"},
		{name: "added vertex present", base: directed, d: GraphDiff[string]{AddedVertices: []VertexEntry[string]{vx("A", 0)}}, err: ErrDiffMismatch, text: "vertex A is already in the graph"},
		{name: "added edge from a missing vertex", base: directed, d: GraphDiff[string]{AddedEdges: []EdgeEntry[string]{ed("D", "A", 0, "")}}, err: ErrDiffMismatch, text: "vertex D of edge D -> A isn't in the graph"},
		{name: "added edge to a missing vertex", base: directed, d: GraphDiff[string]{AddedEdges: []EdgeEntry[string]{ed("A", "D", 0, "")}}, err: ErrDiffMismatch, text: "vertex D of edge A -> D isn't in the graph"},
		{name: "added edge present", base: directed, d: GraphDiff[string]{AddedEdges: []EdgeEntry[string]{ed("A", "B", 0, "")}}, err: ErrDiffMismatch, text: "edge A -> B is already in the graph"},
		{name: "changed vertex missing", base: directed, d: GraphDiff[string]{ChangedVertices: []VertexChange[string]{{Label: "D"}}}, err: ErrDiffMismatch, text: "vertex D isn't in the graph"},
		{name: "changed edge missing", base: directed, d: GraphDiff[string]{ChangedEdges: []EdgeChange[string]{{From: "B", To: "A"}}}, err: ErrDiffMismatch, text: "edge B -> A isn't in the graph"},
		{
			name: "removed vertex removed twice",
			base: directed,
			d:    GraphDiff[string]{RemovedVertices: []VertexEntry[string]{vx("C", 0), vx("C", 0)}},
			err:  ErrDiffMismatch,
			text: "vertex C isn't in the graph",
		},
		{name: "only the new properties", base: directed, d: GraphDiff[string]{NewProperties: &PropertiesSummary{Directed: true}}, err: ErrDiffMismatch, text: "aren't both set"},
		{name: "only the old properties", base: directed, d: GraphDiff[string]{OldProperties: &PropertiesSummary{Directed: true}}, err: ErrDiffMismatch, text: "aren't both set"},
		{
			name: "other old properties",
			base: directed,
			d:    GraphDiff[string]{OldProperties: &PropertiesSummary{Directed: true, Weighted: true}, NewProperties: &PropertiesSummary{Directed: true}},
			err:  ErrDiffMismatch,
			text: "the graph has the properties {Directed:true Weighted:false Acyclic:false}, not {Directed:true Weighted:true Acyclic:false}",
		},
		{
			name: "properties that change direction",
			base: directed,
			d:    GraphDiff[string]{OldProperties: &PropertiesSummary{Directed: true}, NewProperties: &PropertiesSummary{}},
			err:  ErrIncompatibleGraphs,
		},
		{
			name: "acyclic and undirected properties",
			base: graphSpec{},
			d:    GraphDiff[string]{OldProperties: &PropertiesSummary{}, NewProperties: &PropertiesSummary{Acyclic: true}},
			err:  ErrIncompatibleGraphs,
		},
		{name: "edge that closes a cycle", base: acyclic, d: GraphDiff[string]{AddedEdges: []EdgeEntry[string]{ed("B", "A", 0, "")}}, err: ErrDAGCycle},
		{name: "self-loop in an acyclic graph", base: acyclic, d: GraphDiff[string]{AddedEdges: []EdgeEntry[string]{ed("A", "A", 0, "")}}, err: ErrDAGCycle},
		{
			name: "edges that close a cycle in a graph made acyclic",
			base: directed,
			d: GraphDiff[string]{
				AddedEdges:    []EdgeEntry[string]{ed("B", "A", 0, "")},
				OldProperties: &PropertiesSummary{Directed: true},
				NewProperties: &PropertiesSummary{Directed: true, Acyclic: true},
			},
			err: ErrDAGHasCycle,
		},
		{
			name: "acyclic graph with a cycle",
			base: cycle,
			d:    GraphDiff[string]{OldProperties: &PropertiesSummary{Directed: true}, NewProperties: &PropertiesSummary{Directed: true, Acyclic: true}},
			err:  ErrDAGHasCycle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := tt.base.build(t)
			before := fullState(base)
			got, err := tt.d.Apply(base)
			if !errors.Is(err, tt.err) || got != nil {
				t.Fatalf("got %v, %v, want %v", got, err, tt.err)
			}
			if !strings.Contains(err.Error(), tt.text) {
				t.Fatalf("error %q doesn't contain %q", err, tt.text)
			}
			if fullState(base) != before {
				t.Fatal("Apply changed base")
			}
		})
	}
}

func TestGraphDiff_ApplyNilDiff(t *testing.T) {
	base := graphSpec{opts: []GraphOptionFunc{Acyclic(), Weighted()}, edges: []EdgeEntry[string]{ed("A", "B", 2, "x")}}.build(t)
	var d *GraphDiff[string]
	for _, source := range []Graph[string]{base, cloneWrapper[string]{base}} {
		got, err := d.Apply(source)
		if err != nil || got == base || fullState(got) != fullState(base) {
			t.Fatalf("got %v, %v", got, err)
		}
	}

	defer func() {
		if recover() == nil {
			t.Fatal("no panic for a nil base")
		}
	}()
	_, _ = (&GraphDiff[string]{}).Apply(nil)
}

func TestGraphDiff_JSON(t *testing.T) {
	a := graphSpec{opts: []GraphOptionFunc{Directed()}, edges: []EdgeEntry[string]{ed("A", "B", 1, "")}}.build(t)
	b := graphSpec{opts: []GraphOptionFunc{Directed(), Weighted()}, edges: []EdgeEntry[string]{ed("A", "C", 2, "x")}}.build(t)
	d, err := Diff(a, b)
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"AddedVertices":[{"Label":"C","Weight":0}],"RemovedVertices":[{"Label":"B","Weight":0}],` +
		`"AddedEdges":[{"From":"A","To":"C","Weight":2,"Label":"x"}],"RemovedEdges":[{"From":"A","To":"B","Weight":1,"Label":""}],` +
		`"OldProperties":{"Directed":true,"Weighted":false,"Acyclic":false},"NewProperties":{"Directed":true,"Weighted":true,"Acyclic":false}}`
	if string(data) != want {
		t.Fatalf("got  %s\nwant %s", data, want)
	}

	var decoded GraphDiff[string]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&decoded, d) {
		t.Fatalf("decoded %+v, want %+v", decoded, d)
	}
	c, err := decoded.Apply(a)
	if err != nil {
		t.Fatal(err)
	}
	if rest, _ := Diff(c, b); !rest.IsEmpty() {
		t.Fatalf("the decoded diff leaves:\n%s", rest)
	}
}

// diffBenchGraphs returns a directed graph with 10,000 vertices and 50,000
// edges, and a copy with 1% of its vertex weights changed, 500 of its edges
// removed and 500 new ones. The edges of an acyclic graph go from a higher
// label to a lower one, against the order the vertices were added in.
func diffBenchGraphs(opts ...GraphOptionFunc) (a, b Graph[int]) {
	const n, m = 10000, 50000
	rng := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // seeded so runs are comparable
	a = New[int](append([]GraphOptionFunc{Directed(), Weighted()}, opts...)...)
	for i := range n {
		a.AddVertexByLabel(i)
	}
	addEdge := func(g Graph[int]) bool {
		u, v := rng.IntN(n), rng.IntN(n)
		if g.IsAcyclic() && u <= v {
			return false
		}
		_, err := g.AddEdge(g.GetVertexByID(u), g.GetVertexByID(v), WithEdgeWeight(float64(rng.IntN(9))))
		return err == nil
	}
	for a.Size() < m {
		addEdge(a)
	}

	b = Clone(a)
	for i := 0; i < n; i += 100 {
		b.GetVertexByID(i).properties.weight = 1
	}
	edges := b.AllEdges()
	for i := 0; i < 500; i++ {
		b.RemoveEdges(edges[i*97])
	}
	for added := 0; added < 500; {
		if addEdge(b) {
			added++
		}
	}
	return a, b
}

func BenchmarkDiff(b *testing.B) {
	x, y := diffBenchGraphs()
	b.Run("identical", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if d, _ := Diff(x, x); !d.IsEmpty() {
				b.Fatal("not empty")
			}
		}
	})
	b.Run("changed", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if d, _ := Diff(x, y); d.IsEmpty() {
				b.Fatal("empty")
			}
		}
	})
}

func BenchmarkGraphDiff_Apply(b *testing.B) {
	for _, bench := range []struct {
		name  string
		opts  []GraphOptionFunc
		empty bool
	}{
		{name: "empty", empty: true},
		{name: "directed"},
		{name: "acyclic", opts: []GraphOptionFunc{Acyclic()}},
	} {
		x, y := diffBenchGraphs(bench.opts...)
		d, err := Diff(x, y)
		if err != nil {
			b.Fatal(err)
		}
		if bench.empty {
			d = &GraphDiff[int]{}
		}
		b.Run(bench.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := d.Apply(x); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
