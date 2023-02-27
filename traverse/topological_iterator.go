package traverse

import (
	"github.com/hmdsefi/gograph"
)

type topologicalIterator[T comparable] struct {
	graph gograph.Graph[T]
	queue []*gograph.Vertex[T]
	head  int
}

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

func (t *topologicalIterator[T]) HasNext() bool {
	return t.head < len(t.queue)-1
}

func (t *topologicalIterator[T]) Next() *gograph.Vertex[T] {
	t.head = t.head + 1
	return t.queue[t.head]
}

func (t *topologicalIterator[T]) Iterate(f func(v *gograph.Vertex[T]) error) error {
	for t.HasNext() {
		if err := f(t.Next()); err != nil {
			return err
		}
	}

	return nil
}

func (t *topologicalIterator[T]) Reset() {
	t.head = -1
}
