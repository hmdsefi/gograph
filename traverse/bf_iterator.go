package traverse

import (
	"github.com/hmdsefi/gograph"
)

// breadthFirstIterator is an implementation of the Iterator interface
// for traversing a graph using a breadth-first search (BFS) algorithm.
type breadthFirstIterator[T comparable] struct {
	graph   gograph.Graph[T] // the graph being traversed.
	start   T                // the label of the starting vertex for the BFS traversal.
	queue   []T              // a slice that represents the queue of vertices to visit in BFS traversal order.
	visited map[T]bool       // a map that keeps track of whether a vertex has been visited or not.
	head    int              // the current head of the queue.
}

// NewBreadthFirstIterator creates a new instance of breadthFirstIterator
// and returns it as the Iterator interface.
func NewBreadthFirstIterator[T comparable](g gograph.Graph[T], start T) (Iterator[T], error) {
	v := g.GetVertexByID(start)
	if v == nil {
		return nil, gograph.ErrVertexDoesNotExist
	}

	return newBreadthFirstIterator[T](g, start), nil
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

// HasNext returns a boolean indicating whether there are more vertices
// to be visited in the BFS traversal. It returns true if the head index
// is in the range of the queue indices.
func (d *breadthFirstIterator[T]) HasNext() bool {
	return d.head < len(d.queue)-1
}

// Next returns the next vertex to be visited in the BFS traversal.
// It dequeues the next vertex from the queue and updates the head field.
// If the HasNext is false, returns nil.
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

// Iterate iterates through all the vertices in the BFS traversal order
// and applies the given function to each vertex. If the function returns
// an error, the iteration stops and the error is returned.
func (d *breadthFirstIterator[T]) Iterate(f func(v *gograph.Vertex[T]) error) error {
	for d.HasNext() {
		if err := f(d.Next()); err != nil {
			return err
		}
	}

	return nil
}

// Reset resets the iterator by setting the initial state of the iterator.
func (d *breadthFirstIterator[T]) Reset() {
	d.queue = []T{d.start}
	d.head = -1
	d.visited = map[T]bool{d.start: true}
}
