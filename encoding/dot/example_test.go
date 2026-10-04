package dot_test

import (
	"fmt"
	"os"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/encoding/dot"
)

func Example() {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())

	a := g.AddVertexByLabel("A")
	_, _ = g.AddEdge(a, gograph.NewVertex("B"), gograph.WithEdgeWeight(4))

	out, err := dot.Marshal(g)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(string(out))

	// Output:
	// digraph {
	// 	n0 [label="A"];
	// 	n1 [label="B"];
	// 	n0 -> n1 [label="4"];
	// }
}

func ExampleWithVertexAttributes() {
	g := gograph.New[string](gograph.Directed())

	_, _ = g.AddEdge(gograph.NewVertex("checkout"), gograph.NewVertex("build"))
	_, _ = g.AddEdge(g.GetVertexByID("build"), gograph.NewVertex("test"))

	// highlight the vertices that failed
	failed := map[string]bool{"test": true}

	err := dot.Write(os.Stdout, g,
		dot.WithName[string]("pipeline"),
		dot.WithGraphAttributes[string](map[string]string{"rankdir": "LR"}),
		dot.WithVertexAttributes(func(v *gograph.Vertex[string]) map[string]string {
			if failed[v.Label()] {
				return map[string]string{"color": "red", "style": "filled", "fillcolor": "#ffdddd"}
			}
			return map[string]string{"shape": "box"}
		}),
	)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// digraph pipeline {
	// 	rankdir="LR";
	// 	n0 [label="build", shape="box"];
	// 	n1 [label="checkout", shape="box"];
	// 	n2 [color="red", fillcolor="#ffdddd", label="test", style="filled"];
	// 	n0 -> n2;
	// 	n1 -> n0;
	// }
}
