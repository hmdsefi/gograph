package partition

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hmdsefi/gograph"
)

func normalizeVertexCliques[T comparable](cliques [][]*gograph.Vertex[T]) [][]*gograph.Vertex[T] {
	for _, c := range cliques {
		sort.Slice(
			c, func(i, j int) bool {
				return fmt.Sprintf("%v", c[i].Label()) < fmt.Sprintf("%v", c[j].Label())
			},
		)
	}

	sort.Slice(
		cliques, func(i, j int) bool {
			a, b := cliques[i], cliques[j]
			for k := 0; k < len(a) && k < len(b); k++ {
				if a[k].Label() != b[k].Label() {
					return fmt.Sprintf("%v", a[k].Label()) < fmt.Sprintf("%v", b[k].Label())
				}
			}
			return len(a) < len(b)
		},
	)

	return cliques
}

func TestMaximalCliques_Triangle(t *testing.T) {
	g := gograph.New[string]()

	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")

	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(c, a)

	cliques := MaximalCliques(g)

	cliques = normalizeVertexCliques(cliques)

	want := [][]*gograph.Vertex[string]{{a, b, c}}
	want = normalizeVertexCliques(want)

	if !reflect.DeepEqual(cliques, want) {
		t.Fatalf("expected %v, got %v", want, cliques)
	}
}

func TestMaximalCliques_SquareWithDiagonal(t *testing.T) {
	g := gograph.New[string]()

	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	d := g.AddVertexByLabel("D")

	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(c, d)
	_, _ = g.AddEdge(d, a)
	_, _ = g.AddEdge(a, c) // diagonal

	cliques := MaximalCliques(g)
	cliques = normalizeVertexCliques(cliques)

	want := [][]*gograph.Vertex[string]{
		{a, b, c},
		{a, c, d},
	}
	want = normalizeVertexCliques(want)

	if !reflect.DeepEqual(cliques, want) {
		t.Fatalf("expected %v, got %v", want, cliques)
	}
}
