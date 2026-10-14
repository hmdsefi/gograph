package connectivity

import (
	"errors"
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

func directedGraph(n int, edges ...[2]int) gograph.Graph[int] {
	g := gograph.New[int](gograph.Directed())
	for i := 0; i < n; i++ {
		g.AddVertexByLabel(i)
	}
	for _, e := range edges {
		_, _ = g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1]))
	}
	return g
}

func mustCondense(t *testing.T, g gograph.Graph[int], opts ...CondenseOption[int]) *Condensation[int] {
	t.Helper()
	c, err := Condense(g, opts...)
	if err != nil {
		t.Fatalf("Condense: unexpected error: %v", err)
	}
	checkCondensation(t, g, c)
	return c
}

// checkCondensation checks the parts of the result that hold for any graph:
// the members partition the vertices, ComponentOf agrees with them, the
// edges are exactly the pairs of components that an edge of g crosses
// between, and every edge goes from a lower label to a higher one.
func checkCondensation(t *testing.T, g gograph.Graph[int], c *Condensation[int]) {
	t.Helper()
	if int(c.Graph.Order()) != len(c.Members) {
		t.Fatalf("Graph has %d vertices for %d components", c.Graph.Order(), len(c.Members))
	}
	if !c.Graph.IsDirected() {
		t.Fatal("Graph is not directed")
	}

	seen := 0
	for i, members := range c.Members {
		if c.Graph.GetVertexByID(i) == nil {
			t.Fatalf("Graph has no vertex %d", i)
		}
		if len(members) == 0 {
			t.Fatalf("component %d has no members", i)
		}
		for _, v := range members {
			if v != g.GetVertexByID(v.Label()) {
				t.Fatalf("Members has a copy of %d", v.Label())
			}
			if comp, ok := c.ComponentOf[v.Label()]; !ok || comp != i {
				t.Fatalf("%d is a member of %d, but ComponentOf has %d", v.Label(), i, comp)
			}
			seen++
		}
	}
	if seen != int(g.Order()) || len(c.ComponentOf) != int(g.Order()) {
		t.Fatalf("components have %d vertices and ComponentOf %d, want %d", seen, len(c.ComponentOf), g.Order())
	}

	want := make(map[[2]int]bool)
	for _, e := range g.AllEdges() {
		from, to := c.ComponentOf[e.Source().Label()], c.ComponentOf[e.Destination().Label()]
		if from != to {
			want[[2]int{from, to}] = true
		}
	}
	got := make(map[[2]int]bool)
	for _, e := range c.Graph.AllEdges() {
		from, to := e.Source().Label(), e.Destination().Label()
		if from >= to {
			t.Fatalf("edge %d -> %d doesn't go from a lower label to a higher one", from, to)
		}
		got[[2]int{from, to}] = true
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Graph has edges %v, want %v", got, want)
	}

	if _, err := gograph.TopologySort(c.Graph); err != nil {
		t.Fatalf("TopologySort(Graph): %v", err)
	}
}

func memberLabels(c *Condensation[int]) [][]int {
	return sccLabels(c.Members)
}

func TestCondense_DAG(t *testing.T) {
	g := directedGraph(4, [2]int{0, 1}, [2]int{0, 2}, [2]int{1, 3}, [2]int{2, 3})
	c := mustCondense(t, g)

	for _, members := range c.Members {
		if len(members) != 1 {
			t.Fatalf("Members = %v, want one vertex per component", memberLabels(c))
		}
	}
	if len(c.Members) != 4 || c.Graph.Size() != g.Size() {
		t.Fatalf("Graph has %d edges, want %d", c.Graph.Size(), g.Size())
	}
}

func TestCondense_OneCycle(t *testing.T) {
	g := directedGraph(4, [2]int{2, 3}, [2]int{3, 0}, [2]int{0, 1}, [2]int{1, 2})
	c := mustCondense(t, g)

	if got := memberLabels(c); !reflect.DeepEqual(got, [][]int{{0, 1, 2, 3}}) {
		t.Fatalf("Members = %v, want [[0 1 2 3]]", got)
	}
	if c.Graph.Size() != 0 {
		t.Fatalf("Graph has %d edges, want 0", c.Graph.Size())
	}
}

func TestCondense_TwoCyclesJoined(t *testing.T) {
	// two edges cross from the first cycle to the second
	g := directedGraph(5,
		[2]int{3, 4}, [2]int{4, 3},
		[2]int{0, 1}, [2]int{1, 2}, [2]int{2, 0},
		[2]int{2, 3}, [2]int{1, 4},
	)
	c := mustCondense(t, g)

	if got := memberLabels(c); !reflect.DeepEqual(got, [][]int{{0, 1, 2}, {3, 4}}) {
		t.Fatalf("Members = %v, want [[0 1 2] [3 4]]", got)
	}
	if c.Graph.Size() != 1 {
		t.Fatalf("Graph has %d edges, want 1", c.Graph.Size())
	}
}

func TestCondense_CycleWithTail(t *testing.T) {
	g := directedGraph(5, [2]int{0, 1}, [2]int{1, 2}, [2]int{2, 0}, [2]int{3, 0}, [2]int{2, 4})
	c := mustCondense(t, g)

	if got := memberLabels(c); !reflect.DeepEqual(got, [][]int{{3}, {0, 1, 2}, {4}}) {
		t.Fatalf("Members = %v, want [[3] [0 1 2] [4]]", got)
	}
	if want := map[int]int{0: 1, 1: 1, 2: 1, 3: 0, 4: 2}; !reflect.DeepEqual(c.ComponentOf, want) {
		t.Fatalf("ComponentOf = %v, want %v", c.ComponentOf, want)
	}
}

func TestCondense_SelfLoops(t *testing.T) {
	g := directedGraph(2, [2]int{0, 0}, [2]int{0, 1}, [2]int{1, 1})
	c := mustCondense(t, g)

	if got := memberLabels(c); !reflect.DeepEqual(got, [][]int{{0}, {1}}) {
		t.Fatalf("Members = %v, want [[0] [1]]", got)
	}
	if c.Graph.Size() != 1 {
		t.Fatalf("Graph has %d edges, want 1", c.Graph.Size())
	}
}

func TestCondense_Weights(t *testing.T) {
	g := gograph.New[int](gograph.Directed(), gograph.Weighted())
	for i, w := range []float64{1, 2, 4, 8} {
		g.AddVertexByLabel(i, gograph.WithVertexWeight(w))
	}
	for _, e := range []struct {
		from, to int
		weight   float64
	}{{0, 1, 5}, {1, 0, 1}, {0, 2, 7}, {1, 2, 3}, {2, 3, 2}} {
		_, _ = g.AddEdge(g.GetVertexByID(e.from), g.GetVertexByID(e.to), gograph.WithEdgeWeight(e.weight))
	}

	var crossings [][]float64
	c := mustCondense(t, g,
		WithComponentWeight(func(members []*gograph.Vertex[int]) float64 {
			var sum float64
			for _, v := range members {
				sum += v.Weight()
			}
			return sum
		}),
		WithCrossingWeight(func(crossing []*gograph.Edge[int]) float64 {
			weights := make([]float64, len(crossing))
			for i, e := range crossing {
				weights[i] = e.Weight()
			}
			crossings = append(crossings, weights)
			return math.Min(weights[0], weights[len(weights)-1])
		}),
	)

	if got := memberLabels(c); !reflect.DeepEqual(got, [][]int{{0, 1}, {2}, {3}}) {
		t.Fatalf("Members = %v, want [[0 1] [2] [3]]", got)
	}
	if !c.Graph.IsWeighted() {
		t.Fatal("Graph is not weighted")
	}
	for label, want := range []float64{3, 4, 8} {
		if got := c.Graph.GetVertexByID(label).Weight(); got != want {
			t.Fatalf("component %d has weight %v, want %v", label, got, want)
		}
	}
	if want := [][]float64{{7, 3}, {2}}; !reflect.DeepEqual(crossings, want) {
		t.Fatalf("crossing edges have weights %v, want %v", crossings, want)
	}
	for _, e := range []struct {
		from, to int
		weight   float64
	}{{0, 1, 3}, {1, 2, 2}} {
		edge := c.Graph.GetEdge(c.Graph.GetVertexByID(e.from), c.Graph.GetVertexByID(e.to))
		if edge == nil || edge.Weight() != e.weight {
			t.Fatalf("edge %d -> %d = %v, want weight %v", e.from, e.to, edge, e.weight)
		}
	}
}

func TestCondense_NoOptions(t *testing.T) {
	g := gograph.New[int](gograph.Directed(), gograph.Weighted())
	_, _ = g.AddEdge(gograph.NewVertex(0, gograph.WithVertexWeight(3)), gograph.NewVertex(1), gograph.WithEdgeWeight(5))
	c := mustCondense(t, g)

	if c.Graph.IsWeighted() {
		t.Fatal("Graph is weighted without WithCrossingWeight")
	}
	if w := c.Graph.GetVertexByID(0).Weight(); w != 0 {
		t.Fatalf("component 0 has weight %v without WithComponentWeight", w)
	}
}

func TestCondense_Empty(t *testing.T) {
	c := mustCondense(t, gograph.New[int](gograph.Directed()))
	if len(c.Members) != 0 || c.Graph.Order() != 0 {
		t.Fatalf("Condense of an empty graph has %d components", len(c.Members))
	}
}

func TestCondense_Undirected(t *testing.T) {
	g := gograph.New[int]()
	_, _ = g.AddEdge(gograph.NewVertex(0), gograph.NewVertex(1))
	if c, err := Condense(g); !errors.Is(err, gograph.ErrNotDirected) || c != nil {
		t.Fatalf("Condense = %v, %v, want nil, %v", c, err, gograph.ErrNotDirected)
	}
}

func TestCondense_RandomGraphs(t *testing.T) {
	rng := rand.New(rand.NewSource(3)) //nolint:gosec // seeded so failures are reproducible
	for run := 0; run < 300; run++ {
		n := 1 + rng.Intn(15)
		g := gograph.New[int](gograph.Directed())
		for _, label := range rng.Perm(n) {
			g.AddVertexByLabel(label)
		}
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if rng.Intn(6) == 0 {
					_, _ = g.AddEdge(g.GetVertexByID(i), g.GetVertexByID(j))
				}
			}
		}

		c := mustCondense(t, g)
		if got, want := partitionKey(c.Members), partitionKey(Kosaraju(g)); got != want {
			t.Fatalf("run %d: Members = %s, Kosaraju = %s", run, got, want)
		}

		// members follow GetAllVertices
		position := make(map[int]int)
		for i, v := range g.GetAllVertices() {
			position[v.Label()] = i
		}
		for _, members := range c.Members {
			for i := 1; i < len(members); i++ {
				if position[members[i-1].Label()] > position[members[i].Label()] {
					t.Fatalf("run %d: members %v don't follow GetAllVertices", run, sccLabels(c.Members))
				}
			}
		}

		again := mustCondense(t, g)
		if !reflect.DeepEqual(memberLabels(again), memberLabels(c)) {
			t.Fatalf("run %d: two calls returned %v and %v", run, memberLabels(c), memberLabels(again))
		}
	}
}

func TestCondense_LongCycleAndChain(t *testing.T) {
	const n = 100_000
	g := gograph.New[int](gograph.Directed())
	prev := g.AddVertexByLabel(0)
	for i := 1; i < 2*n; i++ {
		v := g.AddVertexByLabel(i)
		_, _ = g.AddEdge(prev, v)
		prev = v
	}
	// the first n vertices form a cycle, followed by a chain of n vertices
	_, _ = g.AddEdge(g.GetVertexByID(n-1), g.GetVertexByID(0))

	c, err := Condense(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Members) != n+1 || len(c.Members[0]) != n || c.Graph.Size() != n {
		t.Fatalf("got %d components, %d members in the first, %d edges",
			len(c.Members), len(c.Members[0]), c.Graph.Size())
	}
}
