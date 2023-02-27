package gograph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddVertex(t *testing.T) {
	g := newBaseGraph[string](newProperties(Directed()))
	g.AddVertex(nil)
	g.AddVertex(NewVertex("morocco"))
	g.AddVertexByLabel("london")
	g.AddVertexByLabel("berlin")
	g.AddVertexByLabel("paris")
	assert.Len(t, g.vertices, 4)
}

func TestFindVertex(t *testing.T) {
	g := newBaseGraph[string](newProperties(Directed()))
	v1 := g.AddVertexByLabel("morocco")
	v2 := g.AddVertexByLabel("paris")
	_, err := g.AddEdge(v1, v2)
	assert.NoError(t, err)
	v := g.findVertex("morocco")
	assert.Equal(t, v1.label, v.label)
	v = g.findVertex("london")
	assert.Nil(t, v)
}

func TestBaseGraph_AddEdgeDirected(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))
	_, err := g.AddEdge(NewVertex(0), nil)
	assert.Error(t, err)

	v1 := g.AddVertexByLabel(1)
	v2 := g.AddVertexByLabel(2)
	_, err = g.AddEdge(v1, v2)
	assert.NoError(t, err)
	assert.Len(t, g.vertices[v1.label].neighbors, 1)
	assert.Len(t, g.vertices[v2.label].neighbors, 0)
	assert.Len(t, g.edges, 1)

	destMapV1, existsV1 := g.edges[v1.label]
	assert.True(t, existsV1)
	assert.Len(t, destMapV1, 1)
	assert.Equal(t, v1, destMapV1[v2.label].source)
	assert.Equal(t, v2, destMapV1[v2.label].dest)

	// create the vertices if they don't exist
	edge, err := g.AddEdge(NewVertex(3), NewVertex(4))
	assert.NoError(t, err)
	assert.Len(t, g.vertices[edge.source.label].neighbors, 1)
	assert.Len(t, g.vertices[edge.dest.label].neighbors, 0)
	assert.Len(t, g.edges, 2)

	destMapV3, existsV3 := g.edges[edge.source.label]
	assert.True(t, existsV3)
	assert.Len(t, destMapV3, 1)
	assert.Equal(t, edge.source, destMapV3[edge.dest.label].source)
	assert.Equal(t, edge.dest, destMapV3[edge.dest.label].dest)
}

func TestBaseGraph_AddEdgeAcyclic(t *testing.T) {
	// Create a new dag
	g := newBaseGraph[int](newProperties(Acyclic()))

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

func TestBaseGraph_EdgesOf(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))
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

func TestBaseGraph_RemoveEdges(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))
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

	assert.Equal(t, 0, v5.inDegree)
	assert.Len(t, v4.neighbors, 0)
	assert.Len(t, v4.neighbors, 0)

	_, existsV4 := g.edges[v4.label]
	assert.False(t, existsV4)

	g.RemoveEdges(NewEdge(v1, v2), NewEdge(v3, v4))
	assert.Equal(t, v3, v1.neighbors[0])
	assert.Equal(t, 0, v2.inDegree)
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
	assert.Len(t, v4.neighbors, 0)

	_, existsV1 = g.edges[v1.label]
	assert.False(t, existsV1)

	_, existsV4 := g.edges[v4.label]
	assert.False(t, existsV4)
}

func TestBaseGraph_ContainsEdge(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))

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

func TestBaseGraph_ContainsVertex(t *testing.T) {
	g := newBaseGraph[int](newProperties(Directed()))
	v1 := g.AddVertexByLabel(1)

	assert.False(t, g.ContainsVertex(nil))
	assert.False(t, g.ContainsVertex(NewVertex(0)))
	assert.True(t, g.ContainsVertex(v1))
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
