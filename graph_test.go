package gograph

import (
	"reflect"
	"testing"
)

func TestVertex(t *testing.T) {
	g := New[string]()
	vA := g.AddVertexByLabel("A")
	vB := g.AddVertexByLabel("B")
	vC := g.AddVertexByLabel("C")
	_, err := g.AddEdge(vA, vB)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	_, err = g.AddEdge(vA, vC)
	if err != nil {
		t.Errorf(testErrMsgError, err)
	}

	v := vA.NeighborByLabel("B")
	if !reflect.DeepEqual(vB, v) {
		t.Errorf(testErrMsgNotEqual, vB, v)
	}

	if !vA.HasNeighbor(vC) {
		t.Error(testErrMsgNotTrue)
	}

	if vA.HasNeighbor(NewVertex("D")) {
		t.Error(testErrMsgNotFalse)
	}

	if vA.HasNeighbor(nil) {
		t.Error(testErrMsgNotFalse)
	}

	if vA.OutDegree() != 2 {
		t.Errorf(testErrMsgNotEqual, 2, vA.OutDegree())
	}

	// test cloning neighbors
	neighbors := vA.Neighbors()
	if len(neighbors) != len(vA.neighbors) {
		t.Errorf(testErrMsgNotEqual, len(neighbors), len(vA.neighbors))
	}

	neighbors[0].label = "D"
	if neighbors[0].Label() == vA.neighbors[0].Label() {
		t.Error(testErrMsgNotFalse)
	}
}

func TestNewVertex_Options(t *testing.T) {
	if w := NewVertex("A").Weight(); w != 0 {
		t.Errorf(testErrMsgNotEqual, 0.0, w)
	}

	if w := NewVertex("A", WithVertexWeight(3)).Weight(); w != 3 {
		t.Errorf(testErrMsgNotEqual, 3.0, w)
	}

	if w := NewVertex("A", WithVertexWeight(1), WithVertexWeight(2)).Weight(); w != 2 {
		t.Errorf(testErrMsgNotEqual, 2.0, w)
	}

	g := New[string]()
	g.AddVertex(NewVertex("A", WithVertexWeight(3)))
	if w := g.GetVertexByID("A").Weight(); w != 3 {
		t.Errorf("AddVertex: "+testErrMsgNotEqual, 3.0, w)
	}

	if _, err := g.AddEdge(NewVertex("B", WithVertexWeight(5)), NewVertex("C", WithVertexWeight(7))); err != nil {
		t.Fatalf(testErrMsgError, err)
	}
	if w := g.GetVertexByID("B").Weight(); w != 5 {
		t.Errorf("AddEdge source: "+testErrMsgNotEqual, 5.0, w)
	}
	if w := g.GetVertexByID("C").Weight(); w != 7 {
		t.Errorf("AddEdge destination: "+testErrMsgNotEqual, 7.0, w)
	}
}

func TestEdge_OtherVertex(t *testing.T) {
	edge := NewEdge[int](NewVertex(1), NewVertex(2))

	if edge.OtherVertex(3) != nil {
		t.Errorf("Expect OtherVertex return nil, but get %+v", edge.OtherVertex(3))
	}

	if edge.OtherVertex(1).label != edge.Destination().Label() {
		t.Errorf("Expect OtherVertex return 2, but get %+v", edge.OtherVertex(1))
	}

	if edge.OtherVertex(2).label != edge.Source().Label() {
		t.Errorf("Expect OtherVertex return 1, but get %+v", edge.OtherVertex(2))
	}
}
