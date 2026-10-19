// Command gomodgraph loads the output of "go mod graph" into a gograph graph
// and answers a few questions about the module requirements:
//
//	go mod graph | go run github.com/hmdsefi/gograph/examples/gomodgraph cycles
//	go mod graph | go run github.com/hmdsefi/gograph/examples/gomodgraph versions
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
	"strconv"
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
  versions [module]    print the versions of each module the graph has more than one of,
                       or of <module>, and mark the version that builds
  diff [-selected] <old> <new>
                       compare two saved "go mod graph" outputs, or with -selected,
                       the version of each module that builds
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
	wantArgs := map[string][2]int{
		"cycles":     {0, 0},
		"dependents": {1, 1},
		"why":        {1, 1},
		"versions":   {0, 1},
		"diff":       {2, 2},
		"mermaid":    {1, 1},
	}
	n, ok := wantArgs[command]
	if !ok {
		return fmt.Errorf("unknown command %q\n\n%s", command, usage)
	}

	selected := false
	if command == "diff" {
		diffFlags := flag.NewFlagSet("diff", flag.ContinueOnError)
		diffFlags.SetOutput(io.Discard)
		diffFlags.BoolVar(&selected, "selected", false, "compare the version of each module that builds")
		if err := diffFlags.Parse(args); err != nil {
			return fmt.Errorf("%w\n\n%s", err, usage)
		}
		args = diffFlags.Args()
	}

	if len(args) < n[0] || len(args) > n[1] {
		want := strconv.Itoa(n[0])
		if n[1] != n[0] {
			want += " or " + strconv.Itoa(n[1])
		}
		return fmt.Errorf("%s takes %s arguments, got %d\n\n%s", command, want, len(args), usage)
	}

	if command == "diff" {
		if selected {
			return diffSelected(stdout, args[0], args[1])
		}
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
	case "versions":
		return printVersions(stdout, g, mains, args)
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

// printVersions prints each module path that the graph has at more than one
// version, or only the path of the module in args, followed by its versions
// in order. The version that builds is marked with "*", and each version
// lists the modules that require it directly. A main module builds from its
// own code, so its path is printed with every version it's required at, and
// none of them is marked.
func printVersions(w io.Writer, g gograph.Graph[string], mains, args []string) error {
	versions := moduleVersions(g)
	selected := selectedVersions(g, mains)
	isMain := make(map[string]bool, len(mains))
	for _, m := range mains {
		isMain[m] = true
	}

	var paths []string
	if len(args) == 1 {
		if len(matching(g, args[0])) == 0 {
			return notInGraph(args[0])
		}
		path, _, _ := strings.Cut(args[0], "@")
		paths = []string{path}
	} else {
		for path, list := range versions {
			if len(list) > 1 || isMain[path] {
				paths = append(paths, path)
			}
		}
		sort.Strings(paths)
	}

	requiredBy := make(map[string][]string)
	for _, e := range g.AllEdges() {
		to := e.Destination().Label()
		requiredBy[to] = append(requiredBy[to], e.Source().Label())
	}

	for _, path := range paths {
		if isMain[path] {
			_, _ = fmt.Fprintf(w, "%s (main module)\n", path)
		} else {
			_, _ = fmt.Fprintln(w, path)
		}

		width := 0
		for _, version := range versions[path] {
			width = max(width, len(version))
		}
		for _, version := range versions[path] {
			mark := " "
			if !isMain[path] && version == selected[path] {
				mark = "*"
			}
			line := fmt.Sprintf("%s %-*s", mark, width, version)
			if by := requiredBy[path+"@"+version]; len(by) > 0 {
				sort.Strings(by)
				line += "  required by " + strings.Join(by, ", ")
			}
			_, _ = fmt.Fprintln(w, strings.TrimRight(line, " "))
		}
	}

	return nil
}

// moduleVersions returns the versions of each module path in the graph, from
// the lowest to the highest. Main modules have no version, so their own
// vertices aren't included.
func moduleVersions(g gograph.Graph[string]) map[string][]string {
	versions := make(map[string][]string)
	for _, v := range g.GetAllVertices() {
		if path, version, ok := strings.Cut(v.Label(), "@"); ok {
			versions[path] = append(versions[path], version)
		}
	}
	for _, list := range versions {
		sort.Slice(list, func(i, j int) bool {
			if c := compareVersions(list[i], list[j]); c != 0 {
				return c < 0
			}
			return list[i] < list[j]
		})
	}
	return versions
}

// selectedVersions returns the version of each module path that builds,
// which is the highest version of it in the graph: minimal version selection
// picks the highest version that any module requires. A main module builds
// from its own code, so its path is left out even when other modules
// require a version of it.
func selectedVersions(g gograph.Graph[string], mains []string) map[string]string {
	isMain := make(map[string]bool, len(mains))
	for _, m := range mains {
		isMain[m] = true
	}

	selected := make(map[string]string)
	for path, list := range moduleVersions(g) {
		if !isMain[path] {
			selected[path] = list[len(list)-1]
		}
	}
	return selected
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
	oldGraph, _, err := loadFile(oldFile)
	if err != nil {
		return err
	}
	newGraph, _, err := loadFile(newFile)
	if err != nil {
		return err
	}

	printMissing(w, "- module", oldGraph, newGraph, modules)
	printMissing(w, "+ module", newGraph, oldGraph, modules)
	printMissing(w, "- require", oldGraph, newGraph, requirements)
	printMissing(w, "+ require", newGraph, oldGraph, requirements)

	return nil
}

// diffSelected compares the version of each module path that builds in the
// old graph and in the new one. It prints "-" with the old version for a
// path only the old graph has, "+" with the new version for a path only the
// new graph has, and "~" with both versions for a path whose version changed.
func diffSelected(w io.Writer, oldFile, newFile string) error {
	oldGraph, oldMains, err := loadFile(oldFile)
	if err != nil {
		return err
	}
	newGraph, newMains, err := loadFile(newFile)
	if err != nil {
		return err
	}
	oldVersions := selectedVersions(oldGraph, oldMains)
	newVersions := selectedVersions(newGraph, newMains)

	var removed, added, changed []string
	for path, oldVersion := range oldVersions {
		newVersion, ok := newVersions[path]
		switch {
		case !ok:
			removed = append(removed, fmt.Sprintf("- %s %s", path, oldVersion))
		case newVersion != oldVersion:
			changed = append(changed, fmt.Sprintf("~ %s %s -> %s", path, oldVersion, newVersion))
		}
	}
	for path, newVersion := range newVersions {
		if _, ok := oldVersions[path]; !ok {
			added = append(added, fmt.Sprintf("+ %s %s", path, newVersion))
		}
	}

	for _, lines := range [][]string{removed, added, changed} {
		sort.Strings(lines)
		for _, line := range lines {
			_, _ = fmt.Fprintln(w, line)
		}
	}

	return nil
}

// loadFile loads a saved "go mod graph" output, see load.
func loadFile(name string) (gograph.Graph[string], []string, error) {
	f, err := os.Open(name) //nolint:gosec // the user picks the file to read
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	return load(f)
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
