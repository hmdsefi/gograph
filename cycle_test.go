package gograph

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

func TestHasCycle(t *testing.T) {
	tests := []struct {
		name     string
		directed bool
		edges    [][2]int
		want     bool
	}{
		{name: "undirected empty"},
		{name: "undirected edge", edges: [][2]int{{1, 2}}},
		{name: "undirected tree", edges: [][2]int{{1, 2}, {1, 3}, {3, 4}}},
		{name: "undirected triangle", edges: [][2]int{{1, 2}, {2, 3}, {3, 1}}, want: true},
		{name: "undirected cycle in a later component", edges: [][2]int{{1, 2}, {3, 4}, {4, 5}, {5, 3}}, want: true},
		{name: "undirected self-loop", edges: [][2]int{{2, 2}}, want: true},
		{name: "directed empty", directed: true},
		{name: "directed path", directed: true, edges: [][2]int{{1, 2}, {2, 3}}},
		{name: "directed diamond", directed: true, edges: [][2]int{{1, 2}, {1, 3}, {2, 4}, {3, 4}}},
		{name: "directed two-vertex cycle", directed: true, edges: [][2]int{{1, 2}, {2, 1}}, want: true},
		{name: "directed cycle the first vertex can't reach", directed: true, edges: [][2]int{{1, 2}, {3, 4}, {4, 5}, {5, 3}}, want: true},
		{name: "directed self-loop", directed: true, edges: [][2]int{{2, 2}}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := graphWithEdges(tt.directed, 5, tt.edges)
			if got := HasCycle(g); got != tt.want {
				t.Fatalf("HasCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasCycleAcyclic(t *testing.T) {
	g := New[int](Acyclic())
	_, _ = g.AddEdge(NewVertex(1), NewVertex(2))
	_, _ = g.AddEdge(NewVertex(2), NewVertex(3))
	_, _ = g.AddEdge(NewVertex(1), NewVertex(3))
	if _, err := g.AddEdge(NewVertex(3), NewVertex(1)); err == nil {
		t.Fatal("the acyclic graph accepted an edge that closes a cycle")
	}
	if HasCycle(g) {
		t.Fatal("HasCycle() = true for an acyclic graph")
	}
}

func TestHasCycleAfterRemovals(t *testing.T) {
	for _, directed := range []bool{false, true} {
		g := graphWithEdges(directed, 3, [][2]int{{1, 2}, {2, 3}, {3, 1}})
		g.RemoveVertices(g.GetVertexByID(2))
		if HasCycle(g) {
			t.Fatalf("directed=%v: HasCycle() = true after the cycle lost a vertex", directed)
		}
	}
}

func TestHasCyclePanics(t *testing.T) {
	for _, g := range []Graph[int]{nil, (*baseGraph[int])(nil)} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			_ = HasCycle(g)
		}()
	}
}

// TestHasCycleRandom checks directed graphs against TopologySort. A simple
// undirected graph has a cycle when it has a self-loop, or more edges than
// vertices minus components.
func TestHasCycleRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(111, 251)) //nolint:gosec // seeded so failures are reproducible
	for i := range 4000 {
		g := randomTestGraph(r, i%2 == 0)

		var want bool
		if g.IsDirected() {
			_, err := TopologySort(g)
			want = err != nil
		} else {
			edges := 0
			for _, e := range g.AllEdges() {
				if e.source == e.dest {
					want = true
				}
				edges++
			}
			want = want || edges/2 > len(g.GetAllVertices())-len(referenceComponents(g))
		}

		if got := HasCycle(g); got != want {
			t.Fatalf("graph %d, directed=%v: HasCycle() = %v, want %v", i, g.IsDirected(), got, want)
		}
	}
}

func ExampleHasCycle() {
	g := New[string](Directed())
	_, _ = g.AddEdge(NewVertex("build"), NewVertex("test"))
	_, _ = g.AddEdge(NewVertex("test"), NewVertex("deploy"))
	fmt.Println(HasCycle(g))

	_, _ = g.AddEdge(g.GetVertexByID("deploy"), g.GetVertexByID("build"))
	fmt.Println(HasCycle(g))

	// Output:
	// false
	// true
}
