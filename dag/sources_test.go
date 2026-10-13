package dag

import (
	"errors"
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestSourcesAndSinks(t *testing.T) {
	t.Run("empty graph", func(t *testing.T) {
		g := gograph.New[string](gograph.Directed())

		sources, err := Sources(g)
		if err != nil {
			t.Fatalf("Sources(empty): unexpected error: %v", err)
		}
		if sources == nil || len(sources) != 0 {
			t.Fatalf("Sources(empty) = %v, want empty slice", labels(sources))
		}

		sinks, err := Sinks(g)
		if err != nil {
			t.Fatalf("Sinks(empty): unexpected error: %v", err)
		}
		if sinks == nil || len(sinks) != 0 {
			t.Fatalf("Sinks(empty) = %v, want empty slice", labels(sinks))
		}
	})

	t.Run("chain", func(t *testing.T) {
		// A -> B -> C
		g := newGraph(t, [2]string{"A", "B"}, [2]string{"B", "C"})

		sources, err := Sources(g)
		if err != nil {
			t.Fatalf("Sources: unexpected error: %v", err)
		}
		assertLabels(t, "Sources(chain)", labels(sources), []string{"A"})

		sinks, err := Sinks(g)
		if err != nil {
			t.Fatalf("Sinks: unexpected error: %v", err)
		}
		assertLabels(t, "Sinks(chain)", labels(sinks), []string{"C"})
	})

	t.Run("diamond", func(t *testing.T) {
		// A -> B -> D, A -> C -> D
		g := newGraph(t,
			[2]string{"A", "B"},
			[2]string{"A", "C"},
			[2]string{"B", "D"},
			[2]string{"C", "D"},
		)

		sources, err := Sources(g)
		if err != nil {
			t.Fatalf("Sources: unexpected error: %v", err)
		}
		assertLabels(t, "Sources(diamond)", labels(sources), []string{"A"})

		sinks, err := Sinks(g)
		if err != nil {
			t.Fatalf("Sinks: unexpected error: %v", err)
		}
		assertLabels(t, "Sinks(diamond)", labels(sinks), []string{"D"})
	})

	t.Run("multiple sources and sinks", func(t *testing.T) {
		// S1 -> M -> K1, S2 -> M -> K2
		g := newGraph(t,
			[2]string{"S1", "M"},
			[2]string{"S2", "M"},
			[2]string{"M", "K1"},
			[2]string{"M", "K2"},
		)

		sources, err := Sources(g)
		if err != nil {
			t.Fatalf("Sources: unexpected error: %v", err)
		}
		assertLabels(t, "Sources", labels(sources), []string{"S1", "S2"})

		sinks, err := Sinks(g)
		if err != nil {
			t.Fatalf("Sinks: unexpected error: %v", err)
		}
		assertLabels(t, "Sinks", labels(sinks), []string{"K1", "K2"})
	})

	t.Run("graph with pure cycle", func(t *testing.T) {
		// A -> B -> C -> A
		g := newGraph(t,
			[2]string{"A", "B"},
			[2]string{"B", "C"},
			[2]string{"C", "A"},
		)

		sources, err := Sources(g)
		if err != nil {
			t.Fatalf("Sources(cycle): unexpected error: %v", err)
		}
		if len(sources) != 0 {
			t.Fatalf("Sources(cycle) = %v, want empty slice", labels(sources))
		}

		sinks, err := Sinks(g)
		if err != nil {
			t.Fatalf("Sinks(cycle): unexpected error: %v", err)
		}
		if len(sinks) != 0 {
			t.Fatalf("Sinks(cycle) = %v, want empty slice", labels(sinks))
		}
	})

	t.Run("cycle with source and sink", func(t *testing.T) {
		// S -> A -> B -> A, B -> K
		g := newGraph(t,
			[2]string{"S", "A"},
			[2]string{"A", "B"},
			[2]string{"B", "A"},
			[2]string{"B", "K"},
		)

		sources, err := Sources(g)
		if err != nil {
			t.Fatalf("Sources: unexpected error: %v", err)
		}
		assertLabels(t, "Sources(cycle with branches)", labels(sources), []string{"S"})

		sinks, err := Sinks(g)
		if err != nil {
			t.Fatalf("Sinks: unexpected error: %v", err)
		}
		assertLabels(t, "Sinks(cycle with branches)", labels(sinks), []string{"K"})
	})

	t.Run("isolated vertex", func(t *testing.T) {
		g := gograph.New[string](gograph.Directed())
		vA := g.AddVertexByLabel("A")

		sources, err := Sources(g)
		if err != nil {
			t.Fatalf("Sources: unexpected error: %v", err)
		}
		assertLabels(t, "Sources(isolated)", labels(sources), []string{"A"})

		sinks, err := Sinks(g)
		if err != nil {
			t.Fatalf("Sinks: unexpected error: %v", err)
		}
		assertLabels(t, "Sinks(isolated)", labels(sinks), []string{"A"})

		// Pointers check
		if sources[0] != vA {
			t.Fatalf("Sources returned pointer %p, want %p", sources[0], vA)
		}
		if sinks[0] != vA {
			t.Fatalf("Sinks returned pointer %p, want %p", sinks[0], vA)
		}
	})

	t.Run("isolated vertex with chain", func(t *testing.T) {
		// A -> B, and isolated C
		g := newGraph(t, [2]string{"A", "B"})
		g.AddVertexByLabel("C")

		sources, err := Sources(g)
		if err != nil {
			t.Fatalf("Sources: unexpected error: %v", err)
		}
		assertLabels(t, "Sources", labels(sources), []string{"A", "C"})

		sinks, err := Sinks(g)
		if err != nil {
			t.Fatalf("Sinks: unexpected error: %v", err)
		}
		assertLabels(t, "Sinks", labels(sinks), []string{"B", "C"})
	})

	t.Run("self loop only", func(t *testing.T) {
		g := gograph.New[string](gograph.Directed())
		vA := g.AddVertexByLabel("A")
		if _, err := g.AddEdge(vA, vA); err != nil {
			t.Fatalf("AddEdge self-loop: %v", err)
		}

		sources, err := Sources(g)
		if err != nil {
			t.Fatalf("Sources(self-loop): unexpected error: %v", err)
		}
		if len(sources) != 0 {
			t.Fatalf("Sources(self-loop) = %v, want empty slice", labels(sources))
		}

		sinks, err := Sinks(g)
		if err != nil {
			t.Fatalf("Sinks(self-loop): unexpected error: %v", err)
		}
		if len(sinks) != 0 {
			t.Fatalf("Sinks(self-loop) = %v, want empty slice", labels(sinks))
		}
	})

	t.Run("self loop in graph", func(t *testing.T) {
		// A -> B (with B -> B self-loop) -> C
		g := newGraph(t, [2]string{"A", "B"}, [2]string{"B", "C"})
		vB := g.GetVertexByID("B")
		if _, err := g.AddEdge(vB, vB); err != nil {
			t.Fatalf("AddEdge self-loop: %v", err)
		}

		sources, err := Sources(g)
		if err != nil {
			t.Fatalf("Sources: unexpected error: %v", err)
		}
		assertLabels(t, "Sources", labels(sources), []string{"A"})

		sinks, err := Sinks(g)
		if err != nil {
			t.Fatalf("Sinks: unexpected error: %v", err)
		}
		assertLabels(t, "Sinks", labels(sinks), []string{"C"})
	})

	t.Run("undirected graph returns ErrNotDirected", func(t *testing.T) {
		g := gograph.New[string]() // undirected by default
		vA := g.AddVertexByLabel("A")
		vB := g.AddVertexByLabel("B")
		_, _ = g.AddEdge(vA, vB)

		if _, err := Sources(g); !errors.Is(err, gograph.ErrNotDirected) {
			t.Fatalf("Sources(undirected): err = %v, want ErrNotDirected", err)
		}
		if _, err := Sinks(g); !errors.Is(err, gograph.ErrNotDirected) {
			t.Fatalf("Sinks(undirected): err = %v, want ErrNotDirected", err)
		}
	})

	t.Run("order follows GetAllVertices", func(t *testing.T) {
		// Add vertices in order S3, S1, S2, K2, K1
		g := gograph.New[string](gograph.Directed())
		s3 := g.AddVertexByLabel("S3")
		s1 := g.AddVertexByLabel("S1")
		s2 := g.AddVertexByLabel("S2")
		k2 := g.AddVertexByLabel("K2")
		k1 := g.AddVertexByLabel("K1")

		_, _ = g.AddEdge(s1, k1)
		_, _ = g.AddEdge(s2, k1)
		_, _ = g.AddEdge(s3, k2)

		sources, err := Sources(g)
		if err != nil {
			t.Fatalf("Sources: unexpected error: %v", err)
		}
		assertLabels(t, "Sources order", labels(sources), []string{"S3", "S1", "S2"})

		sinks, err := Sinks(g)
		if err != nil {
			t.Fatalf("Sinks: unexpected error: %v", err)
		}
		assertLabels(t, "Sinks order", labels(sinks), []string{"K2", "K1"})
	})

	t.Run("returned vertices are graph's own pointers", func(t *testing.T) {
		g := newGraph(t, [2]string{"A", "B"})
		vA := g.GetVertexByID("A")
		vB := g.GetVertexByID("B")

		sources, err := Sources(g)
		if err != nil || len(sources) != 1 {
			t.Fatalf("Sources = %v, %v; want [A]", labels(sources), err)
		}
		if sources[0] != vA {
			t.Fatalf("Sources returned %p, want %p", sources[0], vA)
		}

		sinks, err := Sinks(g)
		if err != nil || len(sinks) != 1 {
			t.Fatalf("Sinks = %v, %v; want [B]", labels(sinks), err)
		}
		if sinks[0] != vB {
			t.Fatalf("Sinks returned %p, want %p", sinks[0], vB)
		}
	})
}

func BenchmarkSources(b *testing.B) {
	g := gograph.New[int](gograph.Directed())
	for i := 0; i < 1000; i++ {
		g.AddVertexByLabel(i)
	}
	for i := 0; i < 999; i++ {
		v1 := g.GetVertexByID(i)
		v2 := g.GetVertexByID(i + 1)
		_, _ = g.AddEdge(v1, v2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Sources(g)
	}
}

func BenchmarkSinks(b *testing.B) {
	g := gograph.New[int](gograph.Directed())
	for i := 0; i < 1000; i++ {
		g.AddVertexByLabel(i)
	}
	for i := 0; i < 999; i++ {
		v1 := g.GetVertexByID(i)
		v2 := g.GetVertexByID(i + 1)
		_, _ = g.AddEdge(v1, v2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Sinks(g)
	}
}
