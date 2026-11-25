package connectivity_test

import (
	"fmt"
	"strings"

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

func ExampleFindCycle() {
	g := gograph.New[string](gograph.Directed())
	for _, e := range [][2]string{{"api", "auth"}, {"auth", "db"}, {"db", "cache"}, {"cache", "auth"}} {
		_, _ = g.AddEdge(gograph.NewVertex(e[0]), gograph.NewVertex(e[1]))
	}

	cycle, ok := connectivity.FindCycle(g)
	if !ok {
		fmt.Println("no cycle")
		return
	}
	var path []string
	for _, v := range cycle {
		path = append(path, v.Label())
	}
	fmt.Println("import cycle:", strings.Join(path, " -> "), "->", path[0])
	// Output: import cycle: auth -> db -> cache -> auth
}
