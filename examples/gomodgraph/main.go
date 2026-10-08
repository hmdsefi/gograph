// Command gomodgraph loads the output of "go mod graph" into a gograph graph
// and answers a few questions about the module requirements:
//
//	go mod graph | go run github.com/hmdsefi/gograph/examples/gomodgraph cycles
//	go mod graph | go run github.com/hmdsefi/gograph/examples/gomodgraph mermaid example.com/lib
//
// Each line of "go mod graph" is "from to", meaning that from requires to.
// The main module has no version, and the other modules are written as
// path@version. Lines for the go and toolchain versions are skipped.
//
// The edges are loaded as given, from -> to, so the dependencies of a module
// are the vertices it can reach, and the modules that depend on it are the
// vertices that can reach it. This is the opposite of a build graph, where
// an edge A -> B means A has to happen before B.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/hmdsefi/gograph"
	"github.com/hmdsefi/gograph/connectivity"
	"github.com/hmdsefi/gograph/dag"
	"github.com/hmdsefi/gograph/encoding/mermaid"
	"github.com/hmdsefi/gograph/traverse"
)

const usage = `usage: gomodgraph [-f file] <command> [arguments]

Reads the output of "go mod graph" from stdin, or from file with -f.

Commands:
  cycles               print groups of module versions that require each other
  dependents <module>  print every module that requires <module>, directly or indirectly
  why <module>         print the shortest requirement path from the main module to <module>
  diff <old> <new>     compare two saved "go mod graph" outputs
  mermaid <module>     draw <module> with its direct dependencies and dependents

<module> is either a module path, which matches every version, or path@version.`

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	flags := flag.NewFlagSet("gomodgraph", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	file := flags.String("f", "", "read the graph from this file instead of stdin")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("%w\n\n%s", err, usage)
	}

	args = flags.Args()
	if len(args) == 0 {
		return fmt.Errorf("missing command\n\n%s", usage)
	}

	command, args := args[0], args[1:]
	wantArgs := map[string]int{"cycles": 0, "dependents": 1, "why": 1, "diff": 2, "mermaid": 1}
	n, ok := wantArgs[command]
	if !ok {
		return fmt.Errorf("unknown command %q\n\n%s", command, usage)
	}
	if len(args) != n {
		return fmt.Errorf("%s takes %d arguments, got %d\n\n%s", command, n, len(args), usage)
	}

	if command == "diff" {
		return diff(stdout, args[0], args[1])
	}

	in := stdin
	if *file != "" {
		f, err := os.Open(*file) //nolint:gosec // the user picks the file to read
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	g, mains, err := load(in)
	if err != nil {
		return err
	}

	switch command {
	case "cycles":
		printCycles(stdout, g)
		return nil
	case "dependents":
		return printDependents(stdout, g, args[0])
	case "mermaid":
		return printMermaid(stdout, g, args[0])
	default:
		printWhy(stdout, g, mains, args[0])
		return nil
	}
}

// load reads "go mod graph" output into a directed graph with an edge from
// each module to each module it requires. It also returns the main modules,
// which are the ones without a version, in the order they appear.
func load(r io.Reader) (gograph.Graph[string], []string, error) {
	g := gograph.New[string](gograph.Directed())
	var mains []string

	scanner := bufio.NewScanner(r)
	for line := 1; scanner.Scan(); line++ {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 2 {
			return nil, nil, fmt.Errorf("line %d: want \"from to\", got %q", line, scanner.Text())
		}

		from, to := fields[0], fields[1]
		if isGoVersion(from) || isGoVersion(to) {
			continue
		}

		if !strings.Contains(from, "@") && g.GetVertexByID(from) == nil {
			mains = append(mains, from)
		}
		_, _ = g.AddEdge(vertex(g, from), vertex(g, to))
	}

	return g, mains, scanner.Err()
}

func isGoVersion(module string) bool {
	return strings.HasPrefix(module, "go@") || strings.HasPrefix(module, "toolchain@")
}

// vertex returns the vertex with the given label, adding it if needed.
func vertex(g gograph.Graph[string], label string) *gograph.Vertex[string] {
	if v := g.GetVertexByID(label); v != nil {
		return v
	}
	return g.AddVertexByLabel(label)
}

// matches reports whether the vertex label is the module, either as a path
// that matches every version or as path@version.
func matches(label, module string) bool {
	path, _, _ := strings.Cut(label, "@")
	return label == module || path == module
}

// matching returns the labels of every version of module, in vertex order.
func matching(g gograph.Graph[string], module string) []string {
	var labels []string
	for _, v := range g.GetAllVertices() {
		if matches(v.Label(), module) {
			labels = append(labels, v.Label())
		}
	}
	return labels
}

func notInGraph(module string) error {
	return fmt.Errorf("module %s is not in the graph", module)
}

// printMermaid writes a flowchart of module, the modules it requires
// directly, the modules that require it directly, and the requirement
// edges between those vertices.
func printMermaid(w io.Writer, g gograph.Graph[string], module string) error {
	sub, err := neighborhood(g, module)
	if err != nil {
		return err
	}
	return mermaid.Write(w, sub)
}

// neighborhood is the induced subgraph on the module and the modules one
// requirement away from it: the ones it requires, and the ones that require it.
func neighborhood(g gograph.Graph[string], module string) (gograph.Graph[string], error) {
	chosen := matching(g, module)
	if len(chosen) == 0 {
		return nil, notInGraph(module)
	}

	keep := make(map[string]bool, len(chosen))
	for _, label := range chosen {
		keep[label] = true
	}
	for _, e := range g.AllEdges() {
		from, to := e.Source().Label(), e.Destination().Label()
		if matches(from, module) {
			keep[to] = true
		}
		if matches(to, module) {
			keep[from] = true
		}
	}

	sub := gograph.New[string](gograph.Directed())
	for _, v := range g.GetAllVertices() {
		if keep[v.Label()] {
			sub.AddVertexByLabel(v.Label())
		}
	}
	for _, e := range g.AllEdges() {
		from, to := e.Source().Label(), e.Destination().Label()
		if keep[from] && keep[to] {
			_, _ = sub.AddEdge(sub.GetVertexByID(from), sub.GetVertexByID(to))
		}
	}
	return sub, nil
}

// printCycles prints each strongly connected component with more than one
// module version, one per line.
func printCycles(w io.Writer, g gograph.Graph[string]) {
	var cycles []string
	for _, component := range connectivity.Tarjan(g) {
		if len(component) < 2 {
			continue
		}
		labels := make([]string, len(component))
		for i, v := range component {
			labels[i] = v.Label()
		}
		sort.Strings(labels)
		cycles = append(cycles, strings.Join(labels, " "))
	}

	sort.Strings(cycles)
	for _, c := range cycles {
		_, _ = fmt.Fprintln(w, c)
	}
}

// printDependents prints every module that requires a version of module,
// directly or indirectly. Ancestors are the modules that can reach it.
func printDependents(w io.Writer, g gograph.Graph[string], module string) error {
	targets := matching(g, module)
	if len(targets) == 0 {
		return notInGraph(module)
	}

	isTarget := make(map[string]bool, len(targets))
	for _, label := range targets {
		isTarget[label] = true
	}

	dependents := make(map[string]bool)
	for _, label := range targets {
		ancestors, err := dag.Ancestors(g, label)
		if err != nil {
			return err
		}
		for _, v := range ancestors {
			if !isTarget[v.Label()] {
				dependents[v.Label()] = true
			}
		}
	}

	labels := make([]string, 0, len(dependents))
	for label := range dependents {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	for _, label := range labels {
		_, _ = fmt.Fprintln(w, label)
	}

	return nil
}

// printWhy prints the shortest path from each main module to the first
// version of module that a breadth-first search reaches, in the format of
// "go mod why -m".
func printWhy(w io.Writer, g gograph.Graph[string], mains []string, module string) {
	_, _ = fmt.Fprintf(w, "# %s\n", module)

	found := false
	for _, mainModule := range mains {
		if path := shortestPath(g, mainModule, module); path != nil {
			_, _ = fmt.Fprintln(w, strings.Join(path, "\n"))
			found = true
		}
	}
	if !found {
		_, _ = fmt.Fprintf(w, "(main module does not need module %s)\n", module)
	}
}

// shortestPath returns the labels on a shortest path from start to a version
// of module, or nil if there is none. Vertices come out of the breadth-first
// iterator in order of distance, so the first vertex that has an edge to a
// new vertex is on a shortest path to it.
func shortestPath(g gograph.Graph[string], start, module string) []string {
	it, err := traverse.NewBreadthFirstIterator(g, start)
	if err != nil {
		return nil
	}

	parent := map[string]string{start: ""}
	for it.HasNext() {
		v := it.Next()
		if v.Label() != start && matches(v.Label(), module) {
			var path []string
			for label := v.Label(); label != ""; label = parent[label] {
				path = append([]string{label}, path...)
			}
			return path
		}

		for _, n := range v.Neighbors() {
			if _, ok := parent[n.Label()]; !ok {
				parent[n.Label()] = v.Label()
			}
		}
	}

	return nil
}

// diff prints the module versions and requirements that are only in the old
// graph with a "-", and the ones only in the new graph with a "+".
func diff(w io.Writer, oldFile, newFile string) error {
	oldGraph, err := loadFile(oldFile)
	if err != nil {
		return err
	}
	newGraph, err := loadFile(newFile)
	if err != nil {
		return err
	}

	printMissing(w, "- module", oldGraph, newGraph, modules)
	printMissing(w, "+ module", newGraph, oldGraph, modules)
	printMissing(w, "- require", oldGraph, newGraph, requirements)
	printMissing(w, "+ require", newGraph, oldGraph, requirements)

	return nil
}

func loadFile(name string) (gograph.Graph[string], error) {
	f, err := os.Open(name) //nolint:gosec // the user picks the file to read
	if err != nil {
		return nil, err
	}
	defer f.Close()

	g, _, err := load(f)
	return g, err
}

// modules returns each module version in a, and whether b has it too.
func modules(a, b gograph.Graph[string]) map[string]bool {
	out := make(map[string]bool)
	for _, v := range a.GetAllVertices() {
		out[v.Label()] = b.ContainsVertex(v)
	}
	return out
}

// requirements returns each "from to" requirement in a, and whether b has
// it too.
func requirements(a, b gograph.Graph[string]) map[string]bool {
	out := make(map[string]bool)
	for _, e := range a.AllEdges() {
		key := e.Source().Label() + " " + e.Destination().Label()
		out[key] = b.ContainsEdge(e.Source(), e.Destination())
	}
	return out
}

// printMissing prints, in sorted order, the items of a that b doesn't have.
func printMissing(
	w io.Writer,
	prefix string,
	a, b gograph.Graph[string],
	items func(a, b gograph.Graph[string]) map[string]bool,
) {
	var missing []string
	for item, inB := range items(a, b) {
		if !inB {
			missing = append(missing, item)
		}
	}

	sort.Strings(missing)
	for _, item := range missing {
		_, _ = fmt.Fprintf(w, "%s %s\n", prefix, item)
	}
}
