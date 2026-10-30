package path

import (
	"errors"
	"math/rand/v2"
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

func TestBellmanFord_MissingStart(t *testing.T) {
	g := gograph.New[string](gograph.Weighted(), gograph.Directed())

	vA := g.AddVertexByLabel("A")
	vB := g.AddVertexByLabel("B")

	_, _ = g.AddEdge(vA, vB, gograph.WithEdgeWeight(1))

	dist, err := BellmanFord(g, "X")
	if !errors.Is(err, gograph.ErrVertexDoesNotExist) {
		t.Errorf("Expected error \"%s\", but got \"%v\"", gograph.ErrVertexDoesNotExist, err)
	}

	if dist != nil {
		t.Errorf("Expected nil distances, but got %v", dist)
	}
}

// randomWeighted returns a directed graph with n vertices and m random edges
// with whole weights from 1 to 9.
func randomWeighted(n, m int) gograph.Graph[int] {
	rng := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // seeded so runs are comparable
	g := gograph.New[int](gograph.Directed(), gograph.Weighted())
	vertices := make([]*gograph.Vertex[int], n)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for int(g.Size()) < m {
		from, to := vertices[rng.IntN(n)], vertices[rng.IntN(n)]
		_, _ = g.AddEdge(from, to, gograph.WithEdgeWeight(float64(1+rng.IntN(9))))
	}
	return g
}

func BenchmarkBellmanFord(b *testing.B) {
	// the vertices of the chain come in reverse, so each pass over the edges
	// in AllEdges order moves the distances one edge further
	const chain = 1000
	reversed := gograph.New[int](gograph.Directed(), gograph.Weighted())
	for i := chain - 1; i >= 0; i-- {
		reversed.AddVertexByLabel(i)
	}
	for i := range chain - 1 {
		_, _ = reversed.AddEdge(reversed.GetVertexByID(i), reversed.GetVertexByID(i+1), gograph.WithEdgeWeight(-1))
	}

	graphs := []struct {
		name string
		g    gograph.Graph[int]
	}{
		{name: "random-1000x5000", g: randomWeighted(1000, 5000)},
		{name: "random-10000x50000", g: randomWeighted(10000, 50000)},
		{name: "reversed-chain-1000", g: reversed},
	}
	for _, tt := range graphs {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := BellmanFord(tt.g, 0); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
