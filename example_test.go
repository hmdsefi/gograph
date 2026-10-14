package gograph_test

import (
	"errors"
	"fmt"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/path"
	"github.com/hmdsefi/gograph/traverse"
)

func Example() {
	// An edge A -> B means A has to happen before B.
	g := gograph.New[string](gograph.Acyclic())

	checkout := g.AddVertexByLabel("checkout")
	build := g.AddVertexByLabel("build")
	test := g.AddVertexByLabel("test")
	release := g.AddVertexByLabel("release")

	_, _ = g.AddEdge(checkout, build)
	_, _ = g.AddEdge(build, test)
	_, _ = g.AddEdge(test, release)

	// Acyclic graphs reject any edge that would create a cycle.
	_, err := g.AddEdge(release, checkout)
	fmt.Println(errors.Is(err, gograph.ErrDAGCycle))

	order, _ := gograph.TopologySort(g)
	for _, v := range order {
		fmt.Println(v.Label())
	}

	// Output:
	// true
	// checkout
	// build
	// test
	// release
}

func ExampleNew_directed() {
	g := gograph.New[int](gograph.Directed())

	_, _ = g.AddEdge(gograph.NewVertex(1), gograph.NewVertex(2))
	_, _ = g.AddEdge(gograph.NewVertex(1), gograph.NewVertex(3))
	_, _ = g.AddEdge(gograph.NewVertex(2), gograph.NewVertex(4))
	_, _ = g.AddEdge(gograph.NewVertex(3), gograph.NewVertex(4))
	_, _ = g.AddEdge(gograph.NewVertex(4), gograph.NewVertex(5))
	_, _ = g.AddEdge(gograph.NewVertex(5), gograph.NewVertex(6))

	fmt.Println(g.Order(), g.Size())
	for _, e := range g.AllEdges() {
		fmt.Println(e.Source().Label(), "->", e.Destination().Label())
	}

	// Output:
	// 6 6
	// 1 -> 2
	// 1 -> 3
	// 2 -> 4
	// 3 -> 4
	// 4 -> 5
	// 5 -> 6
}

func ExampleNew_acyclic() {
	g := gograph.New[int](gograph.Acyclic())

	_, _ = g.AddEdge(gograph.NewVertex(1), gograph.NewVertex(2))
	_, _ = g.AddEdge(gograph.NewVertex(2), gograph.NewVertex(3))

	_, err := g.AddEdge(gograph.NewVertex(3), gograph.NewVertex(1))
	fmt.Println(err)

	// Output:
	// edges would create cycle
}

func ExampleNew_undirected() {
	// Graphs are undirected by default.
	g := gograph.New[string]()

	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	d := g.AddVertexByLabel("D")

	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(a, d)
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(b, d)

	// Every undirected edge can be followed both ways.
	fmt.Println(g.ContainsEdge(a, b), g.ContainsEdge(b, a))

	// Output:
	// true true
}

func ExampleNew_weighted() {
	g := gograph.New[string](gograph.Weighted())

	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	d := g.AddVertexByLabel("D")

	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(4))
	_, _ = g.AddEdge(a, d, gograph.WithEdgeWeight(3))
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(3))
	_, _ = g.AddEdge(b, d, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(c, d, gograph.WithEdgeWeight(2))

	dist := path.Dijkstra(g, "A")
	fmt.Println(dist["C"])

	// Output:
	// 5
}

func Example_breadthFirst() {
	g := gograph.New[string](gograph.Directed())

	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("C"))
	_, _ = g.AddEdge(gograph.NewVertex("B"), gograph.NewVertex("D"))

	it, err := traverse.NewBreadthFirstIterator(g, "A")
	if err != nil {
		fmt.Println(err)
		return
	}

	for it.HasNext() {
		fmt.Println(it.Next().Label())
	}

	// Output:
	// A
	// B
	// C
	// D
}

func ExampleWithVertexWeight() {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())

	a := g.AddVertexByLabel("A", gograph.WithVertexWeight(3))
	b := g.AddVertexByLabel("B", gograph.WithVertexWeight(2))
	c := g.AddVertexByLabel("C", gograph.WithVertexWeight(4))

	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(b, c)

	fmt.Println(a.Weight(), b.Weight(), c.Weight())

	// Output:
	// 3 2 4
}

func ExampleWithEdgeLabel() {
	g := gograph.New[string](gograph.Directed())

	edge, _ := g.AddEdge(
		gograph.NewVertex("service"),
		gograph.NewVertex("database"),
		gograph.WithEdgeLabel("queries"),
	)

	fmt.Println(edge.Label())

	// Output:
	// queries
}
