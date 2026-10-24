package traverse

import (
	"github.com/hmdsefi/gograph"
)

// breadthFirstIterator is an implementation of the Iterator interface
// for traversing a graph using a breadth-first search (BFS) algorithm.
type breadthFirstIterator[T comparable] struct {
	graph   gograph.Graph[T]         // the graph being traversed.
	start   T                        // the label of the starting vertex for the BFS traversal.
	queue   []*gograph.Vertex[T]     // a slice that represents the queue of vertices to visit in BFS traversal order.
	visited map[T]*gograph.Vertex[T] // the vertex queued for each label.
	head    int                      // the current head of the queue.
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
	d := &breadthFirstIterator[T]{
		graph: g,
		start: start,
	}
	d.Reset()

	return d
}

// HasNext returns a boolean indicating whether there are more vertices
// to be visited in the BFS traversal. It drops queued vertices that were
// removed from the graph, and returns true if one is left.
func (d *breadthFirstIterator[T]) HasNext() bool {
	for d.head < len(d.queue)-1 {
		if v := d.queue[d.head+1]; hasEdges(v) || stored(d.graph, v) {
			return true
		}
		d.head++
	}

	return false
}

// Next returns the next vertex to be visited in the BFS traversal.
// It dequeues the next vertex from the queue and updates the head field.
// A vertex removed from the graph after it was queued is skipped, and a
// vertex added later with the same label is visited if the traversal
// reaches it.
// If the HasNext is false, returns nil.
func (d *breadthFirstIterator[T]) Next() *gograph.Vertex[T] {
	for d.head < len(d.queue)-1 {
		d.head++

		// get the next vertex from the queue
		currentNode := d.queue[d.head]
		if !hasEdges(currentNode) && !stored(d.graph, currentNode) {
			continue
		}

		// add unvisited neighbors to the queue
		neighbors := currentNode.Neighbors()
		for _, neighbor := range neighbors {
			label := neighbor.Label()
			prev := d.visited[label]
			if prev != nil && prev.InDegree() > 0 {
				continue
			}
			if v := reached(d.graph, prev, label); v != nil {
				d.visited[label] = v
				d.queue = append(d.queue, v)
			}
		}

		return currentNode
	}

	return nil
}

// Iterate iterates through all the vertices in the BFS traversal order
// and applies the given function to each vertex. If the function returns
// an error, the iteration stops and the error is returned.
func (d *breadthFirstIterator[T]) Iterate(f func(v *gograph.Vertex[T]) error) error {
	for v := d.Next(); v != nil; v = d.Next() {
		if err := f(v); err != nil {
			return err
		}
	}

	return nil
}

// Reset resets the iterator by setting the initial state of the iterator.
// If the start vertex was removed from the graph, there is nothing left to
// visit.
func (d *breadthFirstIterator[T]) Reset() {
	d.queue = nil
	d.head = -1
	d.visited = make(map[T]*gograph.Vertex[T])

	if v := d.graph.GetVertexByID(d.start); v != nil {
		d.queue = append(d.queue, v)
		d.visited[d.start] = v
	}
}
