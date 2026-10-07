package connectivity

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

func ExampleBridges() {
	g := gograph.New[string]()
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))
	_, _ = g.AddEdge(gograph.NewVertex("B"), gograph.NewVertex("C"))
	edges, _ := Bridges(g)
	for _, e := range edges {
		fmt.Println(e.Source().Label(), e.Destination().Label())
	}
	// Output:
	// B C
	// A B
}

func ExampleArticulationPoints() {
	g := gograph.New[string]()
	_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))
	_, _ = g.AddEdge(gograph.NewVertex("B"), gograph.NewVertex("C"))
	vertices, _ := ArticulationPoints(g)
	for _, v := range vertices {
		fmt.Println(v.Label())
	}
	// Output: B
}

func TestCutsLongChain(t *testing.T) {
	g := gograph.New[int]()
	const n = 10000
	for i := 0; i < n; i++ {
		g.AddVertexByLabel(i)
	}
	for i := 1; i < n; i++ {
		_, _ = g.AddEdge(g.GetVertexByID(i-1), g.GetVertexByID(i))
	}
	bs, err := Bridges(g)
	if err != nil || len(bs) != n-1 {
		t.Fatalf("bridges %d, error %v", len(bs), err)
	}
	vs, err := ArticulationPoints(g)
	if err != nil || len(vs) != n-2 {
		t.Fatalf("points %d, error %v", len(vs), err)
	}
}

func TestCuts(t *testing.T) {
	for _, tt := range []struct {
		name    string
		n       int
		edges   [][2]int
		bridges [][2]int
		points  []int
	}{
		{"empty", 0, nil, [][2]int{}, []int{}},
		{"isolated", 3, nil, [][2]int{}, []int{}},
		{"chain", 4, [][2]int{{0, 1}, {1, 2}, {2, 3}}, [][2]int{{2, 3}, {1, 2}, {0, 1}}, []int{1, 2}},
		{"cycle with tail", 4, [][2]int{{0, 1}, {1, 2}, {2, 0}, {2, 3}}, [][2]int{{2, 3}}, []int{2}},
		{"root and disconnected", 6, [][2]int{{0, 1}, {0, 2}, {3, 4}, {4, 5}, {5, 3}}, [][2]int{{0, 1}, {0, 2}}, []int{0}},
		{"self loops", 2, [][2]int{{0, 0}, {0, 1}, {1, 1}}, [][2]int{{0, 1}}, []int{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := gograph.New[int]()
			for i := 0; i < tt.n; i++ {
				g.AddVertexByLabel(i)
			}
			for _, e := range tt.edges {
				if _, err := g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1])); err != nil {
					t.Fatal(err)
				}
			}
			bs, err := Bridges(g)
			if err != nil {
				t.Fatal(err)
			}
			got := make([][2]int, 0, len(bs))
			for _, e := range bs {
				got = append(got, [2]int{e.Source().Label(), e.Destination().Label()})
				if e != g.GetEdge(e.Source(), e.Destination()) {
					t.Fatal("copied edge")
				}
			}
			if !reflect.DeepEqual(got, tt.bridges) {
				t.Fatalf("bridges %v, want %v", got, tt.bridges)
			}
			vs, err := ArticulationPoints(g)
			if err != nil {
				t.Fatal(err)
			}
			labels := make([]int, 0, len(vs))
			for _, v := range vs {
				labels = append(labels, v.Label())
				if v != g.GetVertexByID(v.Label()) {
					t.Fatal("copied vertex")
				}
			}
			if !reflect.DeepEqual(labels, tt.points) {
				t.Fatalf("points %v, want %v", labels, tt.points)
			}
		})
	}
}

func TestCutsDirected(t *testing.T) {
	g := gograph.New[int](gograph.Directed())
	if _, err := Bridges(g); err != ErrNotUndirected {
		t.Fatalf("Bridges error: %v", err)
	}
	if _, err := ArticulationPoints(g); err != ErrNotUndirected {
		t.Fatalf("ArticulationPoints error: %v", err)
	}
}

func TestCutsAgainstRemoval(t *testing.T) {
	pairs := [][2]int{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3}}
	for mask := 0; mask < 1<<len(pairs); mask++ {
		g := gograph.New[int]()
		for i := 0; i < 4; i++ {
			g.AddVertexByLabel(i)
		}
		var edges [][2]int
		for i, pair := range pairs {
			if mask&(1<<i) != 0 {
				edges = append(edges, pair)
				_, _ = g.AddEdge(g.GetVertexByID(pair[0]), g.GetVertexByID(pair[1]))
			}
		}
		base := countCutComponents(edges, -1, -1)
		bs, _ := Bridges(g)
		gotEdges := make(map[[2]int]bool)
		for _, e := range bs {
			a, b := e.Source().Label(), e.Destination().Label()
			if a > b {
				a, b = b, a
			}
			gotEdges[[2]int{a, b}] = true
		}
		for i, e := range edges {
			if gotEdges[e] != (countCutComponents(edges, -1, i) > base) {
				t.Fatalf("mask %d: bridge %v", mask, e)
			}
		}
		vs, _ := ArticulationPoints(g)
		gotPoints := make(map[int]bool)
		for _, v := range vs {
			gotPoints[v.Label()] = true
		}
		for v := 0; v < 4; v++ {
			if gotPoints[v] != (countCutComponents(edges, v, -1) > base) {
				t.Fatalf("mask %d: articulation %d", mask, v)
			}
		}
	}
}

func countCutComponents(edges [][2]int, removedVertex, removedEdge int) int {
	seen := make(map[int]bool)
	count := 0
	for root := 0; root < 4; root++ {
		if root == removedVertex || seen[root] {
			continue
		}
		count++
		queue := []int{root}
		seen[root] = true
		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			for i, e := range edges {
				if i == removedEdge || e[0] == removedVertex || e[1] == removedVertex {
					continue
				}
				neighbor := -1
				if e[0] == v {
					neighbor = e[1]
				} else if e[1] == v {
					neighbor = e[0]
				}
				if neighbor >= 0 && !seen[neighbor] {
					seen[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}
	}
	return count
}
