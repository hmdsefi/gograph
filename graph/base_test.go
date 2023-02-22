package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddVertex(t *testing.T) {
	g := newDirectedGraph()
	g.AddVertex(nil)
	g.AddVertex(NewVertex(0))
	g.AddVertexWithID(1)
	g.AddVertexWithID(2)
	g.AddVertexWithID(3)
	assert.Len(t, g.vertices, 4)
}

func TestFindVertex(t *testing.T) {
	g := newDirectedGraph()
	v1 := g.AddVertexWithID(1)
	v2 := g.AddVertexWithID(2)
	_, err := g.AddEdge(v1, v2)
	assert.NoError(t, err)
	v := g.findVertex(1)
	assert.Equal(t, 1, v.id)
	v = g.findVertex(3)
	assert.Nil(t, v)
}

func TestDirectedGraph_EdgesOf(t *testing.T) {
	g := NewDirectedGraph()
	v1 := g.AddVertexWithID(1)
	v2 := g.AddVertexWithID(2)
	v3 := g.AddVertexWithID(3)
	v4 := g.AddVertexWithID(4)
	v5 := g.AddVertexWithID(5)
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

	edgesV2 := g.EdgesOf(v2)
	require.Len(t, edgesV2, 2)
	assert.Equal(t, v2, edgesV2[0].source)
	assert.Equal(t, v4, edgesV2[0].dest)
	assert.Equal(t, v1, edgesV2[1].source)
	assert.Equal(t, v2, edgesV2[1].dest)

	edgesV4 := g.EdgesOf(v4)
	require.Len(t, edgesV4, 3)
	edgeMap := make(map[int]*Edge)
	for i := range edgesV4 {
		edgeMap[edgesV4[i].source.id] = edgesV4[i]
	}

	assert.Equal(t, v4, edgeMap[v4.id].source)
	assert.Equal(t, v5, edgeMap[v4.id].dest)
	assert.Equal(t, v2, edgeMap[v2.id].source)
	assert.Equal(t, v4, edgeMap[v2.id].dest)
	assert.Equal(t, v3, edgeMap[v3.id].source)
	assert.Equal(t, v4, edgeMap[v3.id].dest)

	edges := g.EdgesOf(nil)
	assert.Nil(t, edges)

	v6 := NewVertex(6)
	edges = g.EdgesOf(v6)
	assert.Nil(t, edges)
}

func TestDirectedGraph_RemoveEdges(t *testing.T) {
	g := newDirectedGraph()
	v1 := g.AddVertexWithID(1)
	v2 := g.AddVertexWithID(2)
	v3 := g.AddVertexWithID(3)
	v4 := g.AddVertexWithID(4)
	v5 := g.AddVertexWithID(5)
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

	g.RemoveEdges(NewEdge(v4, v5))

	assert.Equal(t, 0, v4.outDegree)
	assert.Equal(t, 0, v5.inDegree)
	assert.Len(t, v4.neighbors, 0)

	_, existsV4 := g.edges[v4.id]
	assert.False(t, existsV4)

	g.RemoveEdges(NewEdge(v1, v2), NewEdge(v3, v4))
	assert.Equal(t, 1, v1.outDegree)
	assert.Equal(t, v3, v1.neighbors[0])
	assert.Equal(t, 0, v2.inDegree)
	assert.Equal(t, 0, v3.outDegree)
	assert.Equal(t, 1, v4.inDegree)
	assert.Len(t, v1.neighbors, 1)
	assert.Len(t, v3.neighbors, 0)

	_, existsV3 := g.edges[v3.id]
	assert.False(t, existsV3)

	destMapV1, existsV1 := g.edges[v1.id]
	assert.True(t, existsV1)
	assert.Len(t, destMapV1, 1)

	_, existsV2 := destMapV1[v2.id]
	assert.False(t, existsV2)
}

func TestDirectedGraph_RemoveVertices(t *testing.T) {
	g := newDirectedGraph()

	g.RemoveVertices(nil)
	g.RemoveVertices(NewVertex(0))

	v1 := g.AddVertexWithID(1)
	v2 := g.AddVertexWithID(2)
	v3 := g.AddVertexWithID(3)
	v4 := g.AddVertexWithID(4)
	v5 := g.AddVertexWithID(5)
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
	assert.Equal(t, 1, v1.outDegree)
	assert.Equal(t, v3, v1.neighbors[0])
	assert.Equal(t, 1, v4.inDegree)
	assert.Len(t, v1.neighbors, 1)

	_, existsV2 := g.edges[v2.id]
	assert.False(t, existsV2)

	destMapV1, existsV1 := g.edges[v1.id]
	assert.True(t, existsV1)
	assert.Equal(t, v3, destMapV1[v3.id].dest)
	assert.Len(t, destMapV1, 1)

	g.RemoveVertices(v1, v5)
	assert.Equal(t, 0, v3.inDegree)
	assert.Equal(t, 0, v4.outDegree)

	_, existsV1 = g.edges[v1.id]
	assert.False(t, existsV1)

	_, existsV4 := g.edges[v4.id]
	assert.False(t, existsV4)
}

func TestDirectedGraph_ContainsEdge(t *testing.T) {
	g := newDirectedGraph()

	assert.False(t, g.ContainsEdge(nil, nil))

	v1 := g.AddVertexWithID(1)

	assert.False(t, g.ContainsEdge(NewVertex(0), v1))
	assert.False(t, g.ContainsEdge(nil, v1))

	v2 := g.AddVertexWithID(2)
	v3 := g.AddVertexWithID(3)
	v4 := g.AddVertexWithID(4)
	_, err := g.AddEdge(v1, v2)
	require.NoError(t, err)
	_, err = g.AddEdge(v1, v3)
	require.NoError(t, err)
	_, err = g.AddEdge(v2, v4)
	require.NoError(t, err)
	_, err = g.AddEdge(v3, v4)
	require.NoError(t, err)

	assert.True(t, g.ContainsEdge(v1, v2))
	assert.True(t, g.ContainsEdge(v1, v3))
	assert.True(t, g.ContainsEdge(v2, v4))
	assert.True(t, g.ContainsEdge(v3, v4))

	assert.False(t, g.ContainsEdge(v1, v4))
	assert.False(t, g.ContainsEdge(v2, v3))
	assert.False(t, g.ContainsEdge(v3, v1))

}

func TestDirectedGraph_ContainsVertex(t *testing.T) {
	g := newDirectedGraph()
	v1 := g.AddVertexWithID(1)

	assert.False(t, g.ContainsVertex(nil))
	assert.False(t, g.ContainsVertex(NewVertex(0)))
	assert.True(t, g.ContainsVertex(v1))
}
