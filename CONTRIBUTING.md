# Contributing to gograph

Thanks for your interest in gograph. Bug reports, documentation fixes, tests and new
features are all welcome.

## Finding something to work on

- The [open milestones](https://github.com/hmdsefi/gograph/milestones?state=open&sort=due_date&direction=asc)
  group the issues by release, the next release first. Please start with an issue from
  the earliest one, since that's the work needed soonest.
- The [roadmap](https://github.com/hmdsefi/gograph/issues/136) lists planned work in the order it's likely to land.
- Issues labeled [good first issue](https://github.com/hmdsefi/gograph/labels/good%20first%20issue)
  are small and well described.
- Issues labeled [help wanted](https://github.com/hmdsefi/gograph/labels/help%20wanted)
  are larger, and each one describes the proposed API, the expected behavior and the tests.

Before starting on an issue, leave a comment so others know it's taken. For a new
feature that doesn't have an issue yet, please open one first so we can agree on the
API before you write the code.

## Development setup

You need Go 1.24.2 or newer (see `go.mod`) and, for linting,
[golangci-lint](https://golangci-lint.run/) v2.

```shell
git clone https://github.com/<your-username>/gograph.git
cd gograph

go test -race ./...     # run all tests, same as make test
golangci-lint run ./... # run the linters, same as make lint
make check-coverage     # run the tests and check test coverage
```

CI runs the linters and the tests on every pull request, and fails if total test
coverage drops below 95% (see `.testcoverage.yml`).

## Backward compatibility

gograph is used by other projects, so changes must not break existing code. In
practice:

- **Don't add methods to exported interfaces** such as `Graph`, `GraphType` and
  `traverse.Iterator`. Code outside this repository implements and mocks them, and a new
  method breaks all of it. Add a package-level function instead. Adding methods to the
  `Vertex` and `Edge` structs is fine.
- **Don't change the signature of an exported function or method**, even to add a
  variadic parameter. That changes the function's type, which breaks code that stores
  it in a variable.
- **Keep sentinel errors as they are.** Callers compare errors such as `ErrDAGCycle`
  with `==`, so return them unwrapped. A new error can be an alias of an existing one,
  the way `path.ErrNotDAG` is `gograph.ErrDAGHasCycle`.
- **Discuss behavior changes first.** If a fix changes what an existing function
  returns, open an issue before sending the pull request.

## Code guidelines

- Format with `gofmt` and `goimports`. The linters check both.
- Every exported identifier needs a doc comment. Say what it does, what it returns for
  edge cases (missing vertices, empty graphs, undirected graphs) and its time complexity
  if it's an algorithm.
- Every change needs tests. Bug fixes need a test that fails without the fix.
  Table-driven tests fit most of the existing code.
- New public API should come with an `Example` function, so it shows up on
  pkg.go.dev and `go test` keeps it working.
- No new dependencies. The library, including its tests, uses only the standard library.

## Pull requests

- Keep each pull request to one change. Several small pull requests are easier to
  review than one large one.
- Reference the issue in the description, for example `Fixes #123`.
- Describe how you tested the change.
- Make sure `go test -race ./...` and `golangci-lint run ./...` pass locally.
- Be ready to explain every line of your change during review.

Pull requests get a first review within about a week. If yours hasn't, feel free to
leave a comment on it.

## Reporting bugs

Please use the bug report template and include a minimal program that reproduces the
problem. The smaller the reproduction, the faster the fix.

For security issues, don't open a public issue. See [SECURITY.md](SECURITY.md).

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). By taking part,
you agree to follow it.

## License

By contributing, you agree that your contributions are licensed under the
[Apache License 2.0](LICENSE), the same license as the project.
