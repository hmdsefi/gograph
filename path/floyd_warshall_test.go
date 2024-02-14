package path

import (
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestFloydWarshall(t *testing.T) {
	g := gograph.New[string](gograph.Weighted(), gograph.Directed())

	vA := g.AddVertexByLabel("A")
	vB := g.AddVertexByLabel("B")
	vC := g.AddVertexByLabel("C")
	vD := g.AddVertexByLabel("D")
	vE := g.AddVertexByLabel("E")
	vF := g.AddVertexByLabel("F")

	_, _ = g.AddEdge(vA, vB, gograph.WithEdgeWeight(5))
	_, _ = g.AddEdge(vB, vC, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(vB, vD, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(vC, vE, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(vE, vD, gograph.WithEdgeWeight(-1))
	_, _ = g.AddEdge(vD, vF, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(vF, vE, gograph.WithEdgeWeight(3))

	dist, err := FloydWarshall(g)
	if err != nil {
		t.Errorf("Expected no errors, but get an err: %s", err)
	}

	for s, d := range dist {
		t.Log(s, d)
	}
}
