package traverse

import (
	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/util"
)

// closestFirstIterator implements the Iterator interface to traverse
// a graph by increasing distance from the start vertex.
//
// Closest-first traversal visits the next reachable vertex with the
// smallest total distance from the start vertex. The distance is the
// sum of edge weights along the shortest path found so far. Edge
// weights are expected to be non-negative.
type closestFirstIterator[T comparable] struct {
	graph    gograph.Graph[T]             // the graph that being traversed.
	start    T                            // the label of starting point of the traversal.
	visited  map[T]bool                   // a map that keeps track of whether a vertex has been visited or not.
	pq       *util.VertexPriorityQueue[T] // a slice of util.VertexWithPriority that represents a min heap.
	currDist float64                      // the current distance from the start node.
}

// NewClosestFirstIterator creates a new instance of closestFirstIterator
// and returns it as the Iterator interface.
//
// if the start node doesn't exist, returns error.
func NewClosestFirstIterator[T comparable](graph gograph.Graph[T], start T) (Iterator[T], error) {
	v := graph.GetVertexByID(start)
	if v == nil {
		return nil, gograph.ErrVertexDoesNotExist
	}

	pq := util.NewVertexPriorityQueue[T]()
	pq.Push(util.NewVertexWithPriority[T](v, 0))
	return &closestFirstIterator[T]{
		graph:    graph,
		start:    start,
		visited:  make(map[T]bool),
		pq:       pq,
		currDist: 0,
	}, nil
}

// HasNext returns a boolean indicating whether there are more vertices
// to be visited or not. Queued vertices that have been removed from the
// graph are skipped.
func (c *closestFirstIterator[T]) HasNext() bool {
	for c.pq.Len() > 0 {
		v := c.pq.Peek().Vertex()
		if !c.visited[v.Label()] && c.graph.GetVertexByID(v.Label()) == v {
			return true
		}

		c.pq.Pop()
	}

	return false
}

// Next returns the next vertex to be visited in closest-first order.
//
// If the HasNext is false, returns nil.
func (c *closestFirstIterator[T]) Next() *gograph.Vertex[T] {
	if !c.HasNext() {
		return nil
	}

	vp := c.pq.Pop()
	c.currDist = vp.Priority()
	currNode := vp.Vertex()
	c.visited[currNode.Label()] = true

	neighbors := currNode.Neighbors()
	for _, neighbor := range neighbors {
		edge := c.graph.GetEdge(currNode, neighbor)
		if !c.visited[neighbor.Label()] {
			dist := c.currDist + edge.Weight()
			c.pq.Push(util.NewVertexWithPriority(c.graph.GetVertexByID(neighbor.Label()), dist))
		}
	}

	return currNode
}

// Iterate iterates through the vertices in closest-first order and applies
// the given function to each vertex. If the function returns an error,
// the iteration stops and the error is returned.
func (c *closestFirstIterator[T]) Iterate(f func(v *gograph.Vertex[T]) error) error {
	for c.HasNext() {
		if err := f(c.Next()); err != nil {
			return err
		}
	}

	return nil
}

// Reset resets the iterator by setting the initial state of the iterator.
// If the start vertex has been removed from the graph, the iterator has
// nothing left to visit.
func (c *closestFirstIterator[T]) Reset() {
	c.visited = make(map[T]bool)
	c.currDist = 0

	c.pq = util.NewVertexPriorityQueue[T]()
	if v := c.graph.GetVertexByID(c.start); v != nil {
		c.pq.Push(util.NewVertexWithPriority(v, 0))
	}
}
