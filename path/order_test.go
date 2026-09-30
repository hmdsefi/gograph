package path

import (
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestTransitiveReduction_StableOrder(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	edges := [][2]string{
		{"a", "b"}, {"a", "c"}, {"a", "e"}, {"a", "d"},
		{"b", "e"}, {"c", "e"}, {"d", "e"},
	}
	for _, e := range edges {
		_, _ = g.AddEdge(gograph.NewVertex(e[0]), gograph.NewVertex(e[1]))
	}

	// the edges of the result follow the order of the input edges
	want := []string{"a>b", "a>c", "a>d", "b>e", "c>e", "d>e"}
	for range 50 {
		reduced, err := TransitiveReduction(g)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var got []string
		for _, e := range reduced.AllEdges() {
			got = append(got, e.Source().Label()+">"+e.Destination().Label())
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("TransitiveReduction() edges = %v, want %v", got, want)
		}
	}
}
