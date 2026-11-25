package partition

import (
	"testing"
	"time"

	"github.com/hmdsefi/gograph"
)

// pathGraph returns the path 0 - 1 - ... - n-1.
func pathGraph(n int) gograph.Graph[int] {
	g := gograph.New[int]()
	prev := g.AddVertexByLabel(0)
	for i := 1; i < n; i++ {
		v := g.AddVertexByLabel(i)
		_, _ = g.AddEdge(prev, v)
		prev = v
	}
	return g
}

// girvanNewmanWithin fails the test when GirvanNewman doesn't return within
// limit, so a search that never ends shows up as a failure.
func girvanNewmanWithin(t *testing.T, g gograph.Graph[int], k int, limit time.Duration) ([]gograph.Graph[int], error) {
	t.Helper()
	type result struct {
		communities []gograph.Graph[int]
		err         error
	}
	done := make(chan result, 1)
	go func() {
		communities, err := GirvanNewman(g, k)
		done <- result{communities, err}
	}()
	select {
	case r := <-done:
		return r.communities, r.err
	case <-time.After(limit):
		t.Fatalf("GirvanNewman(k=%d) didn't return within %v", k, limit)
		return nil, nil
	}
}

func TestGirvanNewman_LongPath(t *testing.T) {
	// a search from either end goes deeper than 1,024 levels
	const n = 1100
	communities, err := girvanNewmanWithin(t, pathGraph(n), 2, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(communities) != 2 {
		t.Fatalf("expected 2 communities, got %d", len(communities))
	}

	// the middle edge has the highest betweenness
	for i, c := range communities {
		vertices := c.GetAllVertices()
		if len(vertices) != n/2 {
			t.Fatalf("community %d: expected %d vertices, got %d", i, n/2, len(vertices))
		}
		if first := vertices[0].Label(); first != i*n/2 {
			t.Errorf("community %d: expected it to start at %d, got %d", i, i*n/2, first)
		}
	}
}

func TestEdgeBetweenness_LongPath(t *testing.T) {
	// the search from vertex 0 reaches vertex 1,024 at depth 1,024
	const n = 1025
	g := pathGraph(n)
	betweenness, err := calculateEdgeBetweenness(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The edge from i to i+1 is on the shortest paths between the i+1
	// vertices on one side and the n-1-i on the other, so its betweenness
	// is proportional to (i+1)(n-1-i). n-1 is a power of two, so the
	// expected values are exact.
	edge := func(i int) *gograph.Edge[int] {
		return g.GetEdge(g.GetVertexByID(i), g.GetVertexByID(i+1))
	}
	// The first edge is on the shortest paths from vertex 0 to the 1,024
	// others, and each stored direction gets half of them.
	first := betweenness[edge(0)]
	if first != 512 {
		t.Fatalf("edge 0-1: expected betweenness 512, got %v", first)
	}
	for i := 0; i < n-1; i++ {
		want := first * float64((i+1)*(n-1-i)) / float64(n-1)
		if got := betweenness[edge(i)]; got != want {
			t.Fatalf("edge %d-%d: expected betweenness %v, got %v", i, i+1, want, got)
		}
	}
}

func TestGirvanNewman_TooManyShortestPaths(t *testing.T) {
	// Each diamond doubles the number of shortest paths, so after 1,024 of
	// them there are 2^1024 shortest paths from 0 to the last vertex, more
	// than a float64 can count.
	const diamonds = 1024
	g := gograph.New[int]()
	x := g.AddVertexByLabel(0)
	for i := 0; i < diamonds; i++ {
		a := g.AddVertexByLabel(3*i + 1)
		b := g.AddVertexByLabel(3*i + 2)
		next := g.AddVertexByLabel(3*i + 3)
		_, _ = g.AddEdge(x, a)
		_, _ = g.AddEdge(x, b)
		_, _ = g.AddEdge(a, next)
		_, _ = g.AddEdge(b, next)
		x = next
	}

	if _, err := girvanNewmanWithin(t, g, 2, time.Minute); err == nil {
		t.Fatal("expected an error when the shortest paths can't be counted")
	}
}
