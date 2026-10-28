package gograph_test

import (
	"fmt"

	"github.com/hmdsefi/gograph"
)

func ExampleEdges() {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"), gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("C"), gograph.WithEdgeWeight(5))
	_, _ = g.AddEdge(gograph.NewVertex("B"), gograph.NewVertex("C"), gograph.WithEdgeWeight(1))

	total := 0.0
	for e := range gograph.Edges(g) {
		total += e.Weight()
	}
	fmt.Println(total)
	// Output: 8
}

func ExampleVertices() {
	g := gograph.New[string](gograph.Directed())
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("C"))
	_, _ = g.AddEdge(gograph.NewVertex("C"), gograph.NewVertex("B"))

	// the vertices with no outgoing edges
	for v := range gograph.Vertices(g) {
		if v.OutDegree() == 0 {
			fmt.Println(v.Label())
		}
	}
	// Output: B
}
