package partition

import (
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

	if uint32(len(supernodes[0])+len(supernodes[1])) != g.Order() {
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

	if len(result.CutEdges) != 0 {
		t.Errorf("expected 0 edges when k == number of vertices, got %d", len(result.CutEdges))
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
