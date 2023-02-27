package gograph

// GraphOptionFunc represent an alias of function type that
// modifies the specified graph properties.
type GraphOptionFunc func(properties *GraphProperties)

// GraphProperties represents the properties of a graph.
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

// Acyclic returns a GraphOptionFunc that modifies the specified
// graph properties. It sets the isAcyclic and isDirected to true.
// Only direct graph can be acyclic.
func Acyclic() GraphOptionFunc {
	return func(properties *GraphProperties) {
		properties.isAcyclic = true
		properties.isDirected = true
	}
}

// Directed returns a GraphOptionFunc that modifies the specified
// graph properties. It sets the isDirected to true.
func Directed() GraphOptionFunc {
	return func(properties *GraphProperties) {
		properties.isDirected = true
	}
}

// Weighted returns a GraphOptionFunc that modifies the specified
// graph properties. It sets the isWeighted to true.
func Weighted() GraphOptionFunc {
	return func(properties *GraphProperties) {
		properties.isWeighted = true
	}
}

// EdgeOptionFunc represent an alias of function type that
// modifies the specified edge properties.
type EdgeOptionFunc func(properties *EdgeProperties)

// EdgeProperties represents the properties of an edge.
type EdgeProperties struct {
	weight float64
}

// WithEdgeWeight sets the edge weight for the specified edge
// properties in the returned EdgeOptionFunc.
func WithEdgeWeight(weight float64) EdgeOptionFunc {
	return func(properties *EdgeProperties) {
		properties.weight = weight
	}
}

// VertexOptionFunc represent an alias of function type that
// modifies the specified vertex properties.
type VertexOptionFunc func(properties *VertexProperties)

// VertexProperties represents the properties of an edge.
type VertexProperties struct {
	weight float64
}

// WithVertexWeight sets the edge weight for the specified vertex
// properties in the returned VertexOptionFunc.
func WithVertexWeight(weight float64) VertexOptionFunc {
	return func(properties *VertexProperties) {
		properties.weight = weight
	}
}
