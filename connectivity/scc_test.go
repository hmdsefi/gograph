package connectivity

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime/debug"
	"sort"
	"testing"

	"github.com/hmdsefi/gograph"
)

var sccFuncs = map[string]func(gograph.Graph[int]) [][]*gograph.Vertex[int]{
	"Tarjan":   Tarjan[int],
	"Gabow":    Gabow[int],
	"Kosaraju": Kosaraju[int],
}

// componentOf maps each vertex label to the index of its component.
func componentOf(sccs [][]*gograph.Vertex[int]) map[int]int {
	comp := make(map[int]int)
	for i, scc := range sccs {
		for _, v := range scc {
			comp[v.Label()] = i
		}
	}

	return comp
}

// partitionKey describes the components independently of their order.
func partitionKey(sccs [][]*gograph.Vertex[int]) string {
	var parts []string
	for _, scc := range sccs {
		labels := make([]int, 0, len(scc))
		for _, v := range scc {
			labels = append(labels, v.Label())
		}
		sort.Ints(labels)
		parts = append(parts, fmt.Sprint(labels))
	}
	sort.Strings(parts)

	return fmt.Sprint(parts)
}

func reachable(g gograph.Graph[int], from int) map[int]bool {
	seen := map[int]bool{from: true}
	queue := []int{from}
	for len(queue) > 0 {
		v := g.GetVertexByID(queue[0])
		queue = queue[1:]
		for _, n := range v.Neighbors() {
			if !seen[n.Label()] {
				seen[n.Label()] = true
				queue = append(queue, n.Label())
			}
		}
	}

	return seen
}

// bruteForceSCCs groups vertices that can reach each other.
func bruteForceSCCs(g gograph.Graph[int]) string {
	vertices := g.GetAllVertices()
	reach := make(map[int]map[int]bool, len(vertices))
	for _, v := range vertices {
		reach[v.Label()] = reachable(g, v.Label())
	}

	assigned := make(map[int]bool)
	var sccs [][]*gograph.Vertex[int]
	for _, v := range vertices {
		if assigned[v.Label()] {
			continue
		}
		var scc []*gograph.Vertex[int]
		for _, w := range vertices {
			if reach[v.Label()][w.Label()] && reach[w.Label()][v.Label()] {
				assigned[w.Label()] = true
				scc = append(scc, w)
			}
		}
		sccs = append(sccs, scc)
	}

	return partitionKey(sccs)
}

func TestSCCs_RandomGraphs(t *testing.T) {
	rng := rand.New(rand.NewSource(1)) //nolint:gosec // seeded so failures are reproducible
	for run := 0; run < 300; run++ {
		n := 1 + rng.Intn(12)
		g := gograph.New[int](gograph.Directed())
		for i := 0; i < n; i++ {
			g.AddVertexByLabel(i)
		}
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if rng.Intn(5) == 0 {
					_, _ = g.AddEdge(g.GetVertexByID(i), g.GetVertexByID(j))
				}
			}
		}

		want := bruteForceSCCs(g)
		for name, scc := range sccFuncs {
			sccs := scc(g)
			if got := partitionKey(sccs); got != want {
				t.Fatalf("run %d, %s: expected %s, got %s", run, name, want, got)
			}

			if name == "Kosaraju" {
				continue
			}

			// Tarjan and Gabow emit a component only after every component
			// it has an edge to, so they come out in reverse topological order.
			comp := componentOf(sccs)
			for _, e := range g.AllEdges() {
				from, to := comp[e.Source().Label()], comp[e.Destination().Label()]
				if from < to {
					t.Fatalf("run %d, %s: component %d comes before component %d", run, name, from, to)
				}
			}
		}
	}
}

func TestGabow_SameAsTarjan(t *testing.T) {
	// Both algorithms keep the unassigned vertices on a stack in the order the
	// search reaches them, and emit a component when its root is done, so
	// they return the same components in the same order.
	rng := rand.New(rand.NewSource(2)) //nolint:gosec // seeded so failures are reproducible
	for run := 0; run < 300; run++ {
		n := 1 + rng.Intn(15)
		g := gograph.New[int](gograph.Directed())
		for i := 0; i < n; i++ {
			g.AddVertexByLabel(i)
		}
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if rng.Intn(6) == 0 {
					_, _ = g.AddEdge(g.GetVertexByID(i), g.GetVertexByID(j))
				}
			}
		}

		want := sccLabels(Tarjan(g))
		if got := sccLabels(Gabow(g)); !reflect.DeepEqual(got, want) {
			t.Fatalf("run %d: Gabow() = %v, Tarjan() = %v", run, got, want)
		}
	}
}

func TestSCCs_DisconnectedWithSelfLoops(t *testing.T) {
	g := gograph.New[int](gograph.Directed())
	for i := 1; i <= 6; i++ {
		g.AddVertexByLabel(i)
	}
	edges := [][2]int{{1, 2}, {2, 1}, {4, 4}, {5, 4}, {6, 6}, {6, 1}}
	for _, e := range edges {
		_, _ = g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1]))
	}

	// 3 is isolated, and 4 and 6 have self-loops but no cycle with another vertex
	want := "[[1 2] [3] [4] [5] [6]]"
	for name, scc := range sccFuncs {
		if got := partitionKey(scc(g)); got != want {
			t.Errorf("%s: expected %s, got %s", name, want, got)
		}
	}
}

func TestSCCs_UndirectedGraph(t *testing.T) {
	g := gograph.New[int]()
	for i := 1; i <= 5; i++ {
		g.AddVertexByLabel(i)
	}
	_, _ = g.AddEdge(g.GetVertexByID(1), g.GetVertexByID(2))
	_, _ = g.AddEdge(g.GetVertexByID(2), g.GetVertexByID(3))
	_, _ = g.AddEdge(g.GetVertexByID(4), g.GetVertexByID(5))

	// every undirected edge goes both ways, so the components are the connected components
	want := "[[1 2 3] [4 5]]"
	for name, scc := range sccFuncs {
		if got := partitionKey(scc(g)); got != want {
			t.Errorf("%s: expected %s, got %s", name, want, got)
		}
	}
}

func TestSCCs_LongPath(t *testing.T) {
	const n = 200_000

	chain := gograph.New[int](gograph.Directed())
	prev := chain.AddVertexByLabel(0)
	for i := 1; i < n; i++ {
		v := chain.AddVertexByLabel(i)
		_, _ = chain.AddEdge(prev, v)
		prev = v
	}

	cycle := gograph.New[int](gograph.Directed())
	prev = cycle.AddVertexByLabel(0)
	for i := 1; i < n; i++ {
		v := cycle.AddVertexByLabel(i)
		_, _ = cycle.AddEdge(prev, v)
		prev = v
	}
	_, _ = cycle.AddEdge(prev, cycle.GetVertexByID(0))

	// A search that recurses once per vertex needs about 48 MB of stack here.
	defer debug.SetMaxStack(debug.SetMaxStack(8 << 20))

	for name, scc := range sccFuncs {
		if got := len(scc(chain)); got != n {
			t.Errorf("%s on a chain: expected %d components, got %d", name, n, got)
		}

		sccs := scc(cycle)
		if len(sccs) != 1 || len(sccs[0]) != n {
			t.Errorf("%s on a cycle: expected 1 component of %d vertices, got %d components", name, n, len(sccs))
		}
	}
}
