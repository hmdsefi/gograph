package connectivity

import (
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestSCCs_ReturnGraphVertices(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	a := g.AddVertexByLabel("A", gograph.WithVertexWeight(5))
	b := g.AddVertexByLabel("B", gograph.WithVertexWeight(7))
	c := g.AddVertexByLabel("C", gograph.WithVertexWeight(9))
	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(c, b)

	funcs := map[string]func(gograph.Graph[string]) [][]*gograph.Vertex[string]{
		"Kosaraju": Kosaraju[string],
		"Tarjan":   Tarjan[string],
		"Gabow":    Gabow[string],
	}

	for name, scc := range funcs {
		t.Run(name, func(t *testing.T) {
			count := 0
			for _, component := range scc(g) {
				for _, v := range component {
					count++
					if want := g.GetVertexByID(v.Label()); v != want {
						t.Errorf(
							"Expect the graph's vertex %s, but got a copy with weight %v and %d neighbors",
							v.Label(), v.Weight(), len(v.Neighbors()),
						)
					}
				}
			}

			if count != 3 {
				t.Errorf("Expect 3 vertices in the components, but got %d", count)
			}
		})
	}
}
