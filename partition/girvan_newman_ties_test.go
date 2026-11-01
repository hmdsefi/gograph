package partition

import (
	"math"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

// graphFromEdges returns an undirected graph with the vertices 0 to n-1,
// added in order, and the given edges.
func graphFromEdges(n int, edges [][2]int) gograph.Graph[int] {
	g := gograph.New[int]()
	for i := 0; i < n; i++ {
		g.AddVertexByLabel(i)
	}
	for _, e := range edges {
		_, _ = g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1]))
	}
	return g
}

func intCommunities(graphs []gograph.Graph[int]) [][]int {
	out := make([][]int, len(graphs))
	for i, g := range graphs {
		for _, v := range g.GetAllVertices() {
			out[i] = append(out[i], v.Label())
		}
	}
	return out
}

// The expected communities below are the ones the same algorithm gives
// with exact fractions, where the documented rule breaks every tie.
func TestGirvanNewman_SymmetricTies(t *testing.T) {
	// the 4-dimensional hypercube looks the same from every edge
	var hypercube [][2]int
	for i := 0; i < 16; i++ {
		for b := 0; b < 4; b++ {
			if j := i ^ (1 << b); j > i {
				hypercube = append(hypercube, [2]int{i, j})
			}
		}
	}

	var bipartite [][2]int
	for i := 0; i < 3; i++ {
		for j := 3; j < 6; j++ {
			bipartite = append(bipartite, [2]int{i, j})
		}
	}

	var cycle [][2]int
	for i := 0; i < 8; i++ {
		cycle = append(cycle, [2]int{i, (i + 1) % 8})
	}

	tests := []struct {
		name  string
		n     int
		edges [][2]int
		k     int
		want  [][]int
	}{
		{
			name:  "hypercube",
			n:     16,
			edges: hypercube,
			k:     2,
			want:  [][]int{{0, 2, 4, 8, 6, 10, 12, 14}, {1, 3, 5, 9, 7, 11, 13, 15}},
		},
		{
			name:  "complete bipartite, two communities",
			n:     6,
			edges: bipartite,
			k:     2,
			want:  [][]int{{0, 4}, {1, 3, 5, 2}},
		},
		{
			name:  "complete bipartite, three communities",
			n:     6,
			edges: bipartite,
			k:     3,
			want:  [][]int{{0, 4}, {1, 5}, {2, 3}},
		},
		{
			// 0-1 goes first, then the middle edge of the path that's left
			name:  "cycle",
			n:     8,
			edges: cycle,
			k:     2,
			want:  [][]int{{0, 7, 6, 5}, {1, 2, 3, 4}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			communities, err := GirvanNewman(graphFromEdges(tt.n, tt.edges), tt.k)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := intCommunities(communities); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GirvanNewman() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMostCentralEdge_NoValue(t *testing.T) {
	g := graphFromEdges(3, [][2]int{{0, 1}, {1, 2}})
	if e := mostCentralEdge(g, nil); e != nil {
		t.Fatalf("expected no edge without betweenness values, got %v-%v", e.Source().Label(), e.Destination().Label())
	}

	// a NaN is never the highest, so the loop in GirvanNewman ends
	nan := make(map[*gograph.Edge[int]]float64)
	for _, e := range g.AllEdges() {
		nan[e] = math.NaN()
	}
	if e := mostCentralEdge(g, nan); e != nil {
		t.Fatalf("expected no edge when every value is NaN, got %v-%v", e.Source().Label(), e.Destination().Label())
	}
}
