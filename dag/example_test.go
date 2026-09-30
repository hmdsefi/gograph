package dag_test

import (
	"fmt"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/dag"
)

func labels(vertices []*gograph.Vertex[string]) []string {
	out := make([]string, len(vertices))
	for i, v := range vertices {
		out[i] = v.Label()
	}
	return out
}

func Example() {
	// An edge A -> B means B depends on A.
	g := gograph.New[string](gograph.Acyclic())

	checkout := g.AddVertexByLabel("checkout")
	build := g.AddVertexByLabel("build")
	test := g.AddVertexByLabel("test")
	lint := g.AddVertexByLabel("lint")

	_, _ = g.AddEdge(checkout, build)
	_, _ = g.AddEdge(build, test)
	_, _ = g.AddEdge(build, lint)

	descendants, _ := dag.Descendants(g, "build")
	fmt.Println("depends on build:", labels(descendants))

	ancestors, _ := dag.Ancestors(g, "test")
	fmt.Println("test depends on:", labels(ancestors))

	affected, _ := dag.Affected(g, "build")
	fmt.Println("rebuild after a build change:", labels(affected))

	// Output:
	// depends on build: [test lint]
	// test depends on: [build checkout]
	// rebuild after a build change: [build test lint]
}
