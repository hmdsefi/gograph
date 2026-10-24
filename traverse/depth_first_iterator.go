package traverse

import (
	"github.com/hmdsefi/gograph"
)

// depthFirstIterator  is an implementation of the Iterator interface
// for traversing a graph using a depth-first search (DFS) algorithm.
type depthFirstIterator[T comparable] struct {
	graph   gograph.Graph[T] // the graph being traversed.
	start   T                // the label of the starting vertex for the DFS traversal.
	stack   []T              // a slice that represents the stack of vertices to visit in DFS traversal order.
	visited map[T]bool       // a map that keeps track of whether a vertex has been visited or not.
}

// NewDepthFirstIterator creates a new instance of depthFirstIterator
// and returns it as the Iterator interface.
func NewDepthFirstIterator[T comparable](g gograph.Graph[T], start T) (Iterator[T], error) {
	v := g.GetVertexByID(start)
	if v == nil {
		return nil, gograph.ErrVertexDoesNotExist
	}

	return newDepthFirstIterator[T](g, start), nil
}

func newDepthFirstIterator[T comparable](g gograph.Graph[T], start T) *depthFirstIterator[T] {
	return &depthFirstIterator[T]{
		graph:   g,
		start:   start,
		stack:   []T{start},
		visited: map[T]bool{start: true},
	}
}

// HasNext returns a boolean indicating whether there are more vertices
// to be visited in the DFS traversal. It drops stacked vertices that were
// removed from the graph, and returns true if one is left.
func (d *depthFirstIterator[T]) HasNext() bool {
	for len(d.stack) > 0 {
		if d.graph.GetVertexByID(d.stack[len(d.stack)-1]) != nil {
			return true
		}
		d.stack = d.stack[:len(d.stack)-1]
	}

	return false
}

// Next returns the next vertex to be visited in the DFS traversal. It
// pops the latest vertex that has been added to the stack. A vertex
// removed from the graph after it was stacked is skipped.
// If the HasNext is false, returns nil.
func (d *depthFirstIterator[T]) Next() *gograph.Vertex[T] {
	for len(d.stack) > 0 {
		// get the next vertex from the stack
		label := d.stack[len(d.stack)-1]
		d.stack = d.stack[:len(d.stack)-1]
		currentNode := d.graph.GetVertexByID(label)
		if currentNode == nil {
			continue
		}

		// add unvisited neighbors to the stack
		neighbors := currentNode.Neighbors()
		for _, neighbor := range neighbors {
			if !d.visited[neighbor.Label()] {
				d.stack = append(d.stack, neighbor.Label())
				d.visited[neighbor.Label()] = true
			}
		}

		return currentNode
	}

	return nil
}

// Iterate iterates through all the vertices in the DFS traversal order
// and applies the given function to each vertex. If the function returns
// an error, the iteration stops and the error is returned.
func (d *depthFirstIterator[T]) Iterate(f func(v *gograph.Vertex[T]) error) error {
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
func (d *depthFirstIterator[T]) Reset() {
	d.stack = []T{d.start}
	d.visited = map[T]bool{d.start: true}
}
