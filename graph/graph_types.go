package graph

type GraphType interface {
	IsDirected() bool
	IsAcyclic() bool
	IsWeighted() bool
}

func (g *baseGraph[T]) IsDirected() bool {
	return g.properties.isDirected
}

func (g *baseGraph[T]) IsAcyclic() bool {
	return g.properties.isAcyclic
}

func (g *baseGraph[T]) IsWeighted() bool {
	return g.properties.isWeighted
}
