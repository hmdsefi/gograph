package traverse

import (
	"crypto/rand"
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
type randomWalkIterator[T comparable] struct {
	graph       gograph.Graph[T]   // the graph that being traversed.
	start       *gograph.Vertex[T] // the starting point of the traversal.
	current     *gograph.Vertex[T] // the latest node that has been returned by the iterator.
	steps       int                // the maximum number of steps to be taken during the traversal.
	currentStep int                // the step counter.
}

func NewRandomWalkIterator[T comparable](graph gograph.Graph[T], start *gograph.Vertex[T], steps int) Iterator[T] {
	return &randomWalkIterator[T]{
		graph:   graph,
		start:   start,
		current: start,
		steps:   steps,
	}
}

func (r *randomWalkIterator[T]) HasNext() bool {
	return r.current.OutDegree() > 0 && r.currentStep < r.steps
}

func (r *randomWalkIterator[T]) Next() *gograph.Vertex[T] {
	if !r.HasNext() {
		return nil
	}

	neighbors := r.current.Neighbors()
	if len(neighbors) == 0 {
		// there is no vertex to continue
		r.currentStep = r.steps
		return r.current
	}

	i, _ := rand.Int(rand.Reader, big.NewInt(int64(len(neighbors))))
	r.current = neighbors[i.Int64()]
	r.currentStep++

	return r.current
}

func (r *randomWalkIterator[T]) Iterate(f func(v *gograph.Vertex[T]) error) error {
	for r.HasNext() {
		if err := f(r.Next()); err != nil {
			return err
		}
	}

	return nil
}

func (r *randomWalkIterator[T]) Reset() {
	r.current = r.start
	r.currentStep = 0
}
