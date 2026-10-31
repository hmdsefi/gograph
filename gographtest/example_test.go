package gographtest_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/connectivity"
	"github.com/hmdsefi/gograph/gographtest"
)

// printer is a testing.TB that prints each failure, standing in for the t
// of a real test.
type printer struct{ testing.TB }

func (printer) Helper() {}

func (printer) Errorf(format string, args ...any) { fmt.Printf(format+"\n", args...) }

func ExampleIsTopologicalOrder() {
	t := printer{}

	g := gograph.New[string](gograph.Directed())
	shirt, tie, jacket := gograph.NewVertex("shirt"), gograph.NewVertex("tie"), gograph.NewVertex("jacket")
	_, _ = g.AddEdge(shirt, tie)
	_, _ = g.AddEdge(tie, jacket)

	order, _ := gograph.TopologySort(g)
	gographtest.IsTopologicalOrder(t, g, order)

	gographtest.IsTopologicalOrder(t, g, []*gograph.Vertex[string]{tie, shirt, jacket})

	// Output:
	// gographtest: not a topological order of the graph:
	// 	edge shirt -> tie goes from position 1 to 0
}

func ExampleSameGroups() {
	t := printer{}

	g := gograph.New[string]()
	_, _ = g.AddEdge(gograph.NewVertex("a"), gograph.NewVertex("b"))
	_, _ = g.AddEdge(gograph.NewVertex("c"), gograph.NewVertex("d"))
	g.AddVertexByLabel("e")

	components := connectivity.ConnectedComponents(g)
	gographtest.SameGroups(t, [][]string{{"e"}, {"d", "c"}, {"b", "a"}}, components)

	gographtest.SameGroups(t, [][]string{{"a", "b", "c", "d"}, {"e"}}, components)

	// Output:
	// gographtest: groups differ:
	// 	missing [a b c d]
	// 	unexpected [a b]
	// 	unexpected [c d]
}

func ExampleEqual() {
	t := printer{}

	want := gograph.New[string](gograph.Directed(), gograph.Weighted())
	_, _ = want.AddEdge(gograph.NewVertex("a"), gograph.NewVertex("b"), gograph.WithEdgeWeight(0.3))
	_, _ = want.AddEdge(gograph.NewVertex("b"), gograph.NewVertex("c"), gograph.WithEdgeWeight(1))

	// 0.1 + 0.2 at run time is 0.30000000000000004, close to 0.3
	x, y := 0.1, 0.2
	got := gograph.Clone(want)
	got.RemoveEdges(got.GetEdge(got.GetVertexByID("a"), got.GetVertexByID("b")))
	_, _ = got.AddEdge(got.GetVertexByID("a"), got.GetVertexByID("b"), gograph.WithEdgeWeight(x+y))

	gographtest.Equal(t, want, got, gograph.WithWeightTolerance(1e-9))

	got.RemoveVertices(got.GetVertexByID("c"))
	gographtest.Equal(t, want, got, gograph.WithWeightTolerance(1e-9))

	// Output:
	// gographtest: graphs differ:
	// 	- vertex c
	// 	- edge b -> c weight 1
}

func ExampleGrid() {
	g := gographtest.Grid(3, 4, gograph.Directed())
	fmt.Println(g.Order(), "vertices,", g.Size(), "edges")

	order, _ := gograph.TopologySort(g)
	gographtest.IsTopologicalOrder(printer{}, g, order)

	// Output:
	// 12 vertices, 17 edges
}

func ExampleRandomDAG() {
	rng := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // seeded so the output is the same on every run
	g := gographtest.RandomDAG(1000, 0.01, rng)

	fmt.Println(g.Order(), "vertices, acyclic:", !gograph.HasCycle(g))
	order, _ := gograph.TopologySort(g)
	gographtest.IsTopologicalOrder(printer{}, g, order)

	// Output:
	// 1000 vertices, acyclic: true
}
