package partition

import (
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestRandomizedKCut_SmallGraph(t *testing.T) {
	g := gograph.New[string]()
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	d := g.AddVertexByLabel("D")

	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(a, c)
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(b, d)
	_, _ = g.AddEdge(c, d)

	k := 2
	result, err := RandomizedKCut(g, k)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Supernodes) == 0 {
		t.Errorf("expected supernodes in k-cut, got none")
	}

	if len(result.CutEdges) == 0 {
		t.Errorf("expected edges in k-cut, got none")
	}

	var supernodes []string
	for _, supernode := range result.Supernodes {
		var str string
		for _, v := range supernode {
			str += v.Label()
		}
		supernodes = append(supernodes, str)
	}

	t.Log(supernodes)

	if len(supernodes[0])+len(supernodes[1]) != int(g.Order()) {
		t.Errorf("expected total number of nodes to be %d, got %d", g.Order(), len(supernodes[0])+len(supernodes[1]))
	}
}

func TestRandomizedKCut_KEqualsVertices(t *testing.T) {
	g := gograph.New[int]()
	v := []int{1, 2, 3}
	for _, val := range v {
		g.AddVertexByLabel(val)
	}

	_, _ = g.AddEdge(g.GetVertexByID(1), g.GetVertexByID(2))
	_, _ = g.AddEdge(g.GetVertexByID(2), g.GetVertexByID(3))
	_, _ = g.AddEdge(g.GetVertexByID(1), g.GetVertexByID(3))

	result, err := RandomizedKCut(g, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// nothing is contracted, so every edge crosses between two supernodes
	if len(result.CutEdges) != 3 {
		t.Errorf("expected 3 cut edges when k == number of vertices, got %d", len(result.CutEdges))
	}

	var got []int
	for _, supernode := range result.Supernodes {
		if len(supernode) != 1 {
			t.Fatalf("expected one vertex per supernode, got %d", len(supernode))
		}
		got = append(got, supernode[0].Label())
	}
	if !reflect.DeepEqual(got, v) {
		t.Errorf("expected supernodes in vertex order %v, got %v", v, got)
	}
}

func TestRandomizedKCut_Cycle(t *testing.T) {
	g := gograph.New[string]()
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	d := g.AddVertexByLabel("D")
	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(c, d)
	_, _ = g.AddEdge(d, a)

	// every 2-way split of a 4-cycle made by contraction cuts two edges
	for run := 0; run < 50; run++ {
		result, err := RandomizedKCut(g, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.CutEdges) != 2 {
			t.Fatalf("run %d: expected 2 cut edges, got %d", run, len(result.CutEdges))
		}
		checkCut(t, g, result)
	}
}

func TestRandomizedKCut_CutEdgesMatchSupernodes(t *testing.T) {
	for run := 0; run < 50; run++ {
		g := orderTestGraph()
		result, err := RandomizedKCut(g, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Supernodes) != 3 {
			t.Fatalf("run %d: expected 3 supernodes, got %d", run, len(result.Supernodes))
		}
		checkCut(t, g, result)
	}
}

func TestRandomizedKCut_DirectedEdgesBothWays(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(b, a)

	result, err := RandomizedKCut(g, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A->B and B->A are two edges in a directed graph
	if len(result.CutEdges) != 2 {
		t.Fatalf("expected 2 cut edges, got %d", len(result.CutEdges))
	}
}

func TestRandomizedKCut_MoreComponentsThanK(t *testing.T) {
	g := gograph.New[int]()
	for i := range 4 {
		g.AddVertexByLabel(i)
	}
	_, _ = g.AddEdge(g.GetVertexByID(0), g.GetVertexByID(1))

	result, err := RandomizedKCut(g, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 0-1, 2 and 3 are three components, so one supernode each
	if len(result.Supernodes) != 3 {
		t.Fatalf("expected 3 supernodes, got %d", len(result.Supernodes))
	}
	if len(result.CutEdges) != 0 {
		t.Fatalf("expected no cut edges, got %d", len(result.CutEdges))
	}
}

// checkCut verifies that the supernodes partition the vertices of g, and that
// the cut edges are exactly the edges of g between different supernodes, with
// each undirected edge listed once.
func checkCut[T comparable](t *testing.T, g gograph.Graph[T], result *KCutResult[T]) {
	t.Helper()

	supernodeOf := make(map[T]int)
	for i, supernode := range result.Supernodes {
		for _, v := range supernode {
			if _, ok := supernodeOf[v.Label()]; ok {
				t.Fatalf("vertex %v is in more than one supernode", v.Label())
			}
			supernodeOf[v.Label()] = i
		}
	}
	if len(supernodeOf) != int(g.Order()) {
		t.Fatalf("supernodes have %d vertices, graph has %d", len(supernodeOf), g.Order())
	}

	type pair struct{ from, to T }
	want := make(map[pair]bool)
	for _, e := range g.AllEdges() {
		from, to := e.Source().Label(), e.Destination().Label()
		if supernodeOf[from] != supernodeOf[to] && !want[pair{to, from}] {
			want[pair{from, to}] = true
		}
	}

	got := make(map[pair]bool)
	for _, e := range result.CutEdges {
		from, to := e.Source().Label(), e.Destination().Label()
		if got[pair{from, to}] || got[pair{to, from}] {
			t.Fatalf("cut edge %v-%v is listed twice", from, to)
		}
		got[pair{from, to}] = true
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d cut edges, got %d", len(want), len(got))
	}
	for p := range got {
		if !want[p] && !want[pair{p.to, p.from}] {
			t.Fatalf("edge %v-%v is not between two supernodes", p.from, p.to)
		}
	}
}

func TestRandomizedKCut_InvalidK(t *testing.T) {
	g := gograph.New[int]()
	g.AddVertexByLabel(1)
	g.AddVertexByLabel(2)

	_, err := RandomizedKCut(g, 1)
	if err == nil {
		t.Errorf("expected error for k < 2")
	}
}

func TestRandomizedKCut_EmptyGraph(t *testing.T) {
	g := gograph.New[int]()
	_, err := RandomizedKCut(g, 2)
	if err == nil {
		t.Errorf("expected error for empty graph")
	}
}

func TestRandomizedKCut_DisconnectedGraph(t *testing.T) {
	g := gograph.New[int]()
	g.AddVertexByLabel(1)
	g.AddVertexByLabel(2)
	g.AddVertexByLabel(3)

	// No edges, fully disconnected
	result, err := RandomizedKCut(g, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.CutEdges) != 0 {
		t.Errorf("expected 0 edges for disconnected graph, got %d", len(result.CutEdges))
	}
}
