# gograph

## Girvan–Newman Community Detection Algorithm
The Girvan–Newman algorithm is a graph partitioning algorithm used for community detection in undirected networks. Communities (or clusters) are groups of nodes that are more densely connected internally than with the rest of the graph. The algorithm identifies edges that are likely “bridges” between communities and removes them iteratively to reveal clusters.

### Graph Algorithm Type
- **Type:** Community detection / Partitioning algorithm
- **Graph Requirements:**
  - Works primarily on undirected graphs
  - Can be adapted for weighted graphs (edge weights influence betweenness)
  - Input graph should allow iteration over vertices and edges

### Usage / Applications
- **Social Networks:** Detect communities or friend groups.
- **Biological Networks:** Identify functional modules (e.g., protein-protein interaction networks).
- **Recommendation Systems:** Discover user or item clusters.
- **Infrastructure Analysis:** Identify weak points in transportation, electrical, or communication networks.

### Advantages
- Intuitive and easy to implement.
- Detects overlapping communities indirectly through edge removal.
- Works on weighted graphs when betweenness is computed using edge weights.

### Limitations
- High computational cost for large graphs.
- Sensitive to noisy data (edges with similar betweenness may be removed arbitrarily).
- Primarily designed for undirected graphs; adaptations are needed for directed graphs.

### Time Complexity
Let V be the number of vertices and E the number of edges:
- **Edge Betweenness Computation:** O(V * (V + E)) per iteration (BFS from each node using Brandes’ algorithm).
- **Edge Removal Iterations:** In the worst case, up to O(E) edges may be removed one at a time.
- **Total Worst-Case Time Complexity:** O(E * V * (V + E)).

### Space Complexity
- **Graph storage (cloned graph):** O(V + E)
- **BFS queues and dependency maps:** O(V + E)
- **Edge betweenness storage:** O(E)
- **Total Space Complexity:** O(V + E)

### How It Works

1. Compute Edge Betweenness
   - Edge betweenness measures the number of shortest paths that pass through each edge.
   - High-betweenness edges often connect different communities.
   - Computed efficiently using Brandes’ algorithm with BFS from each vertex.

2. Remove High Betweenness Edge(s)
   - Remove the edge with the maximum betweenness centrality.
   - This step gradually disconnects the graph along community boundaries.

3. Update Connected Components
   - After removing edges, determine connected components using non-recursive BFS.
   - Each connected component represents a potential community.

4. Repeat
   - Recompute edge betweenness in the modified graph.
   - Continue removing edges until the desired number of communities k is reached, or until no edges remain.

<div align="center">
<img width="1380" height="3839" alt="Image" src="https://github.com/user-attachments/assets/438a87b9-9c30-46a9-9734-a70cc1f15300" />
</div>

### Example


### References
1. Girvan, M., & Newman, M. E. J. (2002). Community structure in social and biological networks. Proceedings of the National Academy of Sciences, 99(12), 7821–7826.
2. Brandes, U. (2001). A Faster Algorithm for Betweenness Centrality. Journal of Mathematical Sociology, 25(2), 163–177.