package util

import (
	"github.com/hmdsefi/gograph"
)

// VertexPriorityQueue is a priority queue that implements
// heap and sort interfaces. It represents a min heap.
type VertexPriorityQueue[T comparable] []*VertexWithPriority[T]

// VertexWithPriority is a vertex priority queue item that stores
// vertex along with its priority.
type VertexWithPriority[T comparable] struct {
	vertex   *gograph.Vertex[T]
	priority float64
}

func NewVertexWithPriority[T comparable](vertex *gograph.Vertex[T], priority float64) *VertexWithPriority[T] {
	return &VertexWithPriority[T]{vertex: vertex, priority: priority}
}

// Len is the number of elements in the collection.
func (pq VertexPriorityQueue[T]) Len() int { return len(pq) }

// Less reports whether the element with index i
// must sort before the element with index j.
func (pq VertexPriorityQueue[T]) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

// Swap swaps the elements with indexes i and j.
func (pq VertexPriorityQueue[T]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

// Push adds new item to the collection.
func (pq *VertexPriorityQueue[T]) Push(x interface{}) {
	item, ok := x.(*VertexWithPriority[T])
	if !ok {
		return
	}

	*pq = append(*pq, item)
}

// Pop removes and returns the minimum element (according to Less) from the heap.
// The complexity is O(log n) where n = h.Len().
// Pop is equivalent to Remove(h, 0).
func (pq *VertexPriorityQueue[T]) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}
