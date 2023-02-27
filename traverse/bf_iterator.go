package traverse

import (
	"github.com/hmdsefi/gograph"
)

type breadthFirstIterator[T comparable] struct {
	graph   gograph.Graph[T]
	start   T
	queue   []T
	visited map[T]bool
	head    int
}

func NewBreadthFirstIterator[T comparable](g gograph.Graph[T], start T) Iterator[T] {
	return newBreadthFirstIterator[T](g, start)
}

func newBreadthFirstIterator[T comparable](g gograph.Graph[T], start T) *breadthFirstIterator[T] {
	return &breadthFirstIterator[T]{
		graph:   g,
		start:   start,
		queue:   []T{start},
		visited: map[T]bool{start: true},
		head:    -1,
	}
}

func (d *breadthFirstIterator[T]) HasNext() bool {
	return d.head < len(d.queue)-1
}

func (d *breadthFirstIterator[T]) Next() *gograph.Vertex[T] {
	if !d.HasNext() {
		return nil
	}

	d.head++

	// get the next vertex from the queue
	currentNode := d.graph.GetVertexByID(d.queue[d.head])

	// add unvisited neighbors to the queue
	neighbors := currentNode.Neighbors()
	for _, neighbor := range neighbors {
		if !d.visited[neighbor.Label()] {
			d.visited[neighbor.Label()] = true
			d.queue = append(d.queue, neighbor.Label())
		}
	}

	return currentNode
}

func (d *breadthFirstIterator[T]) Iterate(f func(v *gograph.Vertex[T]) error) error {
	for d.HasNext() {
		if err := f(d.Next()); err != nil {
			return err
		}
	}

	return nil
}

func (d *breadthFirstIterator[T]) Reset() {
	d.queue = []T{d.start}
	d.head = -1
	d.visited = map[T]bool{d.start: true}
}
