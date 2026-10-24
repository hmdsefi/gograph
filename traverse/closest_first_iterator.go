package traverse

import (
	"container/heap"
	"math"

	"github.com/hmdsefi/gograph"
)

type closestFirstEntry[T comparable] struct {
	vertex   *gograph.Vertex[T]
	priority float64
	index    int
}

type closestFirstQueue[T comparable] struct {
	items   []*closestFirstEntry[T]
	byLabel map[T]*closestFirstEntry[T]
}

func newClosestFirstQueue[T comparable]() *closestFirstQueue[T] {
	return &closestFirstQueue[T]{byLabel: make(map[T]*closestFirstEntry[T])}
}

func (q closestFirstQueue[T]) Len() int { return len(q.items) }

func (q closestFirstQueue[T]) Less(i, j int) bool {
	return q.items[i].priority < q.items[j].priority
}

func (q closestFirstQueue[T]) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
	q.items[i].index = i
	q.items[j].index = j
}

func (q *closestFirstQueue[T]) Push(x any) {
	entry, ok := x.(*closestFirstEntry[T])
	if !ok || entry == nil {
		return
	}
	entry.index = len(q.items)
	q.items = append(q.items, entry)
}

func (q *closestFirstQueue[T]) Pop() any {
	last := len(q.items) - 1
	entry := q.items[last]
	q.items[last] = nil
	entry.index = -1
	q.items = q.items[:last]
	return entry
}

func (q *closestFirstQueue[T]) pushOrDecrease(vertex *gograph.Vertex[T], priority float64) {
	label := vertex.Label()
	if entry := q.byLabel[label]; entry != nil {
		// A different vertex with the same label replaced a removed one, so
		// the old distance doesn't apply to it.
		if entry.vertex != vertex || priority < entry.priority {
			entry.vertex = vertex
			entry.priority = priority
			heap.Fix(q, entry.index)
		}
		return
	}

	entry := &closestFirstEntry[T]{vertex: vertex, priority: priority}
	q.byLabel[label] = entry
	heap.Push(q, entry)
}

func (q *closestFirstQueue[T]) peek() *closestFirstEntry[T] {
	if q.Len() == 0 {
		return nil
	}
	return q.items[0]
}

func (q *closestFirstQueue[T]) pop() *closestFirstEntry[T] {
	if q.Len() == 0 {
		return nil
	}
	entry, ok := heap.Pop(q).(*closestFirstEntry[T])
	if !ok || entry == nil {
		return nil
	}
	delete(q.byLabel, entry.vertex.Label())
	return entry
}

// closestFirstIterator implements the Iterator interface to traverse
// a graph by increasing distance from the start vertex.
//
// Closest-first traversal visits the next reachable vertex with the
// smallest total distance from the start vertex. The distance is the
// sum of edge weights along the shortest path found so far. Edge
// weights are expected to be non-negative.
type closestFirstIterator[T comparable] struct {
	graph    gograph.Graph[T] // the graph that being traversed.
	start    T                // the label of starting point of the traversal.
	visited  map[T]bool       // a map that keeps track of whether a vertex has been visited or not.
	pq       *closestFirstQueue[T]
	currDist float64 // the current distance from the start node.
}

// NewClosestFirstIterator creates a new instance of closestFirstIterator
// and returns it as the Iterator interface.
//
// Edge weights are expected to be non-negative. An edge with a NaN weight is
// skipped. A vertex reached only through a +Inf weight is visited after every
// vertex at a finite distance.
//
// if the start node doesn't exist, returns error.
func NewClosestFirstIterator[T comparable](graph gograph.Graph[T], start T) (Iterator[T], error) {
	v := graph.GetVertexByID(start)
	if v == nil {
		return nil, gograph.ErrVertexDoesNotExist
	}

	pq := newClosestFirstQueue[T]()
	pq.pushOrDecrease(v, 0)
	return &closestFirstIterator[T]{
		graph:    graph,
		start:    start,
		visited:  make(map[T]bool),
		pq:       pq,
		currDist: 0,
	}, nil
}

// HasNext returns a boolean indicating whether there are more vertices
// to be visited or not. It drops queued vertices that were removed from
// the graph, and returns true if one is left.
func (c *closestFirstIterator[T]) HasNext() bool {
	for c.pq.Len() > 0 {
		if v := c.pq.peek().vertex; c.graph.GetVertexByID(v.Label()) == v {
			return true
		}

		c.pq.pop()
	}

	return false
}

// Next returns the next vertex to be visited in closest-first order.
// A vertex removed from the graph after it was queued is skipped.
//
// If the HasNext is false, returns nil.
func (c *closestFirstIterator[T]) Next() *gograph.Vertex[T] {
	for c.pq.Len() > 0 {
		vp := c.pq.pop()
		currNode := vp.vertex
		if c.graph.GetVertexByID(currNode.Label()) != currNode {
			continue
		}

		c.currDist = vp.priority
		c.visited[currNode.Label()] = true

		neighbors := currNode.Neighbors()
		for _, neighbor := range neighbors {
			if c.visited[neighbor.Label()] {
				continue
			}
			edge := c.graph.GetEdge(currNode, neighbor)
			if edge == nil {
				continue
			}
			// the queue can't order a NaN distance
			if dist := c.currDist + edge.Weight(); !math.IsNaN(dist) {
				c.pq.pushOrDecrease(c.graph.GetVertexByID(neighbor.Label()), dist)
			}
		}

		return currNode
	}

	return nil
}

// Iterate iterates through the vertices in closest-first order and applies
// the given function to each vertex. If the function returns an error,
// the iteration stops and the error is returned.
func (c *closestFirstIterator[T]) Iterate(f func(v *gograph.Vertex[T]) error) error {
	for v := c.Next(); v != nil; v = c.Next() {
		if err := f(v); err != nil {
			return err
		}
	}

	return nil
}

// Reset resets the iterator by setting the initial state of the iterator.
// If the start vertex was removed from the graph, there is nothing left to
// visit.
func (c *closestFirstIterator[T]) Reset() {
	c.visited = make(map[T]bool)
	c.currDist = 0

	c.pq = newClosestFirstQueue[T]()
	if v := c.graph.GetVertexByID(c.start); v != nil {
		c.pq.pushOrDecrease(v, 0)
	}
}
