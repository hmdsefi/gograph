package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddVertex(t *testing.T) {
	g := &Graph{}
	g.AddVertex(1)
	g.AddVertex(2)
	g.AddVertex(3)
	assert.Len(t, g.Vertices, 3)
}

func TestAddEdge(t *testing.T) {
	g := &Graph{}
	g.AddVertex(1)
	g.AddVertex(2)
	g.AddEdge(1, 2)
	assert.Len(t, g.Vertices[0].Neighbors, 1)
	assert.Len(t, g.Vertices[1].Neighbors, 0)
}

func TestFindVertex(t *testing.T) {
	g := &Graph{}
	g.AddVertex(1)
	g.AddVertex(2)
	g.AddEdge(1, 2)
	v := g.findVertex(1)
	assert.Equal(t, 1, v.ID)
	v = g.findVertex(3)
	assert.Nil(t, v)
}
