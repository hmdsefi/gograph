package traverse

import "github.com/hmdsefi/gograph/graph"

type depthFirstIterator[T comparable] struct {
	graph   graph.Graph[T]
	start   T
	stack   []T
	visited map[T]bool
}

func NewDepthFirstIterator[T comparable](g graph.Graph[T], start T) Iterator[T] {
	return newDepthFirstIterator[T](g, start)
}

func newDepthFirstIterator[T comparable](g graph.Graph[T], start T) *depthFirstIterator[T] {
	return &depthFirstIterator[T]{
		graph:   g,
		start:   start,
		stack:   []T{start},
		visited: map[T]bool{start: true},
	}
}

func (d *depthFirstIterator[T]) HasNext() bool {
	return len(d.stack) > 0
}

func (d *depthFirstIterator[T]) Next() *graph.Vertex[T] {
	// get the next vertex from the queue
	label := d.stack[len(d.stack)-1]
	d.stack = d.stack[:len(d.stack)-1]
	currentNode := d.graph.GetVertexByID(label)

	// add unvisited neighbors to the queue
	neighbors := currentNode.Neighbors()
	for _, neighbor := range neighbors {
		if !d.visited[neighbor.Label()] {
			d.stack = append(d.stack, neighbor.Label())
			d.visited[neighbor.Label()] = true
		}
	}

	return currentNode
}

func (d *depthFirstIterator[T]) Iterate(f func(v *graph.Vertex[T]) error) error {
	for d.HasNext() {
		if err := f(d.Next()); err != nil {
			return err
		}
	}

	return nil
}

func (d *depthFirstIterator[T]) Reset() {
	d.stack = []T{d.start}
	d.visited = map[T]bool{d.start: true}
}
