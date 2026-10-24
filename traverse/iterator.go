package traverse

import (
	"github.com/hmdsefi/gograph"
)

// Iterator represents a general purpose iterator for iterating over
// a sequence of graph's vertices. It provides methods for checking if
// there are more elements to be iterated over, getting the next element,
// iterating over all elements using a callback function, and resetting
// the iterator to its initial state.
type Iterator[T comparable] interface {
	// HasNext returns a boolean value indicating whether there are more
	// elements to be iterated over. It returns true if there are more
	// elements. Otherwise, returns false.
	HasNext() bool

	// Next returns the next element in the sequence being iterated over.
	// If there are no more elements, it returns nil. It also advances
	// the iterator to the next element.
	Next() *gograph.Vertex[T]

	// Iterate iterates over all elements in the sequence and calls the
	// provided callback function on each element. The callback function
	// takes a single argument of type *Vertex, representing the current
	// element being iterated over. It returns an error value, which is
	// returned by the Iterate method. If the callback function returns
	// an error, iteration is stopped and the error is returned.
	Iterate(func(v *gograph.Vertex[T]) error) error

	// Reset  resets the iterator to its initial state, allowing the
	// sequence to be iterated over again from the beginning.
	Reset()
}

// hasEdges reports whether v has an edge. A removed vertex reports no
// neighbors and an in-degree of 0 from then on, so a vertex with an edge is
// still in its graph.
func hasEdges[T comparable](v *gograph.Vertex[T]) bool {
	return v.InDegree() > 0 || v.OutDegree() > 0
}

// stored reports whether v is still the graph's vertex for its label.
func stored[T comparable](g gograph.Graph[T], v *gograph.Vertex[T]) bool {
	return g.GetVertexByID(v.Label()) == v
}

// reached returns the graph's vertex for a label that the traversal reached
// through an edge, or nil if that vertex is prev, the one seen before with
// the label. A reached vertex has an in-degree of at least 1, so a seen
// vertex with a positive in-degree needs no lookup, and one with an
// in-degree of 0 was removed.
func reached[T comparable](g gograph.Graph[T], prev *gograph.Vertex[T], label T) *gograph.Vertex[T] {
	v := g.GetVertexByID(label)
	if v == nil || v == prev {
		return nil
	}

	return v
}
