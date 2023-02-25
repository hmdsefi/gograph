package graph

type GraphOptionFunc func(properties *GraphProperties)

type GraphProperties struct {
	isDirected bool
	isWeighted bool
	isAcyclic  bool
}

func Acyclic() GraphOptionFunc {
	return func(properties *GraphProperties) {
		properties.isAcyclic = true
		properties.isDirected = true
	}
}

func Weighted() GraphOptionFunc {
	return func(properties *GraphProperties) {
		properties.isWeighted = true
	}
}
