# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- A random walk from a vertex without outgoing edges returned nothing. It now
  returns the start vertex. ([#147])
- `TransitiveReduction` dropped vertex weights, and dropped edge weights when the
  input graph wasn't created with `Weighted()`. It now keeps both. ([#148])
- `FloydWarshall` reported the weight of a self-loop as a vertex's distance to
  itself. It now reports 0, and a negative loop is still reported as a negative
  cycle. ([#145])
- `BellmanFord` returned distances for a start vertex that isn't in the graph. It
  now returns `gograph.ErrVertexDoesNotExist`. ([#99])
- `VertexPriorityQueue.Pop` panicked on an empty queue, and `Push(nil)` and
  `Vertex.HasNeighbor(nil)` panicked too. `Pop` now returns nil, `Push` ignores
  nil, and `HasNeighbor` returns false. ([#146])
- In directed graphs, `RemoveVertices` didn't subtract the outgoing edges of the
  removed vertex, so `Size()` stayed too high. ([#96])
- In undirected graphs, a self-loop was counted twice in some places and once in
  others, and removing it left the vertex listed as its own neighbor. A loop is
  now stored once, so it adds 1 to `Size()`, `InDegree()` and `OutDegree()`, and
  `GetAllEdges(a, a)` returns it once. ([#142])

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

[Unreleased]: https://github.com/hmdsefi/gograph/compare/v0.7.2...HEAD
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
[#137]: https://github.com/hmdsefi/gograph/pull/137
[#138]: https://github.com/hmdsefi/gograph/pull/138
[#139]: https://github.com/hmdsefi/gograph/pull/139
[#140]: https://github.com/hmdsefi/gograph/pull/140
[#141]: https://github.com/hmdsefi/gograph/issues/141
[#96]: https://github.com/hmdsefi/gograph/issues/96
[#142]: https://github.com/hmdsefi/gograph/issues/142
[#143]: https://github.com/hmdsefi/gograph/issues/143
[#144]: https://github.com/hmdsefi/gograph/issues/144
[#99]: https://github.com/hmdsefi/gograph/issues/99
[#145]: https://github.com/hmdsefi/gograph/issues/145
[#146]: https://github.com/hmdsefi/gograph/issues/146
[#147]: https://github.com/hmdsefi/gograph/issues/147
[#148]: https://github.com/hmdsefi/gograph/issues/148
[#149]: https://github.com/hmdsefi/gograph/issues/149
[#151]: https://github.com/hmdsefi/gograph/pull/151
[#152]: https://github.com/hmdsefi/gograph/pull/152
[#153]: https://github.com/hmdsefi/gograph/pull/153
[#154]: https://github.com/hmdsefi/gograph/pull/154
