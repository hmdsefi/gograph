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
