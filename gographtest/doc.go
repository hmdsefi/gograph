// Package gographtest helps test code that uses gograph, in the style of
// net/http/httptest.
//
// The assertions check results that have more than one right answer:
// IsTopologicalOrder accepts any valid topological order, SameGroups
// compares groups of vertices, such as components, cliques or communities,
// in any order, and Equal compares two graphs whatever order they were
// built in, through gograph.Diff. Each one calls t.Helper, so a failure
// points at the line of the test, and reports one failure through t.Errorf
// that lists up to 10 problems and the number of the rest.
//
// The generators build common graphs with the labels 0 to n-1: Path, Cycle,
// Complete, Grid, Random and RandomDAG. The options go to gograph.New, so
// each generator builds directed or undirected, weighted or unweighted
// graphs. Random and RandomDAG take a caller-owned *rand.Rand from
// math/rand/v2, so a seeded test builds the same graph on every run.
package gographtest
