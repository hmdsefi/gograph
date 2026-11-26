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

// cycleGraph returns a directed graph with the given edges, written as
// "AB" for A -> B. The vertices are added in the order they first appear.
func cycleGraph(edges ...string) gograph.Graph[string] {
	g := gograph.New[string](gograph.Directed())
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

func cycleLabels[T comparable](cycles [][]*gograph.Vertex[T]) [][]T {
	if cycles == nil {
		return nil
	}
	out := make([][]T, len(cycles))
	for i, c := range cycles {
		out[i] = make([]T, len(c))
		for j, v := range c {
			out[i][j] = v.Label()
		}
	}
	return out
}

func TestCycles_Triangle(t *testing.T) {
	// C is added first, so the triangle starts at C.
	g := gograph.New[string](gograph.Directed())
	c, a, b := g.AddVertexByLabel("C"), g.AddVertexByLabel("A"), g.AddVertexByLabel("B")
	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(b, c)
	_, _ = g.AddEdge(c, a)

	cycles, err := Cycles(g, 10)
	if err != nil {
		t.Fatal(err)
	}
	if want := [][]string{{"C", "A", "B"}}; !reflect.DeepEqual(cycleLabels(cycles), want) {
		t.Fatalf("Cycles = %v, want %v", cycleLabels(cycles), want)
	}
	for i, v := range cycles[0] {
		if v != g.GetVertexByID(v.Label()) {
			t.Errorf("vertex %d is not the graph's own pointer", i)
		}
	}
}

func TestCycles_TwoCyclesShareAVertex(t *testing.T) {
	g := cycleGraph("AB", "BA", "AC", "CD", "DA")
	cycles, err := Cycles(g, 10)
	if err != nil {
		t.Fatal(err)
	}
	if want := [][]string{{"A", "B"}, {"A", "C", "D"}}; !reflect.DeepEqual(cycleLabels(cycles), want) {
		t.Fatalf("Cycles = %v, want %v", cycleLabels(cycles), want)
	}

	one, err := Cycles(g, 1)
	if err != nil {
		t.Fatal(err)
	}
	if want := [][]string{{"A", "B"}}; !reflect.DeepEqual(cycleLabels(one), want) {
		t.Fatalf("Cycles with limit 1 = %v, want %v", cycleLabels(one), want)
	}
}

func TestCycles_SelfLoop(t *testing.T) {
	g := cycleGraph("AA", "AB", "BB", "BC")
	cycles, err := Cycles(g, 10)
	if err != nil {
		t.Fatal(err)
	}
	if want := [][]string{{"A"}, {"B"}}; !reflect.DeepEqual(cycleLabels(cycles), want) {
		t.Fatalf("Cycles = %v, want %v", cycleLabels(cycles), want)
	}
}

func TestCycles_NoCycles(t *testing.T) {
	dag := cycleGraph("AB", "BC", "AC", "CD")
	acyclic := gograph.New[string](gograph.Directed(), gograph.Acyclic())
	_, _ = acyclic.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))

	for name, g := range map[string]gograph.Graph[string]{
		"dag":     dag,
		"acyclic": acyclic,
		"empty":   gograph.New[string](gograph.Directed()),
	} {
		t.Run(name, func(t *testing.T) {
			cycles, err := Cycles(g, 10)
			if err != nil || cycles != nil {
				t.Fatalf("Cycles = %v, %v; want nil, nil", cycles, err)
			}
		})
	}
}

func TestCycles_Errors(t *testing.T) {
	undirected := gograph.New[string]()
	_, _ = undirected.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))
	if _, err := Cycles(undirected, 10); !errors.Is(err, ErrNotDirected) || !errors.Is(err, gograph.ErrNotDirected) {
		t.Errorf("undirected graph: err = %v, want ErrNotDirected", err)
	}

	g := cycleGraph("AB", "BA")
	for _, limit := range []int{0, -1} {
		if cycles, err := Cycles(g, limit); !errors.Is(err, ErrInvalidLimit) || cycles != nil {
			t.Errorf("limit %d: got %v, %v; want nil, ErrInvalidLimit", limit, cycles, err)
		}
	}
}

func TestCycles_StopsAtTheLimitAndRunsAgree(t *testing.T) {
	// A complete directed graph on 6 vertices has 409 cycles.
	g := gograph.New[int](gograph.Directed())
	vertices := make([]*gograph.Vertex[int], 6)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for _, a := range vertices {
		for _, b := range vertices {
			if a != b {
				_, _ = g.AddEdge(a, b)
			}
		}
	}

	all, err := Cycles(g, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 409 {
		t.Fatalf("got %d cycles, want 409", len(all))
	}
	for _, limit := range []int{1, 7, 100, 408, 409} {
		first, _ := Cycles(g, limit)
		second, _ := Cycles(g, limit)
		if len(first) != limit {
			t.Fatalf("limit %d: got %d cycles", limit, len(first))
		}
		if !reflect.DeepEqual(first, second) {
			t.Fatalf("limit %d: two runs differ", limit)
		}
		if !reflect.DeepEqual(first, all[:limit]) {
			t.Fatalf("limit %d: not the first cycles of the full list", limit)
		}
	}
}

func TestCycles_BlockedVertexIsUnblockedLater(t *testing.T) {
	// From A, the search reaches D through B first, while C is on the path,
	// so D fails and is blocked. The cycle through C unblocks it again.
	g := cycleGraph("AB", "BC", "CD", "DC", "CA", "BD", "DA")
	cycles, err := Cycles(g, 100)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"A", "B", "C", "D"}, {"A", "B", "C"}, {"A", "B", "D", "C"}, {"A", "B", "D"}, {"C", "D"}}
	if got := cycleLabels(cycles); !reflect.DeepEqual(got, want) {
		t.Fatalf("Cycles = %v, want %v", got, want)
	}
}

// bruteForceCycles lists the cycles of g in the order Cycles promises: for
// each start in vertex order, a depth-first search over the later vertices,
// following the order of Neighbors.
func bruteForceCycles(g gograph.Graph[int]) [][]int {
	vertices := g.GetAllVertices()
	position := map[int]int{}
	for i, v := range vertices {
		position[v.Label()] = i
	}

	var cycles [][]int
	for s, start := range vertices {
		onPath := map[int]bool{start.Label(): true}
		path := []int{start.Label()}
		var search func(v *gograph.Vertex[int])
		search = func(v *gograph.Vertex[int]) {
			for w := range v.AllNeighbors() {
				switch {
				case w == start:
					cycles = append(cycles, append([]int(nil), path...))
				case position[w.Label()] > s && !onPath[w.Label()]:
					onPath[w.Label()] = true
					path = append(path, w.Label())
					search(w)
					path = path[:len(path)-1]
					onPath[w.Label()] = false
				}
			}
		}
		search(start)
	}
	return cycles
}

func TestCycles_MatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewPCG(19, 5)) //nolint:gosec // seeded so failures are reproducible
	for i := range 1500 {
		n := 1 + rng.IntN(9)
		g := gograph.New[int](gograph.Directed())
		vertices := make([]*gograph.Vertex[int], n)
		for j, label := range rng.Perm(n) {
			vertices[j] = g.AddVertexByLabel(label)
		}
		for range rng.IntN(n*n + 1) {
			_, _ = g.AddEdge(vertices[rng.IntN(n)], vertices[rng.IntN(n)])
		}

		want := bruteForceCycles(g)
		got, err := Cycles(g, 1<<30)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cycleLabels(got), want) {
			t.Fatalf("graph %d: Cycles = %v, want %v", i, cycleLabels(got), want)
		}
		if len(want) > 1 {
			limit := 1 + rng.IntN(len(want)-1)
			if got, _ := Cycles(g, limit); !reflect.DeepEqual(cycleLabels(got), want[:limit]) {
				t.Fatalf("graph %d, limit %d: Cycles = %v, want %v", i, limit, cycleLabels(got), want[:limit])
			}
		}
	}
}

func TestCycles_ManySmallComponents(t *testing.T) {
	// A chain of vertices, each with a self-loop. A search that recomputed
	// the components of every later vertex for each start would take
	// quadratic time, and its visit counter would pass 2^31.
	const n = 70_000
	g := gograph.New[int](gograph.Directed())
	vertices := make([]*gograph.Vertex[int], n)
	var prev *gograph.Vertex[int]
	for i := range vertices {
		v := g.AddVertexByLabel(i)
		vertices[i] = v
		_, _ = g.AddEdge(v, v)
		if prev != nil {
			_, _ = g.AddEdge(prev, v)
		}
		prev = v
	}

	cycles, err := Cycles(g, n+1)
	if err != nil {
		t.Fatal(err)
	}
	if len(cycles) != n {
		t.Fatalf("got %d cycles, want %d", len(cycles), n)
	}
	for i, c := range cycles {
		if len(c) != 1 || c[0] != vertices[i] {
			t.Fatalf("cycle %d = %v, want the self-loop of %d", i, cycleLabels(cycles[i:i+1]), i)
		}
	}
}

// cycleWrapper is a graph type other than the one from gograph.New.
type cycleWrapper struct {
	gograph.Graph[int]
}

func TestCycles_OtherGraphTypesAndRemovedVertices(t *testing.T) {
	rng := rand.New(rand.NewPCG(31, 2)) //nolint:gosec // seeded so failures are reproducible
	for i := range 2000 {
		n := 2 + rng.IntN(9)
		g := gograph.New[int](gograph.Directed())
		vertices := make([]*gograph.Vertex[int], n)
		for j, label := range rng.Perm(n) {
			vertices[j] = g.AddVertexByLabel(label)
		}
		for range rng.IntN(n*n + 1) {
			_, _ = g.AddEdge(vertices[rng.IntN(n)], vertices[rng.IntN(n)])
		}
		if rng.IntN(2) == 0 {
			g.RemoveVertices(vertices[rng.IntN(n)])
		}
		graph := g
		if rng.IntN(2) == 0 {
			graph = cycleWrapper{g}
		}

		want := bruteForceCycles(g)
		got, err := Cycles(graph, 1<<30)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cycleLabels(got), want) {
			t.Fatalf("graph %d: Cycles = %v, want %v", i, cycleLabels(got), want)
		}
	}
}

func ExampleCycles() {
	// Imports between packages: two import cycles share the package "app".
	g := gograph.New[string](gograph.Directed())
	app, db, log, cfg := g.AddVertexByLabel("app"), g.AddVertexByLabel("db"),
		g.AddVertexByLabel("log"), g.AddVertexByLabel("cfg")
	_, _ = g.AddEdge(app, db)
	_, _ = g.AddEdge(db, app)
	_, _ = g.AddEdge(app, log)
	_, _ = g.AddEdge(log, cfg)
	_, _ = g.AddEdge(cfg, app)

	cycles, err := Cycles(g, 10)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, c := range cycles {
		fmt.Println(pathLabels(c))
	}
	// Output:
	// [app db]
	// [app log cfg]
}

// cycleRandom returns a random directed graph with n vertices and m edges.
func cycleRandom(n, m int) gograph.Graph[int] {
	rng := rand.New(rand.NewPCG(6, 1)) //nolint:gosec // seeded so runs are comparable
	g := gograph.New[int](gograph.Directed())
	vertices := make([]*gograph.Vertex[int], n)
	for i := range vertices {
		vertices[i] = g.AddVertexByLabel(i)
	}
	for range m {
		_, _ = g.AddEdge(vertices[rng.IntN(n)], vertices[rng.IntN(n)])
	}
	return g
}

func BenchmarkCycles(b *testing.B) {
	g := cycleRandom(1000, 3000)
	for _, limit := range []int{1, 100, 10_000} {
		b.Run("random/limit="+strconv.Itoa(limit), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := Cycles(g, limit); err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	b.Run("self-loop chain", func(b *testing.B) {
		const n = 16_000
		g := gograph.New[int](gograph.Directed())
		prev := g.AddVertexByLabel(0)
		_, _ = g.AddEdge(prev, prev)
		for i := 1; i < n; i++ {
			v := g.AddVertexByLabel(i)
			_, _ = g.AddEdge(v, v)
			_, _ = g.AddEdge(prev, v)
			prev = v
		}
		b.ReportAllocs()
		for b.Loop() {
			if _, err := Cycles(g, n); err != nil {
				b.Fatal(err)
			}
		}
	})
}
