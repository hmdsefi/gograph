// Package traverse visits the vertices of a gograph.Graph in breadth-first,
// depth-first, closest-first, topological and random walk order.
//
// Each traversal except the random walk comes in two forms. An iterator,
// such as the one NewBreadthFirstIterator returns, hands out one vertex per
// Next call and can be Reset. A sequence, such as the one BFS returns, is
// for a range loop. BFS, BFSWithDepth, DFS and ClosestFirst do their work as
// the loop asks for the next vertex, so breaking out of the loop ends the
// traversal there, and ranging over the sequence again runs the traversal
// again from the start. Topological sorts the graph when it is called.
//
// DFSEvents and DFSEventsAll report each step of a depth-first search: when
// a vertex is discovered and finished, and whether each edge is a tree,
// back, forward or cross edge. Algorithms such as cycle finding and
// topological sort by finish time are a loop over these events.
//
// Each vertex is the graph's own vertex, the one GetVertexByID returns for
// its label. Changing the graph during a traversal isn't supported, with one
// exception: a vertex removed from the graph after the traversal queued it
// is skipped.
package traverse
