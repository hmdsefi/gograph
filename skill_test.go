package gograph_test

import (
	"os"
	"strings"
	"testing"
)

func TestSkillRows(t *testing.T) {
	data, err := os.ReadFile(".cursor/skills/gograph/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)

	for _, word := range []string{"—", "–"} {
		if strings.Contains(text, word) {
			t.Fatalf("SKILL.md contains %q", word)
		}
	}

	rows := skillRows(t, text)
	got := map[string]int{}
	for _, row := range rows {
		if row.constraint == "" {
			t.Fatalf("row %q has no constraint", row.call)
		}
		if row.example != "" && !strings.HasPrefix(row.example, "https://gograph.dev/algorithms/") {
			t.Fatalf("row %q example %q is not a gograph.dev algorithm page", row.call, row.example)
		}
		got[row.call]++
	}

	for call, n := range got {
		if n != 1 {
			t.Fatalf("call %q appears %d times", call, n)
		}
	}

	for _, call := range skillCalls {
		if got[call] != 1 {
			t.Fatalf("missing call %q", call)
		}
	}
	if len(got) != len(skillCalls) {
		t.Fatalf("got %d calls, want %d", len(got), len(skillCalls))
	}
}

type skillRow struct {
	call       string
	constraint string
	example    string
}

func skillRows(t *testing.T, text string) []skillRow {
	t.Helper()
	var rows []skillRow
	for _, line := range strings.Split(text, "\n") {
		if !strings.HasPrefix(line, "|") || strings.Contains(line, "Question") || strings.HasPrefix(line, "|---") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) != 6 {
			t.Fatalf("row %q has %d cells", line, len(parts))
		}
		rows = append(rows, skillRow{
			call:       strings.Trim(parts[2], " `"),
			constraint: strings.TrimSpace(parts[3]),
			example:    strings.TrimSpace(parts[4]),
		})
	}
	if len(rows) == 0 {
		t.Fatal("no rows")
	}
	return rows
}

var skillCalls = []string{
	"TopologySort",
	"StableTopologySort",
	"traverse.NewTopologicalIterator",
	"InducedSubgraph",
	"path.Dijkstra",
	"path.DijkstraSimple",
	"path.BellmanFord",
	"path.FloydWarshall",
	"path.DijkstraMultiSource",
	"path.KCenter",
	"path.TransitiveReduction",
	"dag.Descendants",
	"dag.Ancestors",
	"dag.Affected",
	"dag.Levels",
	"dag.NewTracker",
	"dag.CriticalPath",
	"dag.CriticalPathFunc",
	"connectivity.Tarjan",
	"connectivity.Kosaraju",
	"connectivity.Gabow",
	"connectivity.Condense",
	"partition.MaximalCliques",
	"partition.GirvanNewman",
	"partition.RandomizedKCut",
	"traverse.NewBreadthFirstIterator",
	"traverse.NewDepthFirstIterator",
	"traverse.NewClosestFirstIterator",
	"traverse.NewRandomWalkIterator",
	"encoding/dot.Marshal",
	"encoding/mermaid.Marshal",
}
