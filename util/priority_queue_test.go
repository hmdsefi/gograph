package util

import (
	"container/heap"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestVertexPriorityQueue(t *testing.T) {
	// Create a new priority queue
	pq := make(VertexPriorityQueue[string], 0)

	// Push some items with different priorities to the queue
	heap.Push(&pq, NewVertexWithPriority(gograph.NewVertex("A"), 3))
	heap.Push(&pq, NewVertexWithPriority(gograph.NewVertex("B"), 1))
	heap.Push(&pq, NewVertexWithPriority(gograph.NewVertex("C"), 5))
	heap.Push(&pq, NewVertexWithPriority(gograph.NewVertex("D"), 2))
	heap.Push(&pq, NewVertexWithPriority(gograph.NewVertex("E"), 4))

	// Check that the length of the priority queue is 5
	if len(pq) != 5 {
		t.Errorf("PriorityQueue length = %d; want 5", len(pq))
	}

	// Pop items from the queue and check that they are in the correct order
	items := make([]string, 0)
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*VertexWithPriority[string])
		items = append(items, item.vertex.Label())
	}
	expected := []string{"B", "D", "A", "E", "C"}
	if !reflect.DeepEqual(items, expected) {
		t.Errorf("PriorityQueue Pop() order = %v; want %v", items, expected)
	}
}
