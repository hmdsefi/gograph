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

// closestFirstQueue is a heap of entries ordered by priority. An entry stays
// in byVertex once it's popped, with an index of -1, so byVertex also tells
// which vertices were visited.
type closestFirstQueue[T comparable] struct {
	items    []*closestFirstEntry[T]
	byVertex map[*gograph.Vertex[T]]*closestFirstEntry[T]
}

func newClosestFirstQueue[T comparable]() *closestFirstQueue[T] {
	return &closestFirstQueue[T]{byVertex: make(map[*gograph.Vertex[T]]*closestFirstEntry[T])}
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

// push queues vertex at priority. entry is the one byVertex has for vertex,
// or nil. A queued entry only moves to a lower priority, and a popped one
// stays visited.
func (q *closestFirstQueue[T]) push(entry *closestFirstEntry[T], vertex *gograph.Vertex[T], priority float64) {
	switch {
	case entry == nil:
		entry = &closestFirstEntry[T]{vertex: vertex, priority: priority}
		q.byVertex[vertex] = entry
		heap.Push(q, entry)
	case entry.index >= 0 && priority < entry.priority:
		entry.priority = priority
		heap.Fix(q, entry.index)
	}
}

// relax queues each neighbor of v that wasn't visited yet at dist plus the
// weight of the edge to it.
func (q *closestFirstQueue[T]) relax(g gograph.Graph[T], v *gograph.Vertex[T], dist float64) {
	for neighbor := range v.AllNeighbors() {
		entry := q.byVertex[neighbor]
		if entry != nil && entry.index < 0 {
			continue
		}

		edge := g.GetEdge(v, neighbor)
		if edge == nil {
			continue
		}
		// the queue can't order a NaN distance
		if d := dist + edge.Weight(); !math.IsNaN(d) {
			q.push(entry, neighbor, d)
		}
	}
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
	pq.push(nil, v, 0)
	return &closestFirstIterator[T]{
		graph:    graph,
		start:    start,
		pq:       pq,
		currDist: 0,
	}, nil
}

// HasNext returns a boolean indicating whether there are more vertices
// to be visited or not. It drops queued vertices that were removed from
// the graph, and returns true if one is left.
func (c *closestFirstIterator[T]) HasNext() bool {
	for c.pq.Len() > 0 {
		if v := c.pq.peek().vertex; hasEdges(v) || stored(c.graph, v) {
			return true
		}

		c.pq.pop()
	}

	return false
}

// Next returns the next vertex to be visited in closest-first order.
// A vertex removed from the graph after it was queued is skipped, and a
// vertex added later with the same label is visited if the traversal
// reaches it.
//
// If the HasNext is false, returns nil.
func (c *closestFirstIterator[T]) Next() *gograph.Vertex[T] {
	for c.pq.Len() > 0 {
		vp := c.pq.pop()
		currNode := vp.vertex
		if !hasEdges(currNode) && !stored(c.graph, currNode) {
			continue
		}

		c.currDist = vp.priority
		c.pq.relax(c.graph, currNode, c.currDist)

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
	c.currDist = 0

	c.pq = newClosestFirstQueue[T]()
	if v := c.graph.GetVertexByID(c.start); v != nil {
		c.pq.push(nil, v, 0)
	}
}
