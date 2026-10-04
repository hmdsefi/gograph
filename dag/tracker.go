package dag

import (
	"errors"
	"fmt"

	"github.com/hmdsefi/gograph"
)

var (
	// ErrNotReady is returned by Tracker.Done for a vertex that Ready hasn't
	// returned yet.
	ErrNotReady = errors.New("vertex was not returned by Ready")

	// ErrAlreadyDone is returned by Tracker.Done for a vertex that is already
	// done.
	ErrAlreadyDone = errors.New("vertex is already done")
)

// Tracker hands out the vertices of a directed acyclic graph as soon as
// every vertex they depend on is done. Unlike with Levels, a slow vertex
// only holds up the vertices that depend on it.
//
// Ready returns the vertices that can start, and Done marks vertices as
// finished, which can make the vertices that depend on them ready. As in
// the rest of the package, an edge A -> B means B depends on A, so Ready
// returns B once A is done.
//
// A Tracker isn't safe for concurrent use. The usual pattern is a single
// goroutine that starts a worker for each ready vertex, and calls Done as
// the results come back on a channel:
//
//	t, err := dag.NewTracker(g)
//	if err != nil {
//		return err
//	}
//	// buffered, so workers don't block if the loop returns early
//	results := make(chan result, g.Order())
//	for t.Remaining() > 0 {
//		for _, v := range t.Ready() {
//			go func() {
//				results <- result{label: v.Label(), err: run(ctx, v.Label())}
//			}()
//		}
//		r := <-results
//		if r.err != nil {
//			return r.err
//		}
//		if err := t.Done(r.label); err != nil {
//			return err
//		}
//	}
//
// What to do when a vertex fails is up to the caller. To skip the vertices
// that depend on a failed one and keep going with the others, mark the
// failed vertex done and, when Ready returns any of its Descendants, mark
// them done without running them.
type Tracker[T comparable] struct {
	vertices  map[T]*trackedVertex[T]
	ready     []*trackedVertex[T] // vertices with no dependencies left that Ready hasn't returned
	remaining int
}

type trackedVertex[T comparable] struct {
	vertex     *gograph.Vertex[T]
	dependents []*trackedVertex[T]
	waiting    int // dependencies that aren't done yet
	state      trackerState
}

type trackerState uint8

const (
	pending trackerState = iota // not returned by Ready yet
	running                     // returned by Ready, not done yet
	finished
)

// NewTracker returns a Tracker for g. It copies what it needs from g, so
// changes to g afterwards don't affect the Tracker.
//
// It returns gograph.ErrNotDirected for undirected graphs, and
// gograph.ErrDAGHasCycle if the graph has a cycle.
func NewTracker[T comparable](g gograph.Graph[T]) (*Tracker[T], error) {
	if _, err := Levels(g); err != nil {
		return nil, err
	}

	vertices := g.GetAllVertices()
	t := &Tracker[T]{
		vertices:  make(map[T]*trackedVertex[T], len(vertices)),
		remaining: len(vertices),
	}
	for _, v := range vertices {
		t.vertices[v.Label()] = &trackedVertex[T]{vertex: v}
	}
	for _, v := range vertices {
		from := t.vertices[v.Label()]
		for _, n := range v.Neighbors() {
			to := t.vertices[n.Label()]
			from.dependents = append(from.dependents, to)
			to.waiting++
		}
	}
	for _, v := range vertices {
		if tv := t.vertices[v.Label()]; tv.waiting == 0 {
			t.ready = append(t.ready, tv)
		}
	}

	return t, nil
}

// Ready returns the vertices whose dependencies are all done and that Ready
// hasn't returned before, as the graph's own vertex pointers. It returns
// nil if there are none.
//
// The first call returns the vertices with no dependencies, in the order of
// GetAllVertices. After that, vertices come in the order they became ready:
// following the order of the Done calls, and for each vertex marked done,
// the order its edges were added.
func (t *Tracker[T]) Ready() []*gograph.Vertex[T] {
	if len(t.ready) == 0 {
		return nil
	}

	ready := make([]*gograph.Vertex[T], len(t.ready))
	for i, v := range t.ready {
		v.state = running
		ready[i] = v.vertex
	}
	t.ready = t.ready[:0]

	return ready
}

// Done marks the vertices with the given labels as finished, so the
// vertices that depend on them can become ready. Each one must have been
// returned by Ready.
//
// It returns an error that matches gograph.ErrVertexDoesNotExist with
// errors.Is if a label isn't in the graph, ErrNotReady if Ready hasn't
// returned the vertex, and ErrAlreadyDone if the vertex is already done or
// its label is given more than once. If it returns an error, none of the
// vertices are marked done.
func (t *Tracker[T]) Done(labels ...T) error {
	given := make(map[T]bool, len(labels))
	for _, label := range labels {
		v, ok := t.vertices[label]
		switch {
		case !ok:
			return fmt.Errorf("%w: %v", gograph.ErrVertexDoesNotExist, label)
		case v.state == finished || given[label]:
			return fmt.Errorf("%w: %v", ErrAlreadyDone, label)
		case v.state == pending:
			return fmt.Errorf("%w: %v", ErrNotReady, label)
		}
		given[label] = true
	}

	for _, label := range labels {
		v := t.vertices[label]
		v.state = finished
		t.remaining--
		for _, d := range v.dependents {
			d.waiting--
			if d.waiting == 0 {
				t.ready = append(t.ready, d)
			}
		}
	}

	return nil
}

// Remaining returns the number of vertices that aren't done yet.
func (t *Tracker[T]) Remaining() int {
	return t.remaining
}
