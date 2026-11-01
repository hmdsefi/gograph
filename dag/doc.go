// Package dag answers the everyday questions about dependency graphs: what a
// vertex depends on, what depends on it, what a change affects, and which
// vertices can run at the same time.
//
// # Edge direction
//
// The functions follow the convention of gograph.TopologySort: an edge
// A -> B means A comes before B, or "B depends on A". So:
//
//   - Descendants of A are the vertices reachable from A, which is
//     everything that depends on A.
//   - Ancestors of B are the vertices that can reach B, which is
//     everything B depends on.
//   - Affected by a change to A are A and its descendants, which is
//     everything that has to be rebuilt, retested or redeployed.
//
// For example, with the edges checkout -> build, build -> test and
// build -> lint:
//
//   - Descendants(g, "build") returns test and lint.
//   - Ancestors(g, "test") returns build and checkout.
//   - Affected(g, "build") returns build, test and lint.
//
// If a graph stores "A requires B" as A -> B instead, the meanings swap:
// Ancestors returns the dependents and Descendants returns the dependencies.
//
// # Running in dependency order
//
// Levels and Tracker help to run the vertices with plain goroutines or any
// other executor, starting each one after everything it depends on:
//
//   - Levels groups the vertices into levels that can each run at the same
//     time, one level after the other. In the example above, the levels
//     are [checkout], [build] and [test lint].
//   - Tracker hands out each vertex as soon as its own dependencies are
//     done, so a slow vertex doesn't hold up the ones that don't depend
//     on it.
//
// # Critical path
//
// CriticalPath finds the chain of dependent vertices with the largest total
// cost, which is what sets the total time when everything else runs in
// parallel. It adds up vertex and edge weights, and CriticalPathFunc takes
// the costs from functions, for durations kept outside the graph. In the
// example above, with checkout, build, test and lint taking 1, 10, 20 and 2,
// the critical path is checkout, build, test.
//
// # Sources and sinks
//
// Sources and Sinks answer which vertices have no prerequisites (where to
// start) and which have no dependents (final outputs or leftovers):
//
//   - Sources returns the vertices with no incoming edges (InDegree == 0).
//   - Sinks returns the vertices with no outgoing edges (OutDegree == 0).
//
// In the example above, Sources(g) returns checkout, and Sinks(g) returns test
// and lint.
//
// # Behavior
//
// Descendants, Ancestors, Affected, Sources and Sinks work on any directed
// graph, including graphs with cycles. Levels, NewTracker and CriticalPath
// return gograph.ErrDAGHasCycle for a graph with a cycle, since its vertices
// can't run in dependency order. All functions return gograph.ErrNotDirected
// for undirected graphs, and an error that matches
// gograph.ErrVertexDoesNotExist with errors.Is if a label is not in the graph.
//
// Sources and Sinks follow the order of GetAllVertices. Vertices on a cycle
// are never sources or sinks, a self-loop counts as an incoming and an
// outgoing edge, and isolated vertices are both. An empty graph returns an
// empty slice and no error.
//
// Every function returns each vertex once, as the graph's own vertex
// pointers. The results of Descendants, Ancestors and Affected are in
// breadth-first order, nearest first. The search visits the outgoing edges
// of each vertex in the order they were added (Descendants, Affected), or
// its incoming edges in AllEdges order (Ancestors), so vertices at the same
// distance are grouped by the vertex that reached them first.
//
// All searches are iterative, so long chains don't grow the call stack.
package dag
