package graph

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDAG_AddVertex(t *testing.T) {
	dag := &DAG{}
	dag.AddVertexWithID(1)
	dag.AddVertexWithID(2)
	dag.AddVertexWithID(3)

	assert.Equal(t, len(dag.Vertices), 3)
}

func TestDAG_AddEdge(t *testing.T) {
	// Create a new DAG
	dag := NewDAG()

	// Create three vertices with labels 1, 2, and 3
	v1 := &DAGVertex{ID: 1}
	v2 := &DAGVertex{ID: 2}
	v3 := &DAGVertex{ID: 3}

	// Add the vertices to the DAG
	dag.AddVertex(v1)
	dag.AddVertex(v2)
	dag.AddVertex(v3)

	// Add edges from 1 to 2 and from 2 to 3
	err := dag.AddEdge(v1, v2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err = dag.AddEdge(v2, v3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Try to add an edge from 3 to 1, which should result in an error
	err = dag.AddEdge(v3, v1)
	if err == nil {
		t.Error("Expected error, but got none")
	}
}

func TestDAG_TopologySort(t *testing.T) {
	// Create a DAG with 6 vertices and 6 edges
	dag := NewDAG()
	v0 := dag.AddVertexWithID(0)
	v1 := dag.AddVertexWithID(1)
	v2 := dag.AddVertexWithID(2)
	v3 := dag.AddVertexWithID(3)
	v4 := dag.AddVertexWithID(4)
	v5 := dag.AddVertexWithID(5)

	err := dag.AddEdge(v5, v2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = dag.AddEdge(v5, v0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = dag.AddEdge(v4, v0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = dag.AddEdge(v4, v1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = dag.AddEdge(v2, v3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = dag.AddEdge(v3, v1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Perform a topological sort
	sortedVertices, err := dag.TopologySort()

	// Check that there was no error
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check that the sorted order is correct
	expectedOrder := []*DAGVertex{v4, v5, v2, v0, v3, v1}
	if !reflect.DeepEqual(sortedVertices, expectedOrder) {
		t.Errorf("unexpected sort order. Got %v, expected %v", sortedVertices, expectedOrder)
	}
}
