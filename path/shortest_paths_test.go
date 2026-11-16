package path

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"strconv"
	"testing"

	"github.com/hmdsefi/gograph"
)

// pathsFunc is DijkstraPaths, BellmanFordPaths or BFSPaths.
type pathsFunc func(g gograph.Graph[string], start string) (*ShortestPaths[string], error)

var pathsFuncs = []struct {
	name string
	run  pathsFunc
}{
	{"DijkstraPaths", DijkstraPaths[string]},
	{"BellmanFordPaths", BellmanFordPaths[string]},
	{"BFSPaths", BFSPaths[string]},
}

// assertShortestPath checks that the path to label is a simple path from start
// over edges of g, made of the graph's own vertices, and returns it with the
// sum of its weights.
func assertShortestPath(
	t *testing.T,
	g gograph.Graph[string],
	paths *ShortestPaths[string],
	start, label string,
) ([]*gograph.Vertex[string], float64) {
	t.Helper()
	p, ok := paths.PathTo(label)
	if !ok {
		t.Fatalf("PathTo(%s) reports no path", label)
	}
	if len(p) == 0 || p[0].Label() != start || p[len(p)-1].Label() != label {
		t.Fatalf("PathTo(%s) = %v, want it to run from %s to %s", label, pathLabels(p), start, label)
	}
	seen := map[string]bool{}
	var sum float64
	for i, v := range p {
		if g.GetVertexByID(v.Label()) != v {
			t.Errorf("vertex %s is not the graph's own pointer", v.Label())
		}
		if seen[v.Label()] {
			t.Fatalf("PathTo(%s) = %v repeats %s", label, pathLabels(p), v.Label())
		}
		seen[v.Label()] = true
		if i > 0 {
			e := g.GetEdge(p[i-1], v)
			if e == nil {
				t.Fatalf("no edge from %s to %s in %v", p[i-1].Label(), v.Label(), pathLabels(p))
			}
			sum += e.Weight()
		}
	}
	return p, sum
}

func TestShortestPaths_StartUnreachableAndUnknown(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 2}, weightedEdge{"C", "A", 1})

	for _, tt := range pathsFuncs {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := tt.run(g, "A")
			if err != nil {
				t.Fatalf("error = %v", err)
			}

			if d, ok := paths.Distance("A"); !ok || d != 0 {
				t.Errorf("Distance(A) = %v, %v, want 0 and true", d, ok)
			}
			if p, ok := paths.PathTo("A"); !ok || len(p) != 1 || p[0] != g.GetVertexByID("A") {
				t.Errorf("PathTo(A) = %v, %v, want the graph's own A alone", pathLabels(p), ok)
			}

			// C leads to A, but A doesn't reach it, and X is not in the graph
			for _, label := range []string{"C", "X"} {
				if d, ok := paths.Distance(label); ok || d != 0 {
					t.Errorf("Distance(%s) = %v, %v, want 0 and false", label, d, ok)
				}
				if p, ok := paths.PathTo(label); ok || p != nil {
					t.Errorf("PathTo(%s) = %v, %v, want nil and false", label, pathLabels(p), ok)
				}
			}
		})
	}
}

func TestShortestPaths_MissingStart(t *testing.T) {
	g := directedWeighted(weightedEdge{"A", "B", 1})
	for _, tt := range pathsFuncs {
		t.Run(tt.name, func(t *testing.T) {
			if paths, err := tt.run(g, "X"); !errors.Is(err, gograph.ErrVertexDoesNotExist) || paths != nil {
				t.Errorf("got %v and %v, want nil and %v", paths, err, gograph.ErrVertexDoesNotExist)
			}
		})
	}
}

func TestShortestPaths_KeepTheirAnswersWhenTheGraphChanges(t *testing.T) {
	for _, tt := range pathsFuncs {
		t.Run(tt.name, func(t *testing.T) {
			g := directedWeighted(weightedEdge{"A", "B", 1}, weightedEdge{"B", "C", 1})
			paths, err := tt.run(g, "A")
			if err != nil {
				t.Fatalf("error = %v", err)
			}

			_, _ = g.AddEdge(g.GetVertexByID("A"), g.GetVertexByID("C"), gograph.WithEdgeWeight(0))
			g.RemoveVertices(g.GetVertexByID("B"))

			if d, ok := paths.Distance("C"); !ok || d != 2 {
				t.Errorf("Distance(C) = %v, %v, want 2 and true", d, ok)
			}
			p, _ := paths.PathTo("C")
			if want := []string{"A", "B", "C"}; !reflect.DeepEqual(pathLabels(p), want) {
				t.Errorf("PathTo(C) = %v, want %v", pathLabels(p), want)
			}
		})
	}
}

func TestDijkstraPaths_PicksTheCheaperPath(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "G", 10},
		weightedEdge{"S", "B", 4}, weightedEdge{"B", "C", 1}, weightedEdge{"C", "G", 1},
	)
	paths, err := DijkstraPaths(g, "S")
	if err != nil {
		t.Fatalf("DijkstraPaths() error = %v", err)
	}

	p, sum := assertShortestPath(t, g, paths, "S", "G")
	if want := []string{"S", "B", "C", "G"}; !reflect.DeepEqual(pathLabels(p), want) || sum != 6 {
		t.Errorf("PathTo(G) = %v weighing %v, want %v and 6", pathLabels(p), sum, want)
	}
	for label, want := range map[string]float64{"S": 0, "A": 1, "B": 4, "C": 5, "G": 6} {
		if d, ok := paths.Distance(label); !ok || d != want {
			t.Errorf("Distance(%s) = %v, %v, want %v and true", label, d, ok, want)
		}
	}
}

func TestDijkstraPaths_NegativeWeight(t *testing.T) {
	tests := []struct {
		name  string
		edges []weightedEdge
	}{
		{"negative weight", []weightedEdge{{"S", "A", 1}, {"A", "B", -1}}},
		{"-Inf weight", []weightedEdge{{"S", "A", math.Inf(-1)}}},
		{"negative self-loop", []weightedEdge{{"S", "S", -1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := DijkstraPaths(directedWeighted(tt.edges...), "S")
			if !errors.Is(err, ErrNegativeWeight) || paths != nil {
				t.Errorf("got %v and %v, want nil and %v", paths, err, ErrNegativeWeight)
			}
		})
	}
}

func TestDijkstraPaths_NegativeWeightOutOfReachIsNotRead(t *testing.T) {
	g := directedWeighted(weightedEdge{"S", "A", 1}, weightedEdge{"X", "Y", -1}, weightedEdge{"X", "S", -1})
	paths, err := DijkstraPaths(g, "S")
	if err != nil {
		t.Fatalf("DijkstraPaths() error = %v", err)
	}
	if d, ok := paths.Distance("A"); !ok || d != 1 {
		t.Errorf("Distance(A) = %v, %v, want 1 and true", d, ok)
	}
}

func TestDijkstraPaths_OverflowingSumIsNoPath(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", math.MaxFloat64}, weightedEdge{"A", "B", math.MaxFloat64},
	)
	paths, err := DijkstraPaths(g, "S")
	if err != nil {
		t.Fatalf("DijkstraPaths() error = %v", err)
	}
	if d, ok := paths.Distance("A"); !ok || d != math.MaxFloat64 {
		t.Errorf("Distance(A) = %v, %v, want %v and true", d, ok, math.MaxFloat64)
	}
	if d, ok := paths.Distance("B"); ok {
		t.Errorf("Distance(B) = %v, want it out of reach", d)
	}
}

func TestDijkstraPaths_ZeroWeightsAndSelfLoops(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "S", 0}, weightedEdge{"S", "A", 0},
		weightedEdge{"A", "S", 0}, weightedEdge{"A", "B", 0}, weightedEdge{"B", "A", 0},
	)
	paths, err := DijkstraPaths(g, "S")
	if err != nil {
		t.Fatalf("DijkstraPaths() error = %v", err)
	}
	p, sum := assertShortestPath(t, g, paths, "S", "B")
	if want := []string{"S", "A", "B"}; !reflect.DeepEqual(pathLabels(p), want) || sum != 0 {
		t.Errorf("PathTo(B) = %v weighing %v, want %v and 0", pathLabels(p), sum, want)
	}
}

func TestDijkstraPaths_UndirectedAndUnweighted(t *testing.T) {
	g := gograph.New[string]()
	a, b, c := g.AddVertexByLabel("A"), g.AddVertexByLabel("B"), g.AddVertexByLabel("C")
	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(b, c)

	paths, err := DijkstraPaths(g, "C")
	if err != nil {
		t.Fatalf("DijkstraPaths() error = %v", err)
	}
	p, sum := assertShortestPath(t, g, paths, "C", "A")
	if want := []string{"C", "B", "A"}; !reflect.DeepEqual(pathLabels(p), want) || sum != 0 {
		t.Errorf("PathTo(A) = %v weighing %v, want %v and 0", pathLabels(p), sum, want)
	}
	if d, ok := paths.Distance("A"); !ok || d != 0 {
		t.Errorf("Distance(A) = %v, %v, want 0 and true", d, ok)
	}
}

func TestDijkstraPaths_EqualPathsKeepTheFirstFound(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "B", 1}, weightedEdge{"S", "A", 1},
		weightedEdge{"A", "G", 1}, weightedEdge{"B", "G", 1},
	)
	for range 20 {
		paths, err := DijkstraPaths(g, "S")
		if err != nil {
			t.Fatalf("DijkstraPaths() error = %v", err)
		}
		// B's edge from S was added first, so B is settled first
		p, _ := paths.PathTo("G")
		if want := []string{"S", "B", "G"}; !reflect.DeepEqual(pathLabels(p), want) {
			t.Fatalf("PathTo(G) = %v, want %v", pathLabels(p), want)
		}
	}
}

func TestDijkstraPaths_ShorterPathMovesAQueuedVertex(t *testing.T) {
	// G is queued at 10, then at 7, then at 3, while it waits in the queue
	g := directedWeighted(
		weightedEdge{"S", "G", 10}, weightedEdge{"S", "A", 1}, weightedEdge{"S", "B", 2},
		weightedEdge{"A", "G", 6}, weightedEdge{"B", "G", 1}, weightedEdge{"G", "H", 1},
	)
	paths, err := DijkstraPaths(g, "S")
	if err != nil {
		t.Fatalf("DijkstraPaths() error = %v", err)
	}
	p, sum := assertShortestPath(t, g, paths, "S", "H")
	if want := []string{"S", "B", "G", "H"}; !reflect.DeepEqual(pathLabels(p), want) || sum != 4 {
		t.Errorf("PathTo(H) = %v weighing %v, want %v and 4", pathLabels(p), sum, want)
	}
}

// readCounter counts the edges a search reads.
type readCounter struct {
	gograph.Graph[int]
	reads *int
}

func (c readCounter) GetEdge(from, to *gograph.Vertex[int]) *gograph.Edge[int] {
	*c.reads++
	return c.Graph.GetEdge(from, to)
}

func TestDijkstraPaths_ReadsOnlyWhatStartReaches(t *testing.T) {
	// 0, 1 and 2 form a chain that leads nowhere else, next to 10,000 vertices
	// that lead into it
	g := gograph.New[int](gograph.Directed(), gograph.Weighted())
	for i := range 10_003 {
		g.AddVertexByLabel(i)
	}
	for i := range 10_002 {
		from, to := i+1, i
		if i < 2 {
			from, to = i, i+1
		}
		_, _ = g.AddEdge(g.GetVertexByID(from), g.GetVertexByID(to), gograph.WithEdgeWeight(1))
	}

	reads := 0
	paths, err := DijkstraPaths[int](readCounter{g, &reads}, 0)
	if err != nil {
		t.Fatalf("DijkstraPaths() error = %v", err)
	}
	if d, ok := paths.Distance(2); !ok || d != 2 {
		t.Errorf("Distance(2) = %v, %v, want 2 and true", d, ok)
	}
	if _, ok := paths.Distance(3); ok {
		t.Error("3 is out of reach")
	}
	if reads != 2 || len(paths.vertices) != 3 {
		t.Errorf("read %d edges and kept %d vertices, want 2 and 3", reads, len(paths.vertices))
	}
}

// pathsNoEdges is a graph that lists neighbors it has no edge for.
type pathsNoEdges struct {
	gograph.Graph[string]
}

func (pathsNoEdges) GetEdge(_, _ *gograph.Vertex[string]) *gograph.Edge[string] { return nil }

func TestDijkstraPaths_NeighborWithoutAnEdgeIsSkipped(t *testing.T) {
	g := pathsNoEdges{directedWeighted(weightedEdge{"A", "B", 1})}
	paths, err := DijkstraPaths[string](g, "A")
	if err != nil {
		t.Fatalf("DijkstraPaths() error = %v", err)
	}
	if _, ok := paths.Distance("B"); ok {
		t.Error("B has no edge to it")
	}
}

// randomPathGraph returns a random graph with whole weights from 0 to 3, which
// make many equal paths.
func randomPathGraph(rng *rand.Rand, directed bool) (gograph.Graph[string], []string) {
	var g gograph.Graph[string]
	if directed {
		g = gograph.New[string](gograph.Directed(), gograph.Weighted())
	} else {
		g = gograph.New[string](gograph.Weighted())
	}
	n := 1 + rng.IntN(13)
	labels := make([]string, n)
	for i := range labels {
		labels[i] = strconv.Itoa(i)
		g.AddVertexByLabel(labels[i])
	}
	for range rng.IntN(4 * n) {
		from, to := g.GetVertexByID(labels[rng.IntN(n)]), g.GetVertexByID(labels[rng.IntN(n)])
		_, _ = g.AddEdge(from, to, gograph.WithEdgeWeight(float64(rng.IntN(4))))
	}
	return g, labels
}

// TestDijkstraPaths_MatchesDijkstra compares the distances with Dijkstra and
// the paths with DijkstraMultiSource from the same single source.
func TestDijkstraPaths_MatchesDijkstra(t *testing.T) {
	rng := rand.New(rand.NewPCG(11, 0)) //nolint:gosec // seeded so failures are reproducible

	for trial := range 3000 {
		g, labels := randomPathGraph(rng, trial%2 == 0)
		start := labels[0]

		paths, err := DijkstraPaths(g, start)
		if err != nil {
			t.Fatalf("trial %d: error = %v", trial, err)
		}
		dist := Dijkstra(g, start)
		nearest, err := DijkstraMultiSource(g, start)
		if err != nil {
			t.Fatalf("trial %d: DijkstraMultiSource error = %v", trial, err)
		}

		for _, label := range labels {
			d, ok := paths.Distance(label)
			if want := dist[label]; ok != (want != math.MaxFloat64) || (ok && d != want) {
				t.Fatalf("trial %d: Distance(%s) = %v, %v, Dijkstra gives %v", trial, label, d, ok, want)
			}
			if !ok {
				continue
			}
			p, sum := assertShortestPath(t, g, paths, start, label)
			if sum != d {
				t.Fatalf("trial %d: PathTo(%s) = %v weighs %v, want %v", trial, label, pathLabels(p), sum, d)
			}
			if want, _ := nearest.PathTo(label); !reflect.DeepEqual(pathLabels(p), pathLabels(want)) {
				t.Fatalf("trial %d: PathTo(%s) = %v, DijkstraMultiSource gives %v",
					trial, label, pathLabels(p), pathLabels(want))
			}
		}
	}
}

func TestBFSPaths_FewestEdgesWhateverTheWeights(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "B", 1}, weightedEdge{"B", "G", 1},
		weightedEdge{"S", "C", 50}, weightedEdge{"C", "G", math.NaN()}, weightedEdge{"G", "H", -4},
	)
	paths, err := BFSPaths(g, "S")
	if err != nil {
		t.Fatalf("BFSPaths() error = %v", err)
	}

	p, _ := assertShortestPath(t, g, paths, "S", "H")
	if want := []string{"S", "C", "G", "H"}; !reflect.DeepEqual(pathLabels(p), want) {
		t.Errorf("PathTo(H) = %v, want %v", pathLabels(p), want)
	}
	for label, want := range map[string]float64{"S": 0, "A": 1, "C": 1, "B": 2, "G": 2, "H": 3} {
		if d, ok := paths.Distance(label); !ok || d != want {
			t.Errorf("Distance(%s) = %v, %v, want %v and true", label, d, ok, want)
		}
	}
}

func TestBFSPaths_EqualPathsKeepTheFirstFound(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "B", 1}, weightedEdge{"S", "A", 1},
		weightedEdge{"A", "G", 1}, weightedEdge{"B", "G", 1},
	)
	for range 20 {
		paths, err := BFSPaths(g, "S")
		if err != nil {
			t.Fatalf("BFSPaths() error = %v", err)
		}
		p, _ := paths.PathTo("G")
		if want := []string{"S", "B", "G"}; !reflect.DeepEqual(pathLabels(p), want) {
			t.Fatalf("PathTo(G) = %v, want %v", pathLabels(p), want)
		}
	}
}

func TestBFSPaths_Undirected(t *testing.T) {
	g := gograph.New[string]()
	a, b, c := g.AddVertexByLabel("A"), g.AddVertexByLabel("B"), g.AddVertexByLabel("C")
	_, _ = g.AddEdge(a, b)
	_, _ = g.AddEdge(b, c)

	paths, err := BFSPaths(g, "C")
	if err != nil {
		t.Fatalf("BFSPaths() error = %v", err)
	}
	p, _ := assertShortestPath(t, g, paths, "C", "A")
	if want := []string{"C", "B", "A"}; !reflect.DeepEqual(pathLabels(p), want) {
		t.Errorf("PathTo(A) = %v, want %v", pathLabels(p), want)
	}
	if d, ok := paths.Distance("A"); !ok || d != 2 {
		t.Errorf("Distance(A) = %v, %v, want 2 and true", d, ok)
	}
}

// TestBFSPaths_MatchesDijkstraOnUnitWeights compares the edge counts with
// Dijkstra on a copy of each graph whose weights are all 1.
func TestBFSPaths_MatchesDijkstraOnUnitWeights(t *testing.T) {
	rng := rand.New(rand.NewPCG(11, 1)) //nolint:gosec // seeded so failures are reproducible

	for trial := range 3000 {
		g, labels := randomPathGraph(rng, trial%2 == 0)
		var unit gograph.Graph[string]
		if g.IsDirected() {
			unit = gograph.New[string](gograph.Directed(), gograph.Weighted())
		} else {
			unit = gograph.New[string](gograph.Weighted())
		}
		for _, label := range labels {
			unit.AddVertexByLabel(label)
		}
		for e := range gograph.Edges(g) {
			from, to := unit.GetVertexByID(e.Source().Label()), unit.GetVertexByID(e.Destination().Label())
			_, _ = unit.AddEdge(from, to, gograph.WithEdgeWeight(1))
		}

		start := labels[0]
		paths, err := BFSPaths(g, start)
		if err != nil {
			t.Fatalf("trial %d: error = %v", trial, err)
		}
		dist := Dijkstra(unit, start)
		for _, label := range labels {
			d, ok := paths.Distance(label)
			if want := dist[label]; ok != (want != math.MaxFloat64) || (ok && d != want) {
				t.Fatalf("trial %d: Distance(%s) = %v, %v, want %v", trial, label, d, ok, want)
			}
			if !ok {
				continue
			}
			if p, _ := assertShortestPath(t, g, paths, start, label); float64(len(p)-1) != d {
				t.Fatalf("trial %d: PathTo(%s) = %v, want %v edges", trial, label, pathLabels(p), d)
			}
		}
	}
}

func TestBellmanFordPaths_NegativeWeights(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 4}, weightedEdge{"S", "B", 5},
		weightedEdge{"B", "A", -3}, weightedEdge{"A", "G", 2},
	)
	paths, err := BellmanFordPaths(g, "S")
	if err != nil {
		t.Fatalf("BellmanFordPaths() error = %v", err)
	}
	p, sum := assertShortestPath(t, g, paths, "S", "G")
	if want := []string{"S", "B", "A", "G"}; !reflect.DeepEqual(pathLabels(p), want) || sum != 4 {
		t.Errorf("PathTo(G) = %v weighing %v, want %v and 4", pathLabels(p), sum, want)
	}
	if d, ok := paths.Distance("A"); !ok || d != 2 {
		t.Errorf("Distance(A) = %v, %v, want 2 and true", d, ok)
	}
}

func TestBellmanFordPaths_Errors(t *testing.T) {
	directedOnly := gograph.New[string](gograph.Directed())
	directedOnly.AddVertexByLabel("A")
	undirected := gograph.New[string](gograph.Weighted())
	undirected.AddVertexByLabel("A")

	tests := []struct {
		name string
		g    gograph.Graph[string]
		want error
	}{
		{"not weighted", directedOnly, ErrNotWeighted},
		{"not directed", undirected, ErrNotDirected},
		{
			"negative cycle",
			directedWeighted(weightedEdge{"A", "B", 1}, weightedEdge{"B", "C", -2}, weightedEdge{"C", "B", 1}),
			ErrNegativeWeightCycle,
		},
		{
			"cycle through a -Inf weight",
			directedWeighted(weightedEdge{"A", "B", math.Inf(-1)}, weightedEdge{"B", "A", 1}),
			ErrNegativeWeightCycle,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if paths, err := BellmanFordPaths(tt.g, "A"); !errors.Is(err, tt.want) || paths != nil {
				t.Errorf("got %v and %v, want nil and %v", paths, err, tt.want)
			}
		})
	}
}

func TestBellmanFordPaths_FewestEdgesAmongEqualPaths(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "B", 1}, weightedEdge{"B", "G", 1},
		weightedEdge{"S", "C", 2}, weightedEdge{"C", "G", 1}, weightedEdge{"S", "D", 1}, weightedEdge{"D", "G", 2},
	)
	paths, err := BellmanFordPaths(g, "S")
	if err != nil {
		t.Fatalf("BellmanFordPaths() error = %v", err)
	}
	// three paths weigh 3. Two of them have two edges, and C's edge from S
	// comes before D's.
	p, sum := assertShortestPath(t, g, paths, "S", "G")
	if want := []string{"S", "C", "G"}; !reflect.DeepEqual(pathLabels(p), want) || sum != 3 {
		t.Errorf("PathTo(G) = %v weighing %v, want %v and 3", pathLabels(p), sum, want)
	}
}

func TestBellmanFordPaths_ZeroCycle(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 1}, weightedEdge{"A", "B", 2}, weightedEdge{"B", "A", -2}, weightedEdge{"B", "G", 1},
	)
	paths, err := BellmanFordPaths(g, "S")
	if err != nil {
		t.Fatalf("BellmanFordPaths() error = %v", err)
	}
	p, sum := assertShortestPath(t, g, paths, "S", "G")
	if want := []string{"S", "A", "B", "G"}; !reflect.DeepEqual(pathLabels(p), want) || sum != 4 {
		t.Errorf("PathTo(G) = %v weighing %v, want %v and 4", pathLabels(p), sum, want)
	}
}

func TestBellmanFordPaths_NegativeInfinity(t *testing.T) {
	inf := math.Inf(1)
	tests := []struct {
		name  string
		edges []weightedEdge
		want  map[string][]string // the path to each vertex at -Inf
	}{
		{
			name:  "path through a -Inf weight",
			edges: []weightedEdge{{"a", "b", -inf}, {"b", "c", 1}},
			want:  map[string][]string{"b": {"a", "b"}, "c": {"a", "b", "c"}},
		},
		{
			name:  "two -Inf weights in a row",
			edges: []weightedEdge{{"a", "b", -inf}, {"b", "c", -inf}},
			want:  map[string][]string{"b": {"a", "b"}, "c": {"a", "b", "c"}},
		},
		{
			name:  "a finite path to the same vertex is not the shortest",
			edges: []weightedEdge{{"a", "c", -5}, {"a", "b", -inf}, {"b", "c", 1}, {"c", "d", 1}},
			want:  map[string][]string{"b": {"a", "b"}, "c": {"a", "b", "c"}, "d": {"a", "b", "c", "d"}},
		},
		{
			name:  "a -Inf weight behind a finite path",
			edges: []weightedEdge{{"a", "b", 2}, {"b", "c", -inf}, {"c", "d", 1}, {"a", "d", 1}},
			want:  map[string][]string{"c": {"a", "b", "c"}, "d": {"a", "b", "c", "d"}},
		},
		{
			name:  "two -Inf weights into one vertex",
			edges: []weightedEdge{{"a", "b", 1}, {"a", "c", 1}, {"b", "d", -inf}, {"c", "d", -inf}, {"d", "e", 1}},
			want:  map[string][]string{"d": {"a", "b", "d"}, "e": {"a", "b", "d", "e"}},
		},
		{
			// a, c, d has fewer edges, but the path to c is the shortest one
			name:  "the path to the -Inf weight is the shortest, not the one with the fewest edges",
			edges: []weightedEdge{{"a", "c", 10}, {"a", "b", 1}, {"b", "c", 1}, {"c", "d", -inf}},
			want:  map[string][]string{"d": {"a", "b", "c", "d"}},
		},
		{
			name:  "a positive cycle behind a -Inf weight",
			edges: []weightedEdge{{"a", "b", -inf}, {"b", "c", 1}, {"c", "b", 1}},
			want:  map[string][]string{"b": {"a", "b"}, "c": {"a", "b", "c"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := directedWeighted(tt.edges...)
			paths, err := BellmanFordPaths(g, "a")
			if err != nil {
				t.Fatalf("BellmanFordPaths() error = %v", err)
			}
			for label, want := range tt.want {
				if d, ok := paths.Distance(label); !ok || !math.IsInf(d, -1) {
					t.Errorf("Distance(%s) = %v, %v, want -Inf and true", label, d, ok)
				}
				p, sum := assertShortestPath(t, g, paths, "a", label)
				if !reflect.DeepEqual(pathLabels(p), want) || !math.IsInf(sum, -1) {
					t.Errorf("PathTo(%s) = %v weighing %v, want %v and -Inf", label, pathLabels(p), sum, want)
				}
			}
		})
	}
}

// TestBellmanFordPaths_RoundedZeroCycle uses weights that sum to zero in
// decimal and to slightly less in float64. The passes then lower the distances
// around the cycle, and no edge into it adds up to the distance it leads to.
func TestBellmanFordPaths_RoundedZeroCycle(t *testing.T) {
	g := directedWeighted(
		weightedEdge{"S", "A", 0.4}, weightedEdge{"A", "B", 0.1},
		weightedEdge{"B", "C", 0.7}, weightedEdge{"C", "A", -0.8}, weightedEdge{"C", "D", 1},
	)
	paths, err := BellmanFordPaths(g, "S")
	if err != nil {
		t.Fatalf("BellmanFordPaths() error = %v", err)
	}
	dist, err := BellmanFord(g, "S")
	if err != nil {
		t.Fatalf("BellmanFord() error = %v", err)
	}
	if dist["A"] >= 0.4 {
		t.Fatalf("the distance to A is %v, so rounding didn't lower it and the test checks nothing", dist["A"])
	}

	want := map[string][]string{
		"A": {"S", "A"}, "B": {"S", "A", "B"}, "C": {"S", "A", "B", "C"}, "D": {"S", "A", "B", "C", "D"},
	}
	for label, labels := range want {
		d, ok := paths.Distance(label)
		if !ok || d != dist[label] {
			t.Errorf("Distance(%s) = %v, %v, want %v and true", label, d, ok, dist[label])
		}
		p, sum := assertShortestPath(t, g, paths, "S", label)
		if !reflect.DeepEqual(pathLabels(p), labels) || sum < d || sum-d > 1e-12 {
			t.Errorf("PathTo(%s) = %v weighing %v, want %v and just above %v", label, pathLabels(p), sum, labels, d)
		}
	}
}

// TestBellmanFordPaths_MatchesBellmanFord compares random graphs with
// BellmanFord, with negative, infinite and NaN weights among them.
func TestBellmanFordPaths_MatchesBellmanFord(t *testing.T) {
	// whole weights add up exactly, so a path weighs exactly its distance
	weights := []float64{-3, -1, 0, 0, 1, 1, 2, 5, math.Inf(-1), math.Inf(1), math.NaN()}
	rng := rand.New(rand.NewPCG(11, 2)) //nolint:gosec // seeded so failures are reproducible
	searched := 0

	for trial := range 4000 {
		n := 1 + rng.IntN(8)
		g := gograph.New[string](gograph.Directed(), gograph.Weighted())
		for i := range n {
			g.AddVertexByLabel(strconv.Itoa(i))
		}
		for range rng.IntN(2*n + 1) {
			from, to := g.GetVertexByID(strconv.Itoa(rng.IntN(n))), g.GetVertexByID(strconv.Itoa(rng.IntN(n)))
			_, _ = g.AddEdge(from, to, gograph.WithEdgeWeight(weights[rng.IntN(len(weights))]))
		}

		for _, start := range g.GetAllVertices() {
			want, wantErr := BellmanFord(g, start.Label())
			paths, err := BellmanFordPaths(g, start.Label())
			if !errors.Is(err, wantErr) {
				t.Fatalf("trial %d from %s: error = %v, BellmanFord gives %v", trial, start.Label(), err, wantErr)
			}
			if err != nil {
				continue
			}
			searched++

			for label, wantDist := range want {
				d, ok := paths.Distance(label)
				if ok != !math.IsInf(wantDist, 1) || (ok && d != wantDist) {
					t.Fatalf("trial %d from %s: Distance(%s) = %v, %v, BellmanFord gives %v\nedges: %v",
						trial, start.Label(), label, d, ok, wantDist, g.AllEdges())
				}
				if !ok {
					continue
				}
				if p, sum := assertShortestPath(t, g, paths, start.Label(), label); sum != d {
					t.Fatalf("trial %d from %s: PathTo(%s) = %v weighs %v, want %v\nedges: %v",
						trial, start.Label(), label, pathLabels(p), sum, d, g.AllEdges())
				}
			}
		}
	}
	if searched < 4000 {
		t.Errorf("only %d searches had no negative cycle", searched)
	}
}

// TestBellmanFordPaths_RoundedWeightsGiveSimplePaths uses tenths, which round,
// on graphs full of cycles that sum to zero in decimal.
func TestBellmanFordPaths_RoundedWeightsGiveSimplePaths(t *testing.T) {
	rng := rand.New(rand.NewPCG(11, 3)) //nolint:gosec // seeded so failures are reproducible
	lowered := 0

	for trial := range 4000 {
		n := 3 + rng.IntN(5)
		edges := []weightedEdge{{"S", "0", float64(rng.IntN(10)) / 10}}
		sum := 0
		for i := range n {
			tenths := 1 + rng.IntN(9)
			if i == n-1 {
				tenths = -sum
			}
			sum += tenths
			edges = append(edges, weightedEdge{strconv.Itoa(i), strconv.Itoa((i + 1) % n), float64(tenths) / 10})
		}
		edges = append(edges, weightedEdge{strconv.Itoa(rng.IntN(n)), "T", float64(rng.IntN(10)) / 10})
		g := directedWeighted(edges...)

		paths, err := BellmanFordPaths(g, "S")
		if err != nil {
			continue // the rounded cycle never settles, as for BellmanFord
		}
		for _, v := range g.GetAllVertices() {
			d, ok := paths.Distance(v.Label())
			if !ok {
				t.Fatalf("trial %d: %s is out of reach\nedges: %v", trial, v.Label(), edges)
			}
			_, weight := assertShortestPath(t, g, paths, "S", v.Label())
			if weight < d || weight-d > 1e-9 {
				t.Fatalf("trial %d: the path to %s weighs %v, want just above %v\nedges: %v",
					trial, v.Label(), weight, d, edges)
			}
			if weight > d {
				lowered++
			}
		}
	}
	if lowered == 0 {
		t.Error("no path weighs more than its distance, so the test checks nothing")
	}
}

func ExampleDijkstraPaths() {
	g := gograph.New[string](gograph.Weighted())
	home := g.AddVertexByLabel("home")
	park := g.AddVertexByLabel("park")
	shop := g.AddVertexByLabel("shop")
	work := g.AddVertexByLabel("work")
	g.AddVertexByLabel("island")

	_, _ = g.AddEdge(home, park, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(park, work, gograph.WithEdgeWeight(2))
	_, _ = g.AddEdge(home, shop, gograph.WithEdgeWeight(1))
	_, _ = g.AddEdge(shop, work, gograph.WithEdgeWeight(5))

	paths, err := DijkstraPaths(g, "home")
	if err != nil {
		fmt.Println(err)
		return
	}

	route, _ := paths.PathTo("work")
	distance, _ := paths.Distance("work")
	fmt.Println(pathLabels(route), distance)

	_, ok := paths.PathTo("island")
	fmt.Println(ok)
	// Output:
	// [home park work] 4
	// false
}

func ExampleBellmanFordPaths() {
	g := gograph.New[string](gograph.Directed(), gograph.Weighted())
	usd := g.AddVertexByLabel("USD")
	eur := g.AddVertexByLabel("EUR")
	gbp := g.AddVertexByLabel("GBP")

	// a fee to pay, or a rebate to collect, for each exchange
	_, _ = g.AddEdge(usd, gbp, gograph.WithEdgeWeight(5))
	_, _ = g.AddEdge(usd, eur, gograph.WithEdgeWeight(4))
	_, _ = g.AddEdge(eur, gbp, gograph.WithEdgeWeight(-2))

	paths, err := BellmanFordPaths(g, "USD")
	if err != nil {
		fmt.Println(err)
		return
	}

	route, _ := paths.PathTo("GBP")
	distance, _ := paths.Distance("GBP")
	fmt.Println(pathLabels(route), distance)
	// Output:
	// [USD EUR GBP] 2
}

func ExampleBFSPaths() {
	g := gograph.New[string]()
	ann := g.AddVertexByLabel("Ann")
	bob := g.AddVertexByLabel("Bob")
	cat := g.AddVertexByLabel("Cat")
	dan := g.AddVertexByLabel("Dan")

	_, _ = g.AddEdge(ann, bob)
	_, _ = g.AddEdge(bob, cat)
	_, _ = g.AddEdge(cat, dan)
	_, _ = g.AddEdge(ann, cat)

	paths, err := BFSPaths(g, "Ann")
	if err != nil {
		fmt.Println(err)
		return
	}

	route, _ := paths.PathTo("Dan")
	steps, _ := paths.Distance("Dan")
	fmt.Println(pathLabels(route), steps)
	// Output:
	// [Ann Cat Dan] 2
}

func BenchmarkShortestPaths(b *testing.B) {
	// every vertex is in reach
	g := randomWeighted(10_000, 50_000)

	// the same edges, cut so that vertex 0 reaches only the first 100 vertices
	island := gograph.New[int](gograph.Directed(), gograph.Weighted())
	for v := range gograph.Vertices(g) {
		island.AddVertexByLabel(v.Label())
	}
	for e := range gograph.Edges(g) {
		from, to := e.Source().Label(), e.Destination().Label()
		if (from < 100) == (to < 100) {
			_, _ = island.AddEdge(island.GetVertexByID(from), island.GetVertexByID(to), gograph.WithEdgeWeight(e.Weight()))
		}
	}
	for i := range 99 {
		_, _ = island.AddEdge(island.GetVertexByID(i), island.GetVertexByID(i+1), gograph.WithEdgeWeight(1))
	}

	for _, bm := range []struct {
		name string
		g    gograph.Graph[int]
	}{
		{"all 10000 in reach", g},
		{"100 of 10000 in reach", island},
	} {
		for _, fn := range []struct {
			name string
			run  func(gograph.Graph[int], int) (*ShortestPaths[int], error)
		}{
			{"DijkstraPaths", DijkstraPaths[int]},
			{"BFSPaths", BFSPaths[int]},
			{"BellmanFordPaths", BellmanFordPaths[int]},
		} {
			b.Run(bm.name+"/"+fn.name, func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := fn.run(bm.g, 0); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
