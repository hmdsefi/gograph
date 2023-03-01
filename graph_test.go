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
		t.Errorf(testErrMsgNotFalse)
	}
}
