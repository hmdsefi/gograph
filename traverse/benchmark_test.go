package traverse

import (
	"iter"
	"math/rand"
	"testing"

	"github.com/hmdsefi/gograph"
)

// benchGraph returns a weighted directed graph with n vertices and m random
// edges.
func benchGraph(n int, m uint32) gograph.Graph[int] {
	rng := rand.New(rand.NewSource(1)) //nolint:gosec // seeded so runs are comparable
	g := gograph.New[int](gograph.Directed(), gograph.Weighted())
	vertices := make([]*gograph.Vertex[int], n)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for g.Size() < m {
		from, to := vertices[rng.Intn(n)], vertices[rng.Intn(n)]
		_, _ = g.AddEdge(from, to, gograph.WithEdgeWeight(1+rng.Float64()*9))
	}

	return g
}

func benchmarkIterator(b *testing.B, newIterator func(gograph.Graph[int], int) (Iterator[int], error)) {
	g := benchGraph(10000, 50000)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		it, err := newIterator(g, 0)
		if err != nil {
			b.Fatal(err)
		}
		for it.HasNext() {
			it.Next()
		}
	}
}

func BenchmarkBreadthFirstIterator(b *testing.B) {
	benchmarkIterator(b, NewBreadthFirstIterator[int])
}

func BenchmarkDepthFirstIterator(b *testing.B) {
	benchmarkIterator(b, NewDepthFirstIterator[int])
}

func BenchmarkClosestFirstIterator(b *testing.B) {
	benchmarkIterator(b, NewClosestFirstIterator[int])
}

func benchmarkSeq(b *testing.B, seq func(gograph.Graph[int], int) iter.Seq[*gograph.Vertex[int]]) {
	g := benchGraph(10000, 50000)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		n := 0
		for range seq(g, 0) {
			n++
		}
		if n == 0 {
			b.Fatal("no vertices")
		}
	}
}

func BenchmarkBFS(b *testing.B) {
	benchmarkSeq(b, BFS[int])
}

func BenchmarkDFS(b *testing.B) {
	benchmarkSeq(b, DFS[int])
}

func BenchmarkDFSEvents(b *testing.B) {
	g := benchGraph(10000, 50000)
	b.ReportAllocs()
	for b.Loop() {
		n := 0
		for range DFSEvents(g, 0) {
			n++
		}
		if n == 0 {
			b.Fatal("no events")
		}
	}
}

func BenchmarkClosestFirst(b *testing.B) {
	benchmarkSeq(b, func(g gograph.Graph[int], start int) iter.Seq[*gograph.Vertex[int]] {
		return func(yield func(*gograph.Vertex[int]) bool) {
			for v := range ClosestFirst(g, start) {
				if !yield(v) {
					return
				}
			}
		}
	})
}
