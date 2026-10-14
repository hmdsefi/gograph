package path

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestKCenterPathReachesFactorTwo(t *testing.T) {
	g := unitPath("a", "b", "c", "d", "e")

	centers, radius, err := KCenter(g, 2)
	if err != nil {
		t.Fatalf("KCenter: %v", err)
	}
	if got := labelsOf(centers); !slices.Equal(got, []string{"a", "e"}) {
		t.Fatalf("centers %v, want a e", got)
	}
	if radius != 2 {
		t.Fatalf("radius %v, want 2", radius)
	}
	for _, c := range centers {
		if c != g.GetVertexByID(c.Label()) {
			t.Fatalf("center %v is not the graph's own vertex", c.Label())
		}
	}

	// b and d cover the path with radius 1, so the guarantee is tight.
	opt := bestRadius(g, 2)
	if opt != 1 {
		t.Fatalf("best radius %v, want 1", opt)
	}
	if radius > 2*opt {
		t.Fatalf("radius %v is more than twice the best %v", radius, opt)
	}
}

func TestKCenterRadiusMatchesMultiSource(t *testing.T) {
	g := unitPath("a", "b", "c", "d", "e")
	centers, radius, err := KCenter(g, 3)
	if err != nil {
		t.Fatalf("KCenter: %v", err)
	}
	if got := multiSourceRadius(g, centers); got != radius {
		t.Fatalf("radius %v, multi-source radius %v", radius, got)
	}
}

func TestKCenterApproximationOnRandomGraphs(t *testing.T) {
	for seed := uint64(1); seed <= 8; seed++ {
		g := randomUndirected(6, 0.5, seed)
		for _, k := range []int{1, 2, 3} {
			centers, radius, err := KCenter(g, k)
			if err != nil {
				t.Fatalf("seed %d k %d: %v", seed, k, err)
			}
			if got := multiSourceRadius(g, centers); got != radius {
				t.Fatalf("seed %d k %d: radius %v, multi-source %v", seed, k, radius, got)
			}
			opt := bestRadius(g, k)
			if radius < opt {
				t.Fatalf("seed %d k %d: radius %v is below the best %v", seed, k, radius, opt)
			}
			if !math.IsInf(opt, 1) && radius > 2*opt {
				t.Fatalf("seed %d k %d: radius %v is more than twice the best %v", seed, k, radius, opt)
			}
		}
	}
}

func TestKCenterFixedCenters(t *testing.T) {
	g := unitPath("a", "b", "c", "d", "e")

	centers, radius, err := KCenter(g, 2, "b", "d")
	if err != nil {
		t.Fatalf("KCenter: %v", err)
	}
	if got := labelsOf(centers); !slices.Equal(got, []string{"b", "d"}) {
		t.Fatalf("centers %v, want b d", got)
	}
	if radius != 1 {
		t.Fatalf("radius %v, want 1", radius)
	}

	centers, _, err = KCenter(g, 2, "a", "a")
	if err != nil {
		t.Fatalf("duplicate fixed: %v", err)
	}
	if got := labelsOf(centers); len(got) != 2 || got[0] != "a" {
		t.Fatalf("duplicate fixed centers %v, want a then one more", got)
	}

	centers, radius, err = KCenter(g, 1, "c")
	if err != nil {
		t.Fatalf("fixed fills k: %v", err)
	}
	if got := labelsOf(centers); !slices.Equal(got, []string{"c"}) {
		t.Fatalf("centers %v, want c", got)
	}
	if radius != 2 {
		t.Fatalf("radius %v, want 2", radius)
	}
}

func TestKCenterComponents(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a := g.AddVertexByLabel("a")
	b := g.AddVertexByLabel("b")
	c := g.AddVertexByLabel("c")
	d := g.AddVertexByLabel("d")
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(c, d, gograph.WithEdgeWeight(1))

	centers, radius, err := KCenter(g, 2)
	if err != nil {
		t.Fatalf("k=2: %v", err)
	}
	if math.IsInf(radius, 1) {
		t.Fatalf("k=2 radius is infinite, centers %v", labelsOf(centers))
	}
	if len(centers) != 2 {
		t.Fatalf("k=2 centers %v", labelsOf(centers))
	}

	_, radius, err = KCenter(g, 1)
	if err != nil {
		t.Fatalf("k=1: %v", err)
	}
	if !math.IsInf(radius, 1) {
		t.Fatalf("k=1 radius %v, want +Inf", radius)
	}
}

func TestKCenterCoversEveryVertexWhenKIsLarge(t *testing.T) {
	g := unitPath("a", "b", "c")

	centers, radius, err := KCenter(g, 5)
	if err != nil {
		t.Fatalf("KCenter: %v", err)
	}
	if radius != 0 {
		t.Fatalf("radius %v, want 0", radius)
	}
	if got := labelsOf(centers); len(got) != 3 {
		t.Fatalf("centers %v, want all 3 vertices", got)
	}
}

func TestKCenterEmpty(t *testing.T) {
	centers, radius, err := KCenter(gograph.New[string](), 3)
	if err != nil {
		t.Fatalf("empty: %v", err)
	}
	if centers != nil || radius != 0 {
		t.Fatalf("centers %v radius %v, want nil and 0", centers, radius)
	}
}

func TestKCenterErrors(t *testing.T) {
	g := unitPath("a", "b")

	_, _, err := KCenter(g, 0)
	if !errors.Is(err, ErrInvalidK) {
		t.Fatalf("k=0: got %v", err)
	}
	_, _, err = KCenter(gograph.New[int](), 0)
	if !errors.Is(err, ErrInvalidK) {
		t.Fatalf("empty k=0: got %v", err)
	}

	_, _, err = KCenter(g, 1, "a", "b")
	if !errors.Is(err, ErrInvalidK) {
		t.Fatalf("too many fixed: got %v", err)
	}

	_, _, err = KCenter(g, 1, "missing")
	if !errors.Is(err, gograph.ErrVertexDoesNotExist) {
		t.Fatalf("missing fixed: got %v", err)
	}

	neg := gograph.New[string](gograph.Weighted())
	u := neg.AddVertexByLabel("u")
	v := neg.AddVertexByLabel("v")
	_, _ = neg.AddEdge(u, v, gograph.WithEdgeWeight(-1))
	_, _, err = KCenter(neg, 1)
	if !errors.Is(err, ErrNegativeWeight) {
		t.Fatalf("negative: got %v", err)
	}
}

func TestKCenterStable(t *testing.T) {
	g := randomUndirected(8, 0.45, 42)
	first, _, err := KCenter(g, 3)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, _, err := KCenter(g, 3)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if !slices.Equal(labelsOf(first), labelsOf(second)) {
		t.Fatalf("centers %v and %v differ", labelsOf(first), labelsOf(second))
	}
}

func TestKCenterDirected(t *testing.T) {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	a := g.AddVertexByLabel("a")
	b := g.AddVertexByLabel("b")
	c := g.AddVertexByLabel("c")
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(1))

	centers, radius, err := KCenter(g, 1)
	if err != nil {
		t.Fatalf("KCenter: %v", err)
	}
	if labelsOf(centers)[0] != "a" || radius != 2 {
		t.Fatalf("centers %v radius %v, want a and 2", labelsOf(centers), radius)
	}
	if got := multiSourceRadius(g, centers); got != radius {
		t.Fatalf("radius %v, multi-source %v", radius, got)
	}
}

func ExampleKCenter() {
	g := gograph.New[string](gograph.Weighted())
	depot := g.AddVertexByLabel("depot")
	market := g.AddVertexByLabel("market")
	school := g.AddVertexByLabel("school")
	clinic := g.AddVertexByLabel("clinic")
	station := g.AddVertexByLabel("station")
	road := func(from, to *gograph.Vertex[string]) {
		_, _ = g.AddEdge(from, to, gograph.WithEdgeWeight(1))
	}
	road(depot, market)
	road(market, school)
	road(school, clinic)
	road(clinic, station)

	centers, radius, err := KCenter(g, 2)
	if err != nil {
		panic(err)
	}
	fmt.Println("centers", centers[0].Label(), centers[1].Label())
	fmt.Println("radius", radius)
	// Output:
	// centers depot station
	// radius 2
}

func BenchmarkKCenter(b *testing.B) {
	g := gograph.New[int](gograph.Weighted())
	const n = 2000
	verts := make([]*gograph.Vertex[int], n)
	for i := range verts {
		verts[i] = g.AddVertexByLabel(i)
	}
	for i := 1; i < n; i++ {
		_, _ = g.AddEdge(verts[i-1], verts[i], gograph.WithEdgeWeight(1))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, _, err := KCenter(g, 32); err != nil {
			b.Fatal(err)
		}
	}
}

func unitPath(names ...string) gograph.Graph[string] {
	g := gograph.New[string](gograph.Weighted())
	var prev *gograph.Vertex[string]
	for _, name := range names {
		v := g.AddVertexByLabel(name)
		if prev != nil {
			_, _ = g.AddEdge(prev, v, gograph.WithEdgeWeight(1))
		}
		prev = v
	}
	return g
}

func randomUndirected(n int, p float64, seed uint64) gograph.Graph[int] {
	g := gograph.New[int](gograph.Weighted())
	verts := make([]*gograph.Vertex[int], n)
	for i := range verts {
		verts[i] = g.AddVertexByLabel(i)
	}
	rng := rand.New(rand.NewPCG(seed, seed+1)) //nolint:gosec // seeded so failures are reproducible
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if rng.Float64() < p {
				_, _ = g.AddEdge(verts[i], verts[j], gograph.WithEdgeWeight(float64(1+rng.IntN(5))))
			}
		}
	}
	return g
}

func labelsOf[T comparable](vertices []*gograph.Vertex[T]) []T {
	out := make([]T, len(vertices))
	for i, v := range vertices {
		out[i] = v.Label()
	}
	return out
}

// multiSourceRadius is the largest distance from the centers, computed with
// one search from all of them. It is the figure DijkstraMultiSource returns.
func multiSourceRadius[T comparable](g gograph.Graph[T], centers []*gograph.Vertex[T]) float64 {
	vertices := g.GetAllVertices()
	index := make(map[T]int, len(vertices))
	dist := make([]float64, len(vertices))
	for i, v := range vertices {
		index[v.Label()] = i
		dist[i] = math.Inf(1)
	}
	type hop struct {
		to int
		w  float64
	}
	adj := make([][]hop, len(vertices))
	for _, e := range g.AllEdges() {
		adj[index[e.Source().Label()]] = append(adj[index[e.Source().Label()]], hop{index[e.Destination().Label()], e.Weight()})
	}
	for _, c := range centers {
		dist[index[c.Label()]] = 0
	}
	// Linear selection, independent of the heap KCenter uses.
	seen := make([]bool, len(vertices))
	for range vertices {
		u := -1
		for i := range dist {
			if seen[i] {
				continue
			}
			if u < 0 || dist[i] < dist[u] {
				u = i
			}
		}
		seen[u] = true
		for _, e := range adj[u] {
			if nd := dist[u] + e.w; nd < dist[e.to] {
				dist[e.to] = nd
			}
		}
	}
	radius := 0.0
	for _, d := range dist {
		if d > radius {
			radius = d
		}
	}
	return radius
}

func bestRadius[T comparable](g gograph.Graph[T], k int) float64 {
	vertices := g.GetAllVertices()
	best := math.Inf(1)
	combos(len(vertices), k, func(idx []int) {
		chosen := make([]*gograph.Vertex[T], k)
		for i, id := range idx {
			chosen[i] = vertices[id]
		}
		if r := multiSourceRadius(g, chosen); r < best {
			best = r
		}
	})
	return best
}

func combos(n, k int, fn func([]int)) {
	idx := make([]int, k)
	var walk func(start, depth int)
	walk = func(start, depth int) {
		if depth == k {
			fn(idx)
			return
		}
		for i := start; i <= n-(k-depth); i++ {
			idx[depth] = i
			walk(i+1, depth+1)
		}
	}
	walk(0, 0)
}
