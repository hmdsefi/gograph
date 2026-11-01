package traverse_test

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/traverse"
)

func ExampleDFSEvents() {
	g := gograph.New[string](gograph.Directed())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(a, c)
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(c, a)

	for e := range traverse.DFSEvents(g, "A") {
		indent := strings.Repeat("  ", e.Depth)
		if e.Edge != nil {
			fmt.Printf("%s%v %s -> %s\n", indent, e.Kind, e.Edge.Source().Label(), e.Edge.Destination().Label())
			continue
		}
		fmt.Printf("%s%v %s\n", indent, e.Kind, e.Vertex.Label())
	}
	// Output:
	// DiscoverVertex A
	// TreeEdge A -> B
	//   DiscoverVertex B
	//   TreeEdge B -> C
	//     DiscoverVertex C
	//     BackEdge C -> A
	//     FinishVertex C
	//   FinishVertex B
	// ForwardEdge A -> C
	// FinishVertex A
}

func ExampleDFSEventsAll() {
	// in a graph without cycles, the reverse of the finish order is a
	// topological order
	g := gograph.New[string](gograph.Directed())
	for _, e := range [][2]string{
		{"shirt", "tie"}, {"tie", "jacket"}, {"pants", "shoes"}, {"pants", "belt"}, {"belt", "jacket"},
	} {
		_, _ = g.AddEdge(gograph.NewVertex(e[0]), gograph.NewVertex(e[1]))
	}

	var order []string
	for e := range traverse.DFSEventsAll(g) {
		if e.Kind == traverse.FinishVertex {
			order = append(order, e.Vertex.Label())
		}
	}
	slices.Reverse(order)
	fmt.Println(order)
	// Output: [pants belt shoes shirt tie jacket]
}
