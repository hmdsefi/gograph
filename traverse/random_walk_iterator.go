package traverse

import (
	"crypto/rand"
	"math"
	"math/big"

	"github.com/hmdsefi/gograph"
)

// randomWalkIterator implements the Iterator interface to travers
// a graph in a random walk fashion.
//
// Random walk is a stochastic process used to explore a graph, where
// a walker moves through the graph by following random edges. At each
// step, the walker chooses a random neighbor of the current node and
// moves to it, and the process is repeated until a stopping condition
// is met.
//
// In an unweighted graph, each neighboring node has an equal chance of
// being chosen as the next node to visit during the traversal. However,
// in a weighted graph, the probability of choosing a particular neighbor
// as the next node to visit is proportional to the weight of the edge
// connecting the current node and the neighbor. This means that nodes
// connected by heavier edges are more likely to be visited during the
// traversal. Edges with a weight of zero or less are never chosen, unless
// none of the edges of the current node has a positive weight. Then each
// of them has an equal chance.
type randomWalkIterator[T comparable] struct {
	graph       gograph.Graph[T]   // the graph that being traversed.
	start       T                  // the label of starting point of the traversal.
	current     *gograph.Vertex[T] // the latest node that has been returned by the iterator.
	steps       int                // the maximum number of steps to be taken during the traversal.
	currentStep int                // the step counter.
}

// NewRandomWalkIterator creates a new instance of randomWalkIterator
// and returns it as the Iterator interface.
//
// In a weighted graph, each step picks a neighbor with a probability
// proportional to the weight of the edge to it. Edges with a weight of
// zero or less are never picked, unless none of the edges of the current
// vertex has a positive weight. Then each of them has an equal chance. An
// edge with a weight of positive infinity is picked over any edge with a
// finite weight.
func NewRandomWalkIterator[T comparable](graph gograph.Graph[T], start T, steps int) (Iterator[T], error) {
	v := graph.GetVertexByID(start)
	if v == nil {
		return nil, gograph.ErrVertexDoesNotExist
	}

	return &randomWalkIterator[T]{
		graph:   graph,
		start:   start,
		current: v,
		steps:   steps,
	}, nil
}

// HasNext returns a boolean indicating whether there are more vertices
// to be visited or not. The walk returns the start vertex first, and
// stops after steps vertices or at a vertex without outgoing edges.
func (r *randomWalkIterator[T]) HasNext() bool {
	return r.current != nil &&
		r.currentStep < r.steps &&
		(r.currentStep == 0 || r.current.OutDegree() > 0)
}

// Next returns the next vertex to be visited in the random walk traversal.
// It chooses one of the neighbors randomly and returns it.
//
// If the HasNext is false, returns nil.
func (r *randomWalkIterator[T]) Next() *gograph.Vertex[T] {
	if !r.HasNext() {
		return nil
	}

	if r.currentStep == 0 {
		r.currentStep++
		return r.current
	}

	r.currentStep++
	neighbors := r.current.Neighbors()

	if r.graph.IsWeighted() {
		r.current = r.randomVertex(r.current)
		return r.current
	}

	i, _ := rand.Int(rand.Reader, big.NewInt(int64(len(neighbors))))
	r.current = neighbors[i.Int64()]

	return r.current
}

// Iterate iterates through the vertices in random order and applies
// the given function to each vertex. If the function returns an error,
// the iteration stops and the error is returned.
func (r *randomWalkIterator[T]) Iterate(f func(v *gograph.Vertex[T]) error) error {
	for r.HasNext() {
		if err := f(r.Next()); err != nil {
			return err
		}
	}

	return nil
}

// Reset resets the iterator by setting the initial state of the iterator.
func (r *randomWalkIterator[T]) Reset() {
	r.current = r.graph.GetVertexByID(r.start)
	r.currentStep = 0
}

func (r *randomWalkIterator[T]) randomVertex(v *gograph.Vertex[T]) *gograph.Vertex[T] {
	if v == nil {
		return nil
	}

	var maxWeight float64
	var edges []*gograph.Edge[T]
	neighbors := v.Neighbors()

	for _, neighbor := range neighbors {
		if edge := r.graph.GetEdge(v, neighbor); edge != nil {
			edges = append(edges, edge)
			if edge.Weight() > maxWeight {
				maxWeight = edge.Weight()
			}
		}
	}

	if len(edges) == 0 {
		return nil
	}

	if maxWeight <= 0 {
		return edges[randomIndex(len(edges))].OtherVertex(v.Label())
	}

	if math.IsInf(maxWeight, 1) {
		var infinite []*gograph.Edge[T]
		for _, edge := range edges {
			if math.IsInf(edge.Weight(), 1) {
				infinite = append(infinite, edge)
			}
		}
		return infinite[randomIndex(len(infinite))].OtherVertex(v.Label())
	}

	// dividing by the largest weight keeps the sum from overflowing
	var totalWeight float64
	for _, edge := range edges {
		if edge.Weight() > 0 {
			totalWeight += edge.Weight() / maxWeight
		}
	}

	// find the vertex that corresponds to a random weight in [0, totalWeight)
	randWeight := randomFraction() * totalWeight
	var last *gograph.Edge[T]
	for _, edge := range edges {
		if w := edge.Weight(); w > 0 {
			last = edge
			randWeight -= w / maxWeight
			if randWeight < 0 {
				return edge.OtherVertex(v.Label())
			}
		}
	}

	// rounding can leave a remainder after the last positive edge
	return last.OtherVertex(v.Label())
}

// randomIndex returns a random integer in [0, n).
func randomIndex(n int) int {
	i, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(i.Int64())
}

// randomFraction returns a random number in [0, 1).
func randomFraction() float64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(1<<53))
	return float64(n.Int64()) / (1 << 53)
}
