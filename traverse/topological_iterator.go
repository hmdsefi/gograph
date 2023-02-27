package traverse

import (
	"github.com/hmdsefi/gograph"
)

// topologicalIterator  is an implementation of the Iterator interface
// for traversing a graph using a topological sort algorithm.
type topologicalIterator[T comparable] struct {
	graph gograph.Graph[T]     // the graph being traversed.
	queue []*gograph.Vertex[T] // a slice that represents the queue of vertices to visit in topological order.
	head  int                  // the current head of the queue.
}

// NewTopologicalIterator creates a new instance of topologicalIterator
// and returns it as the Iterator interface.
func NewTopologicalIterator[T comparable](g gograph.Graph[T]) (Iterator[T], error) {
	return newTopologicalIterator[T](g)
}

func newTopologicalIterator[T comparable](g gograph.Graph[T]) (*topologicalIterator[T], error) {
	queue, err := gograph.TopologySort[T](g)
	if err != nil {
		return nil, err
	}

	return &topologicalIterator[T]{
		graph: g,
		queue: queue,
		head:  -1,
	}, nil
}

// HasNext returns a boolean indicating whether there are more vertices
// to be visited in the topological traversal. It returns true if the
// head index is in the range of the queue indices.
func (t *topologicalIterator[T]) HasNext() bool {
	return t.head < len(t.queue)-1
}

// Next returns the next vertex to be visited in the topological order.
// If the HasNext is false, returns nil.
func (t *topologicalIterator[T]) Next() *gograph.Vertex[T] {
	if !t.HasNext() {
		return nil
	}

	t.head = t.head + 1
	return t.queue[t.head]
}

// Iterate iterates through all the vertices in the BFS traversal order
// and applies the given function to each vertex. If the function returns
// an error, the iteration stops and the error is returned.
func (t *topologicalIterator[T]) Iterate(f func(v *gograph.Vertex[T]) error) error {
	for t.HasNext() {
		if err := f(t.Next()); err != nil {
			return err
		}
	}

	return nil
}

// Reset resets the iterator by setting the initial state of the iterator.
// It calls the gograph.TopologySort again. If topology sort returns
// error, it panics.
func (t *topologicalIterator[T]) Reset() {
	t.head = -1

	var err error
	t.queue, err = gograph.TopologySort[T](t.graph)
	if err != nil {
		panic(err)
	}
}
