package traverse

import (
	"errors"
	"testing"

	"github.com/hmdsefi/gograph"
)

func initClosestFirstIteratorTestGraph() gograph.Graph[string] {
	g := gograph.New[string]()
	vA := g.AddVertexByLabel("A")
	vB := g.AddVertexByLabel("B")
	vC := g.AddVertexByLabel("C")
	vD := g.AddVertexByLabel("D")

	_, _ = g.AddEdge(vA, vB, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(vA, vC, gograph.WithEdgeWeight(5))
	_, _ = g.AddEdge(vB, vC, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(vB, vD, gograph.WithEdgeWeight(3))
	_, _ = g.AddEdge(vC, vD, gograph.WithEdgeWeight(2))
	return g
}

func TestClosestFirstIterator_HasNext(t *testing.T) {
	// create a graph and a starting vertex for the iterator
	g := initClosestFirstIteratorTestGraph()

	// create an iterator with a vertex that doesn't exist
	_, err := NewClosestFirstIterator(g, "X")
	if err == nil {
		t.Error("Expect NewClosestFirstIterator returns error, but got nil")
	}

	// create the closest-first iterator
	it, err := NewClosestFirstIterator(g, "A")
	if err != nil {
		t.Errorf("Expect NewClosestFirstIterator doesn't return error, but got %s", err)
	}

	// check that the HasNext method returns true before reaching the end of the walk
	vertices := len(g.GetAllVertices())
	for i := 0; i < vertices; i++ {
		if !it.HasNext() {
			t.Errorf("Expected HasNext to return true at step %d, but it returned false", i)
		}
		it.Next()
	}

	// check that the HasNext method returns false after reaching the end of the walk
	if it.HasNext() {
		t.Errorf("Expected HasNext to return false at end of walk, but it returned true")
	}
}

func TestClosestFirstIterator_Next(t *testing.T) {
	// create a graph and a starting vertex for the iterator
	g := initClosestFirstIteratorTestGraph()

	// create the random walk iterator
	it, err := NewClosestFirstIterator(g, "A")
	if err != nil {
		t.Errorf("Expect NewClosestFirstIterator doesn't return error, but got %s", err)
	}
	// check that the Next method returns the expected vertices in the walk
	expected := []string{"A", "B", "C", "D"}
	for i := 0; i < len(expected); i++ {
		v := it.Next()
		if v.Label() != expected[i] {
			t.Errorf(
				"Expected Next to return vertex %s, but it returned vertex %s",
				expected[i], v.Label(),
			)
		}
	}

	v := it.Next()
	if v != nil {
		t.Errorf("Expected Next returns nil, but got %+v", v)
	}
}

func TestClosestFirstIterator_Iterate(t *testing.T) {
	// create a graph and a starting vertex for the iterator
	g := initClosestFirstIteratorTestGraph()

	// create the random walk iterator
	it, err := NewClosestFirstIterator(g, "A")
	if err != nil {
		t.Errorf("Expect NewClosestFirstIterator doesn't return error, but got %s", err)
	}

	// Initialize a slice to hold the visited vertices
	visited := make([]string, 0)

	// Iterate over the closest vertices and add their labels to the visited slice
	err = it.Iterate(
		func(v *gograph.Vertex[string]) error {
			visited = append(visited, v.Label())
			return nil
		},
	)
	if err != nil {
		t.Errorf("Unexpected error during Iterate method call: %v", err)
	}

	// check that the visited slice contains the expected vertices in the walk
	expected := []string{"A", "B", "C", "D"}
	for i := 0; i < len(expected); i++ {
		if visited[i] != expected[i] {
			t.Errorf("Expected visited vertex at step %d to be %s, but it was %s", i, expected[i], visited[i])
		}
	}

	it.Reset()
	if !it.HasNext() {
		t.Errorf("Expected HasNext to return true at step %d, but it returned false", 1)
	}

	v := it.Next()
	if v.Label() != "A" {
		t.Errorf(
			"Expected Next to return vertex %s, but it returned vertex %s",
			"A", v.Label(),
		)
	}

	expectedErr := errors.New("something went wrong")
	err = it.Iterate(
		func(vertex *gograph.Vertex[string]) error {
			return expectedErr
		},
	)
	if err == nil {
		t.Error("Expect iter.Iterate(func) returns error, but got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expect %+v error, but got %+v", expectedErr, err)
	}
}
