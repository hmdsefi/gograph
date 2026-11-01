package gographtest

import (
	"encoding/binary"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/hmdsefi/gograph"
)

// maxProblems is the most problems one failure lists.
const maxProblems = 10

// IsTopologicalOrder fails the test unless order holds every vertex of g
// exactly once, and every edge u -> v has u before v. It compares vertices
// by label, so order can hold copies of the graph's vertices.
//
// A graph with a self-loop has no topological order, so it always fails.
// An undirected graph fails too, the same way gograph.TopologySort returns
// gograph.ErrNotDirected for it. The failure lists the vertices that are
// missing, repeated, nil or not in g, and the edges that point backward, up
// to 10 problems.
//
// It takes O(V+E) time and O(V) space.
func IsTopologicalOrder[T comparable](t testing.TB, g gograph.Graph[T], order []*gograph.Vertex[T]) {
	t.Helper()
	if !g.IsDirected() {
		t.Errorf("gographtest: IsTopologicalOrder needs a directed graph")
		return
	}

	var problems []string
	position := make(map[T]int, len(order))
	for i, v := range order {
		switch {
		case v == nil:
			problems = append(problems, fmt.Sprintf("position %d holds nil", i))
		case g.GetVertexByID(v.Label()) == nil:
			problems = append(problems, fmt.Sprintf("vertex %v isn't in the graph", v.Label()))
		default:
			if _, ok := position[v.Label()]; ok {
				problems = append(problems, fmt.Sprintf("vertex %v comes more than once", v.Label()))
				continue
			}
			position[v.Label()] = i
		}
	}

	for v := range gograph.Vertices(g) {
		if _, ok := position[v.Label()]; !ok {
			problems = append(problems, fmt.Sprintf("vertex %v is missing", v.Label()))
		}
	}

	for e := range gograph.Edges(g) {
		from, okFrom := position[e.Source().Label()]
		to, okTo := position[e.Destination().Label()]
		if okFrom && okTo && from >= to {
			problems = append(problems, fmt.Sprintf("edge %v -> %v goes from position %d to %d",
				e.Source().Label(), e.Destination().Label(), from, to))
		}
	}

	fail(t, "not a topological order of the graph", problems)
}

// SameGroups fails the test unless got has the same groups of labels as
// want, ignoring the order of the groups and of the labels inside each
// group. It compares labels, so the vertices can be copies. It suits
// connected components, strongly connected components, cliques and
// communities.
//
// Groups can share labels, as cliques do, and each group counts as many
// times as it appears: a label twice in one group, or one group twice,
// only matches the same repeat on the other side. The failure lists the
// groups of want missing from got and the groups of got that aren't in want,
// each with its labels in the order given, and any nil vertex in got, up to
// 10 problems.
//
// T is only comparable, so it gives each label a number and sorts the
// numbers of each group. For L labels in all, in groups of at most G labels,
// it takes O(L log G) time and O(L) space.
func SameGroups[T comparable](t testing.TB, want [][]T, got [][]*gograph.Vertex[T]) {
	t.Helper()

	ids := make(map[T]int)
	var nums []int
	key := func(labels []T) string {
		nums = nums[:0]
		for _, label := range labels {
			id, ok := ids[label]
			if !ok {
				id = len(ids)
				ids[label] = id
			}
			nums = append(nums, id)
		}
		slices.Sort(nums)

		// varints mark their own ends, so the key of a group is unique
		buf := make([]byte, 0, len(nums))
		for _, n := range nums {
			buf = binary.AppendVarint(buf, int64(n))
		}
		return string(buf)
	}

	// wanted holds the indexes of the groups of want with each key, and
	// loses one each time a group of got matches
	wanted := make(map[string][]int, len(want))
	for i, group := range want {
		k := key(group)
		wanted[k] = append(wanted[k], i)
	}

	var problems, unexpected []string
	var labels []T
	for i, group := range got {
		labels = labels[:0]
		for _, v := range group {
			if v == nil {
				break
			}
			labels = append(labels, v.Label())
		}
		if len(labels) < len(group) {
			problems = append(problems, fmt.Sprintf("group %d of got holds nil", i))
			continue
		}

		k := key(labels)
		if matches := wanted[k]; len(matches) > 0 {
			wanted[k] = matches[1:]
			continue
		}
		unexpected = append(unexpected, fmt.Sprintf("unexpected %v", labels))
	}

	var missing []int
	for _, matches := range wanted {
		missing = append(missing, matches...)
	}
	slices.Sort(missing)
	for _, i := range missing {
		problems = append(problems, fmt.Sprintf("missing %v", want[i]))
	}
	problems = append(problems, unexpected...)

	fail(t, "groups differ", problems)
}

// fail reports the problems as one failure, if there are any.
func fail(t testing.TB, title string, problems []string) {
	t.Helper()
	if len(problems) == 0 {
		return
	}

	var b strings.Builder
	b.WriteString("gographtest: ")
	b.WriteString(title)
	b.WriteString(":")
	for _, p := range problems[:min(len(problems), maxProblems)] {
		b.WriteString("\n\t")
		b.WriteString(p)
	}
	if len(problems) > maxProblems {
		fmt.Fprintf(&b, "\n\tand %d more", len(problems)-maxProblems)
	}
	t.Errorf("%s", b.String())
}
