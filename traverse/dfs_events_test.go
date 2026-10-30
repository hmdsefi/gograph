package traverse

import (
	"fmt"
	"iter"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/hmdsefi/gograph"
)

// eventString writes an event as "Kind label@depth" for a vertex event and
// "Kind from->to@depth" for an edge event.
func eventString[T comparable](e DFSEvent[T]) string {
	if e.Edge != nil {
		return fmt.Sprintf("%v %v->%v@%d", e.Kind, e.Edge.Source().Label(), e.Edge.Destination().Label(), e.Depth)
	}
	return fmt.Sprintf("%v %v@%d", e.Kind, e.Vertex.Label(), e.Depth)
}

func eventStrings[T comparable](seq iter.Seq[DFSEvent[T]]) []string {
	var out []string
	for e := range seq {
		out = append(out, eventString(e))
	}
	return out
}

// buildGraph adds the vertices in the given order, then the edges.
func buildGraph(directed bool, vertices []string, edges [][2]string) gograph.Graph[string] {
	var options []gograph.GraphOptionFunc
	if directed {
		options = append(options, gograph.Directed())
	}
	g := gograph.New[string](options...)
	for _, label := range vertices {
		g.AddVertexByLabel(label)
	}
	for _, e := range edges {
		_, _ = g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1]))
	}
	return g
}

func TestDFSEventsAll_DirectedEdgeKinds(t *testing.T) {
	// the directed graph in which a depth-first search from u, then w, finds
	// every kind of edge
	g := buildGraph(true, []string{"u", "v", "w", "x", "y", "z"}, [][2]string{
		{"u", "v"}, {"u", "x"}, {"v", "y"}, {"w", "y"}, {"w", "z"}, {"x", "v"}, {"y", "x"}, {"z", "z"},
	})

	want := []string{
		"DiscoverVertex u@0",
		"TreeEdge u->v@0",
		"DiscoverVertex v@1",
		"TreeEdge v->y@1",
		"DiscoverVertex y@2",
		"TreeEdge y->x@2",
		"DiscoverVertex x@3",
		"BackEdge x->v@3",
		"FinishVertex x@3",
		"FinishVertex y@2",
		"FinishVertex v@1",
		"ForwardEdge u->x@0",
		"FinishVertex u@0",
		"DiscoverVertex w@0",
		"CrossEdge w->y@0",
		"TreeEdge w->z@0",
		"DiscoverVertex z@1",
		"BackEdge z->z@1",
		"FinishVertex z@1",
		"FinishVertex w@0",
	}
	if got := eventStrings(DFSEventsAll(g)); !reflect.DeepEqual(got, want) {
		t.Fatalf("events:\n got %q\nwant %q", got, want)
	}
}

func TestDFSEvents_ParentIsTheVertexThatEntersIt(t *testing.T) {
	// A looks at C first, so C, the vertex the search enters B from, is the
	// parent of B, and A -> B is a forward edge
	g := buildGraph(true, []string{"A", "B", "C"}, [][2]string{{"A", "C"}, {"A", "B"}, {"C", "B"}})

	want := []string{
		"DiscoverVertex A@0",
		"TreeEdge A->C@0",
		"DiscoverVertex C@1",
		"TreeEdge C->B@1",
		"DiscoverVertex B@2",
		"FinishVertex B@2",
		"FinishVertex C@1",
		"ForwardEdge A->B@0",
		"FinishVertex A@0",
	}
	if got := eventStrings(DFSEvents(g, "A")); !reflect.DeepEqual(got, want) {
		t.Fatalf("events:\n got %q\nwant %q", got, want)
	}
}

func TestDFSEvents_UndirectedTriangle(t *testing.T) {
	g := buildGraph(false, []string{"a", "b", "c"}, [][2]string{{"a", "b"}, {"b", "c"}, {"c", "a"}})

	want := []string{
		"DiscoverVertex a@0",
		"TreeEdge a->b@0",
		"DiscoverVertex b@1",
		"TreeEdge b->c@1",
		"DiscoverVertex c@2",
		"BackEdge c->a@2",
		"FinishVertex c@2",
		"FinishVertex b@1",
		"FinishVertex a@0",
	}
	if got := eventStrings(DFSEvents(g, "a")); !reflect.DeepEqual(got, want) {
		t.Fatalf("events:\n got %q\nwant %q", got, want)
	}
}

func TestDFSEvents_SelfLoop(t *testing.T) {
	for _, directed := range []bool{true, false} {
		g := buildGraph(directed, []string{"a", "b"}, [][2]string{{"a", "a"}, {"a", "b"}})
		want := []string{
			"DiscoverVertex a@0",
			"BackEdge a->a@0",
			"TreeEdge a->b@0",
			"DiscoverVertex b@1",
			"FinishVertex b@1",
			"FinishVertex a@0",
		}
		if got := eventStrings(DFSEvents(g, "a")); !reflect.DeepEqual(got, want) {
			t.Fatalf("directed %v, events:\n got %q\nwant %q", directed, got, want)
		}
	}
}

func TestDFSEvents_MissingStart(t *testing.T) {
	g := buildGraph(true, []string{"a"}, nil)
	if got := eventStrings(DFSEvents(g, "b")); got != nil {
		t.Fatalf("expected no events, got %q", got)
	}
	if got := eventStrings(DFSEventsAll(gograph.New[string]())); got != nil {
		t.Fatalf("expected no events for an empty graph, got %q", got)
	}
}

func TestDFSEventKind_String(t *testing.T) {
	want := map[DFSEventKind]string{
		DiscoverVertex:   "DiscoverVertex",
		FinishVertex:     "FinishVertex",
		TreeEdge:         "TreeEdge",
		BackEdge:         "BackEdge",
		ForwardEdge:      "ForwardEdge",
		CrossEdge:        "CrossEdge",
		DFSEventKind(-1): "DFSEventKind(-1)",
		DFSEventKind(6):  "DFSEventKind(6)",
	}
	for kind, name := range want {
		if got := kind.String(); got != name {
			t.Errorf("String() = %q, want %q", got, name)
		}
	}
}

// randomGraph returns a graph with n vertices and up to m random edges,
// self-loops included.
func randomGraph(rng *rand.Rand, directed bool, n, m int) gograph.Graph[int] {
	var options []gograph.GraphOptionFunc
	if directed {
		options = append(options, gograph.Directed())
	}
	g := gograph.New[int](options...)
	for i := range n {
		g.AddVertexByLabel(i)
	}
	for range m {
		_, _ = g.AddEdge(g.GetVertexByID(rng.IntN(n)), g.GetVertexByID(rng.IntN(n)))
	}
	return g
}

// recursiveEvents runs the depth-first search of DFSEventsAll with the
// goroutine stack, the plain way to write it.
func recursiveEvents(g gograph.Graph[int]) []string {
	const onPath, finished = 1, 2
	state := map[*gograph.Vertex[int]]int{}
	discovered := map[*gograph.Vertex[int]]int{}
	var out []string

	var visit func(u, parent *gograph.Vertex[int], depth int)
	visit = func(u, parent *gograph.Vertex[int], depth int) {
		discovered[u] = len(discovered)
		state[u] = onPath
		out = append(out, fmt.Sprintf("DiscoverVertex %d@%d", u.Label(), depth))
		for v := range u.AllNeighbors() {
			edge := fmt.Sprintf("%d->%d@%d", u.Label(), v.Label(), depth)
			switch {
			case state[v] == 0:
				out = append(out, "TreeEdge "+edge)
				visit(v, u, depth+1)
			case state[v] == onPath:
				if g.IsDirected() || v != parent {
					out = append(out, "BackEdge "+edge)
				}
			case !g.IsDirected():
			case discovered[v] > discovered[u]:
				out = append(out, "ForwardEdge "+edge)
			default:
				out = append(out, "CrossEdge "+edge)
			}
		}
		state[u] = finished
		out = append(out, fmt.Sprintf("FinishVertex %d@%d", u.Label(), depth))
	}

	for _, root := range g.GetAllVertices() {
		if state[root] == 0 {
			visit(root, nil, 0)
		}
	}
	return out
}

// checkDFSStructure checks the rules of a depth-first search on events that
// cover all of g: the discover and finish events nest, each edge comes
// once, and each edge kind matches the times of its two vertices.
func checkDFSStructure(t *testing.T, g gograph.Graph[int], events []DFSEvent[int]) {
	t.Helper()
	discover := map[*gograph.Vertex[int]]int{}
	finish := map[*gograph.Vertex[int]]int{}
	parent := map[*gograph.Vertex[int]]*gograph.Vertex[int]{}
	var path []*gograph.Vertex[int]
	edges := map[[2]int]int{}

	for i, e := range events {
		switch e.Kind {
		case DiscoverVertex:
			if _, ok := discover[e.Vertex]; ok {
				t.Fatalf("event %d discovers %d again", i, e.Vertex.Label())
			}
			if e.Depth != len(path) {
				t.Fatalf("event %d: depth %d, want %d", i, e.Depth, len(path))
			}
			discover[e.Vertex] = i
			path = append(path, e.Vertex)
		case FinishVertex:
			if len(path) == 0 || path[len(path)-1] != e.Vertex {
				t.Fatalf("event %d finishes %d, which isn't the last vertex on the path", i, e.Vertex.Label())
			}
			finish[e.Vertex] = i
			path = path[:len(path)-1]
		default:
			from, to := e.Edge.Source(), e.Edge.Destination()
			if len(path) == 0 || path[len(path)-1] != from || e.Depth != len(path)-1 {
				t.Fatalf("event %d: edge %d->%d leaves a vertex off the end of the path", i, from.Label(), to.Label())
			}
			key := [2]int{from.Label(), to.Label()}
			if !g.IsDirected() && key[0] > key[1] {
				key[0], key[1] = key[1], key[0]
			}
			edges[key]++

			_, discovered := discover[to]
			_, finished := finish[to]
			var ok bool
			switch e.Kind {
			case TreeEdge:
				ok = !discovered
				parent[to] = from
			case BackEdge:
				ok = discovered && !finished
			case ForwardEdge:
				ok = g.IsDirected() && finished && discover[to] > discover[from]
			case CrossEdge:
				ok = g.IsDirected() && finished && discover[to] < discover[from]
			}
			if !ok {
				t.Fatalf("event %d: %v %d->%d doesn't match the search state", i, e.Kind, from.Label(), to.Label())
			}
		}
	}

	if len(path) != 0 || len(discover) != int(g.Order()) || len(finish) != int(g.Order()) {
		t.Fatalf("discovered %d and finished %d of %d vertices", len(discover), len(finish), g.Order())
	}
	want := map[[2]int]int{}
	for _, e := range g.AllEdges() {
		key := [2]int{e.Source().Label(), e.Destination().Label()}
		if !g.IsDirected() && key[0] > key[1] {
			key[0], key[1] = key[1], key[0]
		}
		want[key] = 1
	}
	if !reflect.DeepEqual(edges, want) {
		t.Fatalf("reported edges %v, want each of %v once", edges, want)
	}
	for child, p := range parent {
		// a child's events nest inside its parent's
		if discover[p] > discover[child] || finish[child] > finish[p] {
			t.Fatalf("events of %d don't nest inside those of its parent %d", child.Label(), p.Label())
		}
	}
}

func TestDFSEventsAll_RandomGraphs(t *testing.T) {
	for seed := range uint64(300) {
		rng := rand.New(rand.NewPCG(seed, 132)) //nolint:gosec // seeded so failures are reproducible
		for _, directed := range []bool{true, false} {
			n := 1 + rng.IntN(12)
			g := randomGraph(rng, directed, n, rng.IntN(2*n+1))

			var events []DFSEvent[int]
			for e := range DFSEventsAll(g) {
				events = append(events, e)
			}
			checkDFSStructure(t, g, events)

			got := make([]string, len(events))
			for i, e := range events {
				got[i] = eventString(e)
			}
			if want := recursiveEvents(g); !reflect.DeepEqual(got, want) {
				t.Fatalf("seed %d, directed %v:\n got %q\nwant %q", seed, directed, got, want)
			}
		}
	}
}

func TestDFSEvents_StopAnywhere(t *testing.T) {
	g := buildGraph(true, []string{"u", "v", "w", "x", "y", "z"}, [][2]string{
		{"u", "v"}, {"u", "x"}, {"v", "y"}, {"w", "y"}, {"w", "z"}, {"x", "v"}, {"y", "x"}, {"z", "z"},
	})
	all := eventStrings(DFSEventsAll(g))
	for stop := 1; stop <= len(all); stop++ {
		var got []string
		for e := range DFSEventsAll(g) {
			got = append(got, eventString(e))
			if len(got) == stop {
				break
			}
		}
		if !reflect.DeepEqual(got, all[:stop]) {
			t.Fatalf("break after %d events: got %q", stop, got)
		}
	}
}

func TestDFSEvents_LongChain(t *testing.T) {
	const n = 100_000
	g := gograph.New[int](gograph.Directed())
	prev := g.AddVertexByLabel(0)
	for i := 1; i < n; i++ {
		next := g.AddVertexByLabel(i)
		_, _ = g.AddEdge(prev, next)
		prev = next
	}

	counts := map[DFSEventKind]int{}
	maxDepth := 0
	for e := range DFSEvents(g, 0) {
		counts[e.Kind]++
		maxDepth = max(maxDepth, e.Depth)
	}
	want := map[DFSEventKind]int{DiscoverVertex: n, FinishVertex: n, TreeEdge: n - 1}
	if !reflect.DeepEqual(counts, want) || maxDepth != n-1 {
		t.Fatalf("got %v and depth %d, want %v and depth %d", counts, maxDepth, want, n-1)
	}
}

func TestDFSEvents_SkipRemovedDuringLoop(t *testing.T) {
	g := buildGraph(true, []string{"a", "b", "c", "d"}, [][2]string{{"a", "b"}, {"a", "c"}, {"a", "d"}})

	var got []string
	for e := range DFSEvents(g, "a") {
		got = append(got, eventString(e))
		if e.Kind == DiscoverVertex && e.Vertex.Label() == "a" {
			g.RemoveVertices(g.GetVertexByID("c"))
			g.RemoveEdges(g.GetEdge(g.GetVertexByID("a"), g.GetVertexByID("d")))
		}
	}
	want := []string{
		"DiscoverVertex a@0",
		"TreeEdge a->b@0",
		"DiscoverVertex b@1",
		"FinishVertex b@1",
		"FinishVertex a@0",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events:\n got %q\nwant %q", got, want)
	}
}

func TestDFSEventsAll_SkipRootRemovedDuringLoop(t *testing.T) {
	// the wrapper isn't a graph from New, so the roots come from
	// GetAllVertices before the first event
	g := buildGraph(true, []string{"a", "b"}, nil)
	var got []string
	for e := range DFSEventsAll[string](&edgeCounter[string]{Graph: g}) {
		got = append(got, eventString(e))
		g.RemoveVertices(g.GetVertexByID("b"))
	}
	want := []string{"DiscoverVertex a@0", "FinishVertex a@0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events:\n got %q\nwant %q", got, want)
	}
}
