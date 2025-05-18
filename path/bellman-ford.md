# gograph

## Shortest Path

### Bellman-Ford

The Bellman-Ford algorithm is a graph algorithm used to find the shortest path from a source vertex to all other
vertices in a weighted graph, even in the presence of negative weight edges (as long as there are no negative weight
cycles). It was developed by Richard Bellman and Lester Ford Jr.

Here's a step-by-step explanation of how the Bellman-Ford algorithm works:

1. **Initialization:** Start by setting the distance of the source vertex to itself as 0,
   and the distance of all other vertices to infinity.

2. **Relaxation:** Iterate through all edges in the graph `|V| - 1` times, where `|V|` is
   the number of vertices. In each iteration, attempt to improve the shortest path estimates
   for all vertices. This is done by relaxing each edge: if the distance to the destination
   vertex through the current edge is shorter than the current estimate, update the estimate with the shorter distance.

3. **Detection of Negative Cycles:** After the `|V| - 1` iterations, perform an additional iteration.
   If during this iteration, any of the distances are further reduced, it indicates the presence of a
   negative weight cycle in the graph. This is because if a vertex's distance can still be improved
   after `|V| - 1` iterations, it means there's a negative weight cycle that can be traversed indefinitely
   to reduce the distance further.

4. **Output:** If there is no negative weight cycle, the algorithm outputs the shortest path
   distances from the source vertex to all other vertices. If there is a negative weight cycle,
   the algorithm typically returns an indication of this fact.

The time complexity of the Bellman-Ford algorithm is `O(V*E)`, where V is the number of vertices and E
is the number of edges. This makes it less efficient than algorithms like Dijkstra's algorithm for
graphs with non-negative edge weights, but its ability to handle negative weight edges (as long as
there are no negative weight cycles) makes it useful in certain scenarios.


