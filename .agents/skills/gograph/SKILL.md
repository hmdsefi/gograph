---
name: gograph
description: Choose a gograph function for a graph question, and the constraint that makes a different function the wrong one.
---

# gograph

Pick the row whose question matches. The constraint is the reason a nearby function is the wrong call.

In the `dag` package, an edge from A to B means B depends on A. Elsewhere an edge is a link from its source to its destination.

A name with no package is in package `gograph`. An empty example means [gograph.dev](https://gograph.dev) has no page for that function.

| Question | Call | Constraint | Example |
|---|---|---|---|
| Order a graph that has no cycles | `TopologySort` | Vertices the edges do not order follow insertion order. `StableTopologySort` takes a compare function for that choice. | https://gograph.dev/algorithms/topological-sort/build-pipeline |
| Order a graph that has no cycles, picking the ready vertex compare prefers | `StableTopologySort` | compare decides among ready vertices. `TopologySort` follows insertion order instead. | https://gograph.dev/algorithms/stable-topological-sort/build-pipeline |
| Walk a graph that has no cycles, one vertex at a time | `traverse.NewTopologicalIterator` | The order is `TopologySort`. Use that function for the whole list. | https://gograph.dev/algorithms/topological-sort/build-pipeline |
| Keep a chosen set of vertices and the edges between them | `InducedSubgraph` | An edge with an endpoint outside the set is dropped. | |
| Distances from one source when no weight is negative | `path.Dijkstra` | Returns distances. A negative weight makes `path.BellmanFord` the call. The path itself is `path.DijkstraMultiSource`. | https://gograph.dev/algorithms/dijkstra/new-york |
| Distances from one source on a small dense graph | `path.DijkstraSimple` | It scans every vertex, O(V^2). `path.Dijkstra` is the call once the graph is large. | |
| Distances from one source when a weight can be negative | `path.BellmanFord` | A negative cycle is an error. It returns distances. All pairs are `path.FloydWarshall`. | https://gograph.dev/algorithms/bellman-ford/school-run |
| Distances between every pair | `path.FloydWarshall` | Directed and weighted. A negative cycle is an error. One source is `path.Dijkstra` or `path.BellmanFord`. | https://gograph.dev/algorithms/floyd-warshall/fx-majors |
| The nearest of several sources, and the path to it | `path.DijkstraMultiSource` | A negative weight is refused. `PathTo` returns the path. Distances from one source are `path.Dijkstra`. | |
| Place k centers so the farthest vertex is as close as this method can make it | `path.KCenter` | On an undirected graph the radius is at most twice the best. That factor does not hold when the graph is directed. | |
| Drop edges that a longer path already covers | `path.TransitiveReduction` | The graph is directed and has no cycle. Reachability stays the same. | https://gograph.dev/algorithms/transitive-reduction/build-pipeline |
| What depends on a vertex | `dag.Descendants` | An edge from A to B means B depends on A. What the vertex depends on is `dag.Ancestors`. | https://gograph.dev/algorithms/descendants/build-pipeline |
| What a vertex depends on | `dag.Ancestors` | An edge from A to B means B depends on A. What depends on the vertex is `dag.Descendants`. | https://gograph.dev/algorithms/ancestors/build-pipeline |
| What a change reaches | `dag.Affected` | The changed vertices come first, then everything reachable from them. The other direction is `dag.Ancestors`. | https://gograph.dev/algorithms/affected/build-pipeline |
| What can run at the same time, one level at a time | `dag.Levels` | A level waits until the whole previous level is done. `dag.NewTracker` starts a task when its own dependencies are done. | |
| Start a task when its own dependencies are done | `dag.NewTracker` | `Ready` lists what can start, `Done` marks it finished, and `Remaining` counts what is left. One goroutine calls them. `dag.Levels` waits for a whole level. | |
| The longest chain of dependent tasks | `dag.CriticalPath` | The cost is the stored vertex and edge weights, and an unset weight is 0. Costs kept outside the graph use `dag.CriticalPathFunc`. | |
| The longest chain when the costs are not stored on the graph | `dag.CriticalPathFunc` | A nil cost function means 0. Stored weights use `dag.CriticalPath`. | |
| Groups of vertices that can all reach each other | `connectivity.Tarjan` | One depth-first search. `connectivity.Kosaraju` and `connectivity.Gabow` find the same groups. | https://gograph.dev/algorithms/tarjan/service-calls |
| The same groups, with a second search on the reversed edges | `connectivity.Kosaraju` | Two searches. `connectivity.Tarjan` does the same job in one search. | https://gograph.dev/algorithms/kosaraju/service-calls |
| The same groups, with two stacks and one search | `connectivity.Gabow` | No low-link numbers. `connectivity.Tarjan` is the usual call. | https://gograph.dev/algorithms/gabow/service-calls |
| One vertex per group that can all reach each other | `connectivity.Condense` | The result is a directed acyclic graph. `connectivity.Tarjan` returns the groups and does not build that graph. | |
| Every set where each vertex is linked to every other | `partition.MaximalCliques` | Undirected. It lists every maximal clique, which is more than the single largest one. | https://gograph.dev/algorithms/maximal-cliques/friend-groups |
| Split an undirected graph by removing the busiest edges | `partition.GirvanNewman` | One edge comes out at a time until there are k components. A random cut is `partition.RandomizedKCut`. | https://gograph.dev/algorithms/girvan-newman/friend-groups |
| Cut an undirected graph into about k pieces at random | `partition.RandomizedKCut` | Another run can cut different edges. A graph already in more than k pieces stays that way. | https://gograph.dev/algorithms/randomized-k-cut/friend-groups |
| Visit vertices in layers from a start | `traverse.NewBreadthFirstIterator` | The edges are unweighted. Weighted distance uses `traverse.NewClosestFirstIterator`. | https://gograph.dev/algorithms/bfs/office-network |
| Follow one branch to its end before the next branch | `traverse.NewDepthFirstIterator` | It is not a shortest path and not layer by layer. | https://gograph.dev/algorithms/dfs/office-network |
| Visit the nearest remaining vertex by edge weight | `traverse.NewClosestFirstIterator` | Weights are not negative. Unweighted layers use `traverse.NewBreadthFirstIterator`. | https://gograph.dev/algorithms/closest-first/school-run |
| Follow random edges for a fixed number of steps | `traverse.NewRandomWalkIterator` | It does not visit every vertex. A full scan uses breadth-first or depth-first. | https://gograph.dev/algorithms/random-walk/service-calls |
| Write the graph as Graphviz DOT | `encoding/dot.Marshal` | `encoding/dot.Write` sends the same text to an io.Writer. A flowchart is `encoding/mermaid.Marshal`. | |
| Write the graph as a Mermaid flowchart | `encoding/mermaid.Marshal` | `encoding/mermaid.Write` sends the same text to an io.Writer. A large graph is too big for the diagram. Graphviz is `encoding/dot.Marshal`. | |
