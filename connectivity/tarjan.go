package connectivity

import "github.com/hmdsefi/gograph"

type tarjanVertex[T comparable] struct {
	*gograph.Vertex[T]
	index   int
	lowLink int
	onStack bool
}

func newTarjanVertex[T comparable](vertex *gograph.Vertex[T]) *tarjanVertex[T] {
	return &tarjanVertex[T]{
		Vertex: vertex,
		index:  -1,
	}
}

type tarjanSCCS[T comparable] struct {
	vertices map[T]*tarjanVertex[T]
}

func newTarjanSCCS[T comparable](vertices map[T]*tarjanVertex[T]) *tarjanSCCS[T] {
	return &tarjanSCCS[T]{vertices: vertices}
}

func tarjan[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T] {
	var (
		index     int
		stack     []*tarjanVertex[T]
		tvertices = make(map[T]*tarjanVertex[T])
		sccs      [][]*tarjanVertex[T]
	)

	vertices := g.GetAllVertices()

	for _, v := range vertices {
		tv := newTarjanVertex(v)
		tvertices[tv.Label()] = tv
	}

	tarj := newTarjanSCCS(tvertices)

	for _, v := range tvertices {
		if v.index < 0 {
			tarj.visit(v, &index, &stack, &sccs)
		}
	}

	result := make([][]*gograph.Vertex[T], len(sccs))
	for i, list := range sccs {
		result[i] = make([]*gograph.Vertex[T], len(list))
		for j := range list {
			result[i][j] = list[j].Vertex
		}

	}
	return result
}

func (t *tarjanSCCS[T]) visit(
	v *tarjanVertex[T],
	index *int,
	stack *[]*tarjanVertex[T],
	sccs *[][]*tarjanVertex[T],
) {
	v.index = *index
	v.lowLink = *index
	*index++
	*stack = append(*stack, v)
	v.onStack = true

	neighbors := v.Neighbors()
	for _, w := range neighbors {
		tv := t.vertices[w.Label()]
		if tv.index == -1 {
			t.visit(tv, index, stack, sccs)
			v.lowLink = min(v.lowLink, tv.lowLink)
		} else if tv.onStack {
			v.lowLink = min(v.lowLink, tv.index)
		}
	}

	if v.lowLink == v.index {
		var scc []*tarjanVertex[T]
		for {
			w := (*stack)[len(*stack)-1]
			*stack = (*stack)[:len(*stack)-1]
			w.onStack = false
			scc = append(scc, w)
			if w == v {
				break
			}
		}
		*sccs = append(*sccs, scc)
	}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
