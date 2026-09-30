package gograph

import "container/heap"

// TopologySort performs a topological sort of the graph using
// Kahn's algorithm. If the sorted list of vertices does not contain
// all vertices in the graph, it means there is a cycle in the graph.
//
// The order is stable: it only depends on the order of GetAllVertices and
// the order the edges were added, so a graph built the same way gives the
// same result on every run. Use StableTopologySort to choose the order of
// the vertices that the edges don't order.
//
// It returns error if it finds a cycle in the graph.
func TopologySort[T comparable](g Graph[T]) ([]*Vertex[T], error) {
	// Initialize a map to store the inDegree of each vertex
	inDegrees := make(map[*Vertex[T]]int)
	vertices := g.GetAllVertices()
	for _, v := range vertices {
		inDegrees[v] = v.inDegree
	}

	// Initialize a queue with vertices of inDegrees zero
	queue := make([]*Vertex[T], 0)
	for _, v := range vertices {
		if inDegrees[v] == 0 {
			queue = append(queue, v)
		}
	}

	// Initialize the sorted list of vertices
	sortedVertices := make([]*Vertex[T], 0)

	// Loop through the vertices with inDegree zero
	for len(queue) > 0 {
		// Get the next vertex with inDegree zero
		curr := queue[0]
		queue = queue[1:]

		// Add the vertex to the sorted list
		sortedVertices = append(sortedVertices, curr)

		// Decrement the inDegree of each of the vertex's neighbors
		for _, neighbor := range curr.neighbors {
			inDegrees[neighbor]--
			if inDegrees[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If the sorted list does not contain all vertices, there is a cycle
	if len(sortedVertices) != len(vertices) {
		return nil, ErrDAGHasCycle
	}

	return sortedVertices, nil
}

// StableTopologySort performs a topological sort of the graph that picks,
// at every step, the smallest vertex with no incoming edges left, using
// compare on the labels. compare returns a negative number if a < b, zero
// if a == b and a positive number if a > b, so cmp.Compare works for
// string and number labels.
//
// If compare only returns zero for equal labels, the result doesn't depend
// on the order the vertices and edges were added. As in TopologySort, an
// edge from A to B puts A before B.
//
// It returns ErrDAGHasCycle if the graph has a cycle. It runs in
// O((V + E) log V) time.
func StableTopologySort[T comparable](g Graph[T], compare func(a, b T) int) ([]*Vertex[T], error) {
	vertices := g.GetAllVertices()
	inDegrees := make(map[*Vertex[T]]int, len(vertices))
	ready := &vertexHeap[T]{compare: compare}
	for _, v := range vertices {
		inDegrees[v] = v.inDegree
		if v.inDegree == 0 {
			ready.vertices = append(ready.vertices, v)
		}
	}
	heap.Init(ready)

	sorted := make([]*Vertex[T], 0, len(vertices))
	for ready.Len() > 0 {
		curr := ready.pop()
		sorted = append(sorted, curr)

		for _, neighbor := range curr.neighbors {
			inDegrees[neighbor]--
			if inDegrees[neighbor] == 0 {
				heap.Push(ready, neighbor)
			}
		}
	}

	if len(sorted) != len(vertices) {
		return nil, ErrDAGHasCycle
	}

	return sorted, nil
}

// vertexHeap is a min-heap of vertices, ordered by compare on their labels.
type vertexHeap[T comparable] struct {
	vertices []*Vertex[T]
	compare  func(a, b T) int
}

func (h *vertexHeap[T]) Len() int { return len(h.vertices) }

func (h *vertexHeap[T]) Less(i, j int) bool {
	return h.compare(h.vertices[i].label, h.vertices[j].label) < 0
}

func (h *vertexHeap[T]) Swap(i, j int) { h.vertices[i], h.vertices[j] = h.vertices[j], h.vertices[i] }

func (h *vertexHeap[T]) Push(x any) {
	if v, ok := x.(*Vertex[T]); ok {
		h.vertices = append(h.vertices, v)
	}
}

func (h *vertexHeap[T]) Pop() any {
	last := len(h.vertices) - 1
	v := h.vertices[last]
	h.vertices[last] = nil
	h.vertices = h.vertices[:last]
	return v
}

func (h *vertexHeap[T]) pop() *Vertex[T] {
	v, _ := heap.Pop(h).(*Vertex[T])
	return v
}
