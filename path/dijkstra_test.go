package path

import (
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestDijkstra(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())

	vA := g.AddVertexByLabel("A")
	vB := g.AddVertexByLabel("B")
	vC := g.AddVertexByLabel("C")
	vD := g.AddVertexByLabel("D")

	_, _ = g.AddEdge(vA, vB, gograph.WithEdgeWeight(4))
	_, _ = g.AddEdge(vA, vC, gograph.WithEdgeWeight(3))
	_, _ = g.AddEdge(vB, vC, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(vB, vD, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(vC, vD, gograph.WithEdgeWeight(4))

	dist := Dijkstra(g, "A")

	if dist[vA.Label()] != 0 {
		t.Errorf("Expected distance from 0 to 0 to be 0, got %f", dist[vA.Label()])
	}
	if dist[vB.Label()] != 4 {
		t.Errorf("Expected distance from 0 to 1 to be 4, got %f", dist[vB.Label()])
	}
	if dist[vC.Label()] != 3 {
		t.Errorf("Expected distance from 0 to 2 to be 3, got %f", dist[vC.Label()])
	}
	if dist[vD.Label()] != 6 {
		t.Errorf("Expected distance from 0 to 3 to be 6, got %f", dist[vD.Label()])
	}
}
