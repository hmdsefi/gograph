# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `Vertex.AllNeighbors` returns an iterator over the neighbors of a vertex.
  For a vertex of a graph, it yields the graph's own vertices without copying
  them. ([#107], [#287])
- `traverse.BFS`, `BFSWithDepth`, `DFS`, `ClosestFirst` and `Topological`
  return traversals as sequences for a `range` loop, in the order of the
  matching iterators. `BFSWithDepth` gives each vertex with its number of
  edges from the start, and `ClosestFirst` with its distance. ([#115], [#289])
- `Vertices` and `Edges` range over all the vertices and edges of a graph. For
  a graph from `New` they allocate nothing: ranging over the edges of a graph
  with 10,000 vertices and 50,000 edges takes 0.44 ms, where `AllEdges` takes
  0.93 ms and 2 MiB. ([#115], [#289])
- `traverse.DFSEvents` and `DFSEventsAll` give the steps of a depth-first
  search as a sequence: each vertex when the search discovers and finishes
  it, and each edge as a tree, back, forward or cross edge, with its depth.
  They keep their own stack, so a path of 100,000 vertices is safe to search,
  and a `range` loop can stop at any event. ([#132], [#295])
- `connectivity.FindCycle` returns the vertices of one cycle in order, such
  as an import cycle to show in an error. It reports a cycle exactly when
  `HasCycle` does, in O(V+E) time. ([#132], [#295])

### Changed

- Each vertex of a graph from `New` keeps the edges that end at it.
  `InDegree` reads that list in O(1), `EdgesOf` takes O(d log d) time for a
  vertex with d edges, and `RemoveVertices` visits only the neighbors of each
  removed vertex. On a graph with 2,000 vertices and 10,000 edges, calling
  `EdgesOf` on every vertex takes 0.3 ms instead of 54 ms, and removing every
  vertex takes 0.7 ms instead of 37 ms. ([#107], [#287])
- On a graph created with `Acyclic`, `AddEdge` keeps a topological order of
  the vertices to check each new edge. An edge that agrees with the order, an
  edge from a vertex with no incoming edges, and an edge to a vertex with no
  outgoing edges need no search. Any other edge searches around its
  endpoints, in at most O(R log R) time for the R vertices and edges reachable
  from its destination. Building a chain of 4,000 edges takes 1.1 ms instead
  of 618 ms, and a random DAG with 2,000 vertices and 10,000 edges 1.8 ms
  instead of 2.2 s. ([#107], [#124], [#287], [#288])
- The traversal iterators, `Dijkstra`, the strongly connected component
  functions, the dag helpers and the encoders read the stored neighbors
  through `AllNeighbors`. The breadth-first and depth-first iterators run
  about 60% faster on a graph of 10,000 vertices and allocate 99% less
  often. ([#107], [#287])
- A vertex takes 16 more bytes, and each edge one more pointer in the list of
  its destination. An acyclic graph also keeps one int per vertex for its
  order. Removing an edge takes time proportional to the in-degree of its
  destination as well as the out-degree of its source. ([#287], [#288])
- `BellmanFord` keeps its distances in slices and stops after a pass that
  changes none of them. `FloydWarshall` keeps its distances in one flat
  matrix and skips each row with no path to the middle vertex. On a random
  graph with 10,000 vertices and 50,000 edges, `BellmanFord` takes 2.7 ms
  instead of 11.9 s, and `FloydWarshall` on 300 vertices takes 9 ms instead
  of 1.66 s. ([#294])

### Fixed

- `BellmanFord` and `FloydWarshall` give `-Inf` for a path through a `-Inf`
  weight on a graph with no cycle, as v0.9.0 did. A cycle through a `-Inf`
  weight, and a negative cycle that a `-Inf` weight leads into, return
  `ErrNegativeWeightCycle`. ([#291], [#294])
- `FloydWarshall` returns `ErrNegativeWeightCycle` only for a negative cycle.
  A check after the main loop added the weights in another order, and its
  rounding reported a cycle on some graphs with none, such as the path
  1 -> 3 -> 2 -> 0 with weights 3.3, 3.3 and 1.1. ([#294])

## [0.9.1] - 2026-10-26

### Added

- `Clone` returns a copy of a graph that can be changed without changing the
  original. It keeps the directed, weighted and acyclic properties, the
  weights, edge labels and metadata, and the order of vertices and edges.
  ([#108], [#251])

### Changed

- On an undirected graph, `GirvanNewman` breaks ties between equally busy
  edges in `AllEdges` order. A graph with such ties can split into different,
  equally valid communities than before. ([#251])
- The closest-first iterator can return vertices at equal distances in a
  different order than before. ([#261], [#273])
- `BellmanFord` and `FloydWarshall` return `ErrNegativeWeightCycle` when a
  distance reaches `-Inf`, from a `-Inf` weight or from weights that add up
  below the float64 range. That includes a path through a `-Inf` weight on a
  graph with no cycle, which [#291] tracks. `DijkstraMultiSource` reports a
  vertex behind a `+Inf` weight as unreachable, as `Dijkstra` does.
  ([#262], [#285])
- The docs say that an undirected edge is stored in both directions and a
  self-loop once, how the `Weighted` option relates to the edge weights in
  each shortest-path function, and what each one returns for a vertex it
  can't reach. Four new examples show them. ([#103], [#248])
- The doc comments of `DijkstraMultiSource`, `KCenter`, `dag.Descendants`,
  `dag.Affected`, `dag.Ancestors`, `GetAllEdges` and `EdgesOf`, and the
  traversal guide in `traverse/README.md`, describe what the code does.
  ([#268], [#278])
- A pull request for an issue in a later milestone goes to that milestone's
  `release/vX.Y.Z` branch. The contributing guide and the pull request
  template say so. ([#257])

### Fixed

- `ConnectedComponents` keeps `GetAllVertices` order for a directed `Graph`
  that isn't from `New`, and reads it in O((V+E)·α(V)) instead of O(V²).
  ([#267], [#271])
- `InducedSubgraph` copies each edge's insertion index, so ties in
  `DijkstraMultiSource` and other functions that use `InsertionIndex` break
  the same way as on the original graph. ([#265], [#272])
- `Dijkstra` no longer expands a vertex again for each outdated queue entry,
  and the closest-first iterator keeps at most one queue entry per vertex. On
  a hub graph with 1,000 branches, `Dijkstra` takes 0.7 ms instead of 63 ms.
  ([#261], [#273])
- `RemoveVertices` clears the neighbors and in-degree of a removed vertex, so
  its `Neighbors`, `OutDegree`, `InDegree` and `HasNeighbor` no longer report
  edges the graph doesn't have. A directed self-loop no longer leaves its
  in-degree at -1. A vertex of another graph with the same label keeps its
  edges. ([#266], [#275], [#290], [#292])
- `TopologySort`, `StableTopologySort` and `traverse.NewTopologicalIterator`
  return `ErrNotDirected` when given an undirected graph. Previously,
  graphs with edges reported `ErrDAGHasCycle`, while edge-free graphs
  sorted successfully. ([#264], [#281])
- `GirvanNewman` counts each shortest path once. On an undirected graph it
  counted each path twice at every level, so a search 1,024 levels deep
  overflowed to `+Inf`, every betweenness value turned into NaN, and
  `GirvanNewman` never returned. It returns an error when two vertices have
  more shortest paths between them than a float64 can count. One betweenness
  pass on a path of 1,025 vertices takes 0.4 s instead of 9.1 s.
  ([#258], [#282])
- `GirvanNewman` removes the first edge in `AllEdges()` order among the edges
  tied for the highest betweenness, as its doc says. Values within one part in
  a billion of the highest count as tied, since float sums added in a
  different order can differ in the last bits. ([#259], [#283])
- `GirvanNewman` returns directed communities for a directed graph. When the
  graph had both `a->b` and `b->a`, the second edge was lost with its weight
  and label. ([#260], [#284])
- `FloydWarshall` skips an edge with a NaN weight, as `BellmanFord` and
  `Dijkstra` do. It kept NaN as the distance even when another path existed.
  `BellmanFord` finds a negative cycle once its distances reach `-Inf`. The
  closest-first iterator skips a NaN weight and visits a vertex behind a
  `+Inf` weight after every vertex at a finite distance, and
  `util.VertexPriorityQueue` orders a NaN priority after every number.
  ([#262], [#285])
- The breadth-first, depth-first and closest-first iterators skip a queued
  vertex that was removed from the graph. Breadth-first and depth-first
  panicked in `Next`, closest-first returned the removed vertex, and `HasNext`
  returned true in all three. After the start vertex is removed, `Reset`
  leaves nothing to visit. ([#263], [#286])
- Mermaid labels write `%` as `#37;` and `:` as `#58;`, since Mermaid reads
  both inside quoted labels. The class names `default` and `node` return the
  invalid class name error, since Mermaid applies them to every node.
  ([#269], [#279])
- `encoding/dot` writes text that Graphviz reads back as given. A value is
  written as is only when Graphviz reads all of it as one HTML-like string.
  Quoted strings write `&` as `&amp;` and NUL as U+FFFD, and IDs such as the
  graph name escape only `"`. A name or attribute name with a backslash at the
  end, or before a `"` or a line break, returns an error. ([#270], [#280])

## [0.9.0] - 2026-10-16

### Added

- `WithEdgeLabel` and `Edge.Label()` to give an edge a short label, such as
  "imports" or "calls". The Mermaid and DOT output show it by default, and
  `TransitiveReduction` and `GirvanNewman` keep it. ([#191], [#217])
- `InducedSubgraph` returns a new graph with the given vertices and every edge
  between them. It has the same directed, weighted and acyclic properties as
  the input, and copies the weights, edge labels and metadata. ([#118], [#234])
- `Edge.InsertionIndex()` reports the order in which an edge was added to its
  graph. ([#233])
- `dag.Levels` groups the vertices of a DAG so that every edge goes from a lower
  level to a higher one, and the vertices of one level can run at the same time.
  `dag.Tracker` hands out each vertex as soon as every vertex it depends on is
  done, so a slow vertex only holds up the ones that depend on it. `Done`
  returns `dag.ErrNotReady` or `dag.ErrAlreadyDone` for a vertex it can't mark
  done. ([#120], [#181])
- `dag.CriticalPath` and `dag.CriticalPathFunc` to find the most expensive chain
  of dependent vertices, such as the steps that set how long a build takes. The
  path runs from a vertex with no incoming edges to one with no outgoing edges,
  and its cost is the sum of the vertex and edge weights. ([#121], [#228])
- `dag.Sources` and `dag.Sinks` to find the vertices with no incoming edges and
  no outgoing edges, in `GetAllVertices()` order. ([#116], [#247])
- `connectivity.Condense` replaces each strongly connected component with one
  vertex and returns the resulting DAG, with the components numbered in
  topological order, the members of each one, and the component of each
  vertex. `WithComponentWeight` and `WithCrossingWeight` set the weights of the
  new vertices and edges. ([#128], [#183])
- `connectivity.ConnectedComponents` returns the connected components of an
  undirected graph and the weakly connected components of a directed one.
  `connectivity.HasCycle` reports whether a graph has a cycle, including a
  self-loop. Both call `gograph.ConnectedComponents` and `gograph.HasCycle`,
  which read the neighbor lists stored on the vertices. ([#111], [#250], [#252])
- `path.DijkstraMultiSource` finds the nearest of several sources for every
  vertex in one search, which costs the same as Dijkstra from a single start.
  `NearestSources` gives the distance, the source and the path for each vertex.
  ([#218], [#233])
- `path.KCenter` chooses k centers so that the vertex farthest from its nearest
  center is as close to one as it can make it, and can start from centers that
  are already open. On an undirected graph the radius is at most twice the
  smallest possible. It returns `path.ErrInvalidK` for a k that's too small and
  `path.ErrNegativeWeight` for a negative edge weight, which
  `DijkstraMultiSource` returns too. ([#219], [#232])
- The `encoding/dot` package writes a graph in the DOT language of Graphviz,
  with options for the graph name and the graph, vertex and edge attributes. A
  graph built the same way gives the same text on every run. ([#123], [#182])
- `mermaid` in `examples/gomodgraph` draws a module with the modules it requires
  directly and the modules that require it directly. ([#127], [#237])
- `versions` and `diff -selected` in `examples/gomodgraph`. `versions` groups
  the module versions by path and marks the one that builds, and `diff -selected`
  compares the version of each module that builds, like `go list -m all` on
  both commits. ([#246], [#249])
- `.agents/skills/gograph/SKILL.md`, a skill that picks the gograph function for
  a graph question and names the constraint that makes a nearby function the
  wrong one. ([#220], [#238])

### Changed

- Raise the minimum Go version to 1.26 and update golangci-lint to v2.14.0.
  Projects using an older local toolchain can stay on v0.8.x. ([#177], [#216])
- The directed graph example in the README added an edge from 2 to 2, where the
  diagram shows an edge from 2 to 4. The README and `ExampleNew_directed` now
  add the edge from 2 to 4, and the example prints every edge. ([#230])

## [0.8.2] - 2026-10-02

### Fixed

- `GirvanNewman` removed every edge tied for the highest betweenness at once, so
  it could split the graph into more than k communities. It now removes one edge
  per step, the first tied one in the order of `AllEdges()`, and returns exactly
  k communities when the graph has at least k vertices and at most k connected
  components. The communities also keep the vertex and edge weights, and are
  weighted when the input graph is. ([#101], [#178])
- `RandomizedKCut` kept only one cut edge between two supernodes, and returned no
  cut edges when k was the number of vertices. It now returns every edge between
  two different supernodes, with each edge of an undirected graph listed once.
  The supernodes follow the order of `GetAllVertices()` instead of a random
  order. ([#102], [#179])
- `Gabow` ran the same search as `Tarjan`. It now runs Gabow's path-based
  algorithm, and the result is the same as before, including the order of the
  components. ([#105], [#180])
- `TransitiveReduction` returned a graph without `Acyclic()` even when the input
  was created with it, so `AddEdge` on the result accepted edges that create a
  cycle. The result is now acyclic when the input is. ([#148], [#197])

### Changed

- CI builds the commit that was pushed instead of the default branch, also
  builds with the latest Go release, and uses current versions of the GitHub
  Actions. ([#106], [#176])
- Fixed doc comments that were out of date or copied from other functions,
  including the closest-first iterator, which was described as a random walk,
  and the note on using a graph from more than one goroutine. The note on
  `Kosaraju` that said it needs a connected graph and doesn't handle self-loops
  is removed, because neither is true. ([#150], [#180], [#184])

## [0.8.1] - 2026-09-30

### Fixed

- `NewVertex` ignored its options, so a weight set with `WithVertexWeight` was
  lost, including for vertices passed to `AddVertex` and `AddEdge`. It now
  applies them. ([#97], [#172])
- The weighted random walk panicked when the edge weights of a vertex added up
  to less than 1, and picked neighbors with the wrong odds when the weights
  weren't whole numbers. Each step now picks a neighbor with a probability
  proportional to the weight of the edge to it. Edges with a weight of zero or
  less are never picked, unless no edge of the vertex has a positive weight.
  Then each edge has an equal chance. ([#98], [#173])
- The closest-first iterator and the random walk returned copies of the
  vertices after the start vertex, and `Kosaraju` returned the vertices of the
  reversed graph it builds, which have no weight and neighbors that point the
  other way. They now return the vertices stored in the graph. ([#100], [#174])

## [0.8.0] - 2026-09-30

### Added

- `StableTopologySort`, a topological sort that takes a compare function and
  picks the smallest ready vertex at each step, so the result doesn't depend on
  the order the graph was built in. ([#114], [#167])
- `dag` package with `Descendants`, `Ancestors` and `Affected`, to find what a
  vertex depends on, what depends on it, and what a change to it affects.
  ([#117], [#168])
- `encoding/mermaid` package that writes a graph as a Mermaid flowchart, with
  options for the direction, the vertex and edge labels, and classes. ([#122],
  [#169])
- `gograph.ErrNotDirected`. `path.ErrNotDirected` is now the same error, so
  existing `errors.Is` checks keep working. ([#168])
- `examples/gomodgraph`, an example program that finds cycles, dependents and
  dependency paths in `go mod graph` output, and compares two of them. ([#127],
  [#170])

### Changed

- `GetAllVertices` returns vertices in the order they were added. `AllEdges`
  groups edges by source vertex in that order, with the edges of each vertex in
  the order they were added. `EdgesOf` returns the edges that start from the
  vertex first, in the order they were added, and then the edges that end at it,
  in the order of their source vertices. The order used to be random. ([#114],
  [#167])
- `TopologySort`, `Tarjan`, `Kosaraju`, `Gabow`, `MaximalCliques`,
  `GirvanNewman` and `TransitiveReduction` give the same result on every run for
  a graph built the same way. ([#114], [#167])

### Fixed

- A random walk from a vertex without outgoing edges returned nothing. It now
  returns the start vertex. ([#147], [#157])
- `MaximalCliques` returned wrong cliques for graphs with more than six vertices.
  It now finds every maximal clique. ([#95], [#166])
- `TransitiveReduction` dropped vertex weights, and dropped edge weights when the
  input graph wasn't created with `Weighted()`. It now keeps both. ([#148],
  [#157])
- `FloydWarshall` reported the weight of a self-loop as a vertex's distance to
  itself. It now reports 0, and a negative loop is still reported as a negative
  cycle. ([#145], [#165])
- `BellmanFord` returned distances for a start vertex that isn't in the graph. It
  now returns `gograph.ErrVertexDoesNotExist`. ([#99], [#165])
- `VertexPriorityQueue.Pop` panicked on an empty queue, and `Push(nil)` and
  `Vertex.HasNeighbor(nil)` panicked too. `Pop` now returns nil, `Push` ignores
  nil, and `HasNeighbor` returns false. ([#146], [#159])
- In directed graphs, `RemoveVertices` didn't subtract the outgoing edges of the
  removed vertex, so `Size()` stayed too high. ([#96], [#164])
- In undirected graphs, a self-loop was counted twice in some places and once in
  others, and removing it left the vertex listed as its own neighbor. A loop is
  now stored once, so it adds 1 to `Size()`, `InDegree()` and `OutDegree()`, and
  `GetAllEdges(a, a)` returns it once. ([#142], [#164])

## [0.7.2] - 2026-09-29

### Fixed

- `RemoveEdges` lowered `Size()` for edges that weren't in the graph, so `Size()`
  could stop matching `len(AllEdges())`. It now only counts the edges it removes.
  ([#141], [#151])
- A vertex added to a second graph was shared by both graphs, so edges added to
  either graph showed up in the other. `AddVertex` and `AddEdge` now store a copy
  when the vertex already belongs to a graph. ([#143], [#152])
- `MaximalCliques` dropped cliques that contain a self-loop and gave random
  results on directed graphs. It now ignores self-loops and treats directed edges
  as undirected. ([#144], [#153])
- `Tarjan`, `Gabow` and `Kosaraju` crashed with a stack overflow on long paths.
  They no longer use recursion, and their results are the same as before.
  ([#149], [#154])

### Changed

- Rewrote the README quick start. The README examples are now checked by tests.
  ([#138])
- Added a contributing guide, issue and pull request templates, and a security
  policy. ([#137])
- CI checks test coverage with go-test-coverage instead of Codecov. ([#140], [#139])

## [0.7.1] - 2026-07-19

### Fixed

- `Dijkstra` skips a neighbor when there is no edge to it instead of panicking.
  ([#94])
- `FloydWarshall` also reports a negative weight cycle when a vertex's distance
  to itself becomes negative. ([#94])
- `VertexPriorityQueue.Pop` returns nil when the popped item isn't a vertex.
  ([#94])

## [0.7.0] - 2025-09-06

### Added

- `partition` package with:
  - Girvan-Newman community detection for undirected graphs.
  - Approximate k-cut of an undirected graph using randomized contraction, a
    generalization of Karger's min-cut.
  - Bron-Kerbosch with pivot selection, degeneracy ordering and bitsets, to find
    all maximal cliques.

## [0.6.0] - 2025-08-22

### Added

- Transitive reduction algorithm.

## [0.5.0] - 2025-05-18

### Changed

- Go version raised to 1.24.2.
- Code cleanup and formatting, and an updated golangci-lint config.

## [0.4.2] - 2024-08-01

### Added

- `Order` and `Size` methods on `Graph`. ([#76])

## [0.4.1] - 2024-07-22

### Fixed

- Adding edges and new vertices to a graph with cycles.

## [0.4.0] - 2023-04-06

### Added

- `path` package with a simple and a standard implementation of Dijkstra's
  algorithm.

## [0.3.0] - 2023-03-16

### Added

- `connectivity` package with Tarjan's, Kosaraju's and Gabow's algorithms to find
  strongly connected components.

## [0.2.0] - 2023-03-08

### Added

- `traverse` package documentation.
- Closest-first iterator.
- Random walk iterator.
- Vertex method for the total degree.
- Edge methods for the source and destination vertices.

## [0.1.0] - 2023-03-01

### Added

- Core graph data structures.
- Basic traversal iterators.

[Unreleased]: https://github.com/hmdsefi/gograph/compare/v0.9.1...HEAD
[0.9.1]: https://github.com/hmdsefi/gograph/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/hmdsefi/gograph/compare/v0.8.2...v0.9.0
[0.8.2]: https://github.com/hmdsefi/gograph/compare/v0.8.1...v0.8.2
[0.8.1]: https://github.com/hmdsefi/gograph/compare/v0.8.0...v0.8.1
[0.8.0]: https://github.com/hmdsefi/gograph/compare/v0.7.2...v0.8.0
[0.7.2]: https://github.com/hmdsefi/gograph/compare/v0.7.1...v0.7.2
[0.7.1]: https://github.com/hmdsefi/gograph/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/hmdsefi/gograph/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/hmdsefi/gograph/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/hmdsefi/gograph/compare/v0.4.2...v0.5.0
[0.4.2]: https://github.com/hmdsefi/gograph/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/hmdsefi/gograph/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/hmdsefi/gograph/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/hmdsefi/gograph/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/hmdsefi/gograph/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/hmdsefi/gograph/releases/tag/v0.1.0

[#76]: https://github.com/hmdsefi/gograph/pull/76
[#94]: https://github.com/hmdsefi/gograph/pull/94
[#95]: https://github.com/hmdsefi/gograph/issues/95
[#96]: https://github.com/hmdsefi/gograph/issues/96
[#97]: https://github.com/hmdsefi/gograph/issues/97
[#98]: https://github.com/hmdsefi/gograph/issues/98
[#99]: https://github.com/hmdsefi/gograph/issues/99
[#100]: https://github.com/hmdsefi/gograph/issues/100
[#101]: https://github.com/hmdsefi/gograph/issues/101
[#102]: https://github.com/hmdsefi/gograph/issues/102
[#105]: https://github.com/hmdsefi/gograph/issues/105
[#106]: https://github.com/hmdsefi/gograph/issues/106
[#114]: https://github.com/hmdsefi/gograph/issues/114
[#116]: https://github.com/hmdsefi/gograph/issues/116
[#117]: https://github.com/hmdsefi/gograph/issues/117
[#121]: https://github.com/hmdsefi/gograph/issues/121
[#122]: https://github.com/hmdsefi/gograph/issues/122
[#127]: https://github.com/hmdsefi/gograph/issues/127
[#137]: https://github.com/hmdsefi/gograph/pull/137
[#138]: https://github.com/hmdsefi/gograph/pull/138
[#139]: https://github.com/hmdsefi/gograph/pull/139
[#140]: https://github.com/hmdsefi/gograph/pull/140
[#141]: https://github.com/hmdsefi/gograph/issues/141
[#142]: https://github.com/hmdsefi/gograph/issues/142
[#143]: https://github.com/hmdsefi/gograph/issues/143
[#144]: https://github.com/hmdsefi/gograph/issues/144
[#145]: https://github.com/hmdsefi/gograph/issues/145
[#146]: https://github.com/hmdsefi/gograph/issues/146
[#147]: https://github.com/hmdsefi/gograph/issues/147
[#148]: https://github.com/hmdsefi/gograph/issues/148
[#149]: https://github.com/hmdsefi/gograph/issues/149
[#150]: https://github.com/hmdsefi/gograph/issues/150
[#151]: https://github.com/hmdsefi/gograph/pull/151
[#152]: https://github.com/hmdsefi/gograph/pull/152
[#153]: https://github.com/hmdsefi/gograph/pull/153
[#154]: https://github.com/hmdsefi/gograph/pull/154
[#157]: https://github.com/hmdsefi/gograph/pull/157
[#159]: https://github.com/hmdsefi/gograph/pull/159
[#164]: https://github.com/hmdsefi/gograph/pull/164
[#165]: https://github.com/hmdsefi/gograph/pull/165
[#166]: https://github.com/hmdsefi/gograph/pull/166
[#167]: https://github.com/hmdsefi/gograph/pull/167
[#168]: https://github.com/hmdsefi/gograph/pull/168
[#169]: https://github.com/hmdsefi/gograph/pull/169
[#170]: https://github.com/hmdsefi/gograph/pull/170
[#172]: https://github.com/hmdsefi/gograph/pull/172
[#173]: https://github.com/hmdsefi/gograph/pull/173
[#174]: https://github.com/hmdsefi/gograph/pull/174
[#176]: https://github.com/hmdsefi/gograph/pull/176
[#177]: https://github.com/hmdsefi/gograph/issues/177
[#216]: https://github.com/hmdsefi/gograph/pull/216
[#178]: https://github.com/hmdsefi/gograph/pull/178
[#179]: https://github.com/hmdsefi/gograph/pull/179
[#180]: https://github.com/hmdsefi/gograph/pull/180
[#184]: https://github.com/hmdsefi/gograph/pull/184
[#191]: https://github.com/hmdsefi/gograph/issues/191
[#197]: https://github.com/hmdsefi/gograph/pull/197
[#217]: https://github.com/hmdsefi/gograph/pull/217
[#228]: https://github.com/hmdsefi/gograph/pull/228
[#247]: https://github.com/hmdsefi/gograph/pull/247
[#246]: https://github.com/hmdsefi/gograph/issues/246
[#249]: https://github.com/hmdsefi/gograph/pull/249
[#118]: https://github.com/hmdsefi/gograph/issues/118
[#120]: https://github.com/hmdsefi/gograph/issues/120
[#123]: https://github.com/hmdsefi/gograph/issues/123
[#128]: https://github.com/hmdsefi/gograph/issues/128
[#218]: https://github.com/hmdsefi/gograph/issues/218
[#219]: https://github.com/hmdsefi/gograph/issues/219
[#220]: https://github.com/hmdsefi/gograph/issues/220
[#181]: https://github.com/hmdsefi/gograph/pull/181
[#182]: https://github.com/hmdsefi/gograph/pull/182
[#183]: https://github.com/hmdsefi/gograph/pull/183
[#230]: https://github.com/hmdsefi/gograph/pull/230
[#232]: https://github.com/hmdsefi/gograph/pull/232
[#233]: https://github.com/hmdsefi/gograph/pull/233
[#234]: https://github.com/hmdsefi/gograph/pull/234
[#237]: https://github.com/hmdsefi/gograph/pull/237
[#238]: https://github.com/hmdsefi/gograph/pull/238
[#252]: https://github.com/hmdsefi/gograph/pull/252
[#111]: https://github.com/hmdsefi/gograph/issues/111
[#250]: https://github.com/hmdsefi/gograph/pull/250
[#108]: https://github.com/hmdsefi/gograph/issues/108
[#251]: https://github.com/hmdsefi/gograph/pull/251
[#261]: https://github.com/hmdsefi/gograph/issues/261
[#264]: https://github.com/hmdsefi/gograph/issues/264
[#265]: https://github.com/hmdsefi/gograph/issues/265
[#266]: https://github.com/hmdsefi/gograph/issues/266
[#267]: https://github.com/hmdsefi/gograph/issues/267
[#271]: https://github.com/hmdsefi/gograph/pull/271
[#272]: https://github.com/hmdsefi/gograph/pull/272
[#273]: https://github.com/hmdsefi/gograph/pull/273
[#275]: https://github.com/hmdsefi/gograph/pull/275
[#281]: https://github.com/hmdsefi/gograph/pull/281
[#103]: https://github.com/hmdsefi/gograph/issues/103
[#258]: https://github.com/hmdsefi/gograph/issues/258
[#259]: https://github.com/hmdsefi/gograph/issues/259
[#260]: https://github.com/hmdsefi/gograph/issues/260
[#262]: https://github.com/hmdsefi/gograph/issues/262
[#263]: https://github.com/hmdsefi/gograph/issues/263
[#268]: https://github.com/hmdsefi/gograph/issues/268
[#269]: https://github.com/hmdsefi/gograph/issues/269
[#270]: https://github.com/hmdsefi/gograph/issues/270
[#290]: https://github.com/hmdsefi/gograph/issues/290
[#291]: https://github.com/hmdsefi/gograph/issues/291
[#248]: https://github.com/hmdsefi/gograph/pull/248
[#257]: https://github.com/hmdsefi/gograph/pull/257
[#278]: https://github.com/hmdsefi/gograph/pull/278
[#279]: https://github.com/hmdsefi/gograph/pull/279
[#280]: https://github.com/hmdsefi/gograph/pull/280
[#282]: https://github.com/hmdsefi/gograph/pull/282
[#283]: https://github.com/hmdsefi/gograph/pull/283
[#284]: https://github.com/hmdsefi/gograph/pull/284
[#285]: https://github.com/hmdsefi/gograph/pull/285
[#286]: https://github.com/hmdsefi/gograph/pull/286
[#292]: https://github.com/hmdsefi/gograph/pull/292
[#107]: https://github.com/hmdsefi/gograph/issues/107
[#287]: https://github.com/hmdsefi/gograph/pull/287
[#124]: https://github.com/hmdsefi/gograph/issues/124
[#288]: https://github.com/hmdsefi/gograph/pull/288
[#115]: https://github.com/hmdsefi/gograph/issues/115
[#289]: https://github.com/hmdsefi/gograph/pull/289
[#294]: https://github.com/hmdsefi/gograph/pull/294
[#132]: https://github.com/hmdsefi/gograph/issues/132
[#295]: https://github.com/hmdsefi/gograph/pull/295
