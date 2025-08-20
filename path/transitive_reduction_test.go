package path

import (
	"errors"
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestTransitiveReduction_Simple(t *testing.T) {
	// Create a simple directed acyclic graph
	g := gograph.New[string](gograph.Directed())

	// Add vertices
	vA := g.AddVertexByLabel("A")
	vB := g.AddVertexByLabel("B")
	vC := g.AddVertexByLabel("C")

	// Add edges A->B, B->C, A->C (A->C is the transitive edge that should be removed)
	_, _ = g.AddEdge(vA, vB)
	_, _ = g.AddEdge(vB, vC)
	_, _ = g.AddEdge(vA, vC)

	// Compute transitive reduction
	reduced, err := TransitiveReduction(g)
	if err != nil {
		t.Errorf("TransitiveReduction returned an error: %v", err)
	}

	// Reduced graph should have 3 vertices and 2 edges (A->B, B->C)
	if len(reduced.GetAllVertices()) != 3 {
		t.Errorf("Expected 3 vertices, got %d", len(reduced.GetAllVertices()))
	}
	if len(reduced.AllEdges()) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(reduced.AllEdges()))
	}

	// Check specific edges
	vAReduced := reduced.GetVertexByID("A")
	vBReduced := reduced.GetVertexByID("B")
	vCReduced := reduced.GetVertexByID("C")

	if reduced.GetEdge(vAReduced, vBReduced) == nil {
		t.Errorf("Edge A->B should exist in reduced graph")
	}
	if reduced.GetEdge(vBReduced, vCReduced) == nil {
		t.Errorf("Edge B->C should exist in reduced graph")
	}
	if reduced.GetEdge(vAReduced, vCReduced) != nil {
		t.Errorf("Edge A->C should not exist in reduced graph")
	}
}

func TestTransitiveReduction_Complex(t *testing.T) {
	// Create a more complex directed acyclic graph
	g := gograph.New[string](gograph.Directed())

	// Add vertices
	vA := g.AddVertexByLabel("A")
	vB := g.AddVertexByLabel("B")
	vC := g.AddVertexByLabel("C")
	vD := g.AddVertexByLabel("D")
	g.AddVertexByLabel("E")

	// Add edges
	// A->B, B->C, C->D
	// A->C, B->D, A->D (these should be reduced)
	// E (isolated vertex)
	_, _ = g.AddEdge(vA, vB)
	_, _ = g.AddEdge(vB, vC)
	_, _ = g.AddEdge(vC, vD)
	_, _ = g.AddEdge(vA, vC) // This should be removed (A->B->C)
	_, _ = g.AddEdge(vB, vD) // This should be removed (B->C->D)
	_, _ = g.AddEdge(vA, vD) // This should be removed (A->B->C->D)

	// Compute transitive reduction
	reduced, err := TransitiveReduction(g)
	if err != nil {
		t.Errorf("TransitiveReduction returned an error: %v", err)
	}

	// Reduced graph should have 5 vertices and 3 edges (A->B, B->C, C->D)
	if len(reduced.GetAllVertices()) != 5 {
		t.Errorf("Expected 5 vertices, got %d", len(reduced.GetAllVertices()))
	}
	if len(reduced.AllEdges()) != 3 {
		t.Errorf("Expected 3 edges, got %d", len(reduced.AllEdges()))
	}

	// Check specific edges
	vAReduced := reduced.GetVertexByID("A")
	vBReduced := reduced.GetVertexByID("B")
	vCReduced := reduced.GetVertexByID("C")
	vDReduced := reduced.GetVertexByID("D")

	if reduced.GetEdge(vAReduced, vBReduced) == nil {
		t.Errorf("Edge A->B should exist in reduced graph")
	}
	if reduced.GetEdge(vBReduced, vCReduced) == nil {
		t.Errorf("Edge B->C should exist in reduced graph")
	}
	if reduced.GetEdge(vCReduced, vDReduced) == nil {
		t.Errorf("Edge C->D should exist in reduced graph")
	}
	if reduced.GetEdge(vAReduced, vCReduced) != nil {
		t.Errorf("Edge A->C should not exist in reduced graph")
	}
	if reduced.GetEdge(vBReduced, vDReduced) != nil {
		t.Errorf("Edge B->D should not exist in reduced graph")
	}
	if reduced.GetEdge(vAReduced, vDReduced) != nil {
		t.Errorf("Edge A->D should not exist in reduced graph")
	}
}

func TestTransitiveReduction_Diamond(t *testing.T) {
	// Create a diamond-shaped graph
	g := gograph.New[string](gograph.Directed())

	// Add vertices
	vA := g.AddVertexByLabel("A")
	vB := g.AddVertexByLabel("B")
	vC := g.AddVertexByLabel("C")
	vD := g.AddVertexByLabel("D")

	// Add edges - Diamond pattern
	// A->B, A->C, B->D, C->D
	_, _ = g.AddEdge(vA, vB)
	_, _ = g.AddEdge(vA, vC)
	_, _ = g.AddEdge(vB, vD)
	_, _ = g.AddEdge(vC, vD)

	// Compute transitive reduction
	// In this case, all edges should remain since there are no transitive edges
	reduced, err := TransitiveReduction(g)
	if err != nil {
		t.Errorf("TransitiveReduction returned an error: %v", err)
	}

	// Reduced graph should have 4 vertices and 4 edges (same as original)
	if len(reduced.GetAllVertices()) != 4 {
		t.Errorf("Expected 4 vertices, got %d", len(reduced.GetAllVertices()))
	}
	if len(reduced.AllEdges()) != 4 {
		t.Errorf("Expected 4 edges, got %d", len(reduced.AllEdges()))
	}

	// Check specific edges - All should still be present
	vAReduced := reduced.GetVertexByID("A")
	vBReduced := reduced.GetVertexByID("B")
	vCReduced := reduced.GetVertexByID("C")
	vDReduced := reduced.GetVertexByID("D")

	if reduced.GetEdge(vAReduced, vBReduced) == nil {
		t.Errorf("Edge A->B should exist in reduced graph")
	}
	if reduced.GetEdge(vAReduced, vCReduced) == nil {
		t.Errorf("Edge A->C should exist in reduced graph")
	}
	if reduced.GetEdge(vBReduced, vDReduced) == nil {
		t.Errorf("Edge B->D should exist in reduced graph")
	}
	if reduced.GetEdge(vCReduced, vDReduced) == nil {
		t.Errorf("Edge C->D should exist in reduced graph")
	}
}

func TestTransitiveReduction_WeightPreservation(t *testing.T) {
	// Create a weighted directed graph
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())

	// Add vertices
	vA := g.AddVertexByLabel("A")
	vB := g.AddVertexByLabel("B")
	vC := g.AddVertexByLabel("C")

	// Add weighted edges
	// A->B (weight 1), B->C (weight 2), A->C (weight 10) - A->C should be removed
	_, _ = g.AddEdge(vA, vB, gograph.WithEdgeWeight(1.0))
	_, _ = g.AddEdge(vB, vC, gograph.WithEdgeWeight(2.0))
	_, _ = g.AddEdge(vA, vC, gograph.WithEdgeWeight(10.0))

	// Compute transitive reduction
	reduced, err := TransitiveReduction(g)
	if err != nil {
		t.Errorf("TransitiveReduction returned an error: %v", err)
	}

	// Check if the reduced graph is also weighted
	if !reduced.IsWeighted() {
		t.Errorf("Expected reduced graph to be weighted")
	}

	// Check if weights are preserved
	vAReduced := reduced.GetVertexByID("A")
	vBReduced := reduced.GetVertexByID("B")
	vCReduced := reduced.GetVertexByID("C")

	edgeAB := reduced.GetEdge(vAReduced, vBReduced)
	edgeBC := reduced.GetEdge(vBReduced, vCReduced)

	if edgeAB == nil {
		t.Errorf("Edge A->B should exist in reduced graph")
	}
	if edgeBC == nil {
		t.Errorf("Edge B->C should exist in reduced graph")
	}

	if edgeAB != nil && edgeAB.Weight() != 1.0 {
		t.Errorf("Edge A->B should have weight 1.0, got %f", edgeAB.Weight())
	}
	if edgeBC != nil && edgeBC.Weight() != 2.0 {
		t.Errorf("Edge B->C should have weight 2.0, got %f", edgeBC.Weight())
	}
}

func TestTransitiveReduction_ErrorCases(t *testing.T) {
	// Test case 1: Undirected graph
	undirectedGraph := gograph.New[string]()
	_, err := TransitiveReduction(undirectedGraph)
	if err == nil {
		t.Errorf("Expected error for undirected graph, got nil")
	}
	if err != ErrNotDirected {
		t.Errorf("Expected ErrNotDirected, got %v", err)
	}

	// Test case 2: Graph with cycle
	cyclicGraph := gograph.New[string](gograph.Directed())
	vA := cyclicGraph.AddVertexByLabel("A")
	vB := cyclicGraph.AddVertexByLabel("B")
	vC := cyclicGraph.AddVertexByLabel("C")

	_, _ = cyclicGraph.AddEdge(vA, vB)
	_, _ = cyclicGraph.AddEdge(vB, vC)
	_, _ = cyclicGraph.AddEdge(vC, vA) // Creates a cycle

	_, err = TransitiveReduction(cyclicGraph)
	if err == nil {
		t.Errorf("Expected error for cyclic graph, got nil")
	}
	if !errors.Is(err, ErrNotDAG) {
		t.Errorf("Expected ErrNotDAG, got %v", err)
	}
}

func TestTransitiveReduction_EmptyGraph(t *testing.T) {
	// Create an empty directed graph
	g := gograph.New[string](gograph.Directed())

	// Compute transitive reduction
	reduced, err := TransitiveReduction(g)
	if err != nil {
		t.Errorf("TransitiveReduction returned an error: %v", err)
	}

	// Reduced graph should be empty too
	if len(reduced.GetAllVertices()) != 0 {
		t.Errorf("Expected 0 vertices, got %d", len(reduced.GetAllVertices()))
	}
	if len(reduced.AllEdges()) != 0 {
		t.Errorf("Expected 0 edges, got %d", len(reduced.AllEdges()))
	}
}
