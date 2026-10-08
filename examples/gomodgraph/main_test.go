package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

const (
	graphFile    = "testdata/modgraph.txt"
	newGraphFile = "testdata/modgraph_new.txt"
)

func runCommand(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	err := run(args, strings.NewReader(stdin), &out)
	return out.String(), err
}

func readFile(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(name) //nolint:gosec // test data path
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestCommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "cycles",
			args: []string{"cycles"},
			want: "example.com/codec@v0.5.0 example.com/util@v0.3.1\n",
		},
		{
			name: "dependents by path",
			args: []string{"dependents", "example.com/util"},
			want: "example.com/app\n" +
				"example.com/codec@v0.5.0\n" +
				"example.com/lib@v1.4.0\n" +
				"example.com/log@v1.1.0\n",
		},
		{
			name: "dependents by version",
			args: []string{"dependents", "example.com/util@v0.2.0"},
			want: "example.com/app\n" +
				"example.com/log@v1.1.0\n",
		},
		{
			name: "why",
			args: []string{"why", "example.com/codec"},
			want: "# example.com/codec\n" +
				"example.com/app\n" +
				"example.com/lib@v1.4.0\n" +
				"example.com/util@v0.3.1\n" +
				"example.com/codec@v0.5.0\n",
		},
		{
			name: "why stops at the first version it reaches",
			args: []string{"why", "example.com/util"},
			want: "# example.com/util\n" +
				"example.com/app\n" +
				"example.com/lib@v1.4.0\n" +
				"example.com/util@v0.3.1\n",
		},
		{
			name: "why not needed",
			args: []string{"why", "example.com/other"},
			want: "# example.com/other\n" +
				"(main module does not need module example.com/other)\n",
		},
		{
			name: "diff",
			args: []string{"diff", graphFile, newGraphFile},
			want: "- module example.com/lib@v1.4.0\n" +
				"- module example.com/log@v1.1.0\n" +
				"- module example.com/util@v0.2.0\n" +
				"+ module example.com/lib@v1.5.0\n" +
				"+ module example.com/metrics@v0.1.0\n" +
				"- require example.com/app example.com/lib@v1.4.0\n" +
				"- require example.com/app example.com/log@v1.1.0\n" +
				"- require example.com/lib@v1.4.0 example.com/util@v0.3.1\n" +
				"- require example.com/log@v1.1.0 example.com/util@v0.2.0\n" +
				"+ require example.com/app example.com/lib@v1.5.0\n" +
				"+ require example.com/app example.com/metrics@v0.1.0\n" +
				"+ require example.com/lib@v1.5.0 example.com/util@v0.3.1\n" +
				"+ require example.com/metrics@v0.1.0 example.com/util@v0.3.1\n",
		},
		{
			name: "diff of the same graph",
			args: []string{"diff", graphFile, graphFile},
			want: "",
		},
		{
			name: "mermaid keeps direct dependencies and dependents",
			args: []string{"mermaid", "example.com/util"},
			want: "flowchart TD\n" +
				"    n0[\"example.com/codec@v0.5.0\"]\n" +
				"    n1[\"example.com/lib@v1.4.0\"]\n" +
				"    n2[\"example.com/log@v1.1.0\"]\n" +
				"    n3[\"example.com/util@v0.2.0\"]\n" +
				"    n4[\"example.com/util@v0.3.1\"]\n" +
				"    n0 --> n4\n" +
				"    n1 --> n4\n" +
				"    n2 --> n3\n" +
				"    n4 --> n0\n",
		},
		{
			name: "mermaid of one version stays on its direct edges",
			args: []string{"mermaid", "example.com/util@v0.2.0"},
			want: "flowchart TD\n" +
				"    n0[\"example.com/log@v1.1.0\"]\n" +
				"    n1[\"example.com/util@v0.2.0\"]\n" +
				"    n0 --> n1\n",
		},
		{
			name: "mermaid of the main module",
			args: []string{"mermaid", "example.com/app"},
			want: "flowchart TD\n" +
				"    n0[\"example.com/app\"]\n" +
				"    n1[\"example.com/lib@v1.4.0\"]\n" +
				"    n2[\"example.com/log@v1.1.0\"]\n" +
				"    n0 --> n1\n" +
				"    n0 --> n2\n",
		},
	}

	stdin := readFile(t, graphFile)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runCommand(t, stdin, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestReadFromFile(t *testing.T) {
	fromStdin, err := runCommand(t, readFile(t, graphFile), "cycles")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fromFile, err := runCommand(t, "", "-f", graphFile, "cycles")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fromStdin != fromFile {
		t.Fatalf("stdin and -f gave different output:\n%s\n%s", fromStdin, fromFile)
	}
}

func TestDependentsOfOneVersionIncludesAnother(t *testing.T) {
	got, err := runCommand(t, "example.com/lib@v1.5.0 example.com/lib@v1.4.0\n", "dependents", "example.com/lib@v1.4.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "example.com/lib@v1.5.0\n" {
		t.Fatalf("got %q", got)
	}
}

func TestMermaidKeepsEdgesBetweenNeighbors(t *testing.T) {
	got, err := runCommand(t, "mid target\ntarget leaf\nmid leaf\n", "mermaid", "target")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "flowchart TD\n" +
		"    n0[\"leaf\"]\n" +
		"    n1[\"mid\"]\n" +
		"    n2[\"target\"]\n" +
		"    n1 --> n2\n" +
		"    n1 --> n0\n" +
		"    n2 --> n0\n"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestNoCycles(t *testing.T) {
	got, err := runCommand(t, "a b@v1\nb@v1 c@v1\n", "cycles")
	if err != nil || got != "" {
		t.Fatalf("got %q, %v, want no output", got, err)
	}
}

func TestErrors(t *testing.T) {
	tests := map[string]struct {
		stdin    string
		args     []string
		contains string
	}{
		"no command":        {args: nil},
		"unknown command":   {args: []string{"graph"}},
		"unknown flag":      {args: []string{"-x", "cycles"}},
		"missing module":    {args: []string{"why"}},
		"extra argument":    {args: []string{"dependents", "a", "b"}},
		"missing diff file": {args: []string{"diff", graphFile}},
		"no such file":      {args: []string{"-f", "testdata/missing.txt", "cycles"}},
		"no such old file":  {args: []string{"diff", "testdata/missing.txt", graphFile}},
		"no such new file":  {args: []string{"diff", graphFile, "testdata/missing.txt"}},
		"malformed line":    {stdin: "a b\nonly-one-field\n", args: []string{"cycles"}},
		"unknown module":    {stdin: "a b@v1\n", args: []string{"dependents", "c"}},
		"mermaid no module": {args: []string{"mermaid"}, contains: "takes 1 arguments"},
		"mermaid missing":   {stdin: "a b@v1\n", args: []string{"mermaid", "c"}, contains: "not in the graph"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := runCommand(t, tt.stdin, tt.args...)
			if err == nil {
				t.Fatal("expected an error")
			}
			if tt.contains != "" && !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("error %q, want it to contain %q", err, tt.contains)
			}
		})
	}
}
