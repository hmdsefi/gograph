package graph

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUndirected_RemoveEdges(t *testing.T) {
	g := newUndirected(newBaseGraph[int](newProperties()))

	g.RemoveVertices(nil)
	g.RemoveVertices(NewVertex(0))

	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)
	_, err := g.AddEdge(v1, v2)
	require.NoError(t, err)
	_, err = g.AddEdge(v1, v3)
	require.NoError(t, err)
	_, err = g.AddEdge(v3, v4)
	require.NoError(t, err)
	_, err = g.AddEdge(v2, v4)
	require.NoError(t, err)

	g.RemoveEdges(NewEdge(v2, v1))
	assert.Equal(t, v3, v1.neighbors[0])
	assert.Equal(t, 2, v4.inDegree)
	assert.Len(t, v1.neighbors, 1)
	assert.Len(t, v4.neighbors, 2)

	destMap, existsV2 := g.edges[v2.label]
	assert.True(t, existsV2)
	_, existsV4 := destMap[v3.label]
	assert.False(t, existsV4)

	destMapV1, existsV1 := g.edges[v1.label]
	assert.True(t, existsV1)
	assert.Equal(t, v3, destMapV1[v3.label].dest)
	assert.Len(t, destMapV1, 1)

	g.RemoveEdges(NewEdge(v1, v3), NewEdge(v4, v3))
	assert.Equal(t, 0, v3.inDegree)
	assert.Equal(t, 1, v4.inDegree)
	assert.Len(t, v4.neighbors, 1)

	_, existsV1 = g.edges[v1.label]
	assert.False(t, existsV1)

	destMap, existsV4 = g.edges[v4.label]
	assert.True(t, existsV4)
	_, existsV3 := destMap[v3.label]
	assert.False(t, existsV3)
}

func TestUndirected_RemoveVertices(t *testing.T) {
	g := newUndirected(newBaseGraph[int](newProperties()))

	g.RemoveVertices(nil)
	g.RemoveVertices(NewVertex(0))

	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)
	v5 := g.AddVertexByLabel(5)
	_, err := g.AddEdge(v1, v2)
	require.NoError(t, err)
	_, err = g.AddEdge(v1, v3)
	require.NoError(t, err)
	_, err = g.AddEdge(v2, v4)
	require.NoError(t, err)
	_, err = g.AddEdge(v3, v4)
	require.NoError(t, err)
	_, err = g.AddEdge(v4, v5)
	require.NoError(t, err)

	g.RemoveVertices(v2)
	assert.Equal(t, v3, v1.neighbors[0])
	assert.Equal(t, 2, v4.inDegree)
	assert.Len(t, v1.neighbors, 1)

	_, existsV2 := g.edges[v2.label]
	assert.False(t, existsV2)

	destMapV1, existsV1 := g.edges[v1.label]
	assert.True(t, existsV1)
	assert.Equal(t, v3, destMapV1[v3.label].dest)
	assert.Len(t, destMapV1, 1)

	g.RemoveVertices(v1, v5)
	assert.Equal(t, 1, v3.inDegree)
	assert.Len(t, v4.neighbors, 1)

	_, existsV1 = g.edges[v1.label]
	assert.False(t, existsV1)

	destMap, existsV4 := g.edges[v4.label]
	assert.True(t, existsV4)
	_, existsV3 := destMap[v3.label]
	assert.True(t, existsV3)
}
