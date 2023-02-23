package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddVertex(t *testing.T) {
	g := newDirectedGraph[string]()
	g.AddVertex(nil)
	g.AddVertex(NewVertex("morocco"))
	g.AddVertexByLabel("london")
	g.AddVertexByLabel("berlin")
	g.AddVertexByLabel("paris")
	assert.Len(t, g.vertices, 4)
}

func TestFindVertex(t *testing.T) {
	g := newDirectedGraph[string]()
	v1 := g.AddVertexByLabel("morocco")
	v2 := g.AddVertexByLabel("paris")
	_, err := g.AddEdge(v1, v2)
	assert.NoError(t, err)
	v := g.findVertex("morocco")
	assert.Equal(t, v1.label, v.label)
	v = g.findVertex("london")
	assert.Nil(t, v)
}

func TestDirectedGraph_EdgesOf(t *testing.T) {
	g := NewDirectedGraph[int]()
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

	edgesV2 := g.EdgesOf(v2)
	require.Len(t, edgesV2, 2)
	assert.Equal(t, v2, edgesV2[0].source)
	assert.Equal(t, v4, edgesV2[0].dest)
	assert.Equal(t, v1, edgesV2[1].source)
	assert.Equal(t, v2, edgesV2[1].dest)

	edgesV4 := g.EdgesOf(v4)
	require.Len(t, edgesV4, 3)
	edgeMap := make(map[int]*Edge[int])
	for i := range edgesV4 {
		edgeMap[edgesV4[i].source.label] = edgesV4[i]
	}

	assert.Equal(t, v4, edgeMap[v4.label].source)
	assert.Equal(t, v5, edgeMap[v4.label].dest)
	assert.Equal(t, v2, edgeMap[v2.label].source)
	assert.Equal(t, v4, edgeMap[v2.label].dest)
	assert.Equal(t, v3, edgeMap[v3.label].source)
	assert.Equal(t, v4, edgeMap[v3.label].dest)

	edges := g.EdgesOf(nil)
	assert.Nil(t, edges)

	v6 := NewVertex(6)
	edges = g.EdgesOf(v6)
	assert.Nil(t, edges)
}

func TestDirectedGraph_RemoveEdges(t *testing.T) {
	g := newDirectedGraph[int]()
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

	g.RemoveEdges(NewEdge(v4, v5))

	assert.Equal(t, 0, v4.outDegree)
	assert.Equal(t, 0, v5.inDegree)
	assert.Len(t, v4.neighbors, 0)

	_, existsV4 := g.edges[v4.label]
	assert.False(t, existsV4)

	g.RemoveEdges(NewEdge(v1, v2), NewEdge(v3, v4))
	assert.Equal(t, 1, v1.outDegree)
	assert.Equal(t, v3, v1.neighbors[0])
	assert.Equal(t, 0, v2.inDegree)
	assert.Equal(t, 0, v3.outDegree)
	assert.Equal(t, 1, v4.inDegree)
	assert.Len(t, v1.neighbors, 1)
	assert.Len(t, v3.neighbors, 0)

	_, existsV3 := g.edges[v3.label]
	assert.False(t, existsV3)

	destMapV1, existsV1 := g.edges[v1.label]
	assert.True(t, existsV1)
	assert.Len(t, destMapV1, 1)

	_, existsV2 := destMapV1[v2.label]
	assert.False(t, existsV2)
}

func TestDirectedGraph_RemoveVertices(t *testing.T) {
	g := newDirectedGraph[int]()

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
	assert.Equal(t, 1, v1.outDegree)
	assert.Equal(t, v3, v1.neighbors[0])
	assert.Equal(t, 1, v4.inDegree)
	assert.Len(t, v1.neighbors, 1)

	_, existsV2 := g.edges[v2.label]
	assert.False(t, existsV2)

	destMapV1, existsV1 := g.edges[v1.label]
	assert.True(t, existsV1)
	assert.Equal(t, v3, destMapV1[v3.label].dest)
	assert.Len(t, destMapV1, 1)

	g.RemoveVertices(v1, v5)
	assert.Equal(t, 0, v3.inDegree)
	assert.Equal(t, 0, v4.outDegree)

	_, existsV1 = g.edges[v1.label]
	assert.False(t, existsV1)

	_, existsV4 := g.edges[v4.label]
	assert.False(t, existsV4)
}

func TestDirectedGraph_ContainsEdge(t *testing.T) {
	g := newDirectedGraph[int]()

	assert.False(t, g.ContainsEdge(nil, nil))

	v1 := g.AddVertexByLabel(1)

	assert.False(t, g.ContainsEdge(NewVertex(0), v1))
	assert.False(t, g.ContainsEdge(nil, v1))

	v2 := g.AddVertexByLabel(2)
	v3 := g.AddVertexByLabel(3)
	v4 := g.AddVertexByLabel(4)
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
	g := newDirectedGraph[int]()
	v1 := g.AddVertexByLabel(1)

	assert.False(t, g.ContainsVertex(nil))
	assert.False(t, g.ContainsVertex(NewVertex(0)))
	assert.True(t, g.ContainsVertex(v1))
}
