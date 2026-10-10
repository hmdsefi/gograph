![build](https://github.com/hmdsefi/gograph/actions/workflows/build.yml/badge.svg)
[![coverage](https://img.shields.io/github/issues/detail/title/hmdsefi/gograph/162?label=coverage&color=brightgreen)](https://github.com/hmdsefi/gograph/actions/workflows/build.yml?query=branch%3Amaster)
[![CodeRabbit Pull Request Reviews](https://img.shields.io/coderabbit/prs/github/hmdsefi/gograph?utm_source=oss&utm_medium=github&utm_campaign=hmdsefi%2Fgograph&labelColor=171717&color=FF570A&link=https%3A%2F%2Fcoderabbit.ai&label=CodeRabbit+Reviews)](https://coderabbit.ai)
[![Go Reference](https://pkg.go.dev/badge/github.com/hmdsefi/gograph.svg)](https://pkg.go.dev/github.com/hmdsefi/gograph)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#science-and-data-analysis)
[![Sponsor](https://img.shields.io/badge/sponsor-hmdsefi-ea4aaa?logo=githubsponsors)](https://github.com/sponsors/hmdsefi)
[![GitHub stars](https://img.shields.io/github/stars/hmdsefi/gograph?style=social)](https://github.com/hmdsefi/gograph)

<p align="center">
  <img alt="golang generic graph package" src="https://github.com/user-attachments/assets/b5728572-9c17-47e8-aa32-28aeeedf1e25" width="480" title="gograph"/>
</p>

# GoGraph

GoGraph is a generic graph library for Go with first-class support for dependency
graphs. Acyclic graphs refuse edges that would create a cycle, and `TopologySort`
gives you an order to run things in. It also covers traversal, shortest paths,
strongly connected components and graph partitioning, with no dependencies outside
the standard library.

- **Generic:** vertex labels can be any comparable type, such as strings, integers or your own structs.
- **Dependency graphs:** `Acyclic()` graphs reject cycles, `TopologySort` returns a valid order, and the `dag` package finds what depends on what and what can run in parallel.
- **Traversal:** BFS, DFS, topological, closest-first and random-walk iterators.
- **Paths:** Dijkstra, Bellman-Ford, Floyd-Warshall and transitive reduction.
- **Connectivity:** strongly connected components with Tarjan, Kosaraju and Gabow, and condensation into a DAG.
- **Partitioning:** maximal cliques (Bron-Kerbosch), Girvan-Newman communities and randomized k-cut.
- **Diagrams:** `encoding/mermaid` writes a graph as a Mermaid flowchart that GitHub renders in Markdown, and `encoding/dot` writes it in the Graphviz DOT language for larger graphs.

See every algorithm run step by step on example graphs, with its result, the Go code and
benchmarks, at [gograph.dev](https://gograph.dev).

Imported by [20+ public Go modules](https://pkg.go.dev/github.com/hmdsefi/gograph?tab=importedby).

<h3 align="center">⭐ If gograph is useful to you, a star on GitHub helps other Go developers find it.</h3>

## Quick start

```shell
go get github.com/hmdsefi/gograph
```

```go
package main

import (
	"errors"
	"fmt"

	"github.com/hmdsefi/gograph"
)

func main() {
	// An edge A -> B means A has to happen before B.
	g := gograph.New[string](gograph.Acyclic())

	checkout := g.AddVertexByLabel("checkout")
	build := g.AddVertexByLabel("build")
	test := g.AddVertexByLabel("test")
	release := g.AddVertexByLabel("release")

	_, _ = g.AddEdge(checkout, build)
	_, _ = g.AddEdge(build, test)
	_, _ = g.AddEdge(test, release)

	// Acyclic graphs reject any edge that would create a cycle.
	_, err := g.AddEdge(release, checkout)
	fmt.Println(errors.Is(err, gograph.ErrDAGCycle)) // true

	order, _ := gograph.TopologySort(g)
	for _, v := range order {
		fmt.Println(v.Label()) // checkout, build, test, release
	}
}
```

When several orders are valid, `TopologySort` follows the order the vertices and edges
were added, so the same graph gives the same result on every run. To choose the order
yourself, `StableTopologySort` takes a compare function such as `cmp.Compare` and always
picks the smallest vertex that is ready.

`GetAllVertices` returns vertices in the order they were added, and `AllEdges` and
`EdgesOf` follow the same order. `Tarjan`, `Kosaraju`, `Gabow`, `MaximalCliques`,
`GirvanNewman` and `TransitiveReduction` give the same result on every run for a graph
built the same way.

## Table of contents

* [Graphs](#graphs)
    * [Directed](#directed)
    * [Acyclic](#acyclic)
    * [Undirected](#undirected)
    * [Weighted](#weighted)
* [Traversal](#traversal)
* [Algorithms](#algorithms)
* [Examples](#examples)
* [Roadmap](#roadmap)
* [Contributing](#contributing)
* [License](#license)

## Graphs

`gograph.New[T]` creates a graph. `T` is the vertex label type and must be
comparable, so slices, maps and functions can't be labels. Options choose the kind
of graph:

- `gograph.Directed()` creates a directed graph. Without it, graphs are undirected.
- `gograph.Acyclic()` creates a directed graph that rejects edges that would create a cycle.
- `gograph.Weighted()` marks the graph as weighted. `BellmanFord` and `FloydWarshall` require it.

Every graph implements the `Graph[T]` interface. See the
[package documentation](https://pkg.go.dev/github.com/hmdsefi/gograph#Graph) for the full list
of methods.

`AddEdge` creates missing vertices, so `gograph.NewVertex` is enough for quick
examples. To keep a reference to a vertex, use `AddVertexByLabel`, which adds the
vertex and returns it.

### Directed

![directed-graph](https://user-images.githubusercontent.com/11541936/221904292-face2083-16da-491f-a339-2164b7040264.png)

```go
g := gograph.New[int](gograph.Directed())

_, _ = g.AddEdge(gograph.NewVertex(1), gograph.NewVertex(2))
_, _ = g.AddEdge(gograph.NewVertex(1), gograph.NewVertex(3))
_, _ = g.AddEdge(gograph.NewVertex(2), gograph.NewVertex(4))
_, _ = g.AddEdge(gograph.NewVertex(3), gograph.NewVertex(4))
_, _ = g.AddEdge(gograph.NewVertex(4), gograph.NewVertex(5))
_, _ = g.AddEdge(gograph.NewVertex(5), gograph.NewVertex(6))
```

### Acyclic

![acyclic-graph](https://user-images.githubusercontent.com/11541936/221911652-ce2dfb5f-5547-4f26-8412-94ad9124d4fa.png)

```go
g := gograph.New[int](gograph.Acyclic())

_, _ = g.AddEdge(gograph.NewVertex(1), gograph.NewVertex(2))
_, _ = g.AddEdge(gograph.NewVertex(2), gograph.NewVertex(3))

_, err := g.AddEdge(gograph.NewVertex(3), gograph.NewVertex(1))
fmt.Println(err) // edges would create cycle
```

### Undirected

![undirected-graph](https://user-images.githubusercontent.com/11541936/221908261-a009049d-2b71-46c3-9026-faa4dcc2a693.png)

```go
// Graphs are undirected by default.
g := gograph.New[string]()

a := g.AddVertexByLabel("A")
b := g.AddVertexByLabel("B")
c := g.AddVertexByLabel("C")
d := g.AddVertexByLabel("D")

_, _ = g.AddEdge(a, b)
_, _ = g.AddEdge(a, d)
_, _ = g.AddEdge(b, c)
_, _ = g.AddEdge(b, d)

// Every undirected edge can be followed both ways.
fmt.Println(g.ContainsEdge(a, b), g.ContainsEdge(b, a)) // true true
```

### Weighted

![weighted-edge](https://user-images.githubusercontent.com/11541936/221908269-b6db15fb-6104-49d9-b9b9-acc062d94e4a.png)

```go
g := gograph.New[string](gograph.Weighted())

a := g.AddVertexByLabel("A")
b := g.AddVertexByLabel("B")
c := g.AddVertexByLabel("C")
d := g.AddVertexByLabel("D")

_, _ = g.AddEdge(a, b, gograph.WithEdgeWeight(4))
_, _ = g.AddEdge(a, d, gograph.WithEdgeWeight(3))
_, _ = g.AddEdge(b, c, gograph.WithEdgeWeight(3))
_, _ = g.AddEdge(b, d, gograph.WithEdgeWeight(1))
_, _ = g.AddEdge(c, d, gograph.WithEdgeWeight(2))

dist := path.Dijkstra(g, "A")
fmt.Println(dist["C"]) // 5
```

Vertices can have weights too:

![weighted-vertex](https://user-images.githubusercontent.com/11541936/221908278-83f3138d-8b28-4c38-825a-627a46d65294.png)

```go
g := gograph.New[string](gograph.Directed(), gograph.Weighted())

a := g.AddVertexByLabel("A", gograph.WithVertexWeight(3))
b := g.AddVertexByLabel("B", gograph.WithVertexWeight(2))
c := g.AddVertexByLabel("C", gograph.WithVertexWeight(4))

_, _ = g.AddEdge(a, b)
_, _ = g.AddEdge(b, c)

fmt.Println(a.Weight(), b.Weight(), c.Weight()) // 3 2 4
```

## Traversal

The `traverse` package provides iterators that all implement the same interface:

```go
type Iterator[T comparable] interface {
	HasNext() bool
	Next() *gograph.Vertex[T]
	Iterate(func(v *gograph.Vertex[T]) error) error
	Reset()
}
```

```go
g := gograph.New[string](gograph.Directed())

_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("B"))
_, _ = g.AddEdge(gograph.NewVertex("A"), gograph.NewVertex("C"))
_, _ = g.AddEdge(gograph.NewVertex("B"), gograph.NewVertex("D"))

it, err := traverse.NewBreadthFirstIterator(g, "A")
if err != nil {
	fmt.Println(err)
	return
}

for it.HasNext() {
	fmt.Println(it.Next().Label()) // A, B, C, D
}
```

Available iterators:

- [Breadth-first](https://github.com/hmdsefi/gograph/tree/master/traverse#bfs)
- [Depth-first](https://github.com/hmdsefi/gograph/tree/master/traverse#dfs)
- [Topological](https://github.com/hmdsefi/gograph/tree/master/traverse#topological-sort)
- [Closest-first](https://github.com/hmdsefi/gograph/tree/master/traverse#closest-first)
- [Random walk](https://github.com/hmdsefi/gograph/tree/master/traverse#random-walk)

## Algorithms

To watch these algorithms run step by step on example graphs, see
[gograph.dev](https://gograph.dev).

- **Ordering:** `gograph.TopologySort` (Kahn's algorithm) and `gograph.StableTopologySort`
  (smallest ready vertex first).
- **Subgraphs:** `gograph.InducedSubgraph` keeps a chosen set of vertices and the edges
  between them.
- **Shortest paths** (`path` package):
  [Dijkstra](https://github.com/hmdsefi/gograph/blob/master/path/dijkstra.md),
  [Bellman-Ford](https://github.com/hmdsefi/gograph/blob/master/path/bellman-ford.md),
  [Floyd-Warshall](https://github.com/hmdsefi/gograph/blob/master/path/floyd-warshall.md),
  `DijkstraMultiSource`, which finds each vertex's nearest source in one search,
  and `KCenter`, which places k centers so the farthest vertex is as close as the
  method can make it.
- **Transitive reduction** (`path` package):
  [TransitiveReduction](https://github.com/hmdsefi/gograph/blob/master/path/transitive-reduction.md).
- **Dependencies** (`dag` package): `Descendants` (what depends on a vertex), `Ancestors`
  (what it depends on), `Affected` (what a change reaches), `Levels` and `Tracker`
  (what can run at the same time), and `CriticalPath` (the longest chain of dependent tasks).
- **Strongly connected components** (`connectivity` package):
  [Tarjan, Kosaraju and Gabow](https://github.com/hmdsefi/gograph/tree/master/connectivity#gograph---connectivity),
  and [condensation](https://github.com/hmdsefi/gograph/tree/master/connectivity#condensation) into a DAG.
- **Partitioning** (`partition` package):
  [maximal cliques (Bron-Kerbosch)](https://github.com/hmdsefi/gograph/blob/master/partition/bron_kerbosch.md),
  [Girvan-Newman](https://github.com/hmdsefi/gograph/blob/master/partition/girvan-newman.md),
  [randomized k-cut](https://github.com/hmdsefi/gograph/blob/master/partition/k-cut.md).

## Examples

- [gomodgraph](examples/gomodgraph): loads `go mod graph` output and finds requirement
  cycles, the modules that depend on a module, why a module is needed, what changed
  between two versions of `go.mod`, and a diagram of a module with its direct
  dependencies and dependents.

## Roadmap

Planned work, in the order it's likely to land, is tracked in
[#136](https://github.com/hmdsefi/gograph/issues/136). Issues labeled
[good first issue](https://github.com/hmdsefi/gograph/labels/good%20first%20issue)
are a good place to start.

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before
opening a pull request. The README examples are also Go examples in
[`example_test.go`](example_test.go), so `go test ./...` checks that they still compile
and print what they claim.

## Sponsoring

If gograph saves you time, you can support its development through
[GitHub Sponsors](https://github.com/sponsors/hmdsefi).

## License

Apache License 2.0, see [LICENSE](LICENSE) for details.
