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
