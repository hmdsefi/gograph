package partition

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

// orderTestGraph has three triangles joined in a ring, so several edges
// share the highest betweenness and there are many maximal cliques.
func orderTestGraph() gograph.Graph[string] {
	g := gograph.New[string]()
	edges := [][2]string{
		{"a1", "a2"}, {"a2", "a3"}, {"a3", "a1"},
		{"b1", "b2"}, {"b2", "b3"}, {"b3", "b1"},
		{"c1", "c2"}, {"c2", "c3"}, {"c3", "c1"},
		{"a1", "b1"}, {"b2", "c2"}, {"c3", "a3"},
	}
	for _, e := range edges {
		_, _ = g.AddEdge(gograph.NewVertex(e[0]), gograph.NewVertex(e[1]))
	}
	return g
}

func communityLabels(graphs []gograph.Graph[string]) []string {
	var out []string
	for _, g := range graphs {
		var labels []string
		for _, v := range g.GetAllVertices() {
			labels = append(labels, v.Label())
		}
		var edges []string
		for _, e := range g.AllEdges() {
			edges = append(edges, e.Source().Label()+">"+e.Destination().Label())
		}
		out = append(out, fmt.Sprint(labels, edges))
	}
	return out
}

func TestMaximalCliques_StableOrder(t *testing.T) {
	want := cliqueLabels(MaximalCliques(orderTestGraph()))
	for range 50 {
		if got := cliqueLabels(MaximalCliques(orderTestGraph())); !reflect.DeepEqual(got, want) {
			t.Fatalf("MaximalCliques() = %v, want %v", got, want)
		}
	}
}

func TestGirvanNewman_StableOrder(t *testing.T) {
	run := func() []string {
		communities, err := GirvanNewman(orderTestGraph(), 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return communityLabels(communities)
	}

	want := run()
	for range 50 {
		if got := run(); !reflect.DeepEqual(got, want) {
			t.Fatalf("GirvanNewman() = %v, want %v", got, want)
		}
	}
}

func TestGirvanNewman_SelfLoop(t *testing.T) {
	// a self-loop is on no shortest path, so it is never removed
	g := orderTestGraph()
	_, _ = g.AddEdge(g.GetVertexByID("b3"), g.GetVertexByID("b3"))

	communities, err := GirvanNewman(g, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got [][]string
	for _, c := range communities {
		var labels []string
		for _, v := range c.GetAllVertices() {
			labels = append(labels, v.Label())
		}
		got = append(got, labels)
	}
	want := [][]string{{"a1", "a2", "a3"}, {"b1", "b2", "b3"}, {"c1", "c2", "c3"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GirvanNewman() = %v, want %v", got, want)
	}

	b3 := communities[1].GetVertexByID("b3")
	if communities[1].GetEdge(b3, b3) == nil {
		t.Fatal("expected the self-loop on b3 to stay")
	}
}
