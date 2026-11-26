package path

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"strconv"
	"testing"

	"github.com/hmdsefi/gograph"
)

// simplePathGraph returns a graph with the given edges, written as "AB". The
// vertices are added in the order they first appear.
func simplePathGraph(directed bool, edges ...string) gograph.Graph[string] {
	var g gograph.Graph[string]
	if directed {
		g = gograph.New[string](gograph.Directed())
	} else {
		g = gograph.New[string]()
	}
	for _, e := range edges {
		from, to := string(e[0]), string(e[1])
		if g.GetVertexByID(from) == nil {
			g.AddVertexByLabel(from)
		}
		if g.GetVertexByID(to) == nil {
			g.AddVertexByLabel(to)
		}
		_, _ = g.AddEdge(g.GetVertexByID(from), g.GetVertexByID(to))
	}
	return g
}

func simplePathLabels[T comparable](paths [][]*gograph.Vertex[T]) [][]T {
	out := make([][]T, len(paths))
	for i, p := range paths {
		out[i] = make([]T, len(p))
		for j, v := range p {
			out[i][j] = v.Label()
		}
	}
	return out
}

func TestAllSimplePaths_Diamond(t *testing.T) {
	g := simplePathGraph(true, "AB", "AC", "BD", "CD")
	paths, err := AllSimplePaths(g, "A", "D")
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"A", "B", "D"}, {"A", "C", "D"}}
	if got := simplePathLabels(paths); !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
	for _, p := range paths {
		for _, v := range p {
			if v != g.GetVertexByID(v.Label()) {
				t.Fatalf("%v is not the graph's own vertex", v.Label())
			}
		}
	}
}

func TestAllSimplePaths_CycleIsNotRepeated(t *testing.T) {
	// B -> C -> B is a cycle on the way to D.
	g := simplePathGraph(true, "AB", "BC", "CB", "CD", "BD", "AA")
	paths, err := AllSimplePaths(g, "A", "D")
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"A", "B", "C", "D"}, {"A", "B", "D"}}
	if got := simplePathLabels(paths); !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
}

func TestAllSimplePaths_UndirectedTriangle(t *testing.T) {
	g := simplePathGraph(false, "AB", "BC", "CA")
	paths, err := AllSimplePaths(g, "A", "B")
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"A", "B"}, {"A", "C", "B"}}
	if got := simplePathLabels(paths); !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
}

func TestAllSimplePaths_MaxLengthAndLimit(t *testing.T) {
	g := simplePathGraph(true, "AB", "BC", "CD", "AC", "BD", "AD")
	all := [][]string{{"A", "B", "C", "D"}, {"A", "B", "D"}, {"A", "C", "D"}, {"A", "D"}}

	tests := []struct {
		name string
		opts []AllPathsOption
		want [][]string
	}{
		{"none", nil, all},
		{"max length 1", []AllPathsOption{WithMaxLength(1)}, [][]string{{"A", "D"}}},
		{"max length 2", []AllPathsOption{WithMaxLength(2)}, all[1:]},
		{"max length 3", []AllPathsOption{WithMaxLength(3)}, all},
		{"limit 1", []AllPathsOption{WithLimit(1)}, all[:1]},
		{"limit 3", []AllPathsOption{WithLimit(3)}, all[:3]},
		{"limit 9", []AllPathsOption{WithLimit(9)}, all},
		{"both", []AllPathsOption{WithMaxLength(2), WithLimit(2)}, all[1:3]},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			paths, err := AllSimplePaths(g, "A", "D", tc.opts...)
			if err != nil {
				t.Fatal(err)
			}
			if got := simplePathLabels(paths); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("paths = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAllSimplePaths_MaxLengthAfterAShortcutWasOnThePath(t *testing.T) {
	// The first path, 0 2 3 1 4 5, reaches 3 while 2 is on the path, so from
	// 3 the only way to 5 takes three edges. Later 3 comes at depth 3 with 2
	// free, and 3 2 5 fits in the two edges left.
	g := gograph.New[int]()
	for _, label := range []int{1, 4, 0, 5, 2, 3} {
		g.AddVertexByLabel(label)
	}
	for _, e := range [][2]int{{4, 5}, {1, 4}, {0, 2}, {4, 0}, {2, 3}, {5, 2}, {1, 3}} {
		_, _ = g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1]))
	}

	paths, err := AllSimplePaths(g, 0, 5, WithMaxLength(5))
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int{{0, 2, 3, 1, 4, 5}, {0, 2, 5}, {0, 4, 5}, {0, 4, 1, 3, 2, 5}}
	if got := simplePathLabels(paths); !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
}

func TestAllSimplePaths_NoPaths(t *testing.T) {
	g := simplePathGraph(true, "AB", "BA", "CA")
	for _, tc := range []struct{ from, to string }{{"A", "A"}, {"A", "C"}, {"B", "B"}} {
		paths, err := AllSimplePaths(g, tc.from, tc.to)
		if err != nil {
			t.Fatal(err)
		}
		if paths == nil || len(paths) != 0 {
			t.Errorf("%s to %s: paths = %v, want an empty slice", tc.from, tc.to, paths)
		}
	}
}

func TestAllSimplePaths_Errors(t *testing.T) {
	g := simplePathGraph(true, "AB")
	tests := []struct {
		name     string
		from, to string
		opts     []AllPathsOption
		want     error
	}{
		{"unknown from", "X", "B", nil, gograph.ErrVertexDoesNotExist},
		{"unknown to", "A", "X", nil, gograph.ErrVertexDoesNotExist},
		{"limit 0", "A", "B", []AllPathsOption{WithLimit(0)}, ErrInvalidLimit},
		{"negative limit", "A", "B", []AllPathsOption{WithLimit(-1)}, ErrInvalidLimit},
		{"max length 0", "A", "B", []AllPathsOption{WithMaxLength(0)}, ErrInvalidLimit},
		{"invalid then valid", "A", "B", []AllPathsOption{WithLimit(0), WithLimit(2)}, ErrInvalidLimit},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			paths, err := AllSimplePaths(g, tc.from, tc.to, tc.opts...)
			if !errors.Is(err, tc.want) || paths != nil {
				t.Errorf("AllSimplePaths = %v, %v; want nil, %v", paths, err, tc.want)
			}
			seq, err := AllSimplePathsSeq(g, tc.from, tc.to, tc.opts...)
			if !errors.Is(err, tc.want) || seq != nil {
				t.Errorf("AllSimplePathsSeq error = %v, want %v and a nil sequence", err, tc.want)
			}
		})
	}
}

func TestAllSimplePathsSeq_StopsEarly(t *testing.T) {
	g := simplePathGraph(true, "AB", "AC", "BD", "CD")
	seq, err := AllSimplePathsSeq(g, "A", "D")
	if err != nil {
		t.Fatal(err)
	}
	var got [][]*gograph.Vertex[string]
	for p := range seq {
		got = append(got, p)
		break
	}
	if want := [][]string{{"A", "B", "D"}}; !reflect.DeepEqual(simplePathLabels(got), want) {
		t.Fatalf("paths = %v, want %v", simplePathLabels(got), want)
	}

	// Ranging again runs the search again.
	count := 0
	for range seq {
		count++
	}
	if count != 2 {
		t.Fatalf("second range gave %d paths, want 2", count)
	}
}

func TestAllSimplePaths_SlicesAreNotShared(t *testing.T) {
	g := simplePathGraph(true, "AB", "AC", "BD", "CD")
	paths, _ := AllSimplePaths(g, "A", "D")
	paths[0][1] = nil
	if paths[1][1] == nil || paths[1][1].Label() != "C" {
		t.Fatal("changing one path changed another")
	}
	again, _ := AllSimplePaths(g, "A", "D")
	if !reflect.DeepEqual(simplePathLabels(again), [][]string{{"A", "B", "D"}, {"A", "C", "D"}}) {
		t.Fatal("changing a result changed the next one")
	}
}

func TestAllSimplePaths_SamePathsOnEveryRun(t *testing.T) {
	rng := rand.New(rand.NewPCG(8, 3)) //nolint:gosec // seeded so failures are reproducible
	g := simplePathRandom(rng, 12, 40, true)
	first, _ := AllSimplePaths(g, 0, 11)
	for range 5 {
		again, _ := AllSimplePaths(g, 0, 11)
		if !reflect.DeepEqual(first, again) {
			t.Fatal("paths differ between runs")
		}
	}
}

// simplePathRandom returns a graph with n vertices and about m random edges,
// self-loops included, with the labels added in a random order.
func simplePathRandom(rng *rand.Rand, n, m int, directed bool) gograph.Graph[int] {
	var g gograph.Graph[int]
	if directed {
		g = gograph.New[int](gograph.Directed())
	} else {
		g = gograph.New[int]()
	}
	vertices := make([]*gograph.Vertex[int], n)
	for i, label := range rng.Perm(n) {
		vertices[i] = g.AddVertexByLabel(label)
	}
	for range m {
		_, _ = g.AddEdge(vertices[rng.IntN(n)], vertices[rng.IntN(n)])
	}
	return g
}

// bruteForceSimplePaths lists the simple paths from `from` to `to` of at most
// maxLength edges in depth-first order, following the order of Neighbors.
func bruteForceSimplePaths(g gograph.Graph[int], from, to, maxLength int) [][]int {
	start, goal := g.GetVertexByID(from), g.GetVertexByID(to)
	var paths [][]int
	if start == goal {
		return paths
	}
	onPath := map[*gograph.Vertex[int]]bool{start: true}
	path := []int{from}
	var search func(v *gograph.Vertex[int])
	search = func(v *gograph.Vertex[int]) {
		if len(path) > maxLength {
			return
		}
		for w := range v.AllNeighbors() {
			switch {
			case w == goal:
				paths = append(paths, append(append([]int(nil), path...), to))
			case !onPath[w]:
				onPath[w] = true
				path = append(path, w.Label())
				search(w)
				path = path[:len(path)-1]
				onPath[w] = false
			}
		}
	}
	search(start)
	return paths
}

func TestAllSimplePaths_MatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewPCG(21, 4)) //nolint:gosec // seeded so failures are reproducible
	for i := range 3000 {
		n := 1 + rng.IntN(8)
		directed := rng.IntN(2) == 0
		g := simplePathRandom(rng, n, rng.IntN(n*n/2+2), directed)
		from, to := rng.IntN(n), rng.IntN(n)

		var opts []AllPathsOption
		maxLength := n
		if rng.IntN(2) == 0 {
			maxLength = 1 + rng.IntN(n)
			opts = append(opts, WithMaxLength(maxLength))
		}
		want := bruteForceSimplePaths(g, from, to, maxLength)
		if len(want) > 1 && rng.IntN(2) == 0 {
			limit := 1 + rng.IntN(len(want)-1)
			want = want[:limit]
			opts = append(opts, WithLimit(limit))
		}

		paths, err := AllSimplePaths(g, from, to, opts...)
		if err != nil {
			t.Fatal(err)
		}
		got := simplePathLabels(paths)
		if len(got) != len(want) || (len(want) > 0 && !reflect.DeepEqual(got, want)) {
			t.Fatalf("graph %d (directed %v), %d to %d, max length %d: paths = %v, want %v",
				i, directed, from, to, maxLength, got, want)
		}
	}
}

func ExampleAllSimplePaths() {
	// Why does the service depend on the library? Through two packages.
	g := gograph.New[string](gograph.Directed())
	service, http, db, lib := g.AddVertexByLabel("service"), g.AddVertexByLabel("http"),
		g.AddVertexByLabel("db"), g.AddVertexByLabel("lib")
	_, _ = g.AddEdge(service, http)
	_, _ = g.AddEdge(service, db)
	_, _ = g.AddEdge(http, lib)
	_, _ = g.AddEdge(db, lib)
	_, _ = g.AddEdge(http, db)

	paths, err := AllSimplePaths(g, "service", "lib")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, p := range paths {
		fmt.Println(pathLabels(p))
	}

	short, _ := AllSimplePaths(g, "service", "lib", WithMaxLength(2))
	fmt.Println(len(short), "paths of at most 2 edges")
	// Output:
	// [service http lib]
	// [service http db lib]
	// [service db lib]
	// 2 paths of at most 2 edges
}

// simplePathTrap returns a directed graph where 0 -> 1 -> goal is the only
// path from 0 to the goal, the last vertex. Vertex 1 also leads into a
// complete directed graph of n vertices whose only way out is back to 1. Each
// of those vertices reaches the goal in the whole graph, but only through 1,
// which is already on the path. A search that only skips vertices that can't
// reach the goal at all tries every path through the complete graph.
func simplePathTrap(n int) gograph.Graph[int] {
	g := gograph.New[int](gograph.Directed())
	vertices := make([]*gograph.Vertex[int], n+3)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	_, _ = g.AddEdge(vertices[0], vertices[1])
	for i := 2; i < n+2; i++ {
		_, _ = g.AddEdge(vertices[1], vertices[i])
		for j := 2; j < n+2; j++ {
			if i != j {
				_, _ = g.AddEdge(vertices[i], vertices[j])
			}
		}
		_, _ = g.AddEdge(vertices[i], vertices[1])
	}
	_, _ = g.AddEdge(vertices[1], vertices[n+2])
	return g
}

func BenchmarkAllSimplePaths(b *testing.B) {
	b.Run("grid 5x5", func(b *testing.B) {
		g := astarGrid(5)
		b.ReportAllocs()
		for b.Loop() {
			if _, err := AllSimplePaths(g, 0, 24); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("grid 10x10, max length 20", func(b *testing.B) {
		g := astarGrid(10)
		b.ReportAllocs()
		for b.Loop() {
			if _, err := AllSimplePaths(g, 0, 99, WithMaxLength(20)); err != nil {
				b.Fatal(err)
			}
		}
	})
	for _, n := range []int{8, 10} {
		b.Run("trap "+strconv.Itoa(n), func(b *testing.B) {
			g := simplePathTrap(n)
			b.ReportAllocs()
			for b.Loop() {
				if _, err := AllSimplePaths(g, 0, n+2); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// simplePathRetries returns a graph with a single path from 0 to the goal,
// the last vertex, that is the slowest case known for WithMaxLength: when the
// search backs out of the spine, each spine vertex unblocks the vertices that
// wait for it, and they are tried again at every depth. The spine a0 ... ak
// z1 ... zr goal has r = k+m. Each ai also leads through a chain of k padding
// vertices into a chain d1 ... dm with shortcuts d1 -> dj, and dm leads back
// to every ai. It returns the graph, the goal and the maximum length that
// keeps the spine path, k+r+1.
func simplePathRetries(k, m int) (gograph.Graph[int], int, int) {
	g := gograph.New[int](gograph.Directed())
	next := 0
	add := func() *gograph.Vertex[int] {
		v := g.AddVertexByLabel(next)
		next++
		return v
	}
	edge := func(a, b *gograph.Vertex[int]) { _, _ = g.AddEdge(a, b) }

	r := k + m
	spine := make([]*gograph.Vertex[int], k+1)
	for i := range spine {
		spine[i] = add()
	}
	tail := make([]*gograph.Vertex[int], r)
	for i := range tail {
		tail[i] = add()
	}
	pad := make([]*gograph.Vertex[int], k)
	for i := range pad {
		pad[i] = add()
	}
	d := make([]*gograph.Vertex[int], m)
	for i := range d {
		d[i] = add()
	}
	goal := add()

	for i := 0; i < k; i++ {
		edge(spine[i], spine[i+1])
	}
	edge(spine[k], tail[0])
	for i := 0; i+1 < r; i++ {
		edge(tail[i], tail[i+1])
	}
	edge(tail[r-1], goal)
	for _, a := range spine {
		edge(a, pad[0])
	}
	for i := 0; i+1 < k; i++ {
		edge(pad[i], pad[i+1])
	}
	edge(pad[k-1], d[0])
	for i := 0; i+1 < m; i++ {
		edge(d[i], d[i+1])
	}
	for j := 2; j < m; j++ {
		edge(d[0], d[j])
	}
	for _, a := range spine {
		edge(d[m-1], a)
	}
	return g, goal.Label(), k + r + 1
}

func BenchmarkAllSimplePathsRetries(b *testing.B) {
	for _, k := range []int{25, 50, 100} {
		g, goal, maxLength := simplePathRetries(k, k)
		b.Run("k="+strconv.Itoa(k), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				paths, err := AllSimplePaths(g, 0, goal, WithMaxLength(maxLength))
				if err != nil || len(paths) != 1 {
					b.Fatal(len(paths), err)
				}
			}
		})
	}
}
