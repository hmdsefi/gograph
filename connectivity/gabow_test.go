package connectivity

import (
	"github.com/hmdsefi/gograph"
	"reflect"
	"sort"
	"testing"
)

func TestGabow(t *testing.T) {
	// Create a dag with 6 vertices and 6 edges
	g := gograph.New[int](gograph.Directed())

	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)
	v5 := g.AddVertexByLabel(5)

	_, _ = g.AddEdge(v1, v2)
	_, _ = g.AddEdge(v2, v3)
	_, _ = g.AddEdge(v3, v1)
	_, _ = g.AddEdge(v3, v4)
	_, _ = g.AddEdge(v4, v5)
	_, _ = g.AddEdge(v5, v4)

	// call the tarjan function
	sccs := gabow(g)

	// check that the function returned the expected number of SCCs
	if len(sccs) != 2 {
		t.Errorf("Expected 2 SCCs, got %d", len(sccs))
	}

	// check that the function returned the correct SCCs
	expectedSCCs := map[int][]int{3: {1, 2, 3}, 2: {4, 5}}

	for i, scc := range sccs {
		var labels []int
		for _, items := range scc {
			labels = append(labels, items.Label())
		}

		sort.Ints(labels)

		if !reflect.DeepEqual(expectedSCCs[len(scc)], labels) {
			t.Errorf("SCC %d: expected %v, got %v", i+1, expectedSCCs[i], scc)
		}
	}
}
