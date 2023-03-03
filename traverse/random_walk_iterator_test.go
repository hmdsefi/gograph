package traverse

import (
	"errors"
	"testing"

	"github.com/hmdsefi/gograph"
)

func initTestGraph() gograph.Graph[int] {
	g := gograph.New[int]()
	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	_, _ = g.AddEdge(v1, v2)
	return g
}

func TestRandomWalkIterator_HasNext(t *testing.T) {
	// create a graph and a starting vertex for the iterator
	g := initTestGraph()

	// create an iterator with a vertex that doesn't exist
	_, err := NewRandomWalkIterator(g, 123, 10)
	if err == nil {
		t.Error("Expect NewRandomWalkIterator returns error, but got nil")
	}

	// create the random walk iterator
	it, err := NewRandomWalkIterator(g, 1, 10)
	if err != nil {
		t.Errorf("Expect NewRandomWalkIterator doesn't return error, but got %s", err)
	}

	// check that the HasNext method returns true before reaching the end of the walk
	for i := 0; i < 10; i++ {
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

func TestRandomWalkIterator_Next(t *testing.T) {
	// create a graph and a starting vertex for the iterator
	g := initTestGraph()

	// create the random walk iterator
	it, err := NewRandomWalkIterator(g, 1, 10)
	if err != nil {
		t.Errorf("Expect NewRandomWalkIterator doesn't return error, but got %s", err)
	}
	// check that the Next method returns the expected vertices in the walk
	expected := []int{1, 2, 1, 2, 1, 2, 1, 2, 1, 2}
	for i := 0; i < len(expected); i++ {
		v := it.Next()
		if v.Label() != expected[i] {
			t.Errorf("Expected Next to return vertex %d at step %d, but it returned vertex %d",
				expected[i], i, v.Label())
		}
	}

	v := it.Next()
	if v != nil {
		t.Errorf("Expected Next returns nil, but got %+v", v)
	}
}

func TestRandomWalkIterator_Iterate(t *testing.T) {
	// create a graph and a starting vertex for the iterator
	g := initTestGraph()

	// create the random walk iterator
	it, err := NewRandomWalkIterator(g, 1, 10)
	if err != nil {
		t.Errorf("Expect NewRandomWalkIterator doesn't return error, but got %s", err)
	}

	// Initialize a slice to hold the visited vertices
	visited := make([]int, 0)

	// Iterate over the vertices in the walk and add their labels to the visited slice
	err = it.Iterate(func(v *gograph.Vertex[int]) error {
		visited = append(visited, v.Label())
		return nil
	})
	if err != nil {
		t.Errorf("Unexpected error during Iterate method call: %v", err)
	}

	// check that the visited slice contains the expected vertices in the walk
	expected := []int{1, 2, 1, 2, 1, 2, 1, 2, 1, 2}
	for i := 0; i < len(expected); i++ {
		if visited[i] != expected[i] {
			t.Errorf("Expected visited vertex at step %d to be %d, but it was %d", i, expected[i], visited[i])
		}
	}

	it.Reset()
	if !it.HasNext() {
		t.Errorf("Expected HasNext to return true at step %d, but it returned false", 1)
	}

	v := it.Next()
	if v.Label() != 1 {
		t.Errorf("Expected Next to return vertex %d at step %d, but it returned vertex %d",
			1, 1, v.Label())
	}

	expectedErr := errors.New("something went wrong")
	err = it.Iterate(func(vertex *gograph.Vertex[int]) error {
		return expectedErr
	})
	if err == nil {
		t.Error("Expect iter.Iterate(func) returns error, but got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expect %+v error, but got %+v", expectedErr, err)
	}
}
