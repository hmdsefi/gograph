package traverse

import (
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

// cycleGraph returns the graph A -> B -> C -> A, with weights on the
// vertices and edges.
func cycleGraph(options ...gograph.GraphOptionFunc) gograph.Graph[string] {
	g := gograph.New[string](append([]gograph.GraphOptionFunc{gograph.Directed()}, options...)...)
	a := g.AddVertexByLabel("A", gograph.WithVertexWeight(1))
	b := g.AddVertexByLabel("B", gograph.WithVertexWeight(2))
	c := g.AddVertexByLabel("C", gograph.WithVertexWeight(3))
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(c, a, gograph.WithEdgeWeight(3))
	return g
}

func TestIterators_ReturnGraphVertices(t *testing.T) {
	tests := []struct {
		name  string
		graph gograph.Graph[string]
		iter  func(g gograph.Graph[string]) (Iterator[string], error)
	}{
		{
			name:  "closest first",
			graph: cycleGraph(gograph.Weighted()),
			iter:  func(g gograph.Graph[string]) (Iterator[string], error) { return NewClosestFirstIterator(g, "A") },
		},
		{
			name:  "random walk",
			graph: cycleGraph(),
			iter:  func(g gograph.Graph[string]) (Iterator[string], error) { return NewRandomWalkIterator(g, "A", 10) },
		},
		{
			name:  "weighted random walk",
			graph: cycleGraph(gograph.Weighted()),
			iter:  func(g gograph.Graph[string]) (Iterator[string], error) { return NewRandomWalkIterator(g, "A", 10) },
		},
		{
			name:  "breadth first",
			graph: cycleGraph(),
			iter:  func(g gograph.Graph[string]) (Iterator[string], error) { return NewBreadthFirstIterator(g, "A") },
		},
		{
			name:  "depth first",
			graph: cycleGraph(),
			iter:  func(g gograph.Graph[string]) (Iterator[string], error) { return NewDepthFirstIterator(g, "A") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it, err := tt.iter(tt.graph)
			if err != nil {
				t.Fatalf("Expect no error, but got %s", err)
			}

			visited := 0
			for it.HasNext() {
				v := it.Next()
				visited++
				if want := tt.graph.GetVertexByID(v.Label()); v != want {
					t.Errorf("Expect the graph's vertex %s, but got a copy with weight %v", v.Label(), v.Weight())
				}
			}

			if visited < 3 {
				t.Errorf("Expect at least 3 vertices, but got %d", visited)
			}
		})
	}
}

func TestIterators_SkipRemovedVertices(t *testing.T) {
	tests := []struct {
		name string
		iter func(g gograph.Graph[string]) (Iterator[string], error)
	}{
		{
			name: "closest first",
			iter: func(g gograph.Graph[string]) (Iterator[string], error) { return NewClosestFirstIterator(g, "A") },
		},
		{
			name: "breadth first",
			iter: func(g gograph.Graph[string]) (Iterator[string], error) { return NewBreadthFirstIterator(g, "A") },
		},
		{
			name: "depth first",
			iter: func(g gograph.Graph[string]) (Iterator[string], error) { return NewDepthFirstIterator(g, "A") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A -> B -> D and A -> C
			g := gograph.New[string](gograph.Directed(), gograph.Weighted())
			a := g.AddVertexByLabel("A")
			b := g.AddVertexByLabel("B")
			c := g.AddVertexByLabel("C")
			d := g.AddVertexByLabel("D")
			_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1))
			_, _ = g.AddEdge(a, c, gograph.WithEdgeWeight(2))
			_, _ = g.AddEdge(b, d, gograph.WithEdgeWeight(1))

			it, err := tt.iter(g)
			if err != nil {
				t.Fatalf("Expect no error, but got %s", err)
			}

			// visit A, which queues B and C, then remove B
			if v := it.Next(); v.Label() != "A" {
				t.Fatalf("Expect A, but got %s", v.Label())
			}
			g.RemoveVertices(b)

			var visited []string
			for it.HasNext() {
				visited = append(visited, it.Next().Label())
			}

			if !reflect.DeepEqual(visited, []string{"C"}) {
				t.Errorf("Expect [C] after removing B, but got %v", visited)
			}

			// remove the start vertex before Reset
			g.RemoveVertices(a)
			it.Reset()
			if it.HasNext() {
				t.Error("Expect HasNext to be false after removing the start vertex, but it was true")
			}

			if v := it.Next(); v != nil {
				t.Errorf("Expect nil, but got %+v", v)
			}
		})
	}
}
