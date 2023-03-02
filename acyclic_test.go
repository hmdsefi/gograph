package gograph

import (
	"reflect"
	"testing"
)

func TestTopologySort(t *testing.T) {
	// Create a dag with 6 vertices and 6 edges
	g := New[int](Acyclic())

	if !g.IsDirected() {
		t.Error(testErrMsgNotTrue)
	}

	if !g.IsAcyclic() {
		t.Error(testErrMsgNotTrue)
	}

	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)
	v5 := g.AddVertexByLabel(5)
	v6 := g.AddVertexByLabel(6)

	_, err := g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v2, v3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v2, v4)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v2, v5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v3, v5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v4, v6)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v5, v6)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Perform a topological sort
	sortedVertices, err := TopologySort[int](g)

	// Check that there was no error
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check that the sorted order is correct
	expectedOrder := []*Vertex[int]{v1, v2, v3, v4, v5, v6}
	if !reflect.DeepEqual(sortedVertices, expectedOrder) {
		t.Errorf("unexpected sort order. Got %v, expected %v", sortedVertices, expectedOrder)
	}
}
