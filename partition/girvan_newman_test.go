package partition

import (
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestGirvanNewman_StringGraph(t *testing.T) {
	g := gograph.New[string]()

	// Create a graph with 6 nodes and cycles
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	d := g.AddVertexByLabel("D")
	e := g.AddVertexByLabel("E")
	f := g.AddVertexByLabel("F")

	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(a, c)
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(c, d)
	_, _ = g.AddEdge(d, e)
	_, _ = g.AddEdge(e, f)
	_, _ = g.AddEdge(d, f)

	// k = 2
	components, err := GirvanNewman(g, 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(components))
	}

	// Check that all original vertices are present
	vertexCount := 0
	for _, comp := range components {
		vertexCount += int(comp.Order())
	}
	if vertexCount != int(g.Order()) {
		t.Fatalf("vertex count mismatch after partitioning")
	}
}

func TestGirvanNewman_IntGraph(t *testing.T) {
	g := gograph.New[int]()

	// Simple triangle graph
	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)

	_, _ = g.AddEdge(v1, v2)
	_, _ = g.AddEdge(v2, v3)
	_, _ = g.AddEdge(v3, v1)

	components, err := GirvanNewman(g, 0) // remove all edges
	if err != nil {
		t.Fatal(err)
	}

	// With all edges removed, each vertex is a separate component
	expected := 3
	if len(components) != expected {
		t.Fatalf("expected %d components, got %d", expected, len(components))
	}

	// Check each component has exactly one vertex
	for _, comp := range components {
		if comp.Order() != 1 {
			t.Fatalf("component should have 1 vertex, got %d", comp.Order())
		}
	}
}

func TestGirvanNewman_RemovesEveryEdge(t *testing.T) {
	// The edge removal order depends on map iteration, so repeat the run.
	for run := 0; run < 50; run++ {
		g := gograph.New[int]()
		v := make([]*gograph.Vertex[int], 5)
		for i := range v {
			v[i] = g.AddVertexByLabel(i)
		}
		_, _ = g.AddEdge(v[0], v[2])
		_, _ = g.AddEdge(v[0], v[4])
		_, _ = g.AddEdge(v[1], v[2])
		_, _ = g.AddEdge(v[2], v[3])

		components, err := GirvanNewman(g, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(components) != len(v) {
			t.Fatalf("run %d: expected %d components, got %d", run, len(v), len(components))
		}
	}
}

func TestGirvanNewman_ComplexGraph(t *testing.T) {
	g := gograph.New[string]()

	// Complex graph with multiple cycles and bridges
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	d := g.AddVertexByLabel("D")
	e := g.AddVertexByLabel("E")
	f := g.AddVertexByLabel("F")
	g1 := g.AddVertexByLabel("G")
	h := g.AddVertexByLabel("H")

	// Core cycle
	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(c, a)

	// Bridge connections
	_, _ = g.AddEdge(c, d)
	_, _ = g.AddEdge(d, e)
	_, _ = g.AddEdge(e, f)
	_, _ = g.AddEdge(f, g1)
	_, _ = g.AddEdge(g1, h)

	components, err := GirvanNewman(g, 3)
	if err != nil {
		t.Fatal(err)
	}

	if len(components) != 3 {
		t.Fatalf("expected 3 components, got %d", len(components))
	}

	// Verify vertices preserved
	vertexCount := 0
	for _, comp := range components {
		vertexCount += int(comp.Order())
	}
	if vertexCount != int(g.Order()) {
		t.Fatalf("vertex count mismatch after partitioning")
	}
}

func TestGirvanNewman_TiedEdges(t *testing.T) {
	// the four outer edges of the square tie for the highest betweenness
	g := gograph.New[string]()
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	d := g.AddVertexByLabel("D")
	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(c, d)
	_, _ = g.AddEdge(d, a)
	_, _ = g.AddEdge(a, c)

	for k := 1; k <= 4; k++ {
		communities, err := GirvanNewman(g, k)
		if err != nil {
			t.Fatalf("k=%d: unexpected error: %v", k, err)
		}
		if len(communities) != k {
			t.Fatalf("k=%d: expected %d communities, got %d", k, k, len(communities))
		}
	}
}

func TestGirvanNewman_TiedBridges(t *testing.T) {
	// the three edges between the triangles tie
	for k := 1; k <= 3; k++ {
		communities, err := GirvanNewman(orderTestGraph(), k)
		if err != nil {
			t.Fatalf("k=%d: unexpected error: %v", k, err)
		}
		if len(communities) != k {
			t.Fatalf("k=%d: expected %d communities, got %d", k, k, len(communities))
		}
	}
}

func TestGirvanNewman_MoreThanVertices(t *testing.T) {
	g := gograph.New[int]()
	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	_, _ = g.AddEdge(v1, v2)

	communities, err := GirvanNewman(g, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(communities) != 2 {
		t.Fatalf("expected one community per vertex, got %d", len(communities))
	}
}

func TestGirvanNewman_AlreadySplit(t *testing.T) {
	g := gograph.New[int]()
	for i := range 3 {
		g.AddVertexByLabel(i)
	}

	communities, err := GirvanNewman(g, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(communities) != 3 {
		t.Fatalf("expected the 3 existing components, got %d", len(communities))
	}
}

func TestGirvanNewman_KeepsWeights(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a := g.AddVertexByLabel("A", gograph.WithVertexWeight(1.5))
	b := g.AddVertexByLabel("B", gograph.WithVertexWeight(2.5))
	c := g.AddVertexByLabel("C", gograph.WithVertexWeight(3.5))
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(4), gograph.WithEdgeLabel("A-B"))
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(5), gograph.WithEdgeLabel("B-C"))

	communities, err := GirvanNewman(g, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(communities) != 2 {
		t.Fatalf("expected 2 communities, got %d", len(communities))
	}

	edges := 0
	for _, community := range communities {
		if !community.IsWeighted() {
			t.Error("expected the communities of a weighted graph to be weighted")
		}
		for _, v := range community.GetAllVertices() {
			if want := g.GetVertexByID(v.Label()).Weight(); v.Weight() != want {
				t.Errorf("vertex %s: expected weight %v, got %v", v.Label(), want, v.Weight())
			}
		}
		for _, e := range community.AllEdges() {
			edges++
			original := g.GetEdge(
				g.GetVertexByID(e.Source().Label()),
				g.GetVertexByID(e.Destination().Label()),
			)
			if e.Weight() != original.Weight() {
				t.Errorf("edge %s-%s: expected weight %v, got %v", e.Source().Label(), e.Destination().Label(), original.Weight(), e.Weight())
			}
			if e.Label() != original.Label() {
				t.Errorf("edge %s-%s: expected label %q, got %q", e.Source().Label(), e.Destination().Label(), original.Label(), e.Label())
			}
		}
	}
	if edges == 0 {
		t.Fatal("expected the community with two vertices to keep its edge")
	}
}

func TestGirvanNewman_UnweightedGraph(t *testing.T) {
	g := gograph.New[string]()
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	_, _ = g.AddEdge(a, b)

	communities, err := GirvanNewman(g, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if communities[0].IsWeighted() {
		t.Error("expected the communities of an unweighted graph to be unweighted")
	}
}

func TestGirvanNewman_EmptyGraph(t *testing.T) {
	g := gograph.New[string]()

	components, err := GirvanNewman(g, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(components) != 0 {
		t.Fatalf("expected 0 components for empty graph, got %d", len(components))
	}
}
