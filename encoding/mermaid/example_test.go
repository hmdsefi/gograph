package mermaid_test

import (
	"fmt"
	"os"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/encoding/mermaid"
)

func Example() {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())

	a := g.AddVertexByLabel("A")
	_, _ = g.AddEdge(a, gograph.NewVertex("B"), gograph.WithEdgeWeight(4))
	_, _ = g.AddEdge(a, gograph.NewVertex("C"), gograph.WithEdgeWeight(3))

	out, err := mermaid.Marshal(g, mermaid.WithDirection[string]("LR"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(string(out))

	// Output:
	// flowchart LR
	//     n0["A"]
	//     n1["B"]
	//     n2["C"]
	//     n0 -->|"4"| n1
	//     n0 -->|"3"| n2
}

func ExampleWithVertexClass() {
	g := gograph.New[string](gograph.Directed())

	_, _ = g.AddEdge(gograph.NewVertex("checkout"), gograph.NewVertex("build"))
	_, _ = g.AddEdge(g.GetVertexByID("build"), gograph.NewVertex("test"))

	// highlight the vertices that failed
	failed := map[string]bool{"test": true}

	err := mermaid.Write(os.Stdout, g,
		mermaid.WithVertexClass(func(v *gograph.Vertex[string]) string {
			if failed[v.Label()] {
				return "failed"
			}
			return ""
		}),
		mermaid.WithClassDef[string]("failed", "fill:#f96"),
	)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// flowchart TD
	//     n0["build"]
	//     n1["checkout"]
	//     n2["test"]
	//     n0 --> n2
	//     n1 --> n0
	//     classDef failed fill:#f96
	//     class n2 failed
}
