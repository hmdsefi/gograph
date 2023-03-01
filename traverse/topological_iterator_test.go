package traverse

import (
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestTopologyOrderIterator(t *testing.T) {
	// create the graph
	g := gograph.New[int](gograph.Acyclic())

	// add vertices to the graph
	vertices := make(map[int]*gograph.Vertex[int])
	for i := 1; i <= 6; i++ {
		vertices[i] = g.AddVertexByLabel(i)
	}

	// add edges to the graph
	_, _ = g.AddEdge(vertices[1], vertices[2])
	_, _ = g.AddEdge(vertices[1], vertices[3])
	_, _ = g.AddEdge(vertices[2], vertices[4])
	_, _ = g.AddEdge(vertices[2], vertices[5])
	_, _ = g.AddEdge(vertices[3], vertices[5])
	_, _ = g.AddEdge(vertices[4], vertices[6])
	_, _ = g.AddEdge(vertices[5], vertices[6])

	// create the topology order iterator
	iterator, err := NewTopologicalIterator[int](g)
	if err != nil {
		t.Errorf("Expect no error by calling NewTopologicalIterator, but got one, %s", err)
	}

	// test the Next method
	expectedOrder := []int{1, 2, 3, 4, 5, 6}
	for i := 0; i < 6; i++ {
		if !iterator.HasNext() {
			t.Errorf("Expected iterator.HasNext() to be true, but it was false for index %d", i)
		}

		v := iterator.Next()
		if v.Label() != expectedOrder[i] {
			t.Errorf("Expected iterator.Next().Label() to be %d, but got %d", expectedOrder[i], v.Label())
		}
	}

	if iterator.HasNext() {
		t.Error("Expected iterator.HasNext() to be false, but it was true")
	}

	// test the Reset method
	iterator.Reset()
	if !iterator.HasNext() {
		t.Error("Expected iterator.HasNext() to be true, but it was false after reset")
	}

	v := iterator.Next()
	if v.Label() != 1 {
		t.Errorf("Expected iterator.Next().Label() to be %d, but got %d", 1, v.Label())
	}

	// test Iterate method
	iterator.Reset()
	var ordered []int
	err = iterator.Iterate(func(vertex *gograph.Vertex[int]) error {
		ordered = append(ordered, vertex.Label())
		return nil
	})
	if err != nil {
		t.Errorf("Expect iterator.Iterate(func) returns no error, but got one %s", err)
	}

	if !reflect.DeepEqual(expectedOrder, ordered) {
		t.Errorf("Expect same vertex order, but got different one expected: %v, actual: %v",
			expectedOrder, ordered)
	}
}
