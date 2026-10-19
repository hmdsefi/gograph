package main

import (
	"bytes"
	"os"
	"path/filepath"
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
			name: "diff of the selected versions",
			args: []string{"diff", "-selected", graphFile, newGraphFile},
			want: "- example.com/log v1.1.0\n" +
				"+ example.com/metrics v0.1.0\n" +
				"~ example.com/lib v1.4.0 -> v1.5.0\n",
		},
		{
			name: "diff of the selected versions of the same graph",
			args: []string{"diff", "-selected", graphFile, graphFile},
			want: "",
		},
		{
			name: "versions",
			args: []string{"versions"},
			want: "example.com/util\n" +
				"  v0.2.0  required by example.com/log@v1.1.0\n" +
				"* v0.3.1  required by example.com/codec@v0.5.0, example.com/lib@v1.4.0\n",
		},
		{
			name: "versions of a module path",
			args: []string{"versions", "example.com/util"},
			want: "example.com/util\n" +
				"  v0.2.0  required by example.com/log@v1.1.0\n" +
				"* v0.3.1  required by example.com/codec@v0.5.0, example.com/lib@v1.4.0\n",
		},
		{
			name: "versions of one version shows the whole path",
			args: []string{"versions", "example.com/util@v0.2.0"},
			want: "example.com/util\n" +
				"  v0.2.0  required by example.com/log@v1.1.0\n" +
				"* v0.3.1  required by example.com/codec@v0.5.0, example.com/lib@v1.4.0\n",
		},
		{
			name: "versions of a module at one version",
			args: []string{"versions", "example.com/lib"},
			want: "example.com/lib\n" +
				"* v1.4.0  required by example.com/app\n",
		},
		{
			name: "versions of the main module",
			args: []string{"versions", "example.com/app"},
			want: "example.com/app (main module)\n",
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

// mainRequiredAtOldVersion is a graph where a dependency requires an older
// version of the main module, and several versions of example.com/x, which
// sort differently as strings and as versions.
const mainRequiredAtOldVersion = `example.com/app example.com/plugin@v1.0.0
example.com/app example.com/x@v1.10.0
example.com/plugin@v1.0.0 example.com/app@v0.8.2
example.com/plugin@v1.0.0 example.com/x@v1.9.0
example.com/plugin@v1.0.0 example.com/tool@v0.3.0
example.com/tool@v0.3.0 example.com/x@v1.10.0-rc.1
example.com/tool@v0.3.0 example.com/x@v0.0.0-20260101000000-abcdefabcdef
`

func TestVersionsOrderAndMainModule(t *testing.T) {
	got, err := runCommand(t, mainRequiredAtOldVersion, "versions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "example.com/app (main module)\n" +
		"  v0.8.2  required by example.com/plugin@v1.0.0\n" +
		"example.com/x\n" +
		"  v0.0.0-20260101000000-abcdefabcdef  required by example.com/tool@v0.3.0\n" +
		"  v1.9.0                              required by example.com/plugin@v1.0.0\n" +
		"  v1.10.0-rc.1                        required by example.com/tool@v0.3.0\n" +
		"* v1.10.0                             required by example.com/app\n"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestVersionsWithoutRequirers(t *testing.T) {
	got, err := runCommand(t, "example.com/a@v1.0.0 example.com/b@v1.0.0\n", "versions", "example.com/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "example.com/a\n* v1.0.0\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDiffSelectedSkipsMainModule(t *testing.T) {
	dir := t.TempDir()
	oldFile := filepath.Join(dir, "old.txt")
	newFile := filepath.Join(dir, "new.txt")
	newGraph := strings.ReplaceAll(mainRequiredAtOldVersion, "plugin@v1.0.0", "plugin@v1.1.0")
	newGraph = strings.ReplaceAll(newGraph, "app@v0.8.2", "app@v0.9.0")
	for name, content := range map[string]string{oldFile: mainRequiredAtOldVersion, newFile: newGraph} {
		if err := os.WriteFile(name, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got, err := runCommand(t, "", "diff", "-selected", oldFile, newFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "~ example.com/plugin v1.0.0 -> v1.1.0\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCompareVersions(t *testing.T) {
	ordered := []string{
		"v0.0.0-20260101000000-abcdefabcdef",
		"v0.0.0-20260102000000-abcdefabcdef",
		"v0.1.0",
		"v1.0.0-99999999999999999999",
		"v1.0.0-100000000000000000000",
		"v1.0.0-alpha",
		"v1.0.0-alpha.1",
		"v1.0.0-alpha.beta",
		"v1.0.0-beta",
		"v1.0.0-beta.2",
		"v1.0.0-beta.11",
		"v1.0.0-rc.1",
		"v1.0.0",
		"v1.2.0",
		"v1.10.0",
		"v2.0.0+incompatible",
		"v10.0.0",
	}
	for i, a := range ordered {
		for j, b := range ordered {
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}
			if got := compareVersions(a, b); got != want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", a, b, got, want)
			}
		}
	}

	equal := [][2]string{
		{"v1", "v1.0.0"},
		{"v1.2", "v1.2.0"},
		{"v1.0.0+a", "v1.0.0+b"},
		{"v2.0.0+incompatible", "v2.0.0"},
	}
	for _, pair := range equal {
		if got := compareVersions(pair[0], pair[1]); got != 0 {
			t.Errorf("compareVersions(%q, %q) = %d, want 0", pair[0], pair[1], got)
		}
	}

	invalid := []string{
		"", "1.0.0", "v", "v01.0.0", "v1.00.0", "v1.0.0-", "v1.0.0-01", "v1.0.0+",
		"v1-pre", "v1.2+build", "v1.2.3.4", "v1.0.0-a..b", "v1.0.0-é", "v1.0.0+b_c", "vx.0.0",
	}
	for _, v := range invalid {
		if got := compareVersions(v, "v0.0.0"); got != -1 {
			t.Errorf("compareVersions(%q, v0.0.0) = %d, want -1", v, got)
		}
		if got := compareVersions("v0.0.0", v); got != 1 {
			t.Errorf("compareVersions(v0.0.0, %q) = %d, want 1", v, got)
		}
		if got := compareVersions(v, "not a version"); got != 0 {
			t.Errorf("compareVersions(%q, \"not a version\") = %d, want 0", v, got)
		}
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
		"versions extra":    {args: []string{"versions", "a", "b"}, contains: "takes 0 or 1 arguments"},
		"versions missing":  {stdin: "a b@v1\n", args: []string{"versions", "b@v2"}, contains: "not in the graph"},
		"diff unknown flag": {args: []string{"diff", "-x", graphFile, graphFile}},
		"selected one file": {args: []string{"diff", "-selected", graphFile}, contains: "takes 2 arguments"},
		"selected old file": {args: []string{"diff", "-selected", "testdata/missing.txt", graphFile}},
		"selected new file": {args: []string{"diff", "-selected", graphFile, "testdata/missing.txt"}},
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
