package connectivity

import (
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestHasCycle_Directed(t *testing.T) {
	tests := []struct {
		name  string
		edges [][2]int
		want  bool
	}{
		{name: "empty"},
		{name: "path", edges: [][2]int{{1, 2}, {2, 3}}},
		{name: "cycle", edges: [][2]int{{1, 2}, {2, 3}, {3, 1}}, want: true},
		{name: "self loop", edges: [][2]int{{2, 2}}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gograph.New[int](gograph.Directed())
			for i := 1; i <= 4; i++ {
				g.AddVertexByLabel(i)
			}
			for _, edge := range tt.edges {
				_, _ = g.AddEdge(g.GetVertexByID(edge[0]), g.GetVertexByID(edge[1]))
			}
			if got := HasCycle(g); got != tt.want {
				t.Fatalf("HasCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasCycle_Undirected(t *testing.T) {
	tests := []struct {
		name  string
		edges [][2]int
		want  bool
	}{
		{name: "empty"},
		{name: "single edge is not a cycle", edges: [][2]int{{1, 2}}},
		{name: "tree", edges: [][2]int{{1, 2}, {1, 3}, {3, 4}}},
		{name: "triangle", edges: [][2]int{{1, 2}, {2, 3}, {3, 1}}, want: true},
		{name: "cycle in another component", edges: [][2]int{{1, 2}, {3, 4}, {4, 5}, {5, 3}}, want: true},
		{name: "self loop", edges: [][2]int{{2, 2}}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gograph.New[int]()
			for i := 1; i <= 5; i++ {
				g.AddVertexByLabel(i)
			}
			for _, edge := range tt.edges {
				_, _ = g.AddEdge(g.GetVertexByID(edge[0]), g.GetVertexByID(edge[1]))
			}
			if got := HasCycle(g); got != tt.want {
				t.Fatalf("HasCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasCycle_AcyclicGraph(t *testing.T) {
	g := gograph.New[int](gograph.Acyclic())
	_, _ = g.AddEdge(gograph.NewVertex(1), gograph.NewVertex(2))
	_, _ = g.AddEdge(gograph.NewVertex(2), gograph.NewVertex(3))
	if HasCycle(g) {
		t.Fatal("HasCycle() = true for an acyclic graph")
	}
}
