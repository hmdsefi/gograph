# gograph - Connectivity

### Why Strong Connectivity Is Important?

The concept of strong connectivity in a graph is important for a variety of reasons. Here are some of the
main reasons why strong connectivity is important:

- **Robustness:** A graph that is strongly connected is more robust and resilient than a graph that is not
  strongly connected. This is because in a strongly connected graph, every vertex can be reached from every
  other vertex, which means that even if some vertices or edges are removed, the graph remains connected.
- **Communication:** In many applications, such as network routing and communication networks, it is important
  to be able to communicate between any pair of vertices in the graph. A strongly connected graph ensures that
  there is a path between any two vertices, which makes communication between them possible.
- **Analysis:** Strong connectivity is often used as a tool for analyzing the structure of a graph. For example,
  identifying strongly connected components in a graph can reveal important patterns and relationships between vertices.
- **Optimization:** In some applications, it may be necessary to find the shortest path or the minimum spanning
  tree that connects all vertices in the graph. Strong connectivity can help to optimize such algorithms by
  reducing the search space and ensuring that all vertices are reachable.
- **Design of algorithms:** Strong connectivity is also an important concept in the design of algorithms for
  various graph problems. For example, many graph algorithms rely on identifying strongly connected components
  or paths in the graph to find optimal solutions.
- **Network reliability:** In network reliability analysis, strong connectivity is used to determine the
  probability that a network will remain connected in the event of failures or disruptions. This is important
  in designing robust networks for critical applications such as transportation, energy distribution, and
  telecommunications.
- **Social network analysis:** Strong connectivity is also important in social network analysis, where it is
  used to identify closely-knit groups or communities within a larger network. This can provide insights into
  social structures and relationships that are not immediately apparent from the graph topology.
- **Testing and verification:** Strong connectivity is an important concept in testing and verification of digital
  circuits, where it is used to ensure that all nodes in the circuit can be reached from all other nodes.
  This is important for ensuring correct functionality and avoiding problems such as deadlocks or race conditions.
- **Transportation and logistics:** Strong connectivity is also important in transportation and logistics,
  where it is used to optimize routing and scheduling of vehicles or goods. In a strongly connected graph,
  it is possible to find the shortest path or the minimum time required to transport goods between any two locations.
- **Graph drawing and visualization:** Strong connectivity is also an important consideration in graph drawing and
  visualization. Many graph layout algorithms aim to minimize edge crossings while maintaining strong connectivity,
  which can improve the readability and aesthetics of the graph.

Strong connectivity is a fundamental concept in graph theory that is used to describe the property of a
directed graph where every vertex is reachable from every other vertex through a directed path. Strong
connectivity has a wide range of practical applications, such as designing robust networks, optimizing
transportation routes, analyzing social structures, and testing digital circuits.

Several algorithms, such as Tarjan's Algorithm, Kosaraju's Algorithm, Path-Based Strong Component Algorithm,
and Gabow's Algorithm, are available to determine the strong connectivity of a graph, and the choice of
algorithm depends on the specific characteristics of the graph and the requirements of the application.

Understanding strong connectivity and its implications is essential for anyone working with graphs or
network data. Strongly connected graphs are more robust and resilient, enable communication between
any pair of vertices, and provide insights into social structures and relationships. Additionally,
strong connectivity is important in optimizing algorithms, network reliability, transportation and
logistics, graph drawing, and verification of digital circuits.

### Tarjan's Algorithm

Tarjan's algorithm is a popular algorithm in graph theory used to find strongly connected components
in a directed graph. The algorithm is named after its inventor, Robert Tarjan. The algorithm is based
on depth-first search (DFS) and is very efficient in both time and space complexity.

The main usage of the Tarjan algorithm is to find strongly connected components in a directed graph.
Strongly connected components are used in many applications, such as finding the shortest path between
two nodes in a graph, identifying the critical paths in a project schedule, and solving problems related
to synchronization in computer science.

The time complexity of Tarjan's algorithm is O(V + E), where V is the number of vertices in the graph
and E is the number of edges in the graph. The space complexity of the algorithm is O(V), where V is
the number of vertices in the graph.

To use Tarjan algorithm, you can call the 'Tarjan[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T]' function
and path your graph to it:

```go
import (
  "github.com/hmdsefi/gograph"
  "github.com/hmdsefi/gograph/connectivity"
)

func main() {
  g := gograph.New[int](gograph.Directed())
  
  v1 := g.AddVertexByLabel(1)
  v2 := g.AddVertexByLabel(2)
  v3 := g.AddVertexByLabel(3)
  v4 := g.AddVertexByLabel(4)
  v5 := g.AddVertexByLabel(5)
  
  g.AddEdge(v1, v2)
  g.AddEdge(v2, v3)
  g.AddEdge(v3, v1)
  g.AddEdge(v3, v4)
  g.AddEdge(v4, v5)
  g.AddEdge(v5, v4)
  
  sccs := connectivity.Tarjan(g)
}
```

It returns a slice of strongly connected component.

### Kosaraju's Algorithm

Kosaraju's algorithm is another popular algorithm in graph theory used to find strongly connected
components in a directed graph. The algorithm is named after its inventor, Sharadha Sharma Kosaraju.
The algorithm is also based on depth-first search (DFS), but it performs two DFS passes over the graph.

The main usage of Kosaraju's algorithm is to find strongly connected components in a directed graph.
Strongly connected components are used in many applications, such as finding the shortest path between
two nodes in a graph, identifying the critical paths in a project schedule, and solving problems related
to synchronization in computer science.

The time complexity of Kosaraju's algorithm is O(V + E), where V is the number of vertices in the graph
and E is the number of edges in the graph. The space complexity of the algorithm is O(V), where V is
the number of vertices in the graph.

To use Kosaraju algorithm, you can call the 'Kosaraju[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T]' function
and path your graph to it:

```go
import (
  "github.com/hmdsefi/gograph"
  "github.com/hmdsefi/gograph/connectivity"
)

func main() {
  g := gograph.New[int](gograph.Directed())
  
  v1 := g.AddVertexByLabel(1)
  v2 := g.AddVertexByLabel(2)
  v3 := g.AddVertexByLabel(3)
  v4 := g.AddVertexByLabel(4)
  v5 := g.AddVertexByLabel(5)
  
  g.AddEdge(v1, v2)
  g.AddEdge(v2, v3)
  g.AddEdge(v3, v1)
  g.AddEdge(v3, v4)
  g.AddEdge(v4, v5)
  g.AddEdge(v5, v4)
  
  sccs := connectivity.Kosaraju(g)
}
```

It returns a slice of strongly connected component.

### Gabow's Algorithm

Gabow's algorithm is another algorithm used to find strongly connected components (SCCs) in a directed graph.
The algorithm was invented by Harold N. Gabow in 1985 and is based on a combination of breadth-first search (BFS)
and depth-first search (DFS).

The main usage of Gabow's algorithm is to find strongly connected components in a directed graph. It can be used
in many applications, such as detecting cycles in a graph, solving problems related to synchronization,
and analyzing network traffic.

The time complexity of Gabow's algorithm is O(V + E), where V is the number of vertices in the graph and E
is the number of edges in the graph. The space complexity of the algorithm is O(V), where V is the number
of vertices in the graph.

To use Gabow algorithm, you can call the 'Gabow[T comparable](g gograph.Graph[T]) [][]*gograph.Vertex[T]' function
and path your graph to it:

```go
import (
  "github.com/hmdsefi/gograph"
  "github.com/hmdsefi/gograph/connectivity"
)

func main() {
  g := gograph.New[int](gograph.Directed())
  
  v1 := g.AddVertexByLabel(1)
  v2 := g.AddVertexByLabel(2)
  v3 := g.AddVertexByLabel(3)
  v4 := g.AddVertexByLabel(4)
  v5 := g.AddVertexByLabel(5)
  
  g.AddEdge(v1, v2)
  g.AddEdge(v2, v3)
  g.AddEdge(v3, v1)
  g.AddEdge(v3, v4)
  g.AddEdge(v4, v5)
  g.AddEdge(v5, v4)
  
  sccs := connectivity.Gabow(g)
}
```

It returns a slice of strongly connected component.
