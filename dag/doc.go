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
// # Behavior
//
// Descendants, Ancestors and Affected work on any directed graph, including
// graphs with cycles. Levels, NewTracker and CriticalPath return
// gograph.ErrDAGHasCycle for a graph with a cycle, since its vertices can't
// run in dependency order. All functions return gograph.ErrNotDirected for undirected graphs,
// and an error that matches gograph.ErrVertexDoesNotExist with errors.Is if
// a label is not in the graph.
//
// Every function returns each vertex once, as the graph's own vertex
// pointers. The results of Descendants, Ancestors and Affected are in
// breadth-first order, nearest first. For vertices at the same distance,
// Descendants and Affected follow the order the edges were added, and
// Ancestors follows the order of the graph's AllEdges.
//
// All searches are iterative, so long chains don't grow the call stack.
package dag
