package gograph

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"strings"
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

func TestCloneChangesLikeOriginal(t *testing.T) {
	r := rand.New(rand.NewPCG(108, 1)) //nolint:gosec // seeded so failures are reproducible
	for trial := range 500 {
		options := [][]GraphOptionFunc{nil, {Directed()}, {Directed(), Weighted()}}[trial%3]
		g := New[int](options...)
		for v := range 8 {
			g.AddVertexByLabel(v)
		}
		for range 16 {
			_, _ = g.AddEdge(g.GetVertexByID(r.IntN(8)), g.GetVertexByID(r.IntN(8)), WithEdgeWeight(float64(r.IntN(5))))
		}
		g.RemoveVertices(g.GetVertexByID(r.IntN(8)))

		wrapped := trial%2 == 1
		source := g
		if wrapped {
			source = cloneWrapper[int]{g}
		}
		clone := Clone(source)
		for step := range 20 {
			from, to, removeVertex := r.IntN(9), r.IntN(9), r.IntN(4) == 0
			for _, h := range []Graph[int]{g, clone} {
				switch step % 4 {
				case 0, 1:
					_, _ = h.AddEdge(NewVertex(from), NewVertex(to))
				case 2:
					if e := h.GetEdge(h.GetVertexByID(from), h.GetVertexByID(to)); e != nil {
						h.RemoveEdges(e)
					}
				default:
					if v := h.GetVertexByID(from); v != nil && removeVertex {
						h.RemoveVertices(v)
					}
				}
			}
			if got, want := graphState(clone, wrapped), graphState(g, wrapped); got != want {
				t.Fatalf("trial %d step %d: copy\n%s\nwant\n%s", trial, step, got, want)
			}
		}
	}
}

func graphState(g Graph[int], skipIndexes bool) string {
	var b strings.Builder
	for _, v := range g.GetAllVertices() {
		fmt.Fprintf(&b, "%d in=%d out=%d;", v.Label(), v.InDegree(), v.OutDegree())
	}
	for _, e := range g.AllEdges() {
		index := e.InsertionIndex()
		if skipIndexes {
			index = 0
		}
		fmt.Fprintf(&b, " %d>%d w=%g i=%d", e.Source().Label(), e.Destination().Label(), e.Weight(), index)
	}
	return b.String()
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
