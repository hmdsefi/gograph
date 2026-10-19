package gograph

import (
	"fmt"
	"sync/atomic"
)

// InducedSubgraph returns a new graph with the given vertices and every edge
// of g whose endpoints are both in that set.
//
// The result has the same directed, weighted, and acyclic properties as g.
// Vertex weights, edge weights, edge labels, and metadata are copied. The
// result does not share vertices or edges with g. Duplicate labels are
// ignored, and vertex order follows the first occurrence of each label.
// No labels returns an empty graph with the same properties.
//
// A label that is not in g returns ErrVertexDoesNotExist wrapped with that
// label, and the returned graph is nil. A nil graph, or a graph that was
// not created by New, panics.
//
// Time is O(k + D), where k is the number of distinct labels and D is the
// sum of the out-degrees of those vertices. Extra space beyond the result
// is O(k).
func InducedSubgraph[T comparable](g Graph[T], labels ...T) (Graph[T], error) {
	bg, ok := g.(*baseGraph[T])
	if !ok || bg == nil {
		panic("gograph: InducedSubgraph requires a graph from New")
	}

	sub, err := inducedSubgraph(bg, labels)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func inducedSubgraph[T comparable](g *baseGraph[T], labels []T) (*baseGraph[T], error) {
	if len(labels) == 0 {
		return newBaseGraph[T](g.properties), nil
	}

	selected, index, err := inducedVertices(g, labels)
	if err != nil {
		return nil, err
	}

	out := &baseGraph[T]{
		vertices:   make(map[T]*Vertex[T], len(selected)),
		properties: g.properties,
	}
	copies := copyInducedVertices(out, selected)
	degree := inducedOutDegree(selected, index)
	allocInducedEdges(out, copies, degree)
	atomic.StoreUint32(&out.edgesCount, copyInducedEdges(g, out, selected, copies, index))

	return out, nil
}

func inducedVertices[T comparable](g *baseGraph[T], labels []T) ([]*Vertex[T], map[T]int, error) {
	index := make(map[T]int, len(labels))
	selected := make([]*Vertex[T], 0, len(labels))
	for _, label := range labels {
		if _, ok := index[label]; ok {
			continue
		}

		v := g.vertices[label]
		if v == nil {
			return nil, nil, fmt.Errorf("%w: %v", ErrVertexDoesNotExist, label)
		}

		index[label] = len(selected)
		selected = append(selected, v)
	}

	return selected, index, nil
}

func copyInducedVertices[T comparable](out *baseGraph[T], selected []*Vertex[T]) []*Vertex[T] {
	copies := make([]*Vertex[T], len(selected))
	for i, src := range selected {
		v := &Vertex[T]{
			label:      src.label,
			properties: src.properties,
			metadata:   src.metadata,
			stored:     true,
			position:   i,
		}
		copies[i] = v
		out.vertices[src.label] = v
		atomic.AddUint32(&out.verticesCount, 1)
	}

	out.order = copies

	return copies
}

// inducedOutDegree counts the kept outgoing edges of each selected vertex.
// A self-loop stored twice in the neighbor list counts once.
func inducedOutDegree[T comparable](selected []*Vertex[T], index map[T]int) []int {
	degree := make([]int, len(selected))
	for i, src := range selected {
		looped := false
		for _, nb := range src.neighbors {
			j, ok := index[nb.label]
			if !ok {
				continue
			}
			if i == j {
				if looped {
					continue
				}
				looped = true
			}
			degree[i]++
		}
	}

	return degree
}

func allocInducedEdges[T comparable](out *baseGraph[T], copies []*Vertex[T], degree []int) {
	sources := 0
	for _, n := range degree {
		if n > 0 {
			sources++
		}
	}

	out.edges = make(map[T]map[T]*Edge[T], sources)
	for i, v := range copies {
		if degree[i] == 0 {
			continue
		}
		v.neighbors = make([]*Vertex[T], 0, degree[i])
		out.edges[v.label] = make(map[T]*Edge[T], degree[i])
	}
}

func copyInducedEdges[T comparable](
	g, out *baseGraph[T],
	selected, copies []*Vertex[T],
	index map[T]int,
) uint32 {
	var n uint32
	for i, src := range selected {
		from := copies[i]
		for _, nb := range src.neighbors {
			j, ok := index[nb.label]
			if !ok {
				continue
			}
			to := copies[j]
			if _, exists := out.edges[from.label][to.label]; exists {
				continue
			}
			putInducedEdge(out, from, to, storedInducedEdge(g, src.label, nb.label))
			n++
		}
	}

	return n
}

func storedInducedEdge[T comparable](g *baseGraph[T], from, to T) *Edge[T] {
	dests := g.edges[from]
	if dests == nil {
		return nil
	}

	return dests[to]
}

func putInducedEdge[T comparable](g *baseGraph[T], from, to *Vertex[T], src *Edge[T]) {
	edge := &Edge[T]{
		source: from,
		dest:   to,
	}
	if src != nil {
		edge.properties = src.properties
		edge.metadata = src.metadata
	}

	from.neighbors = append(from.neighbors, to)
	to.inDegree++
	g.edges[from.label][to.label] = edge
}
