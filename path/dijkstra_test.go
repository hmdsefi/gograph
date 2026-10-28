package path

import (
	"math/rand"
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestDijkstraSimple(t *testing.T) {
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

	// use not existing vertex
	dist := DijkstraSimple(g, "X")
	if len(dist) > 0 {
		t.Errorf("Expected dist map length be 0, got %d", len(dist))
	}

	dist = DijkstraSimple(g, "A")

	if dist[vA.Label()] != 0 {
		t.Errorf("Expected distance from A to %s to be 0, got %f", vA.Label(), dist[vA.Label()])
	}
	if dist[vB.Label()] != 4 {
		t.Errorf("Expected distance from A to %s  to be 4, got %f", vB.Label(), dist[vB.Label()])
	}
	if dist[vC.Label()] != 3 {
		t.Errorf("Expected distance from A to %s  to be 3, got %f", vC.Label(), dist[vC.Label()])
	}
	if dist[vD.Label()] != 6 {
		t.Errorf("Expected distance from A to %s  to be 6, got %f", vD.Label(), dist[vD.Label()])
	}
}

type countingGraph[T comparable] struct {
	gograph.Graph[T]
	getEdgeCalls int
}

func (g *countingGraph[T]) GetEdge(from, to *gograph.Vertex[T]) *gograph.Edge[T] {
	g.getEdgeCalls++
	return g.Graph.GetEdge(from, to)
}

func TestDijkstraDoesNotExpandOutdatedEntries(t *testing.T) {
	const n = 300
	g := gograph.New[int](gograph.Directed())
	s := g.AddVertexByLabel(0)
	hub := g.AddVertexByLabel(1)
	for i := 1; i <= n; i++ {
		x := g.AddVertexByLabel(1000 + i)
		_, _ = g.AddEdge(s, x, gograph.WithEdgeWeight(float64(i)))
		_, _ = g.AddEdge(x, hub, gograph.WithEdgeWeight(float64(3*n-2*i)))
		y := g.AddVertexByLabel(2000 + i)
		_, _ = g.AddEdge(hub, y, gograph.WithEdgeWeight(float64(10*n)))
	}

	cg := &countingGraph[int]{Graph: g}
	distances := Dijkstra[int](cg, 0)
	if got := distances[1]; got != 2*n {
		t.Fatalf("distance to hub: got %v, want %d", got, 2*n)
	}
	if cg.getEdgeCalls > int(g.Size()) {
		t.Fatalf("GetEdge called %d times for %d edges; want at most one call per edge", cg.getEdgeCalls, g.Size())
	}
}

func TestDijkstra(t *testing.T) {
	g := gograph.New[int](gograph.Weighted())

	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)

	_, _ = g.AddEdge(v1, v2, gograph.WithEdgeWeight(4))
	_, _ = g.AddEdge(v1, v3, gograph.WithEdgeWeight(3))
	_, _ = g.AddEdge(v2, v3, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(v2, v4, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(v3, v4, gograph.WithEdgeWeight(4))

	dist := Dijkstra(g, 1)

	if dist[v1.Label()] != 0 {
		t.Errorf("Expected distance from 1 to %d to be 0, got %f", v1.Label(), dist[v1.Label()])
	}
	if dist[v2.Label()] != 4 {
		t.Errorf("Expected distance from 1 to %d to be 4, got %f", v2.Label(), dist[v2.Label()])
	}
	if dist[v3.Label()] != 3 {
		t.Errorf("Expected distance from 1 to %d to be 3, got %f", v3.Label(), dist[v3.Label()])
	}
	if dist[v4.Label()] != 6 {
		t.Errorf("Expected distance from 1 to %d to be 6, got %f", v4.Label(), dist[v4.Label()])
	}

	// use not existing vertex
	dist = Dijkstra(g, 0)
	if len(dist) > 0 {
		t.Errorf("Expected dist map length be 0, got %d", len(dist))
	}
}

func BenchmarkDijkstra(b *testing.B) {
	const n, m = 10000, 50000
	rng := rand.New(rand.NewSource(1)) //nolint:gosec // seeded so runs are comparable
	g := gograph.New[int](gograph.Directed(), gograph.Weighted())
	vertices := make([]*gograph.Vertex[int], n)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for g.Size() < m {
		from, to := vertices[rng.Intn(n)], vertices[rng.Intn(n)]
		_, _ = g.AddEdge(from, to, gograph.WithEdgeWeight(1+rng.Float64()*9))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Dijkstra(g, 0)
	}
}
