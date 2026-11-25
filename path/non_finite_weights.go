package path

import (
	"math"

	"github.com/hmdsefi/gograph"
)

// arc is an edge between two vertex indexes.
type arc struct {
	from, to int
	weight   float64
}

// indexEdges returns the vertices of g, the index of each label in that
// slice, and the edges of g in the order of AllEdges as arcs. An edge with a
// NaN or +Inf weight never shortens a path, so it's left out. A -Inf weight
// is read as 0, and its edge is also listed in negInf. It takes O(V+E) time
// and space.
func indexEdges[T comparable](g gograph.Graph[T]) (vertices []*gograph.Vertex[T], index map[T]int, arcs, negInf []arc) {
	vertices = g.GetAllVertices()
	index = make(map[T]int, len(vertices))
	for i, v := range vertices {
		index[v.Label()] = i
	}

	arcs = make([]arc, 0, g.Size())
	for e := range gograph.Edges(g) {
		w := e.Weight()
		if math.IsNaN(w) || math.IsInf(w, 1) {
			continue
		}

		a := arc{from: index[e.Source().Label()], to: index[e.Destination().Label()], weight: w}
		if math.IsInf(w, -1) {
			a.weight = 0
			negInf = append(negInf, a)
		}
		arcs = append(arcs, a)
	}

	return vertices, index, arcs, negInf
}

func hasNegativeInfinity(dist []float64) bool {
	for _, d := range dist {
		if math.IsInf(d, -1) {
			return true
		}
	}
	return false
}
