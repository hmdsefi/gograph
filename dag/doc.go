// Package dag answers the everyday questions about dependency graphs: what a
// vertex depends on, what depends on it, and what a change affects.
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
// # Behavior
//
// The functions work on any directed graph, including graphs with cycles,
// and return each vertex once, as the graph's own vertex pointers. They
// return gograph.ErrNotDirected for undirected graphs, and an error that
// matches gograph.ErrVertexDoesNotExist with errors.Is if a label is not in
// the graph.
//
// The results are in breadth-first order, nearest first. For vertices at
// the same distance, Descendants and Affected follow the order the edges
// were added, and Ancestors follows the order of the graph's AllEdges.
//
// All searches are iterative, so long chains don't grow the call stack.
package dag
