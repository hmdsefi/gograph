package traverse

import (
	"github.com/hmdsefi/gograph"
)

type Iterator[T comparable] interface {
	HasNext() bool
	Next() *gograph.Vertex[T]
	Iterate(func(v *gograph.Vertex[T]) error) error
	Reset()
}
