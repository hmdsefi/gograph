package graph

import (
	"reflect"
	"testing"
)

func TestDAG_AddEdge(t *testing.T) {
	// Create a new dag
	g := NewDAG[int]()

	// Create three vertices with labels 1, 2, and 3
	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)

	// Add the vertices to the dag
	g.AddVertex(v1)
	g.AddVertex(v2)
	g.AddVertex(v3)

	// Add edges from 1 to 2 and from 2 to 3
	_, err := g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v2, v3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Try to add an edges from 3 to 1, which should result in an error
	_, err = g.AddEdge(v3, v1)
	if err == nil {
		t.Error("Expected error, but got none")
	}
}

func TestDAG_TopologySort(t *testing.T) {
	// Create a dag with 6 vertices and 6 edges
	g := NewDAG[int]()
	v0 := g.AddVertexByLabel(0)
	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)
	v5 := g.AddVertexByLabel(5)

	_, err := g.AddEdge(v5, v2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v5, v0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v4, v0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v4, v1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v2, v3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v3, v1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Perform a topological sort
	sortedVertices, err := TopologySort(g)

	// Check that there was no error
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check that the sorted order is correct
	expectedOrder := []*Vertex[int]{v4, v5, v2, v0, v3, v1}
	if !reflect.DeepEqual(sortedVertices, expectedOrder) {
		t.Errorf("unexpected sort order. Got %v, expected %v", sortedVertices, expectedOrder)
	}
}
