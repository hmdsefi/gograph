package path

import (
	"errors"
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestBellmanFord(t *testing.T) {
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

	dist, err := BellmanFord(g, vA.Label())
	if err != nil {
		t.Errorf("Expected no errors, but get an err: %s", err)
	}

	if dist[vB.Label()] != 5 {
		t.Errorf(
			"Expected %s to %s shortest distance to be %d, but got %f",
			vA.Label(),
			vB.Label(),
			5,
			dist[vB.Label()],
		)
	}

	if dist[vC.Label()] != 6 {
		t.Errorf(
			"Expected %s to %s shortest distance to be %d, but got %f",
			vA.Label(),
			vC.Label(),
			6,
			dist[vC.Label()],
		)
	}

	if dist[vD.Label()] != 6 {
		t.Errorf(
			"Expected %s to %s shortest distance to be %d, but got %f",
			vA.Label(),
			vD.Label(),
			6,
			dist[vD.Label()],
		)
	}

	if dist[vE.Label()] != 7 {
		t.Errorf(
			"Expected %s to %s shortest distance to be %d, but got %f",
			vA.Label(),
			vE.Label(),
			7,
			dist[vE.Label()],
		)
	}

	if dist[vF.Label()] != 8 {
		t.Errorf(
			"Expected %s to %s shortest distance to be %d, but got %f",
			vA.Label(),
			vF.Label(),
			8,
			dist[vF.Label()],
		)
	}
}

func TestBellmanFord_NegativeCycle(t *testing.T) {
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
	_, _ = g.AddEdge(vF, vE, gograph.WithEdgeWeight(-3))

	_, err := BellmanFord(g, vA.Label())
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if !errors.Is(err, ErrNegativeWeightCycle) {
		t.Errorf("Expected error \"%s\", but got \"%s\"", ErrNegativeWeightCycle, err)
	}
}

func TestBellmanFord_NotWeighted(t *testing.T) {
	g := gograph.New[string](gograph.Directed())

	vA := g.AddVertexByLabel("A")
	vB := g.AddVertexByLabel("B")
	vC := g.AddVertexByLabel("C")

	_, _ = g.AddEdge(vA, vB, gograph.WithEdgeWeight(5))
	_, _ = g.AddEdge(vB, vC, gograph.WithEdgeWeight(1))

	_, err := BellmanFord(g, vA.Label())
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if !errors.Is(err, ErrNotWeighted) {
		t.Errorf("Expected error \"%s\", but got \"%s\"", ErrNotWeighted, err)
	}
}

func TestBellmanFord_NotDirected(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())

	vA := g.AddVertexByLabel("A")
	vB := g.AddVertexByLabel("B")
	vC := g.AddVertexByLabel("C")

	_, _ = g.AddEdge(vA, vB, gograph.WithEdgeWeight(5))
	_, _ = g.AddEdge(vB, vC, gograph.WithEdgeWeight(1))

	_, err := BellmanFord(g, vA.Label())
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if !errors.Is(err, ErrNotDirected) {
		t.Errorf("Expected error \"%s\", but got \"%s\"", ErrNotDirected, err)
	}
}
