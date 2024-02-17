# gograph
## Shortest Path
### Floyd-Warshall
The Floyd-Warshall algorithm is a dynamic programming algorithm used to find the shortest paths between 
all pairs of vertices in a weighted graph, even in the presence of negative weight edges (as long as there
are no negative weight cycles). It was proposed by Robert Floyd and Stephen Warshall.

Here's a step-by-step explanation of how the Floyd-Warshall algorithm works:

1. **Initialization:** Create a distance matrix `D[][]` where `D[i][j]` represents the shortest distance between
vertex i and vertex j. Initialize this matrix with the weights of the edges between vertices if there 
is an edge, otherwise set the value to infinity. Also, set the diagonal elements `D[i][i]` to 0.

2. **Shortest Path Calculation:** Iterate through all vertices as intermediate vertices. For each pair of 
 vertices (i, j), check if going through the current intermediate vertex k leads to a shorter path than
 the current known distance from i to j. If so, update the distance matrix `D[i][j]` to the new shorter
 distance `D[i][k] + D[k][j]`.

3. **Detection of Negative Cycles:** After the iterations, if any diagonal element `D[i][i]` of the distance 
 matrix is negative, it indicates the presence of a negative weight cycle in the graph.

4. **Output:** The resulting distance matrix `D[][]` will contain the shortest path distances between all 
 pairs of vertices. If there is a negative weight cycle, it might not produce the correct shortest paths, 
 but it can still detect the presence of such cycles.

The time complexity of the Floyd-Warshall algorithm is O(V^3), where V is the number of vertices in the graph.
Despite its cubic time complexity, it is often preferred over other algorithms like Bellman-Ford for dense
graphs or when the graph has negative weight edges and no negative weight cycles, as it calculates shortest
paths between all pairs of vertices in one go.