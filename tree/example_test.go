package tree_test

import (
	"fmt"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/tree"
)

func ExampleKruskal() {
	// The cost of laying a cable between two offices. The cheapest network
	// that connects every office uses three of the five cables.
	g := gograph.New[string](gograph.Weighted())
	hq := g.AddVertexByLabel("hq")
	north := g.AddVertexByLabel("north")
	south := g.AddVertexByLabel("south")
	east := g.AddVertexByLabel("east")
	_, _ = g.AddEdge(hq, north, gograph.WithEdgeWeight(4))
	_, _ = g.AddEdge(hq, south, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(north, south, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(north, east, gograph.WithEdgeWeight(5))
	_, _ = g.AddEdge(south, east, gograph.WithEdgeWeight(7))

	edges, total, err := tree.Kruskal(g)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, e := range edges {
		fmt.Println(e.Source().Label(), e.Destination().Label(), e.Weight())
	}
	fmt.Println("total", total)
	// Output:
	// hq south 1
	// north south 2
	// north east 5
	// total 8
}

func ExamplePrim() {
	// Two islands with no cable between them give one tree each.
	g := gograph.New[string](gograph.Weighted())
	a, b, c := g.AddVertexByLabel("a"), g.AddVertexByLabel("b"), g.AddVertexByLabel("c")
	x, y := g.AddVertexByLabel("x"), g.AddVertexByLabel("y")
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(3))
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(a, c, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(x, y, gograph.WithEdgeWeight(6))

	edges, total, err := tree.Prim(g)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, e := range edges {
		fmt.Println(e.Source().Label(), e.Destination().Label(), e.Weight())
	}
	fmt.Println("total", total)
	// Output:
	// b c 1
	// a c 2
	// x y 6
	// total 9
}
