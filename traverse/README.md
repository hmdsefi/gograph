# gograph - Traverse

Graph traversal is an important operation in computer science that involves
visiting each node in a graph at least once. There are several graph traversal
algorithms that are commonly used to explore graphs in different ways.
The `gograph/traverse` package is a Go library that provides efficient implementation
of five of these algorithms:

* [BFS](#BFS)
* [DFS](#DFS)
* [Topological Sort](#Topological-Sort)
* [Closest First](#Closest-First)
* [Random Walk](#Random-Walk)

All the traversal algorithms in the 'traverse' package are implemented the following
iterator interface:

```go
// Iterator represents a general purpose iterator for iterating over
// a sequence of graph's vertices. It provides methods for checking if
// there are more elements to be iterated over, getting the next element,
// iterating over all elements using a callback function, and resetting
// the iterator to its initial state.
type Iterator[T comparable] interface {
	// HasNext returns a boolean value indicating whether there are more
	// elements to be iterated over. It returns true if there are more
	// elements. Otherwise, returns false.
	HasNext() bool

	// Next returns the next element in the sequence being iterated over.
	// If there are no more elements, it returns nil. It also advances
	// the iterator to the next element.
	Next() *gograph.Vertex[T]

	// Iterate iterates over all elements in the sequence and calls the
	// provided callback function on each element. The callback function
	// takes a single argument of type *Vertex, representing the current
	// element being iterated over. It returns an error value, which is
	// returned by the Iterate method. If the callback function returns
	// an error, iteration is stopped and the error is returned.
	Iterate(func(v *gograph.Vertex[T]) error) error

	// Reset  resets the iterator to its initial state, allowing the
	// sequence to be iterated over again from the beginning.
	Reset()
}
```

## BFS

BFS iterator is a technique used to implement the Breadth-First Search (BFS)
algorithm for traversing a graph or tree in a systematic way. The BFS iterator
iteratively visits all the vertices of the graph, starting from a given source
vertex and moving outwards in levels. The iterator maintains a queue of vertices
to be visited, and it visits each vertex only once.

One of the most common usages of BFS iterator is to find the shortest path between
two vertices in an unweighted graph. It can also be used to perform level-order
traversal of a binary tree or to find all the vertices at a given distance from a
starting vertex. BFS iterator is also commonly used in graph algorithms, such as
finding connected components, bipartite graphs, and cycle detection.

The time complexity of the BFS iterator algorithm is O(V + E), where V is the
number of vertices and E is the number of edges in the graph. This is because
the algorithm visits each vertex and edge exactly once. The space complexity of
BFS iterator is O(V), where V is the number of vertices in the graph. This is
because the algorithm uses a queue to store the vertices to be visited. The
maximum size of the queue is equal to the number of vertices at the maximum
depth of the BFS traversal.

Here you can see how BFS iterator works:
<img alt="golang generic graph package - BFS traversal" src="https://user-images.githubusercontent.com/11541936/222957305-912411f0-00fe-419e-97f7-5e3fbdab62af.png" title="bfs-traversal"/>

