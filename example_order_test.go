package gograph_test

import (
	"cmp"
	"fmt"

	"github.com/hmdsefi/gograph"
)

func ExampleStableTopologySort() {
	g := gograph.New[string](gograph.Acyclic())

	// "lint" and "build" have no dependencies, so either can go first.
	lint := g.AddVertexByLabel("lint")
	build := g.AddVertexByLabel("build")
	test := g.AddVertexByLabel("test")

	_, _ = g.AddEdge(lint, test)
	_, _ = g.AddEdge(build, test)

	labels := func(vertices []*gograph.Vertex[string]) []string {
		out := make([]string, len(vertices))
		for i, v := range vertices {
			out[i] = v.Label()
		}
		return out
	}

	// TopologySort follows the order the vertices were added.
	order, _ := gograph.TopologySort(g)
	fmt.Println(labels(order))

	// With cmp.Compare, the smallest ready label always comes first.
	order, _ = gograph.StableTopologySort(g, cmp.Compare[string])
	fmt.Println(labels(order))

	// Output:
	// [lint build test]
	// [build lint test]
}
