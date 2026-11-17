package connectivity_test

import (
	"fmt"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/connectivity"
)

func ExampleBridges() {
	g := gograph.New[string]()
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))
	_, _ = g.AddEdge(gograph.NewVertex("B"), gograph.NewVertex("C"))
	edges, _ := connectivity.Bridges(g)
	for _, e := range edges {
		fmt.Println(e.Source().Label(), e.Destination().Label())
	}
	// Output:
	// B C
	// A B
}

func ExampleArticulationPoints() {
	g := gograph.New[string]()
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))
	_, _ = g.AddEdge(gograph.NewVertex("B"), gograph.NewVertex("C"))
	vertices, _ := connectivity.ArticulationPoints(g)
	for _, v := range vertices {
		fmt.Println(v.Label())
	}
	// Output: B
}

func ExampleCondense() {
	g := gograph.New[string](gograph.Directed())
	for _, e := range [][2]string{
		{"api", "auth"},
		{"auth", "users"},
		{"users", "auth"},
		{"users", "db"},
	} {
		_, _ = g.AddEdge(gograph.NewVertex(e[0]), gograph.NewVertex(e[1]))
	}

	c, err := connectivity.Condense(g)
	if err != nil {
		fmt.Println(err)
		return
	}

	// the condensation has no cycles, so it can be sorted
	order, err := gograph.TopologySort(c.Graph)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, comp := range order {
		var members []string
		for _, v := range c.Members[comp.Label()] {
			members = append(members, v.Label())
		}
		fmt.Println(comp.Label(), members)
	}

	// Output:
	// 0 [api]
	// 1 [auth users]
	// 2 [db]
}
