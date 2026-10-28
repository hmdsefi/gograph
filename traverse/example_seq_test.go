package traverse_test

import (
	"fmt"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/traverse"
)

func ExampleBFS() {
	g := gograph.New[string](gograph.Directed())
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("C"))
	_, _ = g.AddEdge(gograph.NewVertex("B"), gograph.NewVertex("D"))
	_, _ = g.AddEdge(gograph.NewVertex("C"), gograph.NewVertex("E"))

	for v := range traverse.BFS(g, "A") {
		fmt.Println(v.Label())
		if v.Label() == "C" {
			break
		}
	}
	// Output:
	// A
	// B
	// C
}

func ExampleDFS() {
	g := gograph.New[string](gograph.Directed())
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("C"))
	_, _ = g.AddEdge(gograph.NewVertex("B"), gograph.NewVertex("D"))

	// the latest vertex found comes next, so C comes before B
	for v := range traverse.DFS(g, "A") {
		fmt.Println(v.Label())
	}
	// Output:
	// A
	// C
	// B
	// D
}

func ExampleBFSWithDepth() {
	g := gograph.New[string](gograph.Directed())
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("C"))
	_, _ = g.AddEdge(gograph.NewVertex("B"), gograph.NewVertex("D"))
	_, _ = g.AddEdge(gograph.NewVertex("C"), gograph.NewVertex("D"))

	for v, depth := range traverse.BFSWithDepth(g, "A") {
		fmt.Println(v.Label(), depth)
	}
	// Output:
	// A 0
	// B 1
	// C 1
	// D 2
}

func ExampleClosestFirst() {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	_, _ = g.AddEdge(gograph.NewVertex("home"), gograph.NewVertex("school"), gograph.WithEdgeWeight(4))
	_, _ = g.AddEdge(gograph.NewVertex("home"), gograph.NewVertex("bakery"), gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(gograph.NewVertex("bakery"), gograph.NewVertex("school"), gograph.WithEdgeWeight(2))

	for v, dist := range traverse.ClosestFirst(g, "home") {
		fmt.Println(v.Label(), dist)
	}
	// Output:
	// home 0
	// bakery 1
	// school 3
}

func ExampleTopological() {
	g := gograph.New[string](gograph.Acyclic())
	_, _ = g.AddEdge(gograph.NewVertex("fetch"), gograph.NewVertex("build"))
	_, _ = g.AddEdge(gograph.NewVertex("build"), gograph.NewVertex("test"))
	_, _ = g.AddEdge(gograph.NewVertex("fetch"), gograph.NewVertex("lint"))

	steps, err := traverse.Topological(g)
	if err != nil {
		fmt.Println(err)
		return
	}
	for v := range steps {
		fmt.Println(v.Label())
	}
	// Output:
	// fetch
	// build
	// lint
	// test
}
