package connectivity

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

// intGraph returns a graph with the vertices 1 to n, added in order, and the
// given edges.
func intGraph(directed bool, n int, edges [][2]int) gograph.Graph[int] {
	var options []gograph.GraphOptionFunc
	if directed {
		options = append(options, gograph.Directed())
	}
	g := gograph.New[int](options...)
	for i := 1; i <= n; i++ {
		g.AddVertexByLabel(i)
	}
	for _, e := range edges {
		_, _ = g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1]))
	}
	return g
}

// checkCycle checks that cycle is a cycle of g made of the graph's own
// vertices: no vertex repeats, and an edge joins each vertex to the next and
// the last vertex to the first.
func checkCycle(t *testing.T, g gograph.Graph[int], cycle []*gograph.Vertex[int]) {
	t.Helper()
	if len(cycle) == 0 {
		t.Fatal("empty cycle")
	}
	if !g.IsDirected() && len(cycle) == 2 {
		t.Fatalf("cycle %v goes back over the edge it came from", labelsOf(cycle))
	}

	seen := map[*gograph.Vertex[int]]bool{}
	for i, v := range cycle {
		if g.GetVertexByID(v.Label()) != v {
			t.Fatalf("vertex %d isn't the graph's own vertex", v.Label())
		}
		if seen[v] {
			t.Fatalf("cycle %v repeats vertex %d", labelsOf(cycle), v.Label())
		}
		seen[v] = true
		next := cycle[(i+1)%len(cycle)]
		if g.GetEdge(v, next) == nil {
			t.Fatalf("cycle %v: no edge %d -> %d", labelsOf(cycle), v.Label(), next.Label())
		}
	}
}

func labelsOf(vertices []*gograph.Vertex[int]) []int {
	out := make([]int, len(vertices))
	for i, v := range vertices {
		out[i] = v.Label()
	}
	return out
}

func TestFindCycle(t *testing.T) {
	tests := []struct {
		name     string
		directed bool
		edges    [][2]int
		want     []int
	}{
		{name: "directed empty", directed: true},
		{name: "directed path", directed: true, edges: [][2]int{{1, 2}, {2, 3}}},
		{name: "directed diamond", directed: true, edges: [][2]int{{1, 2}, {1, 3}, {2, 4}, {3, 4}}},
		{name: "directed cycle", directed: true, edges: [][2]int{{1, 2}, {2, 3}, {3, 1}}, want: []int{1, 2, 3}},
		{name: "directed two-vertex cycle", directed: true, edges: [][2]int{{1, 2}, {2, 1}}, want: []int{1, 2}},
		{name: "directed self-loop", directed: true, edges: [][2]int{{1, 2}, {2, 2}}, want: []int{2}},
		{
			// the search reaches 4 -> 2 before 3 -> 1
			name:     "directed cycles that share vertices",
			directed: true,
			edges:    [][2]int{{1, 2}, {2, 3}, {3, 4}, {4, 2}, {3, 1}},
			want:     []int{2, 3, 4},
		},
		{
			name:     "directed cycle in a later search",
			directed: true,
			edges:    [][2]int{{1, 2}, {3, 4}, {4, 5}, {5, 3}},
			want:     []int{3, 4, 5},
		},
		{name: "undirected empty"},
		{name: "undirected single edge", edges: [][2]int{{1, 2}}},
		{name: "undirected tree", edges: [][2]int{{1, 2}, {1, 3}, {3, 4}, {3, 5}}},
		{name: "undirected triangle", edges: [][2]int{{1, 2}, {2, 3}, {3, 1}}, want: []int{1, 2, 3}},
		{name: "undirected square", edges: [][2]int{{1, 2}, {2, 3}, {3, 4}, {4, 1}}, want: []int{1, 2, 3, 4}},
		{name: "undirected self-loop", edges: [][2]int{{1, 2}, {2, 2}}, want: []int{2}},
		{name: "undirected cycle in a later search", edges: [][2]int{{1, 2}, {3, 4}, {4, 5}, {5, 3}}, want: []int{3, 4, 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := intGraph(tt.directed, 5, tt.edges)
			cycle, ok := FindCycle(g)
			if ok != (tt.want != nil) {
				t.Fatalf("FindCycle() reported %v, want %v", ok, tt.want != nil)
			}
			if !ok {
				if cycle != nil {
					t.Fatalf("expected nil, got %v", labelsOf(cycle))
				}
				return
			}
			checkCycle(t, g, cycle)
			if got := labelsOf(cycle); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("FindCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindCycle_MatchesHasCycle(t *testing.T) {
	for seed := range uint64(500) {
		rng := rand.New(rand.NewPCG(seed, 132)) //nolint:gosec // seeded so failures are reproducible
		for _, directed := range []bool{true, false} {
			n := 1 + rng.IntN(10)
			var edges [][2]int
			for range rng.IntN(n + 3) {
				edges = append(edges, [2]int{1 + rng.IntN(n), 1 + rng.IntN(n)})
			}
			g := intGraph(directed, n, edges)

			cycle, ok := FindCycle(g)
			if ok != HasCycle(g) {
				t.Fatalf("seed %d, directed %v: FindCycle reported %v, HasCycle %v", seed, directed, ok, !ok)
			}
			if ok {
				checkCycle(t, g, cycle)
			}
		}
	}
}

func TestFindCycle_AcyclicGraph(t *testing.T) {
	g := gograph.New[int](gograph.Acyclic())
	_, _ = g.AddEdge(gograph.NewVertex(1), gograph.NewVertex(2))
	_, _ = g.AddEdge(gograph.NewVertex(2), gograph.NewVertex(3))
	if cycle, ok := FindCycle(g); ok || cycle != nil {
		t.Fatalf("FindCycle() = %v, %v for an acyclic graph", cycle, ok)
	}
}

func TestFindCycle_WrappedGraph(t *testing.T) {
	g := intGraph(true, 3, [][2]int{{1, 2}, {2, 3}, {3, 1}})
	cycle, ok := FindCycle[int](wrappedGraph[int]{g})
	if !ok {
		t.Fatal("FindCycle(wrapper) found no cycle")
	}
	checkCycle(t, g, cycle)
}

func BenchmarkCycleSearch(b *testing.B) {
	// without a cycle, both functions search the whole graph
	const n, m = 10000, 50000
	rng := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // seeded so runs are comparable
	dag := gograph.New[int](gograph.Directed())
	for i := range n {
		dag.AddVertexByLabel(i)
	}
	for dag.Size() < m {
		from, to := rng.IntN(n), rng.IntN(n)
		if from < to {
			_, _ = dag.AddEdge(dag.GetVertexByID(from), dag.GetVertexByID(to))
		}
	}

	b.Run("HasCycle", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if HasCycle(dag) {
				b.Fatal("found a cycle")
			}
		}
	})
	b.Run("FindCycle", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, ok := FindCycle(dag); ok {
				b.Fatal("found a cycle")
			}
		}
	})
}

func TestFindCycle_LongCycle(t *testing.T) {
	const n = 100_000
	for _, directed := range []bool{true, false} {
		edges := make([][2]int, n)
		for i := range n {
			edges[i] = [2]int{i + 1, (i+1)%n + 1}
		}
		g := intGraph(directed, n, edges)

		cycle, ok := FindCycle(g)
		if !ok || len(cycle) != n {
			t.Fatalf("directed %v: found %v with %d vertices, want all %d", directed, ok, len(cycle), n)
		}
		checkCycle(t, g, cycle)
	}
}
