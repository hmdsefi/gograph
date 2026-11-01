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
| Order a graph that has no cycles | `TopologySort` | Directed. Vertices the edges do not order follow insertion order. `StableTopologySort` takes a compare function for that choice. | https://gograph.dev/algorithms/topological-sort/build-pipeline |
| Order a graph that has no cycles, picking the ready vertex compare prefers | `StableTopologySort` | Directed. compare decides among ready vertices. `TopologySort` follows insertion order instead. | https://gograph.dev/algorithms/stable-topological-sort/build-pipeline |
| Walk a graph that has no cycles, one vertex at a time | `traverse.NewTopologicalIterator` | Directed. The order is `TopologySort`. Use that function for the whole list. | https://gograph.dev/algorithms/topological-sort/build-pipeline |
| Range over a graph that has no cycles in order | `traverse.Topological` | Directed. It sorts when called and returns the error of `TopologySort` before the loop starts, so stopping the loop early saves nothing. | https://gograph.dev/algorithms/topological-sort/build-pipeline |
| Keep a chosen set of vertices and the edges between them | `InducedSubgraph` | An edge with an endpoint outside the set is dropped. | |
| Change a copy of a graph and leave the original as it is | `Clone` | The vertices and edges are new, so look them up in the copy by label. Metadata is shared, so a map or pointer stored there changes in both. | |
| What changed between two versions of a graph | `Diff` | Matches vertices by label and edges by their two labels, and compares weights, edge labels and the weighted and acyclic properties. Metadata is left out. Both graphs are directed or both undirected. `WithWeightTolerance` accepts close weights. In a test, `gographtest.Equal` reports the same differences. | |
| Make the changes of a diff to another graph | `GraphDiff.Apply` | Returns a changed copy and leaves the graph as it is. Every removed or changed item has to be in the graph and every added one missing from it, or it returns `ErrDiffMismatch`. | |
| Loop over every vertex of a graph | `Vertices` | The graph's own vertices in `GetAllVertices` order, with nothing copied. A change to the graph during the loop can skip or repeat a vertex. `GetAllVertices` returns a slice to keep. | |
| Loop over every edge of a graph | `Edges` | In `AllEdges` order. An undirected edge comes once in each direction, and a self-loop once. A change to the graph during the loop can skip or repeat an edge. | |
| Loop over the vertices a vertex has an edge to | `Vertex.AllNeighbors` | The graph's own vertices, each once, in the order of `Neighbors`. `Neighbors` returns copies in a new slice. | |
| Distances from one source when no weight is negative | `path.Dijkstra` | Returns distances. A negative weight makes `path.BellmanFord` the call. The path itself is `path.DijkstraMultiSource`. | https://gograph.dev/algorithms/dijkstra/new-york |
| Distances from one source on a small dense graph | `path.DijkstraSimple` | It scans every vertex, O(V^2). `path.Dijkstra` is the call once the graph is large. | |
| Distances from one source when a weight can be negative | `path.BellmanFord` | A negative cycle is an error. It returns distances. All pairs are `path.FloydWarshall`. | https://gograph.dev/algorithms/bellman-ford/school-run |
| Distances between every pair | `path.FloydWarshall` | Directed and weighted. A negative cycle is an error. One source is `path.Dijkstra` or `path.BellmanFord`. | https://gograph.dev/algorithms/floyd-warshall/fx-majors |
| The nearest of several sources, and the path to it | `path.DijkstraMultiSource` | A negative weight is refused. `PathTo` returns the path. Distances from one source are `path.Dijkstra`. | https://gograph.dev/algorithms/multi-source-dijkstra/new-york |
| Place k centers so the farthest vertex is as close as this method can make it | `path.KCenter` | On an undirected graph the radius is at most twice the best. That factor does not hold when the graph is directed. | https://gograph.dev/algorithms/k-center/new-york |
| Drop edges that a longer path already covers | `path.TransitiveReduction` | The graph is directed and has no cycle. Reachability stays the same. | https://gograph.dev/algorithms/transitive-reduction/build-pipeline |
| What depends on a vertex | `dag.Descendants` | An edge from A to B means B depends on A. What the vertex depends on is `dag.Ancestors`. | https://gograph.dev/algorithms/descendants/build-pipeline |
| What a vertex depends on | `dag.Ancestors` | An edge from A to B means B depends on A. What depends on the vertex is `dag.Descendants`. | https://gograph.dev/algorithms/ancestors/build-pipeline |
| What a change reaches | `dag.Affected` | The changed vertices come first, then everything reachable from them. The other direction is `dag.Ancestors`. | https://gograph.dev/algorithms/affected/build-pipeline |
| What can run at the same time, one level at a time | `dag.Levels` | A level waits until the whole previous level is done. `dag.NewTracker` starts a task when its own dependencies are done. | https://gograph.dev/algorithms/levels/build-pipeline |
| Start a task when its own dependencies are done | `dag.NewTracker` | `Ready` lists what can start, `Done` marks it finished, and `Remaining` counts what is left. One goroutine calls them. `dag.Levels` waits for a whole level. | https://gograph.dev/algorithms/tracker/build-pipeline |
| The longest chain of dependent tasks | `dag.CriticalPath` | The cost is the stored vertex and edge weights, and an unset weight is 0. Costs kept outside the graph use `dag.CriticalPathFunc`. | https://gograph.dev/algorithms/critical-path/build-pipeline |
| The longest chain when the costs are not stored on the graph | `dag.CriticalPathFunc` | A nil cost function means 0. Stored weights use `dag.CriticalPath`. | |
| The pieces of an undirected graph, or of a directed graph when direction does not matter | `connectivity.ConnectedComponents` | Edge direction is ignored. Groups that can all reach each other are `connectivity.Tarjan`. | https://gograph.dev/algorithms/connected-components/home-network |
| Whether the graph has a cycle | `connectivity.HasCycle` | A self-loop counts. A graph made with `Acyclic()` has none. The cycle itself is `connectivity.FindCycle`. The order of a graph with no cycles is `TopologySort`. | https://gograph.dev/algorithms/has-cycle/build-pipeline |
| Show one cycle, such as an import cycle in an error | `connectivity.FindCycle` | Returns the vertices of the first cycle a depth-first search finds, which can be longer than the shortest one. The yes or no answer alone is `connectivity.HasCycle`, which is faster. | |
| Groups of vertices that can all reach each other | `connectivity.Tarjan` | One depth-first search. `connectivity.Kosaraju` and `connectivity.Gabow` find the same groups. | https://gograph.dev/algorithms/tarjan/service-calls |
| The same groups, with a second search on the reversed edges | `connectivity.Kosaraju` | Two searches. `connectivity.Tarjan` does the same job in one search. | https://gograph.dev/algorithms/kosaraju/service-calls |
| The same groups, with two stacks and one search | `connectivity.Gabow` | No low-link numbers. `connectivity.Tarjan` is the usual call. | https://gograph.dev/algorithms/gabow/service-calls |
| One vertex per group that can all reach each other | `connectivity.Condense` | The result is a directed acyclic graph. `connectivity.Tarjan` returns the groups and does not build that graph. | https://gograph.dev/algorithms/condensation/service-calls |
| Every set where each vertex is linked to every other | `partition.MaximalCliques` | Undirected. It lists every maximal clique, which is more than the single largest one. | https://gograph.dev/algorithms/maximal-cliques/friend-groups |
| Split an undirected graph by removing the busiest edges | `partition.GirvanNewman` | One edge comes out at a time until there are k components. A random cut is `partition.RandomizedKCut`. | https://gograph.dev/algorithms/girvan-newman/friend-groups |
| Cut an undirected graph into about k pieces at random | `partition.RandomizedKCut` | Another run can cut different edges. A graph already in more than k pieces stays that way. | https://gograph.dev/algorithms/randomized-k-cut/friend-groups |
| Visit vertices in layers from a start | `traverse.NewBreadthFirstIterator` | The edges are unweighted. Weighted distance uses `traverse.NewClosestFirstIterator`. | https://gograph.dev/algorithms/bfs/office-network |
| Range over the vertices a start reaches, in layers | `traverse.BFS` | The order of `traverse.NewBreadthFirstIterator`, in a `range` loop that can stop early. The number of edges from the start is `traverse.BFSWithDepth`. | https://gograph.dev/algorithms/bfs/office-network |
| How many edges each vertex a start reaches is from it | `traverse.BFSWithDepth` | Counts edges and ignores weights. Weighted distance is `traverse.ClosestFirst`. | https://gograph.dev/algorithms/bfs/office-network |
| Follow one branch to its end before the next branch | `traverse.NewDepthFirstIterator` | It is not a shortest path and not layer by layer. | https://gograph.dev/algorithms/dfs/office-network |
| Range over the vertices a start reaches, one branch at a time | `traverse.DFS` | The order of `traverse.NewDepthFirstIterator`. When the search enters and leaves each vertex, and the kind of each edge, are `traverse.DFSEvents`. | https://gograph.dev/algorithms/dfs/office-network |
| When a depth-first search enters and leaves each vertex, and the kind of each edge | `traverse.DFSEvents` | From one start. A directed graph has tree, back, forward and cross edges, and an undirected one tree and back edges. It keeps its own stack, so a long path is safe. The whole graph is `traverse.DFSEventsAll`. | |
| The same events over every vertex of a graph | `traverse.DFSEventsAll` | A new search starts from each vertex not found yet, in `GetAllVertices` order, so a cross edge can reach an earlier search. One start is `traverse.DFSEvents`. | |
| Visit the nearest remaining vertex by edge weight | `traverse.NewClosestFirstIterator` | Weights are not negative. Unweighted layers use `traverse.NewBreadthFirstIterator`. | https://gograph.dev/algorithms/closest-first/school-run |
| Range over the vertices by distance from a start, with each distance | `traverse.ClosestFirst` | The order of `traverse.NewClosestFirstIterator`. Weights are not negative, and an edge without a weight counts 0. Edge counts are `traverse.BFSWithDepth`. | https://gograph.dev/algorithms/closest-first/school-run |
| Follow random edges for a fixed number of steps | `traverse.NewRandomWalkIterator` | It does not visit every vertex. A full scan starts a breadth-first or depth-first iterator from each vertex in `GetAllVertices()` that no earlier one reached. | https://gograph.dev/algorithms/random-walk/service-calls |
| Write the graph as Graphviz DOT | `encoding/dot.Marshal` | `encoding/dot.Write` sends the same text to an io.Writer. A flowchart is `encoding/mermaid.Marshal`. | |
| Write the graph as a Mermaid flowchart | `encoding/mermaid.Marshal` | `encoding/mermaid.Write` sends the same text to an io.Writer. A large graph is too big for the diagram. Graphviz is `encoding/dot.Marshal`. | |
| Check in a test that an order is a valid topological order | `gographtest.IsTopologicalOrder` | Accepts any valid order, so the test keeps passing when the algorithm picks another one. The graph is directed. | |
| Check in a test that two lists of groups match in any order | `gographtest.SameGroups` | Compares labels and ignores the order of the groups and of the labels in each group. A repeated group or label has to repeat on the other side too. | |
| Check in a test that two graphs have the same vertices, edges and weights | `gographtest.Equal` | Compares through `Diff`, so the order the graphs were built in does not matter. The failure lists what differs. | |
| A path graph for a test | `gographtest.Path` | The labels 0 to n-1 and n-1 edges. A ring is `gographtest.Cycle`. | |
| A ring graph for a test | `gographtest.Cycle` | The path plus the edge from n-1 to 0. A graph made with `Acyclic()` refuses that edge, so it stays a path. | |
| A graph with an edge between every two vertices for a test | `gographtest.Complete` | A directed graph has both directions of each pair. With `Acyclic()`, only the edges from a lower label to a higher one stay. | |
| A grid graph for a test | `gographtest.Grid` | The vertex in row r and column c has the label r*cols + c. A directed grid points right and down. | |
| A random graph that is the same on every run, for a test | `gographtest.Random` | Each pair gets an edge with probability p, from a seeded `*rand.Rand` of math/rand/v2. A random graph with no cycles is `gographtest.RandomDAG`. | |
| A random graph with no cycles, for a test | `gographtest.RandomDAG` | Edges go from a lower label to a higher one. It is built with `Directed()`, without `Acyclic()`, so it keeps no topological order of its own. | |
