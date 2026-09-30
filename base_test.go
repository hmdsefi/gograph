package gograph

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

const (
	testErrMsgError    = "Expected no error, but got %s"
	testErrMsgNoError  = "Expected error, but got no error"
	testErrMsgWrongLen = "Expected len %d, but got %d"
	testErrMsgNotFalse = "Expected false, but got true"
	testErrMsgNotTrue  = "Expected true, but got false"
	testErrMsgNotEqual = "Expected %+v, but got %+v"
)

func TestAddVertex(t *testing.T) {
	g := newBaseGraph[string](newProperties(Directed(), Weighted()))
	g.AddVertex(nil)

	// default wait is zero for the vertices
	g.AddVertex(NewVertex("morocco"))
	g.AddVertexByLabel("london")
	g.AddVertexByLabel("berlin")
	g.AddVertexByLabel("paris")
	if len(g.GetAllVertices()) != 4 {
		t.Errorf(testErrMsgWrongLen, 4, len(g.vertices))
	}

	v := g.AddVertexByLabel("madrid", WithVertexWeight(1))
	if v.Weight() != 1 {
		t.Errorf(testErrMsgNotEqual, 1, v.Weight())
	}
}

func TestFindVertex(t *testing.T) {
	g := newBaseGraph[string](newProperties(Directed()))
	v1 := g.AddVertexByLabel("morocco")
	v2 := g.AddVertexByLabel("paris")
	_, err := g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	v := g.findVertex("morocco")
	if v1.label != v.Label() {
		t.Errorf(testErrMsgNotEqual, v1.label, v.label)
	}

	v = g.GetVertexByID("london")
	if v != nil {
		t.Errorf("expected nil vertex, but got %+v", v)
	}
}

func TestBaseGraph_AddEdgeDirected(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))
	_, err := g.AddEdge(NewVertex(0), nil)
	if err == nil {
		t.Error(testErrMsgNoError)
	}

	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	_, err = g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	if len(g.vertices[v1.label].neighbors) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(g.vertices[v1.label].neighbors))
	}

	if len(g.vertices[v2.label].neighbors) != 0 {
		t.Errorf(testErrMsgWrongLen, 0, len(g.vertices[v2.label].neighbors))
	}

	if len(g.edges) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(g.edges))
	}

	destMapV1, existsV1 := g.edges[v1.label]
	if !existsV1 {
		t.Error(testErrMsgNotTrue)
	}
	if len(destMapV1) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(destMapV1))
	}

	if !reflect.DeepEqual(v1, destMapV1[v2.label].source) {
		t.Errorf(testErrMsgNotEqual, v1, destMapV1[v2.label].source)
	}
	if !reflect.DeepEqual(v2, destMapV1[v2.label].dest) {
		t.Errorf(testErrMsgNotEqual, v2, destMapV1[v2.label].dest)
	}

	// create the vertices if they don't exist
	edge, err := g.AddEdge(NewVertex(3), NewVertex(4))
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	if len(g.vertices[edge.source.label].neighbors) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(g.vertices[edge.source.label].neighbors))
	}
	if len(g.vertices[edge.dest.label].neighbors) != 0 {
		t.Errorf(testErrMsgWrongLen, 0, len(g.vertices[edge.dest.label].neighbors))
	}
	if len(g.edges) != 2 {
		t.Errorf(testErrMsgWrongLen, 2, len(g.edges))
	}

	destMapV3, existsV3 := g.edges[edge.source.label]
	if !existsV3 {
		t.Error(testErrMsgNotTrue)
	}
	if len(destMapV3) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(destMapV3))
	}

	if !reflect.DeepEqual(edge.source, destMapV3[edge.dest.label].source) {
		t.Errorf(testErrMsgNotEqual, edge.source, destMapV3[edge.dest.label].source)
	}
	if !reflect.DeepEqual(edge.dest, destMapV3[edge.dest.label].dest) {
		t.Errorf(testErrMsgNotEqual, edge.dest, destMapV3[edge.dest.label].dest)
	}
}

func TestBaseGraph_AddEdgeAcyclic(t *testing.T) {
	// Create a new dag
	g := newBaseGraph[int](newProperties(Acyclic()))

	if !g.IsDirected() {
		t.Error(testErrMsgNotTrue)
	}

	if !g.IsAcyclic() {
		t.Error(testErrMsgNotTrue)
	}

	// Create three vertices with labels 1, 2, and 3
	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)

	// Add the vertices to the dag
	g.AddVertex(v1)
	g.AddVertex(v2)
	g.AddVertex(v3)

	expectedVerticesCount := uint32(3)
	actual := g.Order()
	if actual != expectedVerticesCount {
		t.Errorf("expected (%d) but got (%d)", expectedVerticesCount, actual)
	}

	// Add edges from 1 to 2 and from 2 to 3
	_, err := g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = g.AddEdge(v2, v3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Try to add an edges from 3 to 1, which should result in an error
	_, err = g.AddEdge(v3, v1)
	if err == nil {
		t.Error("Expected error, but got none")
	}

	expectedEdgesCount := uint32(2)
	actual = g.Size()
	if actual != expectedEdgesCount {
		t.Errorf("expected (%d) but got (%d)", expectedEdgesCount, actual)
	}
}

func TestBaseGraph_AddEdgeWeighted(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed(), Weighted()))
	_, err := g.AddEdge(NewVertex(0), nil)
	if err == nil {
		t.Error(testErrMsgNoError)
	}

	if !g.IsDirected() {
		t.Error(testErrMsgNotTrue)
	}

	if !g.IsWeighted() {
		t.Error(testErrMsgNotTrue)
	}

	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	_, err = g.AddEdge(v1, v2, WithEdgeWeight(4))
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	if len(g.vertices[v1.label].neighbors) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(g.vertices[v1.label].neighbors))
	}

	if len(g.vertices[v2.label].neighbors) != 0 {
		t.Errorf(testErrMsgWrongLen, 0, len(g.vertices[v2.label].neighbors))
	}

	if len(g.edges) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(g.edges))
	}

	destMapV1, existsV1 := g.edges[v1.label]
	if !existsV1 {
		t.Error(testErrMsgNotTrue)
	}
	if len(destMapV1) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(destMapV1))
	}

	if !reflect.DeepEqual(v1, destMapV1[v2.label].source) {
		t.Errorf(testErrMsgNotEqual, v1, destMapV1[v2.label].source)
	}
	if !reflect.DeepEqual(v2, destMapV1[v2.label].dest) {
		t.Errorf(testErrMsgNotEqual, v2, destMapV1[v2.label].dest)
	}

	if destMapV1[v2.label].Weight() != 4 {
		t.Errorf(testErrMsgNotEqual, 4, destMapV1[v2.label].Weight())
	}
}

func TestBaseGraph_EdgesOf(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))
	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)
	v5 := g.AddVertexByLabel(5)
	_, err := g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v1, v3)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v2, v4)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v3, v4)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v4, v5)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	edgesV2 := g.EdgesOf(v2)
	if len(edgesV2) != 2 {
		t.Errorf(testErrMsgWrongLen, 2, len(edgesV2))
	}

	if !reflect.DeepEqual(v2, edgesV2[0].source) {
		t.Errorf(testErrMsgNotEqual, v2, edgesV2[0].source)
	}
	if !reflect.DeepEqual(v4, edgesV2[0].dest) {
		t.Errorf(testErrMsgNotEqual, v4, edgesV2[0].dest)
	}
	if !reflect.DeepEqual(v1, edgesV2[1].source) {
		t.Errorf(testErrMsgNotEqual, v1, edgesV2[1].source)
	}
	if !reflect.DeepEqual(v2, edgesV2[1].dest) {
		t.Errorf(testErrMsgNotEqual, v2, edgesV2[1].dest)
	}

	edgesV4 := g.EdgesOf(v4)
	if len(edgesV4) != 3 {
		t.Errorf(testErrMsgWrongLen, 3, len(edgesV4))
	}

	edgeMap := make(map[int]*Edge[int])
	for i := range edgesV4 {
		edgeMap[edgesV4[i].source.label] = edgesV4[i]
	}

	if !reflect.DeepEqual(v4, edgeMap[v4.label].source) {
		t.Errorf(testErrMsgNotEqual, v4, edgeMap[v4.label].source)
	}
	if !reflect.DeepEqual(v5, edgeMap[v4.label].dest) {
		t.Errorf(testErrMsgNotEqual, v5, edgeMap[v4.label].dest)
	}
	if !reflect.DeepEqual(v2, edgeMap[v2.label].source) {
		t.Errorf(testErrMsgNotEqual, v2, edgeMap[v2.label].source)
	}
	if !reflect.DeepEqual(v4, edgeMap[v2.label].dest) {
		t.Errorf(testErrMsgNotEqual, v4, edgeMap[v2.label].dest)
	}
	if !reflect.DeepEqual(v3, edgeMap[v3.label].source) {
		t.Errorf(testErrMsgNotEqual, v3, edgeMap[v3.label].source)
	}
	if !reflect.DeepEqual(v4, edgeMap[v3.label].dest) {
		t.Errorf(testErrMsgNotEqual, v4, edgeMap[v3.label].dest)
	}

	edges := g.EdgesOf(nil)
	if edges != nil {
		t.Errorf("expected nil edges, but got %+v", edges)
	}

	v6 := NewVertex(6)
	edges = g.EdgesOf(v6)
	if edges != nil {
		t.Errorf("expected nil edges, but got %+v", edges)
	}
}

func TestBaseGraph_RemoveEdges(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))
	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)
	v5 := g.AddVertexByLabel(5)
	_, err := g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v1, v3)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v2, v4)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v3, v4)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v4, v5)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	g.RemoveEdges(nil, NewEdge[int](v1, nil), NewEdge[int](nil, v1))
	g.RemoveEdges(NewEdge(v4, v5))

	if v5.InDegree() != 0 {
		t.Errorf(testErrMsgNotEqual, 0, v5.InDegree())
	}
	if len(v4.neighbors) != 0 {
		t.Errorf(testErrMsgWrongLen, 0, len(v4.neighbors))
	}

	_, existsV4 := g.edges[v4.label]
	if existsV4 {
		t.Error(t, testErrMsgNotFalse)
	}

	g.RemoveEdges(NewEdge(v1, v2), NewEdge(v3, v4))
	if !reflect.DeepEqual(v3, v1.neighbors[0]) {
		t.Errorf(testErrMsgNotEqual, v3, v1.neighbors[0])
	}
	if v2.InDegree() != 0 {
		t.Errorf(testErrMsgNotEqual, 0, v2.InDegree())
	}
	if v4.InDegree() != 1 {
		t.Errorf(testErrMsgNotEqual, 1, v4.InDegree())
	}
	if len(v1.neighbors) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(v1.neighbors))
	}
	if len(v3.neighbors) != 0 {
		t.Errorf(testErrMsgWrongLen, 0, len(v3.neighbors))
	}

	_, existsV3 := g.edges[v3.label]
	if existsV3 {
		t.Error(t, testErrMsgNotFalse)
	}

	destMapV1, existsV1 := g.edges[v1.label]
	if !existsV1 {
		t.Error(testErrMsgNotTrue)
	}
	if len(destMapV1) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(destMapV1))
	}

	_, existsV2 := destMapV1[v2.label]
	if existsV2 {
		t.Error(t, testErrMsgNotFalse)
	}
}

func assertEdgeCount[T comparable](t *testing.T, g *baseGraph[T], want int) {
	t.Helper()

	if int(g.Size()) != want {
		t.Errorf("Expected Size() %d, but got %d", want, g.Size())
	}
	if len(g.AllEdges()) != want {
		t.Errorf("Expected %d edges in AllEdges(), but got %d", want, len(g.AllEdges()))
	}
}

func TestBaseGraph_RemoveEdgesNotInGraph(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))
	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	e12, _ := g.AddEdge(v1, v2)
	e13, _ := g.AddEdge(v1, v3)
	e23, _ := g.AddEdge(v2, v3)

	g.RemoveEdges(e12)
	g.RemoveEdges(e12)
	assertEdgeCount(t, g, 2)

	g.RemoveEdges(NewEdge(v2, v1))
	assertEdgeCount(t, g, 2)

	if len(v1.neighbors) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(v1.neighbors))
	}
	if len(v2.neighbors) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(v2.neighbors))
	}
	if v3.InDegree() != 2 {
		t.Errorf(testErrMsgNotEqual, 2, v3.InDegree())
	}

	g.RemoveEdges(e13, e23)
	assertEdgeCount(t, g, 0)
}

func TestBaseGraph_RemoveEdgesUndirectedBothDirections(t *testing.T) {
	tests := []struct {
		name   string
		remove func(g *baseGraph[int], v1, v2 *Vertex[int]) []*Edge[int]
		want   int
	}{
		{
			name: "all edges",
			remove: func(g *baseGraph[int], _, _ *Vertex[int]) []*Edge[int] {
				return g.AllEdges()
			},
			want: 0,
		},
		{
			name: "edges of a vertex",
			remove: func(g *baseGraph[int], v1, _ *Vertex[int]) []*Edge[int] {
				return g.EdgesOf(v1)
			},
			want: 2,
		},
		{
			name: "all edges between two vertices",
			remove: func(g *baseGraph[int], v1, v2 *Vertex[int]) []*Edge[int] {
				return g.GetAllEdges(v1, v2)
			},
			want: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newBaseGraph[int](newProperties())
			v1 := g.AddVertexByLabel(1)
			v2 := g.AddVertexByLabel(2)
			v3 := g.AddVertexByLabel(3)
			_, _ = g.AddEdge(v1, v2)
			_, _ = g.AddEdge(v1, v3)
			_, _ = g.AddEdge(v2, v3)
			assertEdgeCount(t, g, 6)

			g.RemoveEdges(tt.remove(g, v1, v2)...)
			assertEdgeCount(t, g, tt.want)
		})
	}
}

func TestBaseGraph_SizeAfterRandomEdgeChanges(t *testing.T) {
	for _, directed := range []bool{true, false} {
		var g *baseGraph[int]
		if directed {
			g = newBaseGraph[int](newProperties(Directed()))
		} else {
			g = newBaseGraph[int](newProperties())
		}

		const order = 8
		for i := 0; i < order; i++ {
			g.AddVertexByLabel(i)
		}

		rng := rand.New(rand.NewSource(1)) //nolint:gosec // seeded so failures are reproducible
		var removed []*Edge[int]
		for step := 0; step < 3000; step++ {
			from := g.GetVertexByID(rng.Intn(order))
			to := g.GetVertexByID(rng.Intn(order))

			switch rng.Intn(5) {
			case 0:
				_, _ = g.AddEdge(from, to)
			case 1:
				if edge := g.GetEdge(from, to); edge != nil {
					removed = append(removed, edge)
					g.RemoveEdges(edge)
				}
			case 2:
				if len(removed) > 0 {
					g.RemoveEdges(removed[rng.Intn(len(removed))])
				}
			case 3:
				g.RemoveEdges(g.EdgesOf(from)...)
			case 4:
				g.RemoveVertices(from)
				g.AddVertexByLabel(from.Label())
			}

			checkEdgeCounts(t, g, fmt.Sprintf("directed=%v step %d", directed, step))
		}
	}
}

// checkEdgeCounts fails the test if Size(), AllEdges() and the degrees of the
// vertices don't agree with the edges stored in the graph.
func checkEdgeCounts[T comparable](t *testing.T, g *baseGraph[T], msg string) {
	t.Helper()

	if int(g.Size()) != len(g.AllEdges()) {
		t.Fatalf("%s: Size() is %d, but AllEdges() has %d edges", msg, g.Size(), len(g.AllEdges()))
	}

	inDegrees := make(map[T]int)
	for _, edge := range g.AllEdges() {
		inDegrees[edge.Destination().Label()]++
	}

	for _, v := range g.GetAllVertices() {
		if v.OutDegree() != len(g.edges[v.Label()]) {
			t.Fatalf(
				"%s: vertex %v has OutDegree() %d, but %d outgoing edges",
				msg, v.Label(), v.OutDegree(), len(g.edges[v.Label()]),
			)
		}

		if v.InDegree() != inDegrees[v.Label()] {
			t.Fatalf(
				"%s: vertex %v has InDegree() %d, but %d incoming edges",
				msg, v.Label(), v.InDegree(), inDegrees[v.Label()],
			)
		}
	}
}

func TestBaseGraph_UndirectedSelfLoop(t *testing.T) {
	for _, withOtherEdge := range []bool{false, true} {
		t.Run(fmt.Sprintf("with other edge %v", withOtherEdge), func(t *testing.T) {
			g := newBaseGraph[string](newProperties())
			a := g.AddVertexByLabel("A")
			b := g.AddVertexByLabel("B")

			// An undirected edge between two vertices is stored in both directions.
			otherEdges := 0
			if withOtherEdge {
				_, _ = g.AddEdge(a, b)
				otherEdges = 2
			}

			loop, err := g.AddEdge(a, a)
			if err != nil {
				t.Fatalf(testErrMsgError, err)
			}

			checkEdgeCounts(t, g, "after adding the loop")
			if int(g.Size()) != otherEdges+1 {
				t.Errorf(testErrMsgNotEqual, otherEdges+1, g.Size())
			}

			if got := len(g.GetAllEdges(a, a)); got != 1 {
				t.Errorf(testErrMsgWrongLen, 1, got)
			}

			if _, err = g.AddEdge(a, a); !errors.Is(err, ErrEdgeAlreadyExists) {
				t.Errorf(testErrMsgNotEqual, ErrEdgeAlreadyExists, err)
			}

			g.RemoveEdges(loop)
			checkEdgeCounts(t, g, "after removing the loop")
			if int(g.Size()) != otherEdges {
				t.Errorf(testErrMsgNotEqual, otherEdges, g.Size())
			}

			if a.HasNeighbor(a) {
				t.Error(testErrMsgNotFalse)
			}

			if _, err = g.AddEdge(a, a); err != nil {
				t.Fatalf(testErrMsgError, err)
			}

			checkEdgeCounts(t, g, "after adding the loop again")
			if int(g.Size()) != otherEdges+1 {
				t.Errorf(testErrMsgNotEqual, otherEdges+1, g.Size())
			}

			g.RemoveVertices(a)
			checkEdgeCounts(t, g, "after removing the vertex")
			if g.Size() != 0 {
				t.Errorf(testErrMsgNotEqual, 0, g.Size())
			}
		})
	}
}

func TestBaseGraph_RemoveVertices(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))

	g.RemoveVertices(nil)
	g.RemoveVertices(NewVertex(0))

	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)
	v5 := g.AddVertexByLabel(5)
	_, err := g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v1, v3)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v2, v4)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v3, v4)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v4, v5)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	g.RemoveVertices(v2)
	checkEdgeCounts(t, g, "after removing 2")
	if g.Size() != 3 {
		t.Errorf(testErrMsgNotEqual, 3, g.Size())
	}

	if !reflect.DeepEqual(v3, v1.neighbors[0]) {
		t.Errorf(testErrMsgNotEqual, v3, v1.neighbors[0])
	}
	if v4.InDegree() != 1 {
		t.Errorf(testErrMsgNotEqual, 0, v4.InDegree())
	}

	if len(v1.neighbors) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(v1.neighbors))
	}

	_, existsV2 := g.edges[v2.label]
	if existsV2 {
		t.Error(t, testErrMsgNotFalse)
	}

	destMapV1, existsV1 := g.edges[v1.label]
	if !existsV1 {
		t.Error(testErrMsgNotTrue)
	}
	if !reflect.DeepEqual(v3, destMapV1[v3.label].dest) {
		t.Errorf(testErrMsgNotEqual, v3, destMapV1[v3.label].dest)
	}
	if len(destMapV1) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(destMapV1))
	}

	g.RemoveVertices(v1, v5)
	checkEdgeCounts(t, g, "after removing 1 and 5")
	if g.Size() != 1 {
		t.Errorf(testErrMsgNotEqual, 1, g.Size())
	}

	if v3.InDegree() != 0 {
		t.Errorf(testErrMsgNotEqual, 0, v3.InDegree())
	}
	if len(v4.neighbors) != 0 {
		t.Errorf(testErrMsgWrongLen, 0, len(v4.neighbors))
	}

	_, existsV1 = g.edges[v1.label]
	if existsV1 {
		t.Error(t, testErrMsgNotFalse)
	}

	_, existsV4 := g.edges[v4.label]
	if existsV4 {
		t.Error(t, testErrMsgNotFalse)
	}
}

func TestBaseGraph_ContainsEdge(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))

	if !g.IsDirected() {
		t.Error(testErrMsgNotTrue)
	}

	if g.ContainsEdge(nil, nil) {
		t.Error(t, testErrMsgNotFalse)
	}

	v1 := g.AddVertexByLabel(1)

	if g.ContainsEdge(NewVertex(0), v1) {
		t.Error(t, testErrMsgNotFalse)
	}

	if g.ContainsEdge(nil, v1) {
		t.Error(t, testErrMsgNotFalse)
	}

	if g.ContainsEdge(v1, nil) {
		t.Error(t, testErrMsgNotFalse)
	}

	if g.ContainsEdge(v1, nil) {
		t.Error(t, testErrMsgNotFalse)
	}

	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)
	_, err := g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v1, v3)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v2, v4)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v3, v4)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	if !g.ContainsEdge(v1, v2) {
		t.Error(testErrMsgNotTrue)
	}
	if !g.ContainsEdge(v1, v3) {
		t.Error(testErrMsgNotTrue)
	}
	if !g.ContainsEdge(v2, v4) {
		t.Error(testErrMsgNotTrue)
	}
	if !g.ContainsEdge(v3, v4) {
		t.Error(testErrMsgNotTrue)
	}

	if g.ContainsEdge(v1, v4) {
		t.Error(t, testErrMsgNotFalse)
	}
	if g.ContainsEdge(v2, v3) {
		t.Error(t, testErrMsgNotFalse)
	}
	if g.ContainsEdge(v3, v1) {
		t.Error(t, testErrMsgNotFalse)
	}
	if g.ContainsEdge(v3, NewVertex(5)) {
		t.Error(t, testErrMsgNotFalse)
	}
}

func TestBaseGraph_ContainsVertex(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))
	v1 := g.AddVertexByLabel(1)

	if g.ContainsVertex(nil) {
		t.Error(t, testErrMsgNotFalse)
	}
	if g.ContainsVertex(NewVertex(0)) {
		t.Error(t, testErrMsgNotFalse)
	}

	if !g.ContainsVertex(v1) {
		t.Error(testErrMsgNotTrue)
	}
}

func TestBaseGraph_RemoveEdgesUndirected(t *testing.T) {
	g := newBaseGraph[int](newProperties())

	g.RemoveVertices(nil)
	g.RemoveVertices(NewVertex(0))

	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)
	_, err := g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v1, v3)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v3, v4)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v2, v4)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	g.RemoveEdges(NewEdge(v2, v1))
	if !reflect.DeepEqual(v3, v1.neighbors[0]) {
		t.Errorf(testErrMsgNotEqual, v3, v1.neighbors[0])
	}
	if v4.InDegree() != 2 {
		t.Errorf(testErrMsgNotEqual, 2, v4.InDegree())
	}
	if len(v1.neighbors) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(v1.neighbors))
	}
	if len(v4.neighbors) != 2 {
		t.Errorf(testErrMsgWrongLen, 2, len(v4.neighbors))
	}

	destMap, existsV2 := g.edges[v2.label]
	if !existsV2 {
		t.Error(testErrMsgNotTrue)
	}

	_, existsV4 := destMap[v3.label]
	if existsV4 {
		t.Error(t, testErrMsgNotFalse)
	}

	destMapV1, existsV1 := g.edges[v1.label]
	if !existsV1 {
		t.Error(testErrMsgNotTrue)
	}
	if !reflect.DeepEqual(v3, destMapV1[v3.label].dest) {
		t.Errorf(testErrMsgNotEqual, v3, destMapV1[v3.label].dest)
	}
	if len(destMapV1) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(destMapV1))
	}

	g.RemoveEdges(NewEdge(v1, v3), NewEdge(v4, v3))
	if v3.InDegree() != 0 {
		t.Errorf(testErrMsgNotEqual, 0, v3.InDegree())
	}
	if v4.InDegree() != 1 {
		t.Errorf(testErrMsgNotEqual, 1, v4.InDegree())
	}
	if len(v4.neighbors) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(v4.neighbors))
	}

	_, existsV1 = g.edges[v1.label]
	if existsV1 {
		t.Error(t, testErrMsgNotFalse)
	}

	destMap, existsV4 = g.edges[v4.label]
	if !existsV4 {
		t.Error(testErrMsgNotTrue)
	}

	_, existsV3 := destMap[v3.label]
	if existsV3 {
		t.Error(t, testErrMsgNotFalse)
	}
}

func TestBaseGraph_RemoveVerticesUndirected(t *testing.T) {
	g := newBaseGraph[int](newProperties())

	g.RemoveVertices(nil)
	g.RemoveVertices(NewVertex(0))

	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)
	v5 := g.AddVertexByLabel(5)
	_, err := g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v1, v3)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v2, v4)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v3, v4)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v4, v5)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	g.RemoveVertices(v2)
	checkEdgeCounts(t, g, "after removing 2")

	if !reflect.DeepEqual(v3, v1.neighbors[0]) {
		t.Errorf(testErrMsgNotEqual, v3, v1.neighbors[0])
	}
	if v4.InDegree() != 2 {
		t.Errorf(testErrMsgNotEqual, 2, v4.InDegree())
	}

	if len(v1.neighbors) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(v1.neighbors))
	}

	_, existsV2 := g.edges[v2.label]
	if existsV2 {
		t.Error(t, testErrMsgNotFalse)
	}

	destMapV1, existsV1 := g.edges[v1.label]
	if !existsV1 {
		t.Error(testErrMsgNotTrue)
	}

	if !reflect.DeepEqual(v3, destMapV1[v3.label].dest) {
		t.Errorf(testErrMsgNotEqual, v3, destMapV1[v3.label].dest)
	}
	if len(destMapV1) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(destMapV1))
	}

	g.RemoveVertices(v1, v5)
	if v3.InDegree() != 1 {
		t.Errorf(testErrMsgNotEqual, 1, v3.InDegree())
	}
	if len(v4.neighbors) != 1 {
		t.Errorf(testErrMsgWrongLen, 1, len(v4.neighbors))
	}

	_, existsV1 = g.edges[v1.label]
	if existsV1 {
		t.Error(t, testErrMsgNotFalse)
	}

	destMap, existsV4 := g.edges[v4.label]
	if !existsV4 {
		t.Error(testErrMsgNotTrue)
	}

	_, existsV3 := destMap[v3.label]
	if !existsV3 {
		t.Error(testErrMsgNotTrue)
	}
}

func TestBaseGraph_GetAllEdges(t *testing.T) {
	g := newBaseGraph[int](newProperties())
	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	_, err := g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}
	_, err = g.AddEdge(v1, v3)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(v1, v3)
	if err == nil {
		t.Error(testErrMsgNoError)
	}

	edges := g.GetAllEdges(v1, v2)
	if len(edges) != 2 {
		t.Errorf(testErrMsgWrongLen, 2, len(edges))
	}

	edges = g.GetAllEdges(v2, v1)
	if len(edges) != 2 {
		t.Errorf(testErrMsgWrongLen, 2, len(edges))
	}

	edges = g.GetAllEdges(v2, v3)
	if len(edges) != 0 {
		t.Errorf(testErrMsgWrongLen, 0, len(edges))
	}

	edges = g.GetAllEdges(v2, nil)
	if edges != nil {
		t.Errorf("Expected nil, but got %+v", edges)
	}

	edges = g.GetAllEdges(nil, v2)
	if edges != nil {
		t.Errorf("Expected nil, but got %+v", edges)
	}

	edges = g.GetAllEdges(v2, NewVertex(4))
	if edges != nil {
		t.Errorf("Expected nil, but got %+v", edges)
	}

	edges = g.GetAllEdges(NewVertex(4), v1)
	if edges != nil {
		t.Errorf("Expected nil, but got %+v", edges)
	}
}

func TestBaseGraph_GetEdge(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))
	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	e, err := g.AddEdge(v1, v2)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	edge := g.GetEdge(v1, v2)
	if !reflect.DeepEqual(e, edge) {
		t.Errorf(testErrMsgNotEqual, e, edge)
	}

	edge = g.GetEdge(v2, v1)
	if edge != nil {
		t.Errorf("Expected nil, but got %+v", edge)
	}

	edge = g.GetEdge(v1, nil)
	if edge != nil {
		t.Errorf("Expected nil, but got %+v", edge)
	}

	edge = g.GetEdge(nil, v2)
	if edge != nil {
		t.Errorf("Expected nil, but got %+v", edge)
	}

	edge = g.GetEdge(v2, NewVertex(4))
	if edge != nil {
		t.Errorf("Expected nil, but got %+v", edge)
	}

	edge = g.GetEdge(NewVertex(4), v1)
	if edge != nil {
		t.Errorf("Expected nil, but got %+v", edge)
	}

	if v1.Degree() != 1 {
		t.Errorf("Expected Degree() returns 1, but got %d", v1.Degree())
	}
}

func TestBaseGraph_GetAllVerticesByID(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))
	expected := []*Vertex[int]{
		g.AddVertexByLabel(1),
		g.AddVertexByLabel(2),
		g.AddVertexByLabel(3),
		g.AddVertexByLabel(4),
	}

	vertices := g.GetAllVerticesByID(1, 2, 3, 4, 5)
	if len(vertices) != 4 {
		t.Errorf(testErrMsgWrongLen, 4, len(vertices))
	}

	for i, vertex := range vertices {
		if expected[i].Label() != vertex.Label() {
			t.Errorf(testErrMsgNotEqual, expected[i].Label(), vertex.Label())
		}
	}
}

func Test_baseGraph_ContainsVertex(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))
	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)

	_, _ = g.AddEdge(v1, v2)
	_, _ = g.AddEdge(v2, v3)
	_, _ = g.AddEdge(v2, v4)

	edges := g.AllEdges()

	if len(edges) != 3 {
		t.Errorf("expected len to be %d, but receive %d", 3, len(edges))
	}
}

func Test_baseGraph_Cyclic(t *testing.T) {
	graph := New[int](Acyclic())

	_, err := graph.AddEdge(NewVertex(1), NewVertex(2))
	if err != nil {
		t.Fatal(err)
	}
	_, err = graph.AddEdge(NewVertex(2), NewVertex(3))
	if err != nil {
		t.Fatal(err)
	}
	_, err = graph.AddEdge(NewVertex(3), NewVertex(1))
	if err == nil {
		t.Fatalf("expected error, but got nil")
	}

	if !errors.Is(err, ErrDAGCycle) {
		t.Errorf("expected error %s, but got %s", ErrDAGCycle, err)
	}
}

// assertNeighborsInGraph fails if a vertex of g lists a neighbor that g
// doesn't contain, or a neighbor that isn't backed by an edge of g.
func assertNeighborsInGraph[T comparable](t *testing.T, name string, g Graph[T]) {
	t.Helper()

	for _, v := range g.GetAllVertices() {
		for _, n := range v.Neighbors() {
			if !g.ContainsVertex(n) {
				t.Errorf("%s: %v lists neighbor %v, which isn't in the graph", name, v.Label(), n.Label())
				continue
			}
			if g.GetEdge(v, n) == nil {
				t.Errorf("%s: %v lists neighbor %v, but there's no edge", name, v.Label(), n.Label())
			}
		}
	}
}

func neighborLabels[T comparable](v *Vertex[T]) []T {
	var labels []T
	for _, n := range v.Neighbors() {
		labels = append(labels, n.Label())
	}

	return labels
}

func TestBaseGraph_AddEdgeWithVerticesFromAnotherGraph(t *testing.T) {
	src := New[string](Directed())
	a := src.AddVertexByLabel("A")
	b := src.AddVertexByLabel("B")
	_, _ = src.AddEdge(a, b)

	dst := New[string](Directed())
	for _, e := range src.AllEdges() {
		if _, err := dst.AddEdge(e.Source(), e.Destination()); err != nil {
			t.Fatal(err)
		}
	}
	if dst.GetVertexByID("A") == a {
		t.Error("dst stores the vertex of src")
	}

	c := dst.AddVertexByLabel("C")
	if _, err := dst.AddEdge(dst.GetVertexByID("A"), c); err != nil {
		t.Fatal(err)
	}

	if src.Size() != 1 || a.OutDegree() != 1 || b.InDegree() != 1 {
		t.Errorf("src changed: Size %d, A out-degree %d, B in-degree %d", src.Size(), a.OutDegree(), b.InDegree())
	}
	if dst.Size() != 2 || dst.GetVertexByID("A").OutDegree() != 2 {
		t.Errorf("dst: Size %d, A out-degree %d", dst.Size(), dst.GetVertexByID("A").OutDegree())
	}
	assertNeighborsInGraph(t, "src", src)
	assertNeighborsInGraph(t, "dst", dst)
}

func TestBaseGraph_AddVertexFromAnotherGraph(t *testing.T) {
	full := New[string](Directed())
	p := full.AddVertexByLabel("P")
	q := full.AddVertexByLabel("Q")
	r := full.AddVertexByLabel("R")
	_, _ = full.AddEdge(p, q)
	_, _ = full.AddEdge(q, r)

	sub := New[string](Directed())
	sub.AddVertex(p)
	sub.AddVertex(q)

	if sub.Order() != 2 || sub.Size() != 0 {
		t.Errorf("sub: Order %d, Size %d", sub.Order(), sub.Size())
	}
	for _, v := range sub.GetAllVertices() {
		if v.OutDegree() != 0 || v.InDegree() != 0 {
			t.Errorf("sub: %v has out-degree %d and in-degree %d", v.Label(), v.OutDegree(), v.InDegree())
		}
	}
	if p.OutDegree() != 1 || q.InDegree() != 1 || q.OutDegree() != 1 {
		t.Error("full changed")
	}
	assertNeighborsInGraph(t, "sub", sub)
}

func TestBaseGraph_AddEdgeAcyclicWithVertexFromAnotherGraph(t *testing.T) {
	full := New[string](Directed())
	p := full.AddVertexByLabel("P")
	q := full.AddVertexByLabel("Q")
	_, _ = full.AddEdge(p, q)

	dag := New[string](Acyclic())
	if _, err := dag.AddEdge(q, NewVertex("S")); err != nil {
		t.Fatalf("AddEdge(Q, S): %v", err)
	}
	if _, err := dag.AddEdge(NewVertex("T"), NewVertex("U")); err != nil {
		t.Fatalf("AddEdge(T, U): %v", err)
	}
	if _, err := TopologySort[string](dag); err != nil {
		t.Fatalf("TopologySort: %v", err)
	}

	if !reflect.DeepEqual(neighborLabels(q), []string(nil)) || q.InDegree() != 1 {
		t.Errorf("full changed: Q neighbors %v, in-degree %d", neighborLabels(q), q.InDegree())
	}
	assertNeighborsInGraph(t, "dag", dag)
}

func TestBaseGraph_AddVertexAfterRemoval(t *testing.T) {
	g := New[string](Directed())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	_, _ = g.AddEdge(a, b)

	g.RemoveVertices(a)
	g.AddVertex(a)

	added := g.GetVertexByID("A")
	if added.OutDegree() != 0 || g.ContainsEdge(added, b) {
		t.Errorf("A came back with neighbors %v", neighborLabels(added))
	}
	if a.OutDegree() != 1 {
		t.Errorf("the removed vertex lost its neighbors: %v", neighborLabels(a))
	}
	assertNeighborsInGraph(t, "g", g)
}

func TestBaseGraph_AddEdgeWithNeighborCopy(t *testing.T) {
	g1 := New[string](Directed())
	x := g1.AddVertexByLabel("X")
	a := g1.AddVertexByLabel("A")
	_, _ = g1.AddEdge(x, a)
	for _, label := range []string{"B", "C", "D"} {
		_, _ = g1.AddEdge(a, g1.AddVertexByLabel(label))
	}

	g2 := New[string](Directed())
	if _, err := g2.AddEdge(x.Neighbors()[0], g2.AddVertexByLabel("E")); err != nil {
		t.Fatal(err)
	}
	_, _ = g1.AddEdge(a, g1.AddVertexByLabel("F"))

	if got := neighborLabels(g2.GetVertexByID("A")); !reflect.DeepEqual(got, []string{"E"}) {
		t.Errorf("A in g2 has neighbors %v, expected [E]", got)
	}

	got := neighborLabels(a)
	sort.Strings(got)
	if !reflect.DeepEqual(got, []string{"B", "C", "D", "F"}) {
		t.Errorf("A in g1 has neighbors %v, expected [B C D F]", got)
	}
	assertNeighborsInGraph(t, "g1", g1)
	assertNeighborsInGraph(t, "g2", g2)
}

func TestBaseGraph_AddVertexCopyKeepsProperties(t *testing.T) {
	g1 := New[string]()
	v := g1.AddVertexByLabel("A", WithVertexWeight(5))
	v.metadata = "data"

	g2 := New[string]()
	g2.AddVertex(v)

	got := g2.GetVertexByID("A")
	if got == v {
		t.Fatal("g2 stores the vertex of g1")
	}
	if got.Weight() != 5 || got.Metadata() != "data" {
		t.Errorf("copy has weight %v and metadata %v", got.Weight(), got.Metadata())
	}
}

func TestBaseGraph_AddVertexKeepsNewVertex(t *testing.T) {
	g := New[int](Directed())
	v := NewVertex(1)
	g.AddVertex(v)
	if g.GetVertexByID(1) != v {
		t.Error("AddVertex didn't store the vertex it was given")
	}

	from, to := NewVertex(2), NewVertex(3)
	edge, err := g.AddEdge(from, to)
	if err != nil {
		t.Fatal(err)
	}
	if g.GetVertexByID(2) != from || g.GetVertexByID(3) != to {
		t.Error("AddEdge didn't store the vertices it was given")
	}
	if edge.Source() != from || edge.Destination() != to {
		t.Error("the edge doesn't point to the graph's vertices")
	}
}
