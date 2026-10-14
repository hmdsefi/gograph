package gograph

import (
	"fmt"
	"reflect"
	"testing"
)

type cloneWrapper[T comparable] struct {
	Graph[T]
}

func TestClone(t *testing.T) {
	for _, options := range [][]GraphOptionFunc{nil, {Directed()}, {Weighted()}, {Directed(), Weighted()}, {Acyclic(), Weighted()}} {
		for _, wrapped := range []bool{false, true} {
			t.Run(fmt.Sprintf("options=%d/wrapped=%v", len(options), wrapped), func(t *testing.T) {
				g := New[string](options...)
				a := g.AddVertexByLabel("a", WithVertexWeight(3))
				b := g.AddVertexByLabel("b", WithVertexWeight(4))
				g.AddVertexByLabel("isolated")
				a.metadata = "vertex"
				e, err := g.AddEdge(a, b, WithEdgeWeight(7), WithEdgeLabel("ab"))
				if err != nil {
					t.Fatal(err)
				}
				e.metadata = "forward"
				if !g.IsDirected() {
					g.GetEdge(b, a).metadata = "reverse"
					g.GetEdge(b, a).properties.weight = 8
				}
				if !g.IsAcyclic() {
					if _, err := g.AddEdge(b, b, WithEdgeLabel("loop")); err != nil {
						t.Fatal(err)
					}
				}
				source := g
				if wrapped {
					source = cloneWrapper[string]{g}
				}
				copy := Clone(source)
				if copy.IsDirected() != g.IsDirected() || copy.IsWeighted() != g.IsWeighted() || copy.IsAcyclic() != g.IsAcyclic() {
					t.Fatal("properties changed")
				}
				if copy.Order() != g.Order() || copy.Size() != g.Size() {
					t.Fatal("counts changed")
				}
				if !reflect.DeepEqual(vertexLabels(copy.GetAllVertices()), vertexLabels(g.GetAllVertices())) {
					t.Fatal("vertex order changed")
				}
				for i, original := range g.AllEdges() {
					got := copy.AllEdges()[i]
					if got == original || got.Source() == original.Source() || got.Destination() == original.Destination() {
						t.Fatal("shared graph objects")
					}
					if got.Source().Label() != original.Source().Label() || got.Destination().Label() != original.Destination().Label() || got.Weight() != original.Weight() || got.Label() != original.Label() || got.Metadata() != original.Metadata() {
						t.Fatal("edge data or order changed")
					}
					if got.InsertionIndex() != original.InsertionIndex() {
						t.Fatal("edge insertion index changed")
					}
				}
				ca := copy.GetVertexByID("a")
				if ca == a || ca.Weight() != 3 || ca.Metadata() != "vertex" {
					t.Fatal("vertex not copied")
				}
				ca.properties.weight = 99
				ca.metadata = "changed"
				copy.RemoveVertices(copy.GetVertexByID("b"))
				if a.Weight() != 3 || a.Metadata() != "vertex" || !g.ContainsEdge(a, b) {
					t.Fatal("copy changed original")
				}
				g.RemoveVertices(a)
				if copy.GetVertexByID("a") == nil {
					t.Fatal("original changed copy")
				}
			})
		}
	}
}

func TestCloneNewEdgeInsertionIndex(t *testing.T) {
	for _, wrapped := range []bool{false, true} {
		g := New[int](Directed())
		a := g.AddVertexByLabel(1)
		b := g.AddVertexByLabel(2)
		e, err := g.AddEdge(a, b)
		if err != nil {
			t.Fatal(err)
		}
		source := g
		if wrapped {
			source = cloneWrapper[int]{g}
		}
		copy := Clone(source)
		added, err := copy.AddEdge(copy.GetVertexByID(2), NewVertex(3))
		if err != nil || added.InsertionIndex() <= e.InsertionIndex() {
			t.Fatalf("new edge index: %v, error: %v", added, err)
		}
	}
}

func TestCloneEmptyAndNil(t *testing.T) {
	for _, g := range []Graph[int]{New[int](Acyclic(), Weighted()), cloneWrapper[int]{New[int](Acyclic(), Weighted())}} {
		copy := Clone(g)
		if copy.Order() != 0 || copy.Size() != 0 || !copy.IsDirected() || !copy.IsAcyclic() || !copy.IsWeighted() {
			t.Fatal("empty graph properties changed")
		}
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil graph did not panic")
		}
	}()
	Clone[int](nil)
}

func TestCloneMetadataReference(t *testing.T) {
	g := New[int]()
	v := g.AddVertexByLabel(1)
	data := &struct{ Value int }{7}
	v.metadata = data
	copy := Clone(g)
	if copy.GetVertexByID(1).Metadata() != data {
		t.Fatal("metadata interface value changed")
	}
}

func ExampleClone() {
	g := New[string](Directed(), Weighted())
	a := g.AddVertexByLabel("a")
	b := g.AddVertexByLabel("b")
	_, err := g.AddEdge(a, b, WithEdgeWeight(5))
	if err != nil {
		panic(err)
	}
	copy := Clone(g)
	copy.RemoveVertices(copy.GetVertexByID("a"))
	fmt.Println(g.Order(), copy.Order(), copy.IsDirected(), copy.IsWeighted())
	// Output: 2 1 true true
}
