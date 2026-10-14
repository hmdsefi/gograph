package gograph

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"slices"
	"testing"
)

func labelsOf[T comparable](components [][]*Vertex[T]) [][]T {
	labels := make([][]T, len(components))
	for i, component := range components {
		labels[i] = make([]T, len(component))
		for j, v := range component {
			labels[i][j] = v.label
		}
	}
	return labels
}

func graphWithEdges(directed bool, vertices int, edges [][2]int) Graph[int] {
	var g Graph[int]
	if directed {
		g = New[int](Directed())
	} else {
		g = New[int]()
	}
	for i := 1; i <= vertices; i++ {
		g.AddVertexByLabel(i)
	}
	for _, e := range edges {
		_, _ = g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1]))
	}
	return g
}

func TestConnectedComponents(t *testing.T) {
	tests := []struct {
		name     string
		directed bool
		vertices int
		edges    [][2]int
		want     [][]int
	}{
		{
			name:     "undirected, in breadth-first order",
			vertices: 6,
			edges:    [][2]int{{1, 3}, {4, 5}, {3, 2}, {1, 6}},
			want:     [][]int{{1, 3, 6, 2}, {4, 5}},
		},
		{
			name:     "directed, weak, in insertion order",
			directed: true,
			vertices: 6,
			edges:    [][2]int{{3, 1}, {2, 3}, {5, 4}},
			want:     [][]int{{1, 2, 3}, {4, 5}, {6}},
		},
		{
			name:     "undirected self-loops join nothing",
			vertices: 2,
			edges:    [][2]int{{1, 1}, {2, 2}},
			want:     [][]int{{1}, {2}},
		},
		{
			name:     "directed self-loops join nothing",
			directed: true,
			vertices: 2,
			edges:    [][2]int{{1, 1}, {2, 2}},
			want:     [][]int{{1}, {2}},
		},
		{
			name:     "directed cycle",
			directed: true,
			vertices: 3,
			edges:    [][2]int{{2, 3}, {3, 1}, {1, 2}},
			want:     [][]int{{1, 2, 3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := graphWithEdges(tt.directed, tt.vertices, tt.edges)
			if got := labelsOf(ConnectedComponents(g)); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ConnectedComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnectedComponentsEmpty(t *testing.T) {
	for _, g := range []Graph[int]{New[int](), New[int](Directed())} {
		if got := ConnectedComponents(g); got == nil || len(got) != 0 {
			t.Fatalf("ConnectedComponents() = %#v, want an empty slice", got)
		}
	}
}

func TestConnectedComponentsAfterRemovals(t *testing.T) {
	for _, directed := range []bool{false, true} {
		g := graphWithEdges(directed, 5, [][2]int{{1, 2}, {2, 3}, {3, 4}, {4, 5}})
		g.RemoveVertices(g.GetVertexByID(3))
		g.AddVertexByLabel(3)
		_, _ = g.AddEdge(g.GetVertexByID(5), g.GetVertexByID(3))

		want := [][]int{{1, 2}, {4, 5, 3}}
		if got := labelsOf(ConnectedComponents(g)); !reflect.DeepEqual(got, want) {
			t.Fatalf("directed=%v: ConnectedComponents() = %v, want %v", directed, got, want)
		}
	}
}

func TestConnectedComponentsStoredVertices(t *testing.T) {
	for _, directed := range []bool{false, true} {
		g := graphWithEdges(directed, 3, [][2]int{{1, 2}})
		components := ConnectedComponents(g)
		for _, component := range components {
			for _, v := range component {
				if v != g.GetVertexByID(v.label) {
					t.Fatalf("directed=%v: vertex %d is a copy", directed, v.label)
				}
			}
		}

		_ = append(components[0], NewVertex(99))
		if components[1][0].label != 3 {
			t.Fatalf("directed=%v: appending to a component changed the next one", directed)
		}
	}
}

func TestConnectedComponentsPanics(t *testing.T) {
	for _, g := range []Graph[int]{nil, (*baseGraph[int])(nil)} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			_ = ConnectedComponents(g)
		}()
	}
}

// referenceComponents is the breadth-first search over EdgesOf that
// GirvanNewman used before ConnectedComponents. For directed graphs, it puts
// the vertices of each component in GetAllVertices order.
func referenceComponents(g Graph[int]) [][]int {
	position := map[int]int{}
	for i, v := range g.GetAllVertices() {
		position[v.label] = i
	}

	visited := map[*Vertex[int]]bool{}
	components := [][]int{}
	for _, root := range g.GetAllVertices() {
		if visited[root] {
			continue
		}
		visited[root] = true
		queue := []*Vertex[int]{root}
		var component []int
		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			component = append(component, v.label)
			for _, e := range g.EdgesOf(v) {
				if w := e.OtherVertex(v.label); w != nil && !visited[w] {
					visited[w] = true
					queue = append(queue, w)
				}
			}
		}
		if g.IsDirected() {
			slices.SortFunc(component, func(a, b int) int { return position[a] - position[b] })
		}
		components = append(components, component)
	}
	return components
}

// randomTestGraph has up to 9 vertices, self-loops, and vertices removed and
// added again, which leave gaps in the insertion order.
func randomTestGraph(r *rand.Rand, directed bool) Graph[int] {
	n := 1 + r.IntN(9)
	g := graphWithEdges(directed, n, nil)
	for range r.IntN(2*n + 1) {
		a, b := 1+r.IntN(n), 1+r.IntN(n)
		if a == b && r.IntN(4) != 0 {
			continue
		}
		_, _ = g.AddEdge(g.GetVertexByID(a), g.GetVertexByID(b))
	}
	for range r.IntN(3) {
		if v := g.GetVertexByID(1 + r.IntN(n)); v != nil {
			g.RemoveVertices(v)
		}
	}
	if r.IntN(2) == 0 {
		v := g.AddVertexByLabel(n + 1)
		for range r.IntN(3) {
			if w := g.GetVertexByID(1 + r.IntN(n)); w != nil {
				_, _ = g.AddEdge(v, w)
			}
		}
	}
	return g
}

func TestConnectedComponentsRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(111, 250)) //nolint:gosec // seeded so failures are reproducible
	for i := range 4000 {
		g := randomTestGraph(r, i%2 == 0)
		want := referenceComponents(g)
		if got := labelsOf(ConnectedComponents(g)); !reflect.DeepEqual(got, want) {
			t.Fatalf("graph %d, directed=%v: ConnectedComponents() = %v, want %v", i, g.IsDirected(), got, want)
		}
	}
}

func ExampleConnectedComponents() {
	g := New[string]()
	_, _ = g.AddEdge(NewVertex("a"), NewVertex("b"))
	_, _ = g.AddEdge(NewVertex("b"), NewVertex("c"))
	_, _ = g.AddEdge(NewVertex("d"), NewVertex("e"))
	g.AddVertexByLabel("f")

	for _, component := range ConnectedComponents(g) {
		var labels []string
		for _, v := range component {
			labels = append(labels, v.Label())
		}
		fmt.Println(labels)
	}

	// Output:
	// [a b c]
	// [d e]
	// [f]
}
