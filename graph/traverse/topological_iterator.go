package traverse

import "github.com/hmdsefi/gograph/graph"

type topologicalIterator[T comparable] struct {
	graph graph.Graph[T]
	queue []*graph.Vertex[T]
	head  int
}

func NewTopologicalIterator[T comparable](g graph.Graph[T]) (Iterator[T], error) {
	return newTopologicalIterator[T](g)
}

func newTopologicalIterator[T comparable](g graph.Graph[T]) (*topologicalIterator[T], error) {
	queue, err := graph.TopologySort[T](g)
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

func (t *topologicalIterator[T]) Next() *graph.Vertex[T] {
	t.head = t.head + 1
	return t.queue[t.head]
}

func (t *topologicalIterator[T]) Iterate(f func(v *graph.Vertex[T]) error) error {
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
