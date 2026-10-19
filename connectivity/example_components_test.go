package connectivity_test

import (
	"fmt"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/connectivity"
)

func ExampleConnectedComponents() {
	g := gograph.New[string](gograph.Directed())
	_, _ = g.AddEdge(gograph.NewVertex("api"), gograph.NewVertex("auth"))
	_, _ = g.AddEdge(gograph.NewVertex("billing"), gograph.NewVertex("auth"))
	_, _ = g.AddEdge(gograph.NewVertex("search"), gograph.NewVertex("index"))

	// Edge directions are ignored, so billing joins api through auth.
	for _, component := range connectivity.ConnectedComponents(g) {
		var labels []string
		for _, v := range component {
			labels = append(labels, v.Label())
		}
		fmt.Println(labels)
	}

	// Output:
	// [api auth billing]
	// [search index]
}

func ExampleHasCycle() {
	g := gograph.New[int]()
	_, _ = g.AddEdge(gograph.NewVertex(1), gograph.NewVertex(2))
	_, _ = g.AddEdge(gograph.NewVertex(2), gograph.NewVertex(3))
	fmt.Println(connectivity.HasCycle(g))

	_, _ = g.AddEdge(g.GetVertexByID(3), g.GetVertexByID(1))
	fmt.Println(connectivity.HasCycle(g))

	// Output:
	// false
	// true
}
