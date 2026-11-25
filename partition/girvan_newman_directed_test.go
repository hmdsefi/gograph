package partition

import (
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestGirvanNewman_DirectedEdgesBothWays(t *testing.T) {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1), gograph.WithEdgeLabel("a to b"))
	_, _ = g.AddEdge(b, a, gograph.WithEdgeWeight(5), gograph.WithEdgeLabel("b to a"))
	_, _ = g.AddEdge(c, c)

	communities, err := GirvanNewman(g, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(communities) != 2 {
		t.Fatalf("expected 2 communities, got %d", len(communities))
	}

	ab := communities[0]
	if !ab.IsDirected() {
		t.Fatal("expected the communities of a directed graph to be directed")
	}
	if ab.Size() != 2 {
		t.Fatalf("expected both edges between A and B, got %d edges", ab.Size())
	}
	for _, want := range []*gograph.Edge[string]{g.GetEdge(a, b), g.GetEdge(b, a)} {
		got := ab.GetEdge(ab.GetVertexByID(want.Source().Label()), ab.GetVertexByID(want.Destination().Label()))
		if got == nil {
			t.Fatalf("edge %s->%s: missing", want.Source().Label(), want.Destination().Label())
		}
		if got.Weight() != want.Weight() || got.Label() != want.Label() {
			t.Errorf("edge %s->%s: expected weight %v and label %q, got %v and %q",
				want.Source().Label(), want.Destination().Label(), want.Weight(), want.Label(), got.Weight(), got.Label())
		}
	}
}

func TestGirvanNewman_DirectedKeepsEveryEdge(t *testing.T) {
	// three directed triangles, each with one edge in both directions,
	// joined in a ring
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	edges := [][2]string{
		{"a1", "a2"}, {"a2", "a3"}, {"a3", "a1"}, {"a2", "a1"},
		{"b1", "b2"}, {"b2", "b3"}, {"b3", "b1"}, {"b3", "b2"},
		{"c1", "c2"}, {"c2", "c3"}, {"c3", "c1"}, {"c1", "c3"},
		{"a1", "b1"}, {"b2", "c2"}, {"c3", "a3"},
	}
	for i, e := range edges {
		_, _ = g.AddEdge(gograph.NewVertex(e[0]), gograph.NewVertex(e[1]),
			gograph.WithEdgeWeight(float64(i+1)), gograph.WithEdgeLabel(e[0]+">"+e[1]))
	}

	for k := 1; k <= 4; k++ {
		communities, err := GirvanNewman(g, k)
		if err != nil {
			t.Fatalf("k=%d: unexpected error: %v", k, err)
		}
		for i, community := range communities {
			if !community.IsDirected() {
				t.Fatalf("k=%d, community %d: expected it to be directed", k, i)
			}

			inside := 0
			for _, want := range g.AllEdges() {
				from := community.GetVertexByID(want.Source().Label())
				to := community.GetVertexByID(want.Destination().Label())
				if from == nil || to == nil {
					continue
				}
				inside++
				got := community.GetEdge(from, to)
				if got == nil {
					t.Fatalf("k=%d, community %d: missing edge %s", k, i, want.Label())
				}
				if got.Weight() != want.Weight() || got.Label() != want.Label() {
					t.Errorf("k=%d, community %d: edge %s has weight %v and label %q, expected %v and %q",
						k, i, want.Label(), got.Weight(), got.Label(), want.Weight(), want.Label())
				}
			}
			if int(community.Size()) != inside {
				t.Errorf("k=%d, community %d: expected %d edges, got %d", k, i, inside, community.Size())
			}
		}
	}
}
