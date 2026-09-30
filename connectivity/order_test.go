package connectivity

import (
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

func sccLabels(sccs [][]*gograph.Vertex[int]) [][]int {
	labels := make([][]int, len(sccs))
	for i, scc := range sccs {
		labels[i] = make([]int, len(scc))
		for j, v := range scc {
			labels[i][j] = v.Label()
		}
	}
	return labels
}

// orderTestGraph has cycles, single vertices and edges between components,
// so there are many valid orders of the components.
func orderTestGraph() gograph.Graph[int] {
	g := gograph.New[int](gograph.Directed())
	for i := range 30 {
		g.AddVertexByLabel(i)
	}
	edges := [][2]int{
		{0, 1}, {1, 2}, {2, 0},
		{3, 4}, {4, 3},
		{6, 7}, {7, 8}, {8, 9}, {9, 6},
		{10, 11}, {11, 12}, {12, 13},
		{2, 3}, {9, 10}, {13, 6}, {20, 21}, {21, 20}, {25, 0},
	}
	for _, e := range edges {
		_, _ = g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1]))
	}
	return g
}

func TestSCCs_StableOrder(t *testing.T) {
	algorithms := []struct {
		name string
		run  func(gograph.Graph[int]) [][]*gograph.Vertex[int]
	}{
		{"Tarjan", Tarjan[int]},
		{"Gabow", Gabow[int]},
		{"Kosaraju", Kosaraju[int]},
	}

	for _, alg := range algorithms {
		t.Run(alg.name, func(t *testing.T) {
			want := sccLabels(alg.run(orderTestGraph()))
			for range 50 {
				if got := sccLabels(alg.run(orderTestGraph())); !reflect.DeepEqual(got, want) {
					t.Fatalf("%s() = %v, want %v", alg.name, got, want)
				}
			}
		})
	}
}
