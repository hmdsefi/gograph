package path

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/hmdsefi/gograph"
)

// A wrong distance from a single source fails this. The expected numbers
// come from Dijkstra, which is a separate search.
func TestDijkstraMultiSourceMatchesDijkstra(t *testing.T) {
	g := weightedPath()
	got, err := DijkstraMultiSource(g, "A")
	if err != nil {
		t.Fatalf("DijkstraMultiSource: %v", err)
	}

	want := Dijkstra(g, "A")
	for _, v := range g.GetAllVertices() {
		d, ok := got.Distance(v.Label())
		if !ok {
			t.Fatalf("distance to %s missing", v.Label())
		}
		if d != want[v.Label()] {
			t.Fatalf("distance to %s: got %v, Dijkstra %v", v.Label(), d, want[v.Label()])
		}
		src, ok := got.Source(v.Label())
		if !ok || src.Label() != "A" || src != g.GetVertexByID("A") {
			t.Fatalf("source of %s: got %v", v.Label(), src)
		}
	}
}

// Keeping a later source on a tie fails this.
func TestDijkstraMultiSourceKeepsTheEarlierSource(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a := g.AddVertexByLabel("A")
	x := g.AddVertexByLabel("X")
	b := g.AddVertexByLabel("B")
	addWeighted(t, g, a, x, 2)
	addWeighted(t, g, x, b, 2)

	got, err := DijkstraMultiSource(g, "A", "B")
	if err != nil {
		t.Fatalf("DijkstraMultiSource: %v", err)
	}
	src, ok := got.Source("X")
	if !ok || src.Label() != "A" {
		t.Fatalf("source of X: got %v, want A", src)
	}
	d, ok := got.Distance("X")
	if !ok || d != 2 {
		t.Fatalf("distance to X: got %v, %v, want 2", d, ok)
	}

	reversed, err := DijkstraMultiSource(g, "B", "A")
	if err != nil {
		t.Fatalf("reversed: %v", err)
	}
	src, ok = reversed.Source("X")
	if !ok || src.Label() != "B" {
		t.Fatalf("source of X with B first: got %v, want B", src)
	}
}

// Taking the path through the earlier vertex, when the earlier edge leaves
// that vertex's sibling, fails this.
func TestDijkstraMultiSourceEqualPathFollowsEarlierEdge(t *testing.T) {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	b := g.AddVertexByLabel("B")
	a := g.AddVertexByLabel("A")
	s := g.AddVertexByLabel("S")
	goal := g.AddVertexByLabel("T")
	addWeighted(t, g, s, a, 1)
	addWeighted(t, g, s, b, 1)
	addWeighted(t, g, a, goal, 1)
	addWeighted(t, g, b, goal, 1)

	got, err := DijkstraMultiSource(g, "S")
	if err != nil {
		t.Fatalf("DijkstraMultiSource: %v", err)
	}
	path, ok := got.PathTo("T")
	if !ok {
		t.Fatal("PathTo(T) missing")
	}
	if fmt.Sprint(labelsOf(path)) != "[S A T]" {
		t.Fatalf("path to T: got %v, want [S A T]", labelsOf(path))
	}
}

// Letting an earlier vertex's later edge win fails this. The edges were
// added along S, A, U, T before the edges along S, B, V, T.
func TestDijkstraMultiSourceEqualPathFollowsAddOrder(t *testing.T) {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	b := g.AddVertexByLabel("B")
	a := g.AddVertexByLabel("A")
	s := g.AddVertexByLabel("S")
	u := g.AddVertexByLabel("U")
	v := g.AddVertexByLabel("V")
	goal := g.AddVertexByLabel("T")
	addWeighted(t, g, s, a, 1)
	addWeighted(t, g, s, b, 1)
	addWeighted(t, g, a, u, 1)
	addWeighted(t, g, b, v, 1)
	addWeighted(t, g, u, goal, 1)
	addWeighted(t, g, v, goal, 1)

	got, err := DijkstraMultiSource(g, "S")
	if err != nil {
		t.Fatalf("DijkstraMultiSource: %v", err)
	}
	path, ok := got.PathTo("T")
	if !ok {
		t.Fatal("PathTo(T) missing")
	}
	if fmt.Sprint(labelsOf(path)) != "[S A U T]" {
		t.Fatalf("path to T: got %v, want [S A U T]", labelsOf(path))
	}
}

// A path that starts at the wrong vertex, or whose weights do not add up
// to the distance, fails this.
func TestDijkstraMultiSourcePathTo(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	s := g.AddVertexByLabel("S")
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	goal := g.AddVertexByLabel("T")
	addWeighted(t, g, s, a, 1)
	addWeighted(t, g, s, b, 1)
	addWeighted(t, g, a, goal, 1)
	addWeighted(t, g, b, goal, 1)

	got, err := DijkstraMultiSource(g, "S")
	if err != nil {
		t.Fatalf("DijkstraMultiSource: %v", err)
	}
	path, ok := got.PathTo("T")
	if !ok {
		t.Fatal("PathTo(T) missing")
	}
	labels := make([]string, len(path))
	for i, v := range path {
		labels[i] = v.Label()
		if v != g.GetVertexByID(v.Label()) {
			t.Fatalf("path vertex %s is not the graph's own pointer", v.Label())
		}
	}
	if fmt.Sprint(labels) != "[S A T]" {
		t.Fatalf("path to T: got %v, want [S A T]", labels)
	}
	if sum := pathWeight(t, g, path); sum != 2 {
		t.Fatalf("path weight: got %v, want 2", sum)
	}
	d, _ := got.Distance("T")
	if d != 2 {
		t.Fatalf("distance to T: got %v, want 2", d)
	}
}

// Reaching a vertex against a directed edge fails this.
func TestDijkstraMultiSourceDirectedReach(t *testing.T) {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	addWeighted(t, g, a, b, 1)

	got, err := DijkstraMultiSource(g, "B")
	if err != nil {
		t.Fatalf("DijkstraMultiSource: %v", err)
	}
	if _, ok := got.Distance("A"); ok {
		t.Fatal("A is reachable against the edge")
	}
	if path, ok := got.PathTo("A"); ok || path != nil {
		t.Fatalf("PathTo(A): got %v, %v", path, ok)
	}
	d, ok := got.Distance("B")
	if !ok || d != 0 {
		t.Fatalf("distance to B: got %v, %v", d, ok)
	}
	src, ok := got.Source("B")
	if !ok || src != g.GetVertexByID("B") {
		t.Fatal("B is not its own source")
	}
}

// Accepting a negative weight the source never reaches fails this.
func TestDijkstraMultiSourceRejectsNegativeWeight(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	d := g.AddVertexByLabel("D")
	addWeighted(t, g, a, b, 1)
	addWeighted(t, g, c, d, -1)

	_, err := DijkstraMultiSource(g, "A")
	if !errors.Is(err, ErrNegativeWeight) {
		t.Fatalf("error: got %v, want %v", err, ErrNegativeWeight)
	}
}

func TestDijkstraMultiSourceMissingSource(t *testing.T) {
	g := weightedPath()
	_, err := DijkstraMultiSource(g, "A", "missing")
	if !errors.Is(err, gograph.ErrVertexDoesNotExist) {
		t.Fatalf("error: got %v, want %v", err, gograph.ErrVertexDoesNotExist)
	}
}

// Counting a repeated source a second time fails this.
func TestDijkstraMultiSourceDuplicateSource(t *testing.T) {
	g := weightedPath()
	once, err := DijkstraMultiSource(g, "A")
	if err != nil {
		t.Fatalf("once: %v", err)
	}
	twice, err := DijkstraMultiSource(g, "A", "A")
	if err != nil {
		t.Fatalf("twice: %v", err)
	}
	for _, v := range g.GetAllVertices() {
		d1, ok1 := once.Distance(v.Label())
		d2, ok2 := twice.Distance(v.Label())
		s1, _ := once.Source(v.Label())
		s2, _ := twice.Source(v.Label())
		if ok1 != ok2 || d1 != d2 || s1 != s2 {
			t.Fatalf("%s: once %v %v %v, twice %v %v %v", v.Label(), d1, ok1, s1, d2, ok2, s2)
		}
	}
}

// Returning an error, or a distance, when no sources were given fails this.
func TestDijkstraMultiSourceNoSources(t *testing.T) {
	g := weightedPath()
	got, err := DijkstraMultiSource(g)
	if err != nil {
		t.Fatalf("DijkstraMultiSource: %v", err)
	}
	for _, v := range g.GetAllVertices() {
		if _, ok := got.Distance(v.Label()); ok {
			t.Fatalf("%s is reachable with no sources", v.Label())
		}
		if _, ok := got.Source(v.Label()); ok {
			t.Fatalf("%s has a source with no sources", v.Label())
		}
	}

	empty := gograph.New[string](gograph.Weighted())
	got, err = DijkstraMultiSource(empty)
	if err != nil {
		t.Fatalf("empty graph: %v", err)
	}
	if _, ok := got.Distance("A"); ok {
		t.Fatal("empty graph returned a distance")
	}
	_, err = DijkstraMultiSource(empty, "A")
	if !errors.Is(err, gograph.ErrVertexDoesNotExist) {
		t.Fatalf("empty graph with a source: got %v", err)
	}
}

// Leaving the later source in place across a zero-weight edge fails this.
func TestDijkstraMultiSourceZeroWeight(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	addWeighted(t, g, a, b, 0)

	got, err := DijkstraMultiSource(g, "A", "B")
	if err != nil {
		t.Fatalf("DijkstraMultiSource: %v", err)
	}
	src, ok := got.Source("B")
	if !ok || src.Label() != "A" {
		t.Fatalf("source of B: got %v, want A", src)
	}
	d, ok := got.Distance("B")
	if !ok || d != 0 {
		t.Fatalf("distance to B: got %v, %v", d, ok)
	}
	path, ok := got.PathTo("B")
	if !ok || len(path) != 2 || path[0].Label() != "A" || path[1].Label() != "B" {
		t.Fatalf("path to B: %v", labelsOf(path))
	}
}

// Treating a missing edge weight as anything other than 0 fails this.
func TestDijkstraMultiSourceUnweightedEdge(t *testing.T) {
	g := gograph.New[string]()
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	if _, err := g.AddEdge(a, b); err != nil {
		t.Fatal(err)
	}
	got, err := DijkstraMultiSource(g, "A")
	if err != nil {
		t.Fatalf("DijkstraMultiSource: %v", err)
	}
	d, ok := got.Distance("B")
	if !ok || d != 0 {
		t.Fatalf("distance to B: got %v, %v, want 0", d, ok)
	}
}

// A second run that disagrees with the first fails this.
func TestDijkstraMultiSourceStable(t *testing.T) {
	g := weightedPath()
	first, err := DijkstraMultiSource(g, "A", "D")
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := DijkstraMultiSource(g, "A", "D")
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	for _, v := range g.GetAllVertices() {
		d1, ok1 := first.Distance(v.Label())
		d2, ok2 := second.Distance(v.Label())
		s1, _ := first.Source(v.Label())
		s2, _ := second.Source(v.Label())
		p1, _ := first.PathTo(v.Label())
		p2, _ := second.PathTo(v.Label())
		if ok1 != ok2 || d1 != d2 || s1 != s2 || fmt.Sprint(labelsOf(p1)) != fmt.Sprint(labelsOf(p2)) {
			t.Fatalf("%s changed between runs", v.Label())
		}
	}
}

// A vertex whose distance is not the best Dijkstra run, or whose source is
// not the earliest source at that distance, fails this.
func TestDijkstraMultiSourceRandomMatchesOracle(t *testing.T) {
	//nolint:gosec // seeded so failures are reproducible
	rng := rand.New(rand.NewPCG(1, 2))
	for trial := 0; trial < 40; trial++ {
		n := 4 + rng.IntN(5)
		g := gograph.New[int](gograph.Weighted())
		for i := 0; i < n; i++ {
			g.AddVertexByLabel(i)
		}
		verts := g.GetAllVertices()
		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				if rng.IntN(2) == 0 {
					continue
				}
				addWeighted(t, g, verts[i], verts[j], float64(rng.IntN(6)))
			}
		}
		k := 1 + rng.IntN(3)
		sources := make([]int, k)
		for i := range sources {
			sources[i] = rng.IntN(n)
		}

		got, err := DijkstraMultiSource(g, sources...)
		if err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		bestDist := make(map[int]float64, n)
		bestSrc := make(map[int]int, n)
		seen := make(map[int]bool, k)
		for _, s := range sources {
			if seen[s] {
				continue
			}
			seen[s] = true
			dist := Dijkstra(g, s)
			for _, v := range verts {
				d := dist[v.Label()]
				if d == math.MaxFloat64 {
					continue
				}
				prev, ok := bestDist[v.Label()]
				if ok && d >= prev {
					continue
				}
				bestDist[v.Label()] = d
				bestSrc[v.Label()] = s
			}
		}
		for _, v := range verts {
			label := v.Label()
			d, ok := got.Distance(label)
			want, reachable := bestDist[label]
			if ok != reachable || (reachable && d != want) {
				t.Fatalf("trial %d vertex %d: distance %v %v, oracle %v %v", trial, label, d, ok, want, reachable)
			}
			src, srcOK := got.Source(label)
			if srcOK != reachable {
				t.Fatalf("trial %d vertex %d: source ok %v, reachable %v", trial, label, srcOK, reachable)
			}
			if reachable && src.Label() != bestSrc[label] {
				t.Fatalf("trial %d vertex %d: source %d, oracle %d", trial, label, src.Label(), bestSrc[label])
			}
			if !reachable {
				continue
			}
			path, pathOK := got.PathTo(label)
			if !pathOK || path[0].Label() != bestSrc[label] || path[len(path)-1].Label() != label {
				t.Fatalf("trial %d vertex %d: path %v", trial, label, labelsOf(path))
			}
			if sum := pathWeight(t, g, path); sum != d {
				t.Fatalf("trial %d vertex %d: path weight %v, distance %v", trial, label, sum, d)
			}
		}
	}
}

// Updating a vertex that was already closer to an earlier source fails this.
func TestContinueNearestKeepsTheCloserSource(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	addWeighted(t, g, a, b, 1)
	addWeighted(t, g, b, c, 1)

	adj, _, index, err := buildSourceAdj(g)
	if err != nil {
		t.Fatal(err)
	}
	dist, rank, origin, pred := newSourceState(len(index), 2)
	ia, ic := index["A"], index["C"]
	dist[ia] = 0
	rank[ia] = 0
	origin[ia] = ia
	continueNearest(adj, dist, rank, origin, pred, []int{ia})

	dist[ic] = 0
	rank[ic] = 1
	origin[ic] = ic
	pred[ic] = -1
	continueNearest(adj, dist, rank, origin, pred, []int{ic})

	if origin[index["A"]] != ia || dist[index["A"]] != 0 {
		t.Fatalf("A changed: origin %d dist %v", origin[index["A"]], dist[index["A"]])
	}
	if origin[index["B"]] != ia || dist[index["B"]] != 1 {
		t.Fatalf("B: origin %d dist %v, want A at 1", origin[index["B"]], dist[index["B"]])
	}
	if origin[ic] != ic || dist[ic] != 0 {
		t.Fatalf("C: origin %d dist %v, want itself at 0", origin[ic], dist[ic])
	}
}

// Leaving B at A's longer distance after C opens fails this.
func TestContinueNearestUpdatesTheCloserSource(t *testing.T) {
	g := gograph.New[string](gograph.Weighted())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	addWeighted(t, g, a, b, 5)
	addWeighted(t, g, b, c, 1)

	adj, _, index, err := buildSourceAdj(g)
	if err != nil {
		t.Fatal(err)
	}
	dist, rank, origin, pred := newSourceState(len(index), 2)
	ia, ic := index["A"], index["C"]
	dist[ia] = 0
	rank[ia] = 0
	origin[ia] = ia
	continueNearest(adj, dist, rank, origin, pred, []int{ia})

	dist[ic] = 0
	rank[ic] = 1
	origin[ic] = ic
	pred[ic] = -1
	continueNearest(adj, dist, rank, origin, pred, []int{ic})

	if origin[index["B"]] != ic || dist[index["B"]] != 1 {
		t.Fatalf("B: origin %d dist %v, want C at 1", origin[index["B"]], dist[index["B"]])
	}
	if origin[ia] != ia || dist[ia] != 0 {
		t.Fatalf("A changed: origin %d dist %v", origin[ia], dist[ia])
	}
}

func ExampleDijkstraMultiSource() {
	g := gograph.New[string](gograph.Weighted())
	depotA := g.AddVertexByLabel("depot-a")
	corner := g.AddVertexByLabel("corner")
	depotB := g.AddVertexByLabel("depot-b")
	_, _ = g.AddEdge(depotA, corner, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(corner, depotB, gograph.WithEdgeWeight(2))

	nearest, err := DijkstraMultiSource(g, "depot-a", "depot-b")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, label := range []string{"depot-a", "corner", "depot-b"} {
		src, _ := nearest.Source(label)
		d, _ := nearest.Distance(label)
		fmt.Printf("%s %s %.0f\n", label, src.Label(), d)
	}
	// Output:
	// depot-a depot-a 0
	// corner depot-a 2
	// depot-b depot-b 0
}

func BenchmarkDijkstraMultiSource(b *testing.B) {
	const n = 2000
	const sourcesN = 32
	g := gograph.New[int](gograph.Directed(), gograph.Weighted())
	verts := make([]*gograph.Vertex[int], n)
	for i := 0; i < n; i++ {
		verts[i] = g.AddVertexByLabel(i)
	}
	for i := 0; i < n-1; i++ {
		if _, err := g.AddEdge(verts[i], verts[i+1], gograph.WithEdgeWeight(1)); err != nil {
			b.Fatal(err)
		}
	}
	sources := make([]int, sourcesN)
	for i := range sources {
		sources[i] = i * (n / sourcesN)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DijkstraMultiSource(g, sources...); err != nil {
			b.Fatal(err)
		}
	}
}

func weightedPath() gograph.Graph[string] {
	g := gograph.New[string](gograph.Weighted())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	d := g.AddVertexByLabel("D")
	_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(4))
	_, _ = g.AddEdge(a, c, gograph.WithEdgeWeight(3))
	_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(b, d, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(c, d, gograph.WithEdgeWeight(4))
	return g
}

func addWeighted[T comparable](t *testing.T, g gograph.Graph[T], from, to *gograph.Vertex[T], w float64) {
	t.Helper()
	if _, err := g.AddEdge(from, to, gograph.WithEdgeWeight(w)); err != nil {
		t.Fatal(err)
	}
}

func pathWeight[T comparable](t *testing.T, g gograph.Graph[T], path []*gograph.Vertex[T]) float64 {
	t.Helper()
	var sum float64
	for i := 0; i < len(path)-1; i++ {
		edge := g.GetEdge(path[i], path[i+1])
		if edge == nil {
			t.Fatalf("no edge %v -> %v", path[i].Label(), path[i+1].Label())
		}
		sum += edge.Weight()
	}
	return sum
}

func labelsOf[T comparable](path []*gograph.Vertex[T]) []T {
	out := make([]T, len(path))
	for i, v := range path {
		out[i] = v.Label()
	}
	return out
}
