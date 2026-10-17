package gograph_test

import (
	"fmt"

	"github.com/hmdsefi/gograph"
)

func Example_undirectedEdgeCounts() {
	g := gograph.New[string]()
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(a, c)
	fmt.Println(g.Size(), len(g.AllEdges()), len(g.EdgesOf(a)), len(g.GetAllEdges(a, b)))
	fmt.Println(a.Degree(), a.InDegree(), a.OutDegree())
	// Output:
	// 4 4 4 2
	// 4 2 2
}
