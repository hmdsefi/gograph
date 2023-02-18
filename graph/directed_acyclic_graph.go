package graph

import "errors"

var (
	ErrDAGVertexNotFound = errors.New("vertex not found")
	ErrDAGHasCycle       = errors.New("edge would create cycle")
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
func (d *DAG) AddVertexWithID(id int64) {
	v := &DAGVertex{ID: id}
	d.Vertices = append(d.Vertices, v)
}

func (d *DAG) AddVertex(v *DAGVertex) {
	d.Vertices = append(d.Vertices, v)
}

// AddEdge adds a directed edge from the vertex with the 'from' id to
// the vertex with the 'to' id, after checking if the edge would create
// a cycle.
//
// It returns error if it finds a cycle between 'from' and 'to'.
func (d *DAG) AddEdge(from, to *DAGVertex) error {
	// Add the new edge
	from.Neighbors = append(from.Neighbors, to)
	to.inDegree++

	// Perform a topological sort to check for cycles
	var sortedVertices []*DAGVertex
	queue := make([]*DAGVertex, 0)

	// Add all vertices with inDegree 0 to the queue
	for _, v := range d.Vertices {
		if v.inDegree == 0 {
			queue = append(queue, v)
		}
	}

	// Traverse the graph using Kahn's algorithm
	for len(queue) > 0 {
		// Dequeue a vertex
		v := queue[0]
		queue = queue[1:]

		// Add the vertex to the sorted list
		sortedVertices = append(sortedVertices, v)

		// Decrement the inDegree of all neighbors of the dequeued vertex
		for _, neighbor := range v.Neighbors {
			neighbor.inDegree--
			if neighbor.inDegree == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If the sorted list does not contain all vertices, there is a cycle
	if len(sortedVertices) != len(d.Vertices) {
		// Remove the new edge
		from.Neighbors = from.Neighbors[:len(from.Neighbors)-1]
		to.inDegree--

		return errors.New("adding this edge would create a cycle in the graph")
	}

	return nil
}

func (d *DAG) hasCycle(current, parent *DAGVertex, visited map[*DAGVertex]bool) bool {
	// Mark the current vertex as visited
	visited[current] = true

	// Check all neighbors of the current vertex
	for _, neighbor := range current.Neighbors {
		// If the neighbor is the parent vertex, continue to the next neighbor
		if neighbor == parent {
			continue
		}

		// If the neighbor has already been visited, there is a cycle
		if visited[neighbor] {
			return true
		}

		// Recursively check for cycles in the neighbor's subtree
		if d.hasCycle(neighbor, current, visited) {
			return true
		}
	}

	// No cycle was found
	return false
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
