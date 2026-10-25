package gograph

import (
	"slices"
	"testing"
)

func seqTestGraphs() map[string]Graph[int] {
	directed, _ := benchGraph(60, benchEdges(60, 240, false), Directed())
	undirected, vertices := benchGraph(60, benchEdges(60, 120, false))
	_, _ = undirected.AddEdge(vertices[5], vertices[5])

	// removals leave empty slots in the insertion order
	for _, g := range []Graph[int]{directed, undirected} {
		g.RemoveVertices(g.GetVertexByID(3), g.GetVertexByID(40))
		g.AddVertexByLabel(3)
	}

	return map[string]Graph[int]{
		"directed":            directed,
		"undirected":          undirected,
		"directed methods":    methodsOnly[int]{directed},
		"undirected methods":  methodsOnly[int]{undirected},
		"empty":               New[int](),
		"empty without edges": New[int](Directed()),
	}
}

func TestVertices_MatchesGetAllVertices(t *testing.T) {
	for name, g := range seqTestGraphs() {
		if got, want := slices.Collect(Vertices(g)), g.GetAllVertices(); !slices.Equal(got, want) {
			t.Fatalf("%s: Vertices = %d vertices, want %d", name, len(got), len(want))
		}
	}
}

func TestEdges_MatchesAllEdges(t *testing.T) {
	for name, g := range seqTestGraphs() {
		if got, want := slices.Collect(Edges(g)), g.AllEdges(); !slices.Equal(got, want) {
			t.Fatalf("%s: Edges = %d edges, want %d", name, len(got), len(want))
		}
	}
}

func TestVerticesAndEdges_StopEarly(t *testing.T) {
	for name, g := range seqTestGraphs() {
		if g.Order() < 3 {
			continue
		}
		n := 0
		for range Vertices(g) {
			if n++; n == 3 {
				break
			}
		}
		m := 0
		for range Edges(g) {
			if m++; m == 3 {
				break
			}
		}
		if n != 3 || m != 3 {
			t.Fatalf("%s: stopped after %d vertices and %d edges, want 3 and 3", name, n, m)
		}
	}
}

func BenchmarkGetAllVertices(b *testing.B) {
	g, _ := benchGraph(10000, benchEdges(10000, 50000, false), Directed())
	b.ReportAllocs()
	for range b.N {
		n := 0
		for range g.GetAllVertices() {
			n++
		}
		if n == 0 {
			b.Fatal("no vertices")
		}
	}
}

func BenchmarkVertices(b *testing.B) {
	g, _ := benchGraph(10000, benchEdges(10000, 50000, false), Directed())
	b.ReportAllocs()
	for range b.N {
		n := 0
		for range Vertices(g) {
			n++
		}
		if n == 0 {
			b.Fatal("no vertices")
		}
	}
}

func BenchmarkAllEdges(b *testing.B) {
	g, _ := benchGraph(10000, benchEdges(10000, 50000, false), Directed())
	b.ReportAllocs()
	for range b.N {
		n := 0
		for range g.AllEdges() {
			n++
		}
		if n == 0 {
			b.Fatal("no edges")
		}
	}
}

func BenchmarkEdges(b *testing.B) {
	g, _ := benchGraph(10000, benchEdges(10000, 50000, false), Directed())
	b.ReportAllocs()
	for range b.N {
		n := 0
		for range Edges(g) {
			n++
		}
		if n == 0 {
			b.Fatal("no edges")
		}
	}
}
