package path

import (
	"errors"
	"math"

	"github.com/hmdsefi/gograph"
)

var (
	ErrNegativeWeightCycle = errors.New("graph contains negative weight cycle")
)

func BellmanFord[T comparable](g gograph.Graph[T], start T) (map[T]float64, error) {
	vertices := g.GetAllVertices()
	edges := g.AllEdges()

	dist := make(map[T]float64)
	maxValue := math.Inf(1)
	for _, v := range vertices {
		dist[v.Label()] = maxValue
	}

	dist[start] = 0
	for i := 1; i < len(vertices); i++ {
		for _, edge := range edges {
			weight := edge.Weight()
			if dist[edge.Source().Label()] != maxValue &&
				dist[edge.Source().Label()]+weight < dist[edge.Destination().Label()] {
				dist[edge.Destination().Label()] = dist[edge.Source().Label()] + weight
			}
		}
	}

	for _, edge := range edges {
		if dist[edge.Source().Label()] != maxValue &&
			dist[edge.Source().Label()]+edge.Weight() < dist[edge.Destination().Label()] {
			return nil, ErrNegativeWeightCycle
		}
	}

	return dist, nil
}
