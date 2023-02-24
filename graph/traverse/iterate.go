package traverse

import "github.com/hmdsefi/gograph/graph"

type Iterator[T comparable] interface {
	HasNext() bool
	Next() *graph.Vertex[T]
	Iterate(func(v *graph.Vertex[T]) error) error
	Reset()
}
