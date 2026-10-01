package connectivity

import "github.com/hmdsefi/gograph"

// Condensation is a graph with one vertex per strongly connected component
// of another graph. It never has a cycle, so the functions that need a DAG,
// such as gograph.TopologySort, work on it.
type Condensation[T comparable] struct {
	// Graph has one vertex per strongly connected component, labeled 0 to
	// n-1 so that every edge goes from a lower label to a higher one. It is
	// directed and has no cycles.
	Graph gograph.Graph[int]

	// Members has the vertices of each component, indexed by component
	// label, as the original graph's own vertex pointers.
	Members [][]*gograph.Vertex[T]

	// ComponentOf maps the label of each original vertex to the label of
	// its component.
	ComponentOf map[T]int
}

type condenseOptions[T comparable] struct {
	componentWeight func(members []*gograph.Vertex[T]) float64
	crossingWeight  func(crossing []*gograph.Edge[T]) float64
}

// CondenseOption configures Condense.
type CondenseOption[T comparable] func(*condenseOptions[T])

// WithComponentWeight sets the weight of each component vertex, for example
// the sum of the weights of its members. fn gets the members in the same
// order as Condensation.Members.
func WithComponentWeight[T comparable](fn func(members []*gograph.Vertex[T]) float64) CondenseOption[T] {
	return func(o *condenseOptions[T]) {
		o.componentWeight = fn
	}
}

// WithCrossingWeight sets the weight of each edge between two components
// from the original edges that cross between them, for example their
// minimum or their sum. fn gets the crossing edges in the order of
// AllEdges. With this option, Condensation.Graph is weighted.
func WithCrossingWeight[T comparable](fn func(crossing []*gograph.Edge[T]) float64) CondenseOption[T] {
	return func(o *condenseOptions[T]) {
		o.crossingWeight = fn
	}
}

// Condense replaces each strongly connected component of g with a single
// vertex. A component with more than one member is a group of vertices that
// depend on each other, for example modules that have to be upgraded in one
// step.
//
// There is an edge from component a to component b if at least one edge of
// g goes from a member of a to a member of b. Edges inside a component,
// including self-loops, are dropped. Without options, Condensation.Graph is
// unweighted and its vertices have no weight.
//
// The components are numbered in topological order, and the members of
// each component follow the order of GetAllVertices. The result is stable:
// a graph built the same way gives the same result on every run.
//
// Condensation.Graph is created with gograph.Directed and not
// gograph.Acyclic, since an acyclic graph sorts itself on every AddEdge. It
// has no cycles all the same.
//
// It runs Tarjan once and takes O(V + E) time. It returns
// gograph.ErrNotDirected for undirected graphs.
func Condense[T comparable](g gograph.Graph[T], opts ...CondenseOption[T]) (*Condensation[T], error) {
	if !g.IsDirected() {
		return nil, gograph.ErrNotDirected
	}

	var o condenseOptions[T]
	for _, opt := range opts {
		opt(&o)
	}

	// Tarjan returns the components in reverse topological order
	sccs := Tarjan(g)
	n := len(sccs)
	componentOf := make(map[T]int, g.Order())
	for i, scc := range sccs {
		for _, v := range scc {
			componentOf[v.Label()] = n - 1 - i
		}
	}

	members := make([][]*gograph.Vertex[T], n)
	for _, v := range g.GetAllVertices() {
		c := componentOf[v.Label()]
		members[c] = append(members[c], v)
	}

	type componentPair struct{ from, to int }
	var pairs []componentPair
	crossing := make(map[componentPair][]*gograph.Edge[T])
	for _, e := range g.AllEdges() {
		p := componentPair{componentOf[e.Source().Label()], componentOf[e.Destination().Label()]}
		if p.from == p.to {
			continue
		}
		if _, ok := crossing[p]; !ok {
			pairs = append(pairs, p)
			crossing[p] = nil
		}
		if o.crossingWeight != nil {
			crossing[p] = append(crossing[p], e)
		}
	}

	graphOptions := []gograph.GraphOptionFunc{gograph.Directed()}
	if o.crossingWeight != nil {
		graphOptions = append(graphOptions, gograph.Weighted())
	}
	condensed := gograph.New[int](graphOptions...)

	vertices := make([]*gograph.Vertex[int], n)
	for i := range n {
		var vertexOptions []gograph.VertexOptionFunc
		if o.componentWeight != nil {
			vertexOptions = append(vertexOptions, gograph.WithVertexWeight(o.componentWeight(members[i])))
		}
		vertices[i] = condensed.AddVertexByLabel(i, vertexOptions...)
	}
	for _, p := range pairs {
		var edgeOptions []gograph.EdgeOptionFunc
		if o.crossingWeight != nil {
			edgeOptions = append(edgeOptions, gograph.WithEdgeWeight(o.crossingWeight(crossing[p])))
		}
		_, _ = condensed.AddEdge(vertices[p.from], vertices[p.to], edgeOptions...)
	}

	return &Condensation[T]{Graph: condensed, Members: members, ComponentOf: componentOf}, nil
}
