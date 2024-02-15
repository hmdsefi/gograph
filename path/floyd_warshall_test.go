package path

import (
	"errors"
	"math"
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

	inf := math.Inf(1)

	expectedDist := map[string]map[string]float64{
		"A": {"A": 0, "B": 5, "C": 6, "D": 6, "E": 7, "F": 8},
		"B": {"A": inf, "B": 0, "C": 1, "D": 1, "E": 2, "F": 3},
		"C": {"A": inf, "B": inf, "C": 0, "D": 0, "E": 1, "F": 2},
		"D": {"A": inf, "B": inf, "C": inf, "D": 0, "E": 5, "F": 2},
		"E": {"A": inf, "B": inf, "C": inf, "D": -1, "E": 0, "F": 1},
		"F": {"A": inf, "B": inf, "C": inf, "D": 2, "E": 3, "F": 0},
	}

	for source, destMap := range dist {
		for dest, value := range destMap {
			if expectedDist[source][dest] != value {
				t.Fatalf(
					"expected distance %f from %s to %s, but got %f",
					expectedDist[source][dest],
					source,
					dest,
					value,
				)
			}
		}
	}
}

func TestFloydWarshall_NotWeighted(t *testing.T) {
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

func TestFloydWarshall_NotDirected(t *testing.T) {
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
