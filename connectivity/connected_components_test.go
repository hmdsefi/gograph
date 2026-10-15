package connectivity

import (
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

func componentLabels[T comparable](components [][]*gograph.Vertex[T]) [][]T {
	labels := make([][]T, len(components))
	for i, component := range components {
		labels[i] = make([]T, len(component))
		for j, vertex := range component {
			labels[i][j] = vertex.Label()
		}
	}
	return labels
}

func TestConnectedComponents_Undirected(t *testing.T) {
	g := gograph.New[int]()
	for i := 1; i <= 7; i++ {
		g.AddVertexByLabel(i)
	}
	for _, edge := range [][2]int{{1, 2}, {2, 3}, {4, 5}, {5, 6}, {6, 4}} {
		_, _ = g.AddEdge(g.GetVertexByID(edge[0]), g.GetVertexByID(edge[1]))
	}

	want := [][]int{{1, 2, 3}, {4, 5, 6}, {7}}
	if got := componentLabels(ConnectedComponents(g)); !reflect.DeepEqual(got, want) {
		t.Fatalf("ConnectedComponents() = %v, want %v", got, want)
	}
}

func TestConnectedComponents_DirectedAreWeak(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	for _, label := range []string{"a", "b", "c", "d", "e", "f"} {
		g.AddVertexByLabel(label)
	}
	for _, edge := range [][2]string{{"b", "a"}, {"c", "b"}, {"d", "e"}} {
		_, _ = g.AddEdge(g.GetVertexByID(edge[0]), g.GetVertexByID(edge[1]))
	}

	want := [][]string{{"a", "b", "c"}, {"d", "e"}, {"f"}}
	if got := componentLabels(ConnectedComponents(g)); !reflect.DeepEqual(got, want) {
		t.Fatalf("ConnectedComponents() = %v, want %v", got, want)
	}
}

type wrappedGraph[T comparable] struct{ gograph.Graph[T] }

func TestConnectedComponents_WrappedGraph(t *testing.T) {
	g := gograph.New[int]()
	for i := 1; i <= 4; i++ {
		g.AddVertexByLabel(i)
	}
	_, _ = g.AddEdge(g.GetVertexByID(1), g.GetVertexByID(2))

	if got, want := componentLabels(ConnectedComponents(wrappedGraph[int]{g})), componentLabels(ConnectedComponents(g)); !reflect.DeepEqual(got, want) {
		t.Fatalf("ConnectedComponents(wrapper) = %v, want %v", got, want)
	}
}

func TestConnectedComponents_Empty(t *testing.T) {
	if got := ConnectedComponents(gograph.New[int]()); len(got) != 0 {
		t.Fatalf("ConnectedComponents() = %v, want no components", got)
	}
}
