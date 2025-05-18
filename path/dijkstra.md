# gograph

## Shortest Path

### Dijkstra

Dijkstra's algorithm is a graph algorithm used to find the shortest path from a single source vertex to all
other vertices in a weighted graph with non-negative edge weights. It was developed by Dutch computer scientist
Edsger W. Dijkstra in 1956.

Here's a step-by-step explanation of how Dijkstra's algorithm works:

1. **Initialization:** Start by selecting a source vertex. Set the distance of the source vertex to itself as 0, and the
   distances of all other vertices to infinity. Maintain a priority queue (or a min-heap) to keep track of vertices
   based on their tentative distances from the source vertex.

2. **Selection of Vertex:** At each step, select the vertex with the smallest tentative distance from the priority
   queue.
   Initially, this will be the source vertex.

3. **Relaxation:** For the selected vertex, iterate through all its neighboring vertices. For each neighboring vertex,
   update its tentative distance if going through the current vertex results in a shorter path than the current
   known distance. If the tentative distance is updated, update the priority queue accordingly.

4. **Repeating Steps:** Repeat steps 2 and 3 until all vertices have been visited or until the priority queue is empty.

5. **Output:** After the algorithm terminates, the distances from the source vertex to all other vertices will be
   finalized.

Dijkstra's algorithm guarantees the shortest path from the source vertex to all other vertices in the graph,
as long as the graph does not contain negative weight edges. It works efficiently for sparse graphs with
non-negative edge weights.

The time complexity of Dijkstra's algorithm is `O((V + E) log V)`, where V is the number of vertices and E is
the number of edges in the graph. This complexity arises from the use of a priority queue to maintain the tentative
distances efficiently. If a simple array-based implementation is used to select the minimum distance vertex in each
step, the time complexity becomes O(V^2), which is more suitable for dense graphs.

#### Implementation with Slices

Steps:

1. Initialize distances with maximum float/integer values except for the source vertex distance, which is set to 0.
2. Iterate numVertices - 1 times (where numVertices is the number of vertices in the graph).
3. Select the vertex with the minimum distance among the unvisited vertices.
4. Relax the distances of its neighboring vertices if a shorter path is found.
5. Mark the selected vertex as visited.

**Time Complexity:** `O(V^2)`, where V is the number of vertices in the graph. This is because finding the
minimum distance vertex in each iteration takes `O(V)` time, and we perform this process V times.

**Space Complexity:** `O(V^2)` for storing the graph and distances.

#### Implementation with Heap

Steps:

* Similar to the slice implementation but instead of linearly searching for the vertex with the minimum distance,
  we use a heap to maintain the priority queue.
* The priority queue ensures that the vertex with the smallest tentative distance is efficiently selected
  in each iteration.

**Time Complexity:** `O((V + E) log V)`, where V is the number of vertices and E is the number of edges in
the graph. This is because each vertex is pushed and popped from the priority queue once, and each
edge is relaxed once.

**Space Complexity:** `O(V)` for storing the priority queue and distances.
