package gograph

import (
	"cmp"
	"errors"
	"math/rand"
	"reflect"
	"testing"
)

func TestStableTopologySort(t *testing.T) {
	g := New[string](Directed())
	for _, label := range []string{"d", "c", "b", "a", "e"} {
		g.AddVertexByLabel(label)
	}
	for _, e := range [][2]string{{"c", "a"}, {"b", "a"}, {"d", "b"}} {
		if _, err := g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1])); err != nil {
			t.Fatalf(testErrMsgError, err)
		}
	}

	want := []string{"c", "d", "b", "a", "e"}
	for range orderRuns {
		sorted, err := StableTopologySort(g, cmp.Compare[string])
		if err != nil {
			t.Fatalf(testErrMsgError, err)
		}
		if got := vertexLabels(sorted); !reflect.DeepEqual(got, want) {
			t.Fatalf("StableTopologySort() = %v, want %v", got, want)
		}
	}
}

func TestStableTopologySort_Reverse(t *testing.T) {
	g := New[int](Acyclic())
	for i := range 6 {
		g.AddVertexByLabel(i)
	}
	if _, err := g.AddEdge(g.GetVertexByID(0), g.GetVertexByID(5)); err != nil {
		t.Fatalf(testErrMsgError, err)
	}

	sorted, err := StableTopologySort(g, func(a, b int) int { return cmp.Compare(b, a) })
	if err != nil {
		t.Fatalf(testErrMsgError, err)
	}
	if got, want := vertexLabels(sorted), []int{4, 3, 2, 1, 0, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("StableTopologySort() = %v, want %v", got, want)
	}
}

func TestStableTopologySort_Cycle(t *testing.T) {
	tests := map[string][][2]string{
		"cycle":     {{"a", "b"}, {"b", "c"}, {"c", "a"}},
		"self-loop": {{"a", "a"}},
	}

	for name, edges := range tests {
		t.Run(name, func(t *testing.T) {
			g := New[string](Directed())
			g.AddVertexByLabel("d")
			for _, e := range edges {
				_, _ = g.AddEdge(NewVertex(e[0]), NewVertex(e[1]))
			}

			sorted, err := StableTopologySort(g, cmp.Compare[string])
			if !errors.Is(err, ErrDAGHasCycle) {
				t.Fatalf("expected %v, got %v", ErrDAGHasCycle, err)
			}
			if sorted != nil {
				t.Fatalf("expected nil result, got %v", vertexLabels(sorted))
			}
		})
	}
}

func TestStableTopologySort_Empty(t *testing.T) {
	sorted, err := StableTopologySort(New[int](Directed()), cmp.Compare[int])
	if err != nil {
		t.Fatalf(testErrMsgError, err)
	}
	if len(sorted) != 0 {
		t.Fatalf("expected no vertices, got %v", vertexLabels(sorted))
	}
}

// smallestTopologicalOrder picks the smallest available vertex at every
// step, checking all vertices each time.
func smallestTopologicalOrder(n int, edges [][2]int) []int {
	inDegree := make([]int, n)
	for _, e := range edges {
		inDegree[e[1]]++
	}

	done := make([]bool, n)
	order := make([]int, 0, n)
	for len(order) < n {
		for v := range n {
			if done[v] || inDegree[v] != 0 {
				continue
			}
			done[v] = true
			order = append(order, v)
			for _, e := range edges {
				if e[0] == v {
					inDegree[e[1]]--
				}
			}
			break
		}
	}
	return order
}

func TestStableTopologySort_SmallestOrder(t *testing.T) {
	rng := rand.New(rand.NewSource(1)) //nolint:gosec // seeded so failures are reproducible
	const n = 40

	for range 20 {
		// edges only go forward in rank, so the graph has no cycle
		rank := rng.Perm(n)
		var edges [][2]int
		for u := range n {
			for v := range n {
				if rank[u] < rank[v] && rng.Intn(10) == 0 {
					edges = append(edges, [2]int{u, v})
				}
			}
		}

		g := New[int](Directed())
		for _, label := range rng.Perm(n) {
			g.AddVertexByLabel(label)
		}
		for _, e := range edges {
			if _, err := g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1])); err != nil {
				t.Fatalf(testErrMsgError, err)
			}
		}

		sorted, err := StableTopologySort(g, cmp.Compare[int])
		if err != nil {
			t.Fatalf(testErrMsgError, err)
		}
		if got, want := vertexLabels(sorted), smallestTopologicalOrder(n, edges); !reflect.DeepEqual(got, want) {
			t.Fatalf("StableTopologySort() = %v, want %v", got, want)
		}
	}
}

// binaryTreeGraph returns a directed graph with n vertices, where vertex i
// has edges to 2i+1 and 2i+2.
func binaryTreeGraph(n int) Graph[int] {
	g := New[int](Directed())
	for i := range n {
		g.AddVertexByLabel(i)
	}
	for i := range n {
		for _, child := range []int{2*i + 1, 2*i + 2} {
			if child < n {
				_, _ = g.AddEdge(g.GetVertexByID(i), g.GetVertexByID(child))
			}
		}
	}
	return g
}

func BenchmarkTopologySort(b *testing.B) {
	g := binaryTreeGraph(100_000)
	b.ResetTimer()
	for range b.N {
		_, _ = TopologySort(g)
	}
}

func BenchmarkStableTopologySort(b *testing.B) {
	g := binaryTreeGraph(100_000)
	b.ResetTimer()
	for range b.N {
		_, _ = StableTopologySort(g, cmp.Compare[int])
	}
}
