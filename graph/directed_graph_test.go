package graph

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddVertex(t *testing.T) {
	g := &directedGraph{}
	g.AddVertexWithID(1)
	g.AddVertexWithID(2)
	g.AddVertexWithID(3)
	assert.Len(t, g.vertices, 3)
}

func TestAddEdge(t *testing.T) {
	g := &directedGraph{}
	v1 := g.AddVertexWithID(1)
	v2 := g.AddVertexWithID(2)
	_, err := g.AddEdge(v1, v2)
	assert.NoError(t, err)
	assert.Len(t, g.vertices[v1.id].neighbors, 1)
	assert.Len(t, g.vertices[v2.id].neighbors, 0)
}

func TestFindVertex(t *testing.T) {
	g := &directedGraph{}
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
	require.Equal(t, 2, len(edgesV2))
	assert.Equal(t, v2, edgesV2[0].source)
	assert.Equal(t, v4, edgesV2[0].dest)
	assert.Equal(t, v1, edgesV2[1].source)
	assert.Equal(t, v2, edgesV2[1].dest)

	edgesV4 := g.EdgesOf(v4)
	require.Equal(t, 3, len(edgesV4))
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
	assert.Equal(t, 0, len(v4.neighbors))
	assert.Equal(t, 0, v5.inDegree)

	_, existsV4 := g.edges[v4.id]
	assert.False(t, existsV4)

	g.RemoveEdges(NewEdge(v1, v2), NewEdge(v3, v4))
	assert.Equal(t, 1, v1.outDegree)
	assert.Equal(t, 1, len(v1.neighbors))
	assert.Equal(t, v3, v1.neighbors[0])
	assert.Equal(t, 0, v2.inDegree)
	assert.Equal(t, 0, v3.outDegree)
	assert.Equal(t, 0, len(v3.neighbors))
	assert.Equal(t, 1, v4.inDegree)

	_, existsV3 := g.edges[v3.id]
	assert.False(t, existsV3)

	destMapV1, existsV1 := g.edges[v1.id]
	assert.True(t, existsV1)
	assert.Equal(t, 1, len(destMapV1))

	_, existsV2 := destMapV1[v2.id]
	assert.False(t, existsV2)
}

func TestDirectedGraph_RemoveVertices(t *testing.T) {
	type fields struct {
		vertices map[int]*Vertex
		edges    map[int]map[int]*Edge
	}
	type args struct {
		vertices []*Vertex
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &directedGraph{
				vertices: tt.fields.vertices,
				edges:    tt.fields.edges,
			}
			g.RemoveVertices(tt.args.vertices...)
		})
	}
}

func TestDirectedGraph_ContainsEdge(t *testing.T) {
	type fields struct {
		vertices map[int]*Vertex
		edges    map[int]map[int]*Edge
	}
	type args struct {
		from *Vertex
		to   *Vertex
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &directedGraph{
				vertices: tt.fields.vertices,
				edges:    tt.fields.edges,
			}
			assert.Equalf(t, tt.want, g.ContainsEdge(tt.args.from, tt.args.to), "ContainsEdge(%v, %v)", tt.args.from, tt.args.to)
		})
	}
}

func TestDirectedGraph_ContainsVertex(t *testing.T) {
	type fields struct {
		vertices map[int]*Vertex
		edges    map[int]map[int]*Edge
	}
	type args struct {
		v *Vertex
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &directedGraph{
				vertices: tt.fields.vertices,
				edges:    tt.fields.edges,
			}
			assert.Equalf(t, tt.want, g.ContainsVertex(tt.args.v), "ContainsVertex(%v)", tt.args.v)
		})
	}
}
