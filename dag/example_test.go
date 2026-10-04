package dag_test

import (
	"fmt"
	"sort"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/dag"
)

func labels(vertices []*gograph.Vertex[string]) []string {
	out := make([]string, len(vertices))
	for i, v := range vertices {
		out[i] = v.Label()
	}
	return out
}

func Example() {
	// An edge A -> B means B depends on A.
	g := gograph.New[string](gograph.Acyclic())

	checkout := g.AddVertexByLabel("checkout")
	build := g.AddVertexByLabel("build")
	test := g.AddVertexByLabel("test")
	lint := g.AddVertexByLabel("lint")

	_, _ = g.AddEdge(checkout, build)
	_, _ = g.AddEdge(build, test)
	_, _ = g.AddEdge(build, lint)

	descendants, _ := dag.Descendants(g, "build")
	fmt.Println("depends on build:", labels(descendants))

	ancestors, _ := dag.Ancestors(g, "test")
	fmt.Println("test depends on:", labels(ancestors))

	affected, _ := dag.Affected(g, "build")
	fmt.Println("rebuild after a build change:", labels(affected))

	// Output:
	// depends on build: [test lint]
	// test depends on: [build checkout]
	// rebuild after a build change: [build test lint]
}

func pipeline() gograph.Graph[string] {
	g := gograph.New[string](gograph.Directed())
	for _, e := range [][2]string{
		{"checkout", "build"},
		{"checkout", "docs"},
		{"build", "test"},
		{"build", "lint"},
	} {
		_, _ = g.AddEdge(gograph.NewVertex(e[0]), gograph.NewVertex(e[1]))
	}
	return g
}

func ExampleLevels() {
	levels, err := dag.Levels(pipeline())
	if err != nil {
		fmt.Println(err)
		return
	}

	// the vertices of a level can run at the same time
	for i, level := range levels {
		fmt.Println(i, labels(level))
	}

	// Output:
	// 0 [checkout]
	// 1 [build docs]
	// 2 [test lint]
}

func ExampleTracker() {
	t, err := dag.NewTracker(pipeline())
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("ready:", labels(t.Ready()))
	_ = t.Done("checkout")
	fmt.Println("ready:", labels(t.Ready()))

	// test and lint start without waiting for docs
	_ = t.Done("build")
	fmt.Println("ready:", labels(t.Ready()))
	fmt.Println("remaining:", t.Remaining())

	// Output:
	// ready: [checkout]
	// ready: [build docs]
	// ready: [test lint]
	// remaining: 3
}

func ExampleTracker_workers() {
	g := pipeline()
	t, err := dag.NewTracker(g)
	if err != nil {
		fmt.Println(err)
		return
	}

	run := func(label string) error {
		return nil // run the task here
	}

	type result struct {
		label string
		err   error
	}
	// buffered, so workers don't block if the loop returns early
	results := make(chan result, g.Order())

	var finished []string
	for t.Remaining() > 0 {
		for _, v := range t.Ready() {
			go func() {
				results <- result{label: v.Label(), err: run(v.Label())}
			}()
		}

		r := <-results
		if r.err != nil {
			fmt.Println(r.label, "failed:", r.err)
			return
		}
		if err := t.Done(r.label); err != nil {
			fmt.Println(err)
			return
		}
		finished = append(finished, r.label)
	}

	sort.Strings(finished)
	fmt.Println(finished)

	// Output:
	// [build checkout docs lint test]
}

func ExampleCriticalPath() {
	// the weight of a vertex is how long the task takes
	g := gograph.New[string](gograph.Directed())
	checkout := g.AddVertexByLabel("checkout", gograph.WithVertexWeight(1))
	build := g.AddVertexByLabel("build", gograph.WithVertexWeight(10))
	test := g.AddVertexByLabel("test", gograph.WithVertexWeight(20))
	lint := g.AddVertexByLabel("lint", gograph.WithVertexWeight(2))

	_, _ = g.AddEdge(checkout, build)
	_, _ = g.AddEdge(build, test)
	_, _ = g.AddEdge(build, lint)

	path, cost, err := dag.CriticalPath(g)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(labels(path), cost)

	// Output:
	// [checkout build test] 31
}

func ExampleCriticalPathFunc() {
	// durations kept outside the graph
	duration := map[string]float64{"checkout": 1, "build": 10, "test": 20, "lint": 2}

	path, cost, err := dag.CriticalPathFunc(pipeline(),
		func(v *gograph.Vertex[string]) float64 { return duration[v.Label()] },
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(labels(path), cost)

	// Output:
	// [checkout build test] 31
}
