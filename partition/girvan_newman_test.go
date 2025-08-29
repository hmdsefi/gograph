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
