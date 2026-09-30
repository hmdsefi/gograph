package gograph

import (
	"fmt"
	"reflect"
	"testing"
)

// orderRuns is how many times the order tests repeat a call. Go randomizes
// map iteration, so an order that comes from a map changes within a few runs.
const orderRuns = 50

func vertexLabels[T comparable](vertices []*Vertex[T]) []T {
	labels := make([]T, len(vertices))
	for i, v := range vertices {
		labels[i] = v.Label()
	}
	return labels
}

func edgePairs[T comparable](edges []*Edge[T]) []string {
	pairs := make([]string, len(edges))
	for i, e := range edges {
		pairs[i] = fmt.Sprintf("%v>%v", e.Source().Label(), e.Destination().Label())
	}
	return pairs
}

func assertVertexOrder[T comparable](t *testing.T, g Graph[T], want []T) {
	t.Helper()
	for range orderRuns {
		if got := vertexLabels(g.GetAllVertices()); !reflect.DeepEqual(got, want) {
			t.Fatalf("GetAllVertices() = %v, want %v", got, want)
		}
	}
	if got := int(g.Order()); got != len(want) {
		t.Fatalf("Order() = %d, want %d", got, len(want))
	}
}

func assertEdgeOrder(t *testing.T, name string, edges func() []*Edge[string], want []string) {
	t.Helper()
	for range orderRuns {
		if got := edgePairs(edges()); !reflect.DeepEqual(got, want) {
			t.Fatalf("%s = %v, want %v", name, got, want)
		}
	}
}

func TestGetAllVertices_InsertionOrder(t *testing.T) {
	g := New[string]()
	for _, label := range []string{"d", "b", "a", "c"} {
		g.AddVertexByLabel(label)
	}
	g.AddVertex(NewVertex("f"))
	if _, err := g.AddEdge(g.GetVertexByID("a"), NewVertex("e")); err != nil {
		t.Fatalf(testErrMsgError, err)
	}

	assertVertexOrder[string](t, g, []string{"d", "b", "a", "c", "f", "e"})
}

func TestGetAllVertices_OrderAfterRemoval(t *testing.T) {
	g := New[int](Directed())
	for i := range 10 {
		g.AddVertexByLabel(i)
	}
	_, _ = g.AddEdge(g.GetVertexByID(3), g.GetVertexByID(4))

	g.RemoveVertices(g.GetVertexByID(3), g.GetVertexByID(7))
	assertVertexOrder[int](t, g, []int{0, 1, 2, 4, 5, 6, 8, 9})

	// a label that is added again goes to the end
	g.AddVertexByLabel(3)
	assertVertexOrder[int](t, g, []int{0, 1, 2, 4, 5, 6, 8, 9, 3})

	// removing most of the vertices drops the empty slots
	g.RemoveVertices(g.GetAllVerticesByID(0, 1, 2, 4, 5, 6)...)
	assertVertexOrder[int](t, g, []int{8, 9, 3})

	g.AddVertexByLabel(10)
	g.RemoveVertices(g.GetVertexByID(9))
	assertVertexOrder[int](t, g, []int{8, 3, 10})

	g.RemoveVertices(g.GetAllVertices()...)
	assertVertexOrder[int](t, g, []int{})

	g.AddVertexByLabel(1)
	assertVertexOrder[int](t, g, []int{1})
}

func TestGetAllVertices_CopiedVertex(t *testing.T) {
	g1 := New[string]()
	a := g1.AddVertexByLabel("a")

	g2 := New[string]()
	g2.AddVertexByLabel("b")
	g2.AddVertex(a)
	g2.AddVertexByLabel("c")
	g2.RemoveVertices(g2.GetVertexByID("b"))

	assertVertexOrder[string](t, g1, []string{"a"})
	assertVertexOrder[string](t, g2, []string{"a", "c"})
}

func TestAllEdges_OrderDirected(t *testing.T) {
	g := New[string](Directed())
	for _, label := range []string{"a", "b", "c", "d"} {
		g.AddVertexByLabel(label)
	}
	for _, e := range [][2]string{{"c", "a"}, {"a", "d"}, {"a", "b"}, {"b", "c"}, {"d", "b"}} {
		if _, err := g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1])); err != nil {
			t.Fatalf(testErrMsgError, err)
		}
	}

	assertEdgeOrder(t, "AllEdges()", g.AllEdges, []string{"a>d", "a>b", "b>c", "c>a", "d>b"})

	g.RemoveEdges(g.GetEdge(g.GetVertexByID("a"), g.GetVertexByID("d")))
	assertEdgeOrder(t, "AllEdges()", g.AllEdges, []string{"a>b", "b>c", "c>a", "d>b"})

	_, _ = g.AddEdge(g.GetVertexByID("a"), g.GetVertexByID("d"))
	assertEdgeOrder(t, "AllEdges()", g.AllEdges, []string{"a>b", "a>d", "b>c", "c>a", "d>b"})
}

func TestAllEdges_OrderUndirected(t *testing.T) {
	g := New[string]()
	for _, e := range [][2]string{{"a", "b"}, {"b", "c"}, {"a", "c"}} {
		if _, err := g.AddEdge(NewVertex(e[0]), NewVertex(e[1])); err != nil {
			t.Fatalf(testErrMsgError, err)
		}
	}

	assertEdgeOrder(t, "AllEdges()", g.AllEdges, []string{"a>b", "a>c", "b>a", "b>c", "c>b", "c>a"})
}

func TestEdgesOf_Order(t *testing.T) {
	g := New[string](Directed())
	for _, label := range []string{"a", "b", "c", "d"} {
		g.AddVertexByLabel(label)
	}
	for _, e := range [][2]string{{"d", "b"}, {"b", "d"}, {"b", "c"}, {"a", "b"}, {"c", "a"}} {
		if _, err := g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1])); err != nil {
			t.Fatalf(testErrMsgError, err)
		}
	}

	b := g.GetVertexByID("b")
	assertEdgeOrder(t, "EdgesOf(b)", func() []*Edge[string] { return g.EdgesOf(b) },
		[]string{"b>d", "b>c", "a>b", "d>b"})

	// a vertex from another graph with the same label works too
	assertEdgeOrder(t, "EdgesOf(b copy)", func() []*Edge[string] { return g.EdgesOf(NewVertex("b")) },
		[]string{"b>d", "b>c", "a>b", "d>b"})
}

func TestTopologySort_Order(t *testing.T) {
	g := New[int](Acyclic())
	for i := range 20 {
		g.AddVertexByLabel(i)
	}
	if _, err := g.AddEdge(g.GetVertexByID(19), g.GetVertexByID(0)); err != nil {
		t.Fatalf(testErrMsgError, err)
	}
	if _, err := g.AddEdge(g.GetVertexByID(5), g.GetVertexByID(2)); err != nil {
		t.Fatalf(testErrMsgError, err)
	}

	// vertices without incoming edges start in insertion order
	want := []int{1, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 2, 0}
	for range orderRuns {
		sorted, err := TopologySort[int](g)
		if err != nil {
			t.Fatalf(testErrMsgError, err)
		}
		if got := vertexLabels(sorted); !reflect.DeepEqual(got, want) {
			t.Fatalf("TopologySort() = %v, want %v", got, want)
		}
	}
}
