package traverse

import (
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
