package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddEdge(t *testing.T) {
	g := newDirectedGraph()
	_, err := g.AddEdge(NewVertex(0), nil)
	assert.Error(t, err)

	v1 := g.AddVertexWithID(1)
	v2 := g.AddVertexWithID(2)
	_, err = g.AddEdge(v1, v2)
	assert.NoError(t, err)
	assert.Len(t, g.vertices[v1.id].neighbors, 1)
	assert.Len(t, g.vertices[v2.id].neighbors, 0)
	assert.Len(t, g.edges, 1)

	destMapV1, existsV1 := g.edges[v1.id]
	assert.True(t, existsV1)
	assert.Len(t, destMapV1, 1)
	assert.Equal(t, v1, destMapV1[v2.id].source)
	assert.Equal(t, v2, destMapV1[v2.id].dest)

	// create the vertices if they don't exist
	edge, err := g.AddEdge(NewVertex(3), NewVertex(4))
	assert.NoError(t, err)
	assert.Len(t, g.vertices[edge.source.id].neighbors, 1)
	assert.Len(t, g.vertices[edge.dest.id].neighbors, 0)
	assert.Len(t, g.edges, 2)

	destMapV3, existsV3 := g.edges[edge.source.id]
	assert.True(t, existsV3)
	assert.Len(t, destMapV3, 1)
	assert.Equal(t, edge.source, destMapV3[edge.dest.id].source)
	assert.Equal(t, edge.dest, destMapV3[edge.dest.id].dest)
}
