package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddVertex(t *testing.T) {
	g := &DirectedGraph{}
	g.AddVertexWithID(1)
	g.AddVertexWithID(2)
	g.AddVertexWithID(3)
	assert.Len(t, g.vertices, 3)
}

func TestAddEdge(t *testing.T) {
	g := &DirectedGraph{}
	v1 := g.AddVertexWithID(1)
	v2 := g.AddVertexWithID(2)
	_, err := g.AddEdge(v1, v2)
	assert.NoError(t, err)
	assert.Len(t, g.vertices[v1.id].neighbors, 1)
	assert.Len(t, g.vertices[v2.id].neighbors, 0)
}

func TestFindVertex(t *testing.T) {
	g := &DirectedGraph{}
	v1 := g.AddVertexWithID(1)
	v2 := g.AddVertexWithID(2)
	_, err := g.AddEdge(v1, v2)
	assert.NoError(t, err)
	v := g.findVertex(1)
	assert.Equal(t, 1, v.id)
	v = g.findVertex(3)
	assert.Nil(t, v)
}
