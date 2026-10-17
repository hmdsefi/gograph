package dag

import (
	"errors"
	"fmt"

	"github.com/hmdsefi/gograph"
)

// ErrNotTree is returned by HeaviestBranch when the graph is not a tree rooted
// at the given vertex: some vertex has more than one incoming edge, the root
// has an incoming edge, or some vertex can't be reached from the root.
var ErrNotTree = errors.New("graph is not a tree rooted at the given vertex")

// HeaviestBranch walks a tree from root to a leaf. At each vertex it follows
// the child whose subtree carries the greatest total weight. This is the walk
// that fork choice rules such as GHOST and LMD-GHOST do on a tree of blocks,
// with the weight being work, stake or the latest votes, whatever the caller
// counts.
//
// The weight of a subtree is the weight of its root vertex plus the weights of
// the subtrees of its children, so every vertex is counted once. weight is
// called once per vertex. A nil weight treats every vertex as 0, and then ties
// decide the walk. Negative weights are allowed, they only change which child
// wins. If two children have the same subtree weight, the one that comes
// first in GetAllVertices wins, so the same graph gives the same path on every
// run.
//
// The graph must be a tree rooted at root: root has no incoming edges, every
// other vertex has exactly one, and every vertex is reachable from root. The
// returned path starts at root and ends at a leaf, as the graph's own vertex
// pointers. A tree of one vertex returns that vertex.
//
// It runs in O(V + E) time.
//
// It returns gograph.ErrNotDirected for undirected graphs, an error that
// matches gograph.ErrVertexDoesNotExist with errors.Is if root is not in the
// graph, gograph.ErrDAGHasCycle if the graph has a cycle, and ErrNotTree if
// the graph is acyclic but not a tree rooted at root. It doesn't validate
// votes, stake or finality, those are up to weight.
func HeaviestBranch[T comparable](
	g gograph.Graph[T],
	root T,
	weight func(*gograph.Vertex[T]) float64,
) ([]*gograph.Vertex[T], error) {
	if !g.IsDirected() {
		return nil, gograph.ErrNotDirected
	}
	if g.GetVertexByID(root) == nil {
		return nil, fmt.Errorf("%w: %v", gograph.ErrVertexDoesNotExist, root)
	}
	if weight == nil {
		weight = func(*gograph.Vertex[T]) float64 { return 0 }
	}

	vertices := g.GetAllVertices()
	index := make(map[T]int, len(vertices))
	for i, v := range vertices {
		index[v.Label()] = i
	}
	rootIdx := index[root]

	out := make([][]int, len(vertices))
	parent := make([]int, len(vertices))
	for i := range parent {
		parent[i] = -1
	}
	for _, e := range g.AllEdges() {
		from, to := index[e.Source().Label()], index[e.Destination().Label()]
		out[from] = append(out[from], to)
		parent[to] = from
	}

	// Kahn's algorithm over the whole graph, to report a cycle anywhere.
	waiting := make([]int, len(vertices))
	queue := make([]int, 0, len(vertices))
	for i, v := range vertices {
		waiting[i] = v.InDegree()
		if waiting[i] == 0 {
			queue = append(queue, i)
		}
	}
	for i := 0; i < len(queue); i++ {
		for _, to := range out[queue[i]] {
			waiting[to]--
			if waiting[to] == 0 {
				queue = append(queue, to)
			}
		}
	}
	if len(queue) != len(vertices) {
		return nil, gograph.ErrDAGHasCycle
	}

	// In an acyclic graph where root is the only vertex without a parent and
	// every other vertex has exactly one, following the parents from any
	// vertex ends at root, so every vertex is reachable from it.
	for i, v := range vertices {
		want := 1
		if i == rootIdx {
			want = 0
		}
		if v.InDegree() != want {
			return nil, ErrNotTree
		}
	}

	// Children in GetAllVertices order, which is the tie-break order.
	children := make([][]int, len(vertices))
	for i := range vertices {
		if i != rootIdx {
			children[parent[i]] = append(children[parent[i]], i)
		}
	}

	// Subtree weights, children before parents: reverse topological order.
	subtree := make([]float64, len(vertices))
	for _, i := range queue {
		subtree[i] = weight(vertices[i])
	}
	for k := len(queue) - 1; k >= 0; k-- {
		i := queue[k]
		if i != rootIdx {
			subtree[parent[i]] += subtree[i]
		}
	}

	path := []*gograph.Vertex[T]{vertices[rootIdx]}
	for at := rootIdx; len(children[at]) > 0; {
		best := children[at][0]
		for _, c := range children[at][1:] {
			if subtree[c] > subtree[best] {
				best = c
			}
		}
		at = best
		path = append(path, vertices[at])
	}

	return path, nil
}
