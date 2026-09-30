package path

import (
	"errors"
	"testing"

	"github.com/hmdsefi/gograph"
)

func TestErrNotDirected_SameAsRoot(t *testing.T) {
	if ErrNotDirected != gograph.ErrNotDirected {
		t.Fatal("path.ErrNotDirected and gograph.ErrNotDirected should be the same error")
	}

	_, err := TransitiveReduction(gograph.New[int]())
	if !errors.Is(err, ErrNotDirected) || !errors.Is(err, gograph.ErrNotDirected) {
		t.Fatalf("expected ErrNotDirected, got %v", err)
	}
}
