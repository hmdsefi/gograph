package traverse

import (
	"reflect"
	"sort"
	"testing"

	"github.com/hmdsefi/gograph"
)

var removalIterators = []struct {
	name string
	iter func(g gograph.Graph[string], start string) (Iterator[string], error)
}{
	{name: "breadth first", iter: NewBreadthFirstIterator[string]},
	{name: "depth first", iter: NewDepthFirstIterator[string]},
	{name: "closest first", iter: NewClosestFirstIterator[string]},
}

// drain returns the labels of the vertices left in the iteration. It fails
// the test if HasNext and Next disagree, or if Next returns a vertex that
// isn't in the graph anymore.
func drain(t *testing.T, g gograph.Graph[string], it Iterator[string]) []string {
	t.Helper()

	var labels []string
	for it.HasNext() {
		v := it.Next()
		if v == nil {
			t.Fatal("Expect a vertex after HasNext returned true, but got nil")
		}
		if g.GetVertexByID(v.Label()) != v {
			t.Fatalf("Expect only vertices in the graph, but got the removed vertex %s", v.Label())
		}
		labels = append(labels, v.Label())
	}

	if v := it.Next(); v != nil {
		t.Fatalf("Expect nil after HasNext returned false, but got %s", v.Label())
	}

	return labels
}

func startIterator(t *testing.T, g gograph.Graph[string], iter func(gograph.Graph[string], string) (Iterator[string], error)) Iterator[string] {
	t.Helper()

	it, err := iter(g, "A")
	if err != nil {
		t.Fatal(err)
	}
	if v := it.Next(); v == nil || v.Label() != "A" {
		t.Fatalf("Expect A first, but got %v", v)
	}

	return it
}

func TestIterators_SkipRemovedQueuedVertex(t *testing.T) {
	for _, tt := range removalIterators {
		t.Run(tt.name, func(t *testing.T) {
			// A -> B (1), A -> C (2), C -> D (1)
			g := gograph.New[string](gograph.Directed())
			a := g.AddVertexByLabel("A")
			b := g.AddVertexByLabel("B")
			c := g.AddVertexByLabel("C")
			d := g.AddVertexByLabel("D")
			_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1))
			_, _ = g.AddEdge(a, c, gograph.WithEdgeWeight(2))
			_, _ = g.AddEdge(c, d, gograph.WithEdgeWeight(1))

			it := startIterator(t, g, tt.iter)
			g.RemoveVertices(b)

			if got, want := drain(t, g, it), []string{"C", "D"}; !reflect.DeepEqual(got, want) {
				t.Errorf("Expect %v, but got %v", want, got)
			}
		})
	}
}

func TestIterators_HasNextFalseWhenOnlyRemovedVerticesAreLeft(t *testing.T) {
	for _, tt := range removalIterators {
		t.Run(tt.name, func(t *testing.T) {
			// A -> B -> C, and C is queued only once B is visited.
			g := gograph.New[string](gograph.Directed())
			a := g.AddVertexByLabel("A")
			b := g.AddVertexByLabel("B")
			c := g.AddVertexByLabel("C")
			_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1))
			_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(1))

			it := startIterator(t, g, tt.iter)
			g.RemoveVertices(b)

			if got := drain(t, g, it); len(got) != 0 {
				t.Errorf("Expect no more vertices, but got %v", got)
			}
		})
	}
}

func TestIterators_NextSkipsVertexRemovedAfterHasNext(t *testing.T) {
	for _, tt := range removalIterators {
		t.Run(tt.name, func(t *testing.T) {
			g := gograph.New[string](gograph.Directed())
			a := g.AddVertexByLabel("A")
			b := g.AddVertexByLabel("B")
			c := g.AddVertexByLabel("C")
			_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1))
			_, _ = g.AddEdge(a, c, gograph.WithEdgeWeight(2))

			it := startIterator(t, g, tt.iter)
			if !it.HasNext() {
				t.Fatal("Expect HasNext to be true while B and C are queued")
			}
			g.RemoveVertices(b, c)

			if v := it.Next(); v != nil {
				t.Errorf("Expect nil, but got the removed vertex %s", v.Label())
			}
		})
	}
}

func TestIterators_ResetAfterStartRemoved(t *testing.T) {
	for _, tt := range removalIterators {
		t.Run(tt.name, func(t *testing.T) {
			g := gograph.New[string](gograph.Directed())
			a := g.AddVertexByLabel("A")
			b := g.AddVertexByLabel("B")
			_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1))

			it := startIterator(t, g, tt.iter)
			drain(t, g, it)

			g.RemoveVertices(a)
			it.Reset()

			if got := drain(t, g, it); len(got) != 0 {
				t.Errorf("Expect no vertices once the start vertex is removed, but got %v", got)
			}
		})
	}
}

func TestIterators_ReturnReaddedVertex(t *testing.T) {
	for _, tt := range removalIterators {
		t.Run(tt.name, func(t *testing.T) {
			// A -> B (10) and A -> C (1). B is replaced, and only C reaches
			// the new B.
			g := gograph.New[string](gograph.Directed())
			a := g.AddVertexByLabel("A")
			b := g.AddVertexByLabel("B")
			c := g.AddVertexByLabel("C")
			_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(10))
			_, _ = g.AddEdge(a, c, gograph.WithEdgeWeight(1))

			it := startIterator(t, g, tt.iter)
			g.RemoveVertices(b)
			b2 := g.AddVertexByLabel("B")
			_, _ = g.AddEdge(c, b2, gograph.WithEdgeWeight(2))

			got := drain(t, g, it)
			sort.Strings(got)
			if want := []string{"B", "C"}; !reflect.DeepEqual(got, want) {
				t.Errorf("Expect %v, but got %v", want, got)
			}
		})
	}
}

func TestIterators_ReturnVertexReaddedAfterItsVisit(t *testing.T) {
	for _, tt := range removalIterators {
		t.Run(tt.name, func(t *testing.T) {
			// A -> B -> C. B is visited, then replaced by a new B that only
			// C reaches.
			g := gograph.New[string](gograph.Directed())
			a := g.AddVertexByLabel("A")
			b := g.AddVertexByLabel("B")
			c := g.AddVertexByLabel("C")
			_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1))
			_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(1))

			it := startIterator(t, g, tt.iter)
			if v := it.Next(); v != b {
				t.Fatalf("Expect B second, but got %v", v)
			}
			g.RemoveVertices(b)
			b2 := g.AddVertexByLabel("B")
			_, _ = g.AddEdge(c, b2, gograph.WithEdgeWeight(1))

			if got, want := drain(t, g, it), []string{"C", "B"}; !reflect.DeepEqual(got, want) {
				t.Errorf("Expect %v, with the new B, but got %v", want, got)
			}
		})
	}
}

func TestIterators_SkipReaddedVertexThatIsNotReached(t *testing.T) {
	for _, tt := range removalIterators {
		t.Run(tt.name, func(t *testing.T) {
			// A -> B, and B is replaced by a vertex with no edges.
			g := gograph.New[string](gograph.Directed())
			a := g.AddVertexByLabel("A")
			b := g.AddVertexByLabel("B")
			_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1))

			it := startIterator(t, g, tt.iter)
			g.RemoveVertices(b)
			g.AddVertexByLabel("B")

			if got := drain(t, g, it); len(got) != 0 {
				t.Errorf("Expect no vertices, since nothing reaches the new B, but got %v", got)
			}
		})
	}
}

func TestIterators_ReturnQueuedVertexWithoutEdges(t *testing.T) {
	for _, tt := range removalIterators {
		t.Run(tt.name, func(t *testing.T) {
			// A -> B, and the edge is removed while B is queued.
			g := gograph.New[string](gograph.Directed())
			a := g.AddVertexByLabel("A")
			b := g.AddVertexByLabel("B")
			edge, _ := g.AddEdge(a, b, gograph.WithEdgeWeight(1))

			it := startIterator(t, g, tt.iter)
			g.RemoveEdges(edge)

			if got, want := drain(t, g, it), []string{"B"}; !reflect.DeepEqual(got, want) {
				t.Errorf("Expect %v, since B is still in the graph, but got %v", want, got)
			}
		})
	}
}

func TestIterators_IterateSkipsVertexRemovedByCallback(t *testing.T) {
	for _, tt := range removalIterators {
		t.Run(tt.name, func(t *testing.T) {
			g := gograph.New[string](gograph.Directed())
			a := g.AddVertexByLabel("A")
			b := g.AddVertexByLabel("B")
			c := g.AddVertexByLabel("C")
			d := g.AddVertexByLabel("D")
			_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1))
			_, _ = g.AddEdge(a, c, gograph.WithEdgeWeight(2))
			_, _ = g.AddEdge(c, d, gograph.WithEdgeWeight(1))

			it, err := tt.iter(g, "A")
			if err != nil {
				t.Fatal(err)
			}

			var got []string
			err = it.Iterate(func(v *gograph.Vertex[string]) error {
				if v == nil {
					t.Fatal("Expect a vertex, but got nil")
				}
				got = append(got, v.Label())
				if v == a {
					g.RemoveVertices(b)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}

			if want := []string{"A", "C", "D"}; !reflect.DeepEqual(got, want) {
				t.Errorf("Expect %v, but got %v", want, got)
			}
		})
	}
}

func TestClosestFirstIterator_ReaddedVertexFartherAway(t *testing.T) {
	// A -> B (10), A -> C (1), A -> D (15). The new B is reached through C at
	// 21, so it comes after D even though the old B was queued at 10.
	g := gograph.New[string](gograph.Directed())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	d := g.AddVertexByLabel("D")
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(10))
	_, _ = g.AddEdge(a, c, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(a, d, gograph.WithEdgeWeight(15))

	it := startIterator(t, g, NewClosestFirstIterator[string])
	g.RemoveVertices(b)
	b2 := g.AddVertexByLabel("B")
	_, _ = g.AddEdge(c, b2, gograph.WithEdgeWeight(20))

	if got, want := drain(t, g, it), []string{"C", "D", "B"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Expect %v, but got %v", want, got)
	}
}
