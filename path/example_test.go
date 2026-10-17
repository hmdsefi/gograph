package path_test

import (
	"errors"
	"fmt"
	"math"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/path"
)

func ExampleDijkstra_unreachableDistances() {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	g.AddVertexByLabel("A")
	g.AddVertexByLabel("B")
	fmt.Println(path.Dijkstra(g, "A")["B"] == math.MaxFloat64)
	fmt.Println(path.DijkstraSimple(g, "A")["B"] == math.MaxFloat64)
	dist, err := path.BellmanFord(g, "A")
	fmt.Println(err == nil, math.IsInf(dist["B"], 1))
	all, err := path.FloydWarshall(g)
	fmt.Println(err == nil, math.IsInf(all["A"]["B"], 1))
	// Output:
	// true
	// true
	// true true
	// true true
}

func ExampleDijkstra_weightedFlag() {
	g := gograph.New[string](gograph.Directed())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(3))
	fmt.Println(path.Dijkstra(g, "A")["B"])
	_, err := path.BellmanFord(g, "A")
	fmt.Println(errors.Is(err, path.ErrNotWeighted))
	// Output:
	// 3
	// true
}

func ExampleBellmanFord_negativeCycle() {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(-1))
	_, _ = g.AddEdge(c, b, gograph.WithEdgeWeight(-1))
	dist, err := path.BellmanFord(g, "A")
	fmt.Println(err == nil, math.IsInf(dist["B"], 1))
	_, err = path.FloydWarshall(g)
	fmt.Println(errors.Is(err, path.ErrNegativeWeightCycle))
	// Output:
	// true true
	// true
}
