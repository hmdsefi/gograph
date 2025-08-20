# gograph

## Path

### Transitive Reduction

"An algorithm for finding a minimal equivalent graph of a digraph." — Harry Hsu, Journal of the ACM, 22(1):11–16, Jan. 1975.

Here's a step-by-step explanation of how the transitive reduction algorithm works:

1. **Verify the graph is a DAG:** Transitive reduction as implemented here applies only to directed acyclic
   graphs. The algorithm checks whether the graph is directed and acyclic before proceeding.

2. **Find descendants efficiently:** For each vertex in the graph, the algorithm identifies its neighbors and
   their descendants using depth-first search traversal. Instead of computing the full transitive closure,
   this approach lazily computes descendants only when needed.

3. **Identify redundant edges:** For each vertex u and its direct neighbor v, the algorithm checks if any of
   v's descendants are also direct neighbors of u. If so, the edge from u to those descendants is redundant
   and can be removed, since there's already a path through v.

4. **Create the reduced graph:** The algorithm constructs a new graph with all vertices from the original
   graph but only includes non-redundant edges.

The time complexity of this transitive reduction algorithm is `O(V(V+E))`, where V is the number of vertices
and E is the number of edges in the graph. This is more efficient than the O(V³) approach using Floyd-Warshall,
especially for sparse graphs where E is much smaller than V².

For a directed acyclic graph (DAG), the transitive reduction is unique. This implementation is specifically
designed for DAGs and will return an error if applied to graphs with cycles or undirected graphs.
