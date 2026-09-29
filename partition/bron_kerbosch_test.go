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

func cliqueLabels(cliques [][]*gograph.Vertex[string]) [][]string {
	out := make([][]string, 0, len(cliques))
	for _, c := range cliques {
		labels := make([]string, 0, len(c))
		for _, v := range c {
			labels = append(labels, v.Label())
		}
		sort.Strings(labels)
		out = append(out, labels)
	}

	sort.Slice(
		out, func(i, j int) bool {
			return fmt.Sprint(out[i]) < fmt.Sprint(out[j])
		},
	)

	return out
}

func TestMaximalCliques_SelfLoop(t *testing.T) {
	g := gograph.New[string]()

	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")

	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(c, a)
	_, _ = g.AddEdge(a, a)

	want := [][]string{{"A", "B", "C"}}
	if got := cliqueLabels(MaximalCliques(g)); !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestMaximalCliques_SingleVertexWithSelfLoop(t *testing.T) {
	g := gograph.New[string]()

	x := g.AddVertexByLabel("X")
	_, _ = g.AddEdge(x, x)

	want := [][]string{{"X"}}
	if got := cliqueLabels(MaximalCliques(g)); !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestMaximalCliques_Directed(t *testing.T) {
	tests := []struct {
		name  string
		edges [][2]string
		want  [][]string
	}{
		{
			name:  "cycle",
			edges: [][2]string{{"A", "B"}, {"B", "C"}, {"C", "A"}},
			want:  [][]string{{"A", "B", "C"}},
		},
		{
			name:  "one direction per pair",
			edges: [][2]string{{"A", "B"}, {"B", "C"}, {"A", "C"}},
			want:  [][]string{{"A", "B", "C"}},
		},
		{
			name:  "both directions",
			edges: [][2]string{{"A", "B"}, {"B", "A"}, {"B", "C"}},
			want:  [][]string{{"A", "B"}, {"B", "C"}},
		},
		{
			name:  "self-loop",
			edges: [][2]string{{"A", "B"}, {"B", "B"}},
			want:  [][]string{{"A", "B"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The result used to depend on map order, so check several runs.
			for run := 0; run < 20; run++ {
				g := gograph.New[string](gograph.Directed())
				for _, e := range tt.edges {
					from := g.GetVertexByID(e[0])
					if from == nil {
						from = g.AddVertexByLabel(e[0])
					}
					to := g.GetVertexByID(e[1])
					if to == nil {
						to = g.AddVertexByLabel(e[1])
					}
					_, _ = g.AddEdge(from, to)
				}

				if got := cliqueLabels(MaximalCliques(g)); !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("run %d: expected %v, got %v", run, tt.want, got)
				}
			}
		})
	}
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
