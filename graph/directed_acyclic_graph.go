package graph

import "errors"

var (
	ErrDAGCycle    = errors.New("edge would create cycle")
	ErrDAGHasCycle = errors.New("the graph contains a cycle")
)

type DAGVertex struct {
	ID        int64
	Neighbors []*DAGVertex
	inDegree  int
}

// DAG represents a directed graph that has no cycles. It is a graph
// where there is no path that starts and ends at the same vertex.
type DAG struct {
	// Vertices represents nodes or points in the graph
	Vertices []*DAGVertex
}

func NewDAG() *DAG {
	return &DAG{}
}

// AddVertexWithID adds a new vertex with the given id to the graph.
func (d *DAG) AddVertexWithID(id int64) *DAGVertex {
	v := &DAGVertex{ID: id}
	d.Vertices = append(d.Vertices, v)

	return v
}

func (d *DAG) AddVertex(v *DAGVertex) {
	d.Vertices = append(d.Vertices, v)
}

// AddEdge adds a directed edge from the vertex with the 'from' id to
// the vertex with the 'to' id, after checking if the edge would create
// a cycle.
//
// AddEdge guarantees that the graph remain DAG after adding new edge.
//
// It returns error if it finds a cycle between 'from' and 'to'.
func (d *DAG) AddEdge(from, to *DAGVertex) error {
	// Add the new edge
	from.Neighbors = append(from.Neighbors, to)
	to.inDegree++

	// If topological sort returns an error, new edge created a cycle
	_, err := d.TopologySort()
	if err != nil {
		// Remove the new edge
		from.Neighbors = from.Neighbors[:len(from.Neighbors)-1]
		to.inDegree--

		return ErrDAGCycle
	}

	return nil
}

// TopologySort performs a topological sort of the graph using
// Kahn's algorithm. If the sorted list of vertices does not contain
// all vertices in the graph, it means there is a cycle in the graph.
//
// It returns error if it finds a cycle in the graph.
func (d *DAG) TopologySort() ([]*DAGVertex, error) {
	// Initialize a map to store the inDegree of each vertex
	inDegrees := make(map[*DAGVertex]int)
	for _, v := range d.Vertices {
		inDegrees[v] = v.inDegree
	}

	// Initialize a queue with vertices of inDegrees zero
	queue := make([]*DAGVertex, 0)
	for v, inDegree := range inDegrees {
		if inDegree == 0 {
			queue = append(queue, v)
		}
	}

	// Initialize the sorted list of vertices
	sortedVertices := make([]*DAGVertex, 0)

	// Loop through the vertices with inDegree zero
	for len(queue) > 0 {
		// Get the next vertex with inDegree zero
		curr := queue[0]
		queue = queue[1:]

		// Add the vertex to the sorted list
		sortedVertices = append(sortedVertices, curr)

		// Decrement the inDegree of each of the vertex's neighbors
		for _, neighbor := range curr.Neighbors {
			inDegrees[neighbor]--
			if inDegrees[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If the sorted list does not contain all vertices, there is a cycle
	if len(sortedVertices) != len(d.Vertices) {
		return nil, ErrDAGHasCycle
	}

	return sortedVertices, nil
}

// findVertex searches for the given id in the vertices. It returns
// a pointer to the vertex if it finds it. Otherwise, returns nil.
func (d *DAG) findVertex(id int64) *DAGVertex {
	for _, v := range d.Vertices {
		if v.ID == id {
			return v
		}
	}
	return nil
}
