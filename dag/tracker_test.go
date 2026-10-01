package dag

import (
	"errors"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/hmdsefi/gograph"
)

func newTracker[T comparable](t *testing.T, g gograph.Graph[T]) *Tracker[T] {
	t.Helper()
	tr, err := NewTracker(g)
	if err != nil {
		t.Fatalf("NewTracker: unexpected error: %v", err)
	}
	return tr
}

func assertReady(t *testing.T, tr *Tracker[string], want ...string) {
	t.Helper()
	assertLabels(t, "Ready()", labels(tr.Ready()), want)
}

func mustDone[T comparable](t *testing.T, tr *Tracker[T], labels ...T) {
	t.Helper()
	if err := tr.Done(labels...); err != nil {
		t.Fatalf("Done(%v): unexpected error: %v", labels, err)
	}
}

func TestTracker_Diamond(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"}, [2]string{"a", "c"}, [2]string{"b", "d"}, [2]string{"c", "d"})
	tr := newTracker(t, g)

	if tr.Remaining() != 4 {
		t.Fatalf("Remaining() = %d, want 4", tr.Remaining())
	}
	assertReady(t, tr, "a")
	assertReady(t, tr)

	mustDone(t, tr, "a")
	assertReady(t, tr, "b", "c")

	// d waits for c, even though b is done
	mustDone(t, tr, "b")
	assertReady(t, tr)
	mustDone(t, tr, "c")
	assertReady(t, tr, "d")

	mustDone(t, tr, "d")
	assertReady(t, tr)
	if tr.Remaining() != 0 {
		t.Fatalf("Remaining() = %d, want 0", tr.Remaining())
	}
}

func TestTracker_SlowVertex(t *testing.T) {
	// with levels, c would wait for slow
	g := newGraph(t, [2]string{"slow", "x"}, [2]string{"fast", "c"})
	tr := newTracker(t, g)

	assertReady(t, tr, "slow", "fast")
	mustDone(t, tr, "fast")
	assertReady(t, tr, "c")
	mustDone(t, tr, "c")
	if tr.Remaining() != 2 {
		t.Fatalf("Remaining() = %d, want 2", tr.Remaining())
	}
}

func TestTracker_DoneSeveral(t *testing.T) {
	g := newGraph(t, [2]string{"a", "c"}, [2]string{"b", "c"}, [2]string{"b", "d"})
	tr := newTracker(t, g)

	assertReady(t, tr, "a", "b")
	mustDone(t, tr, "b", "a")
	assertReady(t, tr, "d", "c")
	mustDone(t, tr)
	if tr.Remaining() != 2 {
		t.Fatalf("Remaining() = %d, want 2", tr.Remaining())
	}
}

func TestTracker_ReadyOrder(t *testing.T) {
	g := gograph.New[string](gograph.Directed())
	for _, label := range []string{"z", "y", "x", "w", "v"} {
		g.AddVertexByLabel(label)
	}
	for _, e := range [][2]string{{"y", "w"}, {"z", "x"}, {"y", "x"}, {"z", "v"}} {
		_, _ = g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1]))
	}
	tr := newTracker(t, g)

	// first in GetAllVertices order, then in the order they became ready
	assertReady(t, tr, "z", "y")
	mustDone(t, tr, "y")
	mustDone(t, tr, "z")
	assertReady(t, tr, "w", "x", "v")
}

func TestTracker_Errors(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"}, [2]string{"a", "c"})
	g.AddVertexByLabel("x")
	tr := newTracker(t, g)

	tests := []struct {
		name   string
		labels []string
		want   error
	}{
		{"source before Ready", []string{"a"}, ErrNotReady},
		{"unknown label", []string{"unknown"}, gograph.ErrVertexDoesNotExist},
	}
	for _, tt := range tests {
		if err := tr.Done(tt.labels...); !errors.Is(err, tt.want) {
			t.Fatalf("%s: Done(%v) = %v, want %v", tt.name, tt.labels, err, tt.want)
		}
	}

	assertReady(t, tr, "a", "x")
	tests = []struct {
		name   string
		labels []string
		want   error
	}{
		{"dependencies not done", []string{"b"}, ErrNotReady},
		{"label given twice", []string{"a", "a"}, ErrAlreadyDone},
		{"valid label then unknown", []string{"a", "unknown"}, gograph.ErrVertexDoesNotExist},
		{"valid label then not ready", []string{"x", "b"}, ErrNotReady},
	}
	for _, tt := range tests {
		if err := tr.Done(tt.labels...); !errors.Is(err, tt.want) {
			t.Fatalf("%s: Done(%v) = %v, want %v", tt.name, tt.labels, err, tt.want)
		}
	}

	// the failed calls didn't mark anything done
	if tr.Remaining() != 4 {
		t.Fatalf("Remaining() = %d, want 4", tr.Remaining())
	}
	assertReady(t, tr)

	mustDone(t, tr, "a")
	if err := tr.Done("a"); !errors.Is(err, ErrAlreadyDone) {
		t.Fatalf("Done(a) twice = %v, want %v", err, ErrAlreadyDone)
	}
	assertReady(t, tr, "b", "c")
}

func TestTracker_GraphChangesAfterward(t *testing.T) {
	g := newGraph(t, [2]string{"a", "b"})
	tr := newTracker(t, g)

	_, _ = g.AddEdge(g.GetVertexByID("b"), gograph.NewVertex("c"))
	g.RemoveEdges(g.GetEdge(g.GetVertexByID("a"), g.GetVertexByID("b")))

	if tr.Remaining() != 2 {
		t.Fatalf("Remaining() = %d, want 2", tr.Remaining())
	}
	assertReady(t, tr, "a")
	mustDone(t, tr, "a")
	assertReady(t, tr, "b")
	mustDone(t, tr, "b")
	assertReady(t, tr)
	if tr.Remaining() != 0 {
		t.Fatalf("Remaining() = %d, want 0", tr.Remaining())
	}
}

func TestTracker_Empty(t *testing.T) {
	tr := newTracker(t, gograph.New[string](gograph.Directed()))
	if tr.Remaining() != 0 || tr.Ready() != nil {
		t.Fatalf("Remaining() = %d, want 0, and Ready() = nil", tr.Remaining())
	}
	mustDone(t, tr)
}

func TestTracker_InvalidGraphs(t *testing.T) {
	tests := []struct {
		name string
		g    gograph.Graph[string]
		want error
	}{
		{"cycle", newGraph(t, [2]string{"a", "b"}, [2]string{"b", "a"}), gograph.ErrDAGHasCycle},
		{"self-loop", newGraph(t, [2]string{"a", "a"}), gograph.ErrDAGHasCycle},
		{"undirected", gograph.New[string](), gograph.ErrNotDirected},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tr, err := NewTracker(tt.g); !errors.Is(err, tt.want) || tr != nil {
				t.Fatalf("NewTracker = %v, %v, want nil, %v", tr, err, tt.want)
			}
		})
	}
}

// TestTracker_RandomDAGs marks the running vertices done in random batches,
// and checks that Ready returns exactly the vertices whose dependencies are
// all done, each one once.
func TestTracker_RandomDAGs(t *testing.T) {
	rng := rand.New(rand.NewSource(1)) //nolint:gosec // seeded so failures are reproducible
	for run := 0; run < 300; run++ {
		g := randomDAG(rng)
		preds := make(map[int][]int)
		for _, e := range g.AllEdges() {
			to := e.Destination().Label()
			preds[to] = append(preds[to], e.Source().Label())
		}
		tr := newTracker(t, g)

		done := make(map[int]bool)
		returned := make(map[int]bool)
		var running []int
		for tr.Remaining() > 0 {
			var want []int
			for _, v := range g.GetAllVertices() {
				if returned[v.Label()] {
					continue
				}
				ready := true
				for _, p := range preds[v.Label()] {
					ready = ready && done[p]
				}
				if ready {
					want = append(want, v.Label())
				}
			}

			got := tr.Ready()
			for _, v := range got {
				if v != g.GetVertexByID(v.Label()) {
					t.Fatalf("run %d: Ready returned a copy of %d", run, v.Label())
				}
				returned[v.Label()] = true
				running = append(running, v.Label())
			}
			gotLabels := labels(got)
			sort.Ints(gotLabels)
			sort.Ints(want)
			if len(gotLabels) != 0 || len(want) != 0 {
				if !reflect.DeepEqual(gotLabels, want) {
					t.Fatalf("run %d: Ready() = %v, want %v", run, gotLabels, want)
				}
			}
			if len(running) == 0 {
				t.Fatalf("run %d: nothing is ready or running, and %d vertices remain", run, tr.Remaining())
			}

			rng.Shuffle(len(running), func(i, j int) { running[i], running[j] = running[j], running[i] })
			n := 1 + rng.Intn(len(running))
			mustDone(t, tr, running[:n]...)
			for _, label := range running[:n] {
				done[label] = true
			}
			running = append([]int(nil), running[n:]...)
		}

		if len(returned) != int(g.Order()) || tr.Ready() != nil {
			t.Fatalf("run %d: Ready returned %d of %d vertices", run, len(returned), g.Order())
		}
	}
}
