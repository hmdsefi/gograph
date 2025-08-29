# gograph

## Girvan–Newman Community Detection Algorithm

The Girvan–Newman algorithm is a graph partitioning algorithm used for community detection in undirected networks.
Communities (or clusters) are groups of nodes that are more densely connected internally than with the rest of the
graph. The algorithm identifies edges that are likely “bridges” between communities and removes them iteratively to
reveal clusters.

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

1. **Compute Edge Betweenness**
    - Edge betweenness measures the number of shortest paths that pass through each edge.
    - High-betweenness edges often connect different communities.
    - Computed efficiently using Brandes’ algorithm with BFS from each vertex.

2. **Remove High Betweenness Edge(s)**
    - Remove the edge with the maximum betweenness centrality.
    - This step gradually disconnects the graph along community boundaries.

3. **Update Connected Components**
    - After removing edges, determine connected components using non-recursive BFS.
    - Each connected component represents a potential community.

4. **Repeat**
    - Recompute edge betweenness in the modified graph.
    - Continue removing edges until the desired number of communities k is reached, or until no edges remain.

<div align="center">
<img width="345" height="959" alt="Image" src="https://github.com/user-attachments/assets/438a87b9-9c30-46a9-9734-a70cc1f15300" />
</div>

### Example

**k=3**
**Vertices:** A, B, C, D, E, F, G, H

#### Step 0: Initial Graph

<div align="center">
<img width="781" height="121" alt="Image" src="https://github.com/user-attachments/assets/198549ac-c8e7-4618-bfa7-7a517a445913" />
</div>

- Action: clone graph to avoid mutating original.
- No traversal yet.

#### Step 1: Compute Edge Betweenness

**Algorithm:** **Brandes’ algorithm** uses **BFS** (Breadth-First Search) to compute the shortest paths **from each
vertex**.

**Process:**

- For each vertex v in V:
    - Run BFS to find shortest paths from v to all other vertices.
    - Keep track of the number of shortest paths reaching each vertex.
    - Accumulate dependency scores backward (like a reverse BFS) to compute contribution to each edge.

Resulting approximate edge betweenness values:
| Edge | Betweenness |
| ---- | ----------- |
| A-B | 1.0 |
| A-C | 1.0 |
| B-C | 1.0 |
| C-D | 7.0 |
| D-E | 5.0 |
| E-F | 4.0 |
| F-G | 3.0 |
| G-H | 2.0 |
| E-H | 3.5 |

**Observation:** C-D has highest betweenness → bridge between triangle A-B-C and linear chain.

#### Step 2: Remove Maximum Betweenness Edge (C-D)

<div align="center">
<img width="501" height="211" alt="Image" src="https://github.com/user-attachments/assets/aeb372fc-3d8b-48f7-9e9e-fcb66534f6ea" />
</div>

**Connected Components:** computed using non-recursive BFS:

- Start from unvisited node, traverse neighbors using a queue.
- Mark visited vertices.
- All vertices visited in the BFS form a component.

**Components:**

1. `{A, B, C}`
2. `{D, E, F, G, H}`

#### Step 3: Recompute Betweenness in Remaining Graphs

**Traversal:** again BFS from each node inside the working graph, recompute shortest paths and edge betweenness.

New betweenness values (approx):
| Edge | Betweenness |
| ---- | ----------- |
| D-E | 5.0 |
| E-F | 4.0 |
| F-G | 3.0 |
| G-H | 2.0 |
| E-H | 5.0 |

**Max edges:** `D-E` and `E-H` → remove one (let’s choose E-H).

#### Step 4: Remove E-H

<div align="center">
<img width="501" height="191" alt="Image" src="https://github.com/user-attachments/assets/22127f8c-0cc8-4826-8801-0eb9de0203ee" />
</div>

`Components:` computed using **BFS** again.
Still connected: `{D, E, F, G, H}`

#### Step 5: Remove D-E

<div align="center">
<img width="396" height="291" alt="Image" src="https://github.com/user-attachments/assets/e6064007-4115-4363-959b-d41abb02cc64" />
</div>

**Components:**

1. `{A, B, C}`
2. `{D}`
3. `{E, F, G, H}`

Components = 3 → stop (k=3).

#### Step 6: Final Partition

**Components:**

1. `{A, B, C}`
2. `{D}`
3. `{E, F, G, H}`

<div align="center">
<img width="551" height="411" alt="Image" src="https://github.com/user-attachments/assets/54380914-3f8c-4b6e-9dfd-e4ac15a51877" />
</div>

#### Step 7: Summary of Traversals

| Step | Purpose                                | Traversal                              |
|------|----------------------------------------|----------------------------------------|
| 1    | Compute edge betweenness               | BFS (Brandes) from each node           |
| 2    | Identify components                    | Non-recursive BFS from unvisited nodes |
| 3    | Recompute betweenness                  | BFS in each component                  |
| 4–5  | Remove max edges and update components | BFS for components                     |

**Key Points for `k=3`:**
- Stop removing edges as soon as the number of connected components reaches k.
- BFS ensures no recursion and no stack overflow, even with larger graphs. 
- High-betweenness edges (C-D, E-H, D-E) are removed first → separates weakly connected communities.

### References

1. Girvan, M., & Newman, M. E. J. (2002). Community structure in social and biological networks. Proceedings of the
   National Academy of Sciences, 99(12), 7821–7826.
2. Brandes, U. (2001). A Faster Algorithm for Betweenness Centrality. Journal of Mathematical Sociology, 25(2), 163–177.