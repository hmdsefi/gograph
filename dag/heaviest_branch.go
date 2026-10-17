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
	for _, v := range vertices {
		want := 1
		if v.Label() == root {
			want = 0
		}
		if v.InDegree() != want {
			if _, err := gograph.TopologySort(g); err != nil {
				return nil, err
			}
			return nil, ErrNotTree
		}
	}

	index := make(map[T]int, len(vertices))
	for i, v := range vertices {
		index[v.Label()] = i
	}
	rootIdx := index[root]
	parent := make([]int, len(vertices))
	parent[rootIdx] = -1
	for _, e := range g.AllEdges() {
		parent[index[e.Destination().Label()]] = index[e.Source().Label()]
	}

	// Subtree weights, leaves first: a vertex is summed once all of its
	// children are. Every vertex but root has one parent, so a vertex that is
	// never summed is on a cycle or below one.
	subtree := make([]float64, len(vertices))
	pending := make([]int, len(vertices))
	queue := make([]int, 0, len(vertices))
	for i, v := range vertices {
		subtree[i] = weight(v)
		pending[i] = v.OutDegree()
		if pending[i] == 0 {
			queue = append(queue, i)
		}
	}
	for k := 0; k < len(queue); k++ {
		i := queue[k]
		if p := parent[i]; p >= 0 {
			subtree[p] += subtree[i]
			pending[p]--
			if pending[p] == 0 {
				queue = append(queue, p)
			}
		}
	}
	if len(queue) != len(vertices) {
		return nil, gograph.ErrDAGHasCycle
	}

	// pending is all zeros now, so it can hold the best child of each vertex.
	// Going in GetAllVertices order keeps the earlier child on a tie.
	best := pending
	for i := range best {
		best[i] = -1
	}
	for i, p := range parent {
		if p >= 0 && (best[p] == -1 || subtree[i] > subtree[best[p]]) {
			best[p] = i
		}
	}

	n := 1
	for at := best[rootIdx]; at != -1; at = best[at] {
		n++
	}
	path := make([]*gograph.Vertex[T], 0, n)
	for at := rootIdx; at != -1; at = best[at] {
		path = append(path, vertices[at])
	}
	return path, nil
}
