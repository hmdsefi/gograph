package gograph

import (
	"reflect"
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

	if v1.Degree() != 1 {
		t.Errorf("Expected Degree() returns 1, but got %d", v1.Degree())
	}
}
