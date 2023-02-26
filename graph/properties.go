package graph

type GraphOptionFunc func(properties *GraphProperties)

type GraphProperties struct {
	isDirected bool
	isWeighted bool
	isAcyclic  bool
}

func newProperties(options ...GraphOptionFunc) GraphProperties {
	var properties GraphProperties
	for _, option := range options {
		option(&properties)
	}

	return properties
}

func Acyclic() GraphOptionFunc {
	return func(properties *GraphProperties) {
		properties.isAcyclic = true
		properties.isDirected = true
	}
}

func Directed() GraphOptionFunc {
	return func(properties *GraphProperties) {
		properties.isDirected = true
	}
}

func Weighted() GraphOptionFunc {
	return func(properties *GraphProperties) {
		properties.isWeighted = true
	}
}

type EdgeOptionFunc func(properties *EdgeProperties)

type EdgeProperties struct {
	weight float64
}

func WithVertexWeight(weight float64) EdgeOptionFunc {
	return func(properties *EdgeProperties) {
		properties.weight = weight
	}
}

type VertexOptionFunc func(properties *VertexProperties)

type VertexProperties struct {
	weight float64
}

func WithEdgeWeight(weight float64) VertexOptionFunc {
	return func(properties *VertexProperties) {
		properties.weight = weight
	}
}
