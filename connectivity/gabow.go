package connectivity

import "github.com/hmdsefi/gograph"

func gabow[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T] {
	var (
		index       int
		components  [][]*gograph.Vertex[T]
		stack       []*tarjanVertex[T]
		strongLinks func(v *tarjanVertex[T])
	)

	graphVertices := g.GetAllVertices()
	vertices := make(map[T]*tarjanVertex[T])
	for _, v := range graphVertices {
		vertices[v.Label()] = newTarjanVertex(v)
	}

	strongLinks = func(v *tarjanVertex[T]) {
		v.index = index
		v.lowLink = index
		index++
		stack = append(stack, v)
		v.onStack = true

		neighbors := v.Neighbors()
		for _, neighbor := range neighbors {
			w := vertices[neighbor.Label()]
			if w.index == -1 {
				strongLinks(w)
				if w.lowLink < v.lowLink {
					v.lowLink = w.lowLink
				}
			} else if w.onStack {
				if w.index < v.lowLink {
					v.lowLink = w.index
				}
			}
		}

		if v.lowLink == v.index {
			var (
				component []*tarjanVertex[T]
				w         *tarjanVertex[T]
			)
			for {
				w, stack = stack[len(stack)-1], stack[:len(stack)-1]
				w.onStack = false
				component = append(component, w)
				if w == v {
					break
				}
			}

			var temp []*gograph.Vertex[T]
			for i := range component {
				temp = append(temp, component[i].Vertex)
			}
			components = append(components, temp)
		}
	}

	for _, v := range vertices {
		v.index = -1
	}

	for _, v := range vertices {
		if v.index == -1 {
			strongLinks(v)
		}
	}

	return components
}
