# gomodgraph

A short example that loads the output of `go mod graph` into gograph and answers a few
questions about a module's requirements. It only uses gograph and the standard library.

```shell
go mod graph | go run github.com/hmdsefi/gograph/examples/gomodgraph@latest cycles
```

Each line of `go mod graph` is `from to`, meaning `from` requires `to`. The example adds an
edge `from -> to` for each line, so the dependencies of a module are the vertices it can
reach, and the modules that depend on it are the vertices that can reach it. That's the
opposite of a build graph, where an edge `A -> B` means `A` has to happen before `B`.
Lines for the `go` and `toolchain` versions are skipped.

## Commands

A `<module>` is either a module path, which matches every version, or `path@version`. The
outputs below come from [`testdata/modgraph.txt`](testdata/modgraph.txt).

`cycles` prints the module versions that require each other, one group per line. It uses
`connectivity.Tarjan`:

```text
$ gomodgraph -f testdata/modgraph.txt cycles
example.com/codec@v0.5.0 example.com/util@v0.3.1
```

`dependents <module>` prints every module that requires a version of `<module>`, directly
or indirectly. It walks a reversed copy of the graph with the breadth-first iterator:

```text
$ gomodgraph -f testdata/modgraph.txt dependents example.com/util@v0.2.0
example.com/app
example.com/log@v1.1.0
```

`why <module>` prints the shortest requirement path from the main module, in the format of
`go mod why -m`:

```text
$ gomodgraph -f testdata/modgraph.txt why example.com/codec
# example.com/codec
example.com/app
example.com/lib@v1.4.0
example.com/util@v0.3.1
example.com/codec@v0.5.0
```

`diff <old> <new>` compares two saved outputs of `go mod graph` and prints the module
versions and requirements that were removed (`-`) or added (`+`):

```text
$ gomodgraph diff testdata/modgraph.txt testdata/modgraph_new.txt
- module example.com/lib@v1.4.0
- module example.com/log@v1.1.0
- module example.com/util@v0.2.0
+ module example.com/lib@v1.5.0
+ module example.com/metrics@v0.1.0
- require example.com/app example.com/lib@v1.4.0
...
```

## Comment on pull requests that change go.mod

This workflow runs `go mod graph` on the base and head commits of a pull request and posts
the `diff` output as a comment. Pull requests from forks get a read-only token, so the
comment step only works for branches in the same repository.

```yaml
name: Module graph

on:
  pull_request:
    paths: [go.mod, go.sum]

permissions:
  contents: read
  pull-requests: write

jobs:
  diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - name: Compare module graphs
        env:
          BASE: ${{ github.event.pull_request.base.sha }}
          HEAD: ${{ github.event.pull_request.head.sha }}
        run: |
          git checkout -q "$BASE" && go mod graph > "$RUNNER_TEMP/old.txt"
          git checkout -q "$HEAD" && go mod graph > "$RUNNER_TEMP/new.txt"
          {
            echo '```diff'
            go run github.com/hmdsefi/gograph/examples/gomodgraph@latest \
              diff "$RUNNER_TEMP/old.txt" "$RUNNER_TEMP/new.txt"
            echo '```'
          } > "$RUNNER_TEMP/comment.md"
      - name: Comment
        env:
          GH_TOKEN: ${{ github.token }}
          PR: ${{ github.event.pull_request.number }}
        run: gh pr comment "$PR" --body-file "$RUNNER_TEMP/comment.md"
```
