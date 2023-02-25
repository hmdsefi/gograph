package traverse

import (
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph/graph"
)

func TestDepthFirstIterator(t *testing.T) {
	// Create a new graph
	g := graph.New[string](graph.Directed())

	// the example graph
	//	A -> B -> C
	//	|    |    |
	//	v    v    v
	//	D -> E -> F

	vertices := map[string]*graph.Vertex[string]{
		"A": g.AddVertexByLabel("A"),
		"B": g.AddVertexByLabel("B"),
		"C": g.AddVertexByLabel("C"),
		"D": g.AddVertexByLabel("D"),
		"E": g.AddVertexByLabel("E"),
		"F": g.AddVertexByLabel("F"),
	}

	// Add some edges
	_, _ = g.AddEdge(vertices["A"], vertices["B"])
	_, _ = g.AddEdge(vertices["A"], vertices["D"])
	_, _ = g.AddEdge(vertices["B"], vertices["C"])
	_, _ = g.AddEdge(vertices["B"], vertices["E"])
	_, _ = g.AddEdge(vertices["C"], vertices["F"])
	_, _ = g.AddEdge(vertices["D"], vertices["E"])
	_, _ = g.AddEdge(vertices["E"], vertices["F"])

	// Test depth first iteration
	iter := NewDepthFirstIterator[string](g, "A")
	expected := []string{"A", "D", "E", "F", "B", "C"}

	for i, label := range expected {
		if !iter.HasNext() {
			t.Errorf("Expected iter.HasNext() to be true, but it was false for label %s", label)
		}

		v := iter.Next()
		if v.Label() != expected[i] {
			t.Errorf("Expected iter.Next().Label() to be %s, but got %s", expected[i], v.Label())
		}
	}

	if iter.HasNext() {
		t.Error("Expected iter.HasNext() to be false, but it was true")
	}

	// test the Reset method
	iter.Reset()
	if !iter.HasNext() {
		t.Error("Expected iter.HasNext() to be true, but it was false after reset")
	}

	v := iter.Next()
	if v.Label() != "A" {
		t.Errorf("Expected iter.Next().Label() to be %s, but got %s", "A", v.Label())
	}

	// test Iterate method
	iter.Reset()
	var ordered []string
	err := iter.Iterate(func(vertex *graph.Vertex[string]) error {
		ordered = append(ordered, vertex.Label())
		return nil
	})
	if err != nil {
		t.Errorf("Expect iter.Iterate(func) returns no error, but got one %s", err)
	}

	if !reflect.DeepEqual(expected, ordered) {
		t.Errorf("Expect same vertex order, but got different one expected: %v, actual: %v",
			expected, ordered)
	}
}
