// Package dot writes graphs in the DOT language of Graphviz, which dot, the
// other Graphviz layout tools and many viewers and editor plugins read.
//
// Every vertex gets a generated node ID (n0, n1, ...) and its label goes in
// the label attribute, so two labels that print the same don't collide.
// Attribute values are written in quotes, with double quotes and
// backslashes escaped and newlines written as \n. A value set by an option
// that starts with < and ends with > is written as is instead, so Graphviz
// reads it as an HTML-like label.
//
// Directed graphs are written as a digraph with -> for edges, and
// undirected graphs as a graph with --. Each undirected edge is written
// once, even though the graph stores it in both directions. Vertices
// without edges are still declared, so they appear in the drawing. Edges
// show their label in the label attribute when one is set, and otherwise
// their weight on weighted graphs. The weight attribute of Graphviz isn't
// set, since it changes the layout and dot only accepts non-negative
// integers for it.
//
// The output is stable: a graph built the same way gives the same text on
// every run, so it can be checked into a repository or compared in tests.
// Vertices are sorted the same way as in the mermaid package, and
// attributes by name.
package dot

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/hmdsefi/gograph"
)

type options[T comparable] struct {
	name        string
	graphAttrs  map[string]string
	vertexAttrs func(*gograph.Vertex[T]) map[string]string
	edgeAttrs   func(*gograph.Edge[T]) map[string]string
}

// Option configures Marshal and Write.
type Option[T comparable] func(*options[T])

// WithName sets the name of the graph, which is written after digraph or
// graph. By default the graph has no name.
func WithName[T comparable](name string) Option[T] {
	return func(o *options[T]) {
		o.name = name
	}
}

// WithGraphAttributes sets attributes for the whole graph, for example
// {"rankdir": "LR"} to lay it out from left to right.
func WithGraphAttributes[T comparable](attrs map[string]string) Option[T] {
	attrs = maps.Clone(attrs)
	return func(o *options[T]) {
		o.graphAttrs = attrs
	}
}

// WithVertexAttributes sets attributes for each vertex, for example its
// shape or color. By default a vertex has a label attribute with fmt.Sprint
// of its label, and fn can replace it by setting label too.
func WithVertexAttributes[T comparable](fn func(*gograph.Vertex[T]) map[string]string) Option[T] {
	return func(o *options[T]) {
		o.vertexAttrs = fn
	}
}

// WithEdgeAttributes sets attributes for each edge, for example its color
// or style. By default an edge's label attribute is its label when one is
// set, or its weight on a weighted graph, and fn can replace it by setting
// label too. For an undirected edge, fn gets the direction that starts at
// the vertex that is written first.
func WithEdgeAttributes[T comparable](fn func(*gograph.Edge[T]) map[string]string) Option[T] {
	return func(o *options[T]) {
		o.edgeAttrs = fn
	}
}

// Marshal returns the graph in the DOT language.
func Marshal[T comparable](g gograph.Graph[T], opts ...Option[T]) ([]byte, error) {
	var buf bytes.Buffer
	if err := Write(&buf, g, opts...); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Write writes the graph to w in the DOT language.
func Write[T comparable](w io.Writer, g gograph.Graph[T], opts ...Option[T]) error {
	if g == nil {
		return errors.New("dot: graph is nil")
	}

	var o options[T]
	for _, opt := range opts {
		opt(&o)
	}

	vertices := sortedVertices(g)
	ids := make(map[T]int, len(vertices))
	for i, v := range vertices {
		ids[v.Label()] = i
	}

	bw := bufio.NewWriter(w)
	header := "digraph"
	if !g.IsDirected() {
		header = "graph"
	}
	if o.name != "" {
		header += " " + id(o.name)
	}
	_, _ = fmt.Fprintf(bw, "%s {\n", header)

	for _, key := range slices.Sorted(maps.Keys(o.graphAttrs)) {
		_, _ = fmt.Fprintf(bw, "\t%s=%s;\n", id(key), value(o.graphAttrs[key]))
	}

	for i, v := range vertices {
		var custom map[string]string
		if o.vertexAttrs != nil {
			custom = o.vertexAttrs(v)
		}
		attrs := attributes(map[string]string{"label": fmt.Sprint(v.Label())}, custom)
		_, _ = fmt.Fprintf(bw, "\tn%d%s;\n", i, attrs)
	}

	writeEdges(bw, g, vertices, ids, o.edgeAttrs)

	_, _ = bw.WriteString("}\n")
	return bw.Flush()
}

// sortedVertices returns the vertices of g ordered by fmt.Sprint of their
// labels, so the output doesn't depend on the order of GetAllVertices.
// Different labels can print the same, like 1 and "1" in a Graph[any], so
// ties are broken by the type and then the Go syntax of the label.
func sortedVertices[T comparable](g gograph.Graph[T]) []*gograph.Vertex[T] {
	type keyed struct {
		text     string
		typ      string
		goSyntax string
		vertex   *gograph.Vertex[T]
	}

	all := g.GetAllVertices()
	keys := make([]keyed, len(all))
	for i, v := range all {
		keys[i] = keyed{
			text:     fmt.Sprint(v.Label()),
			typ:      fmt.Sprintf("%T", v.Label()),
			goSyntax: fmt.Sprintf("%#v", v.Label()),
			vertex:   v,
		}
	}
	slices.SortStableFunc(keys, func(a, b keyed) int {
		return cmp.Or(
			cmp.Compare(a.text, b.text),
			cmp.Compare(a.typ, b.typ),
			cmp.Compare(a.goSyntax, b.goSyntax),
		)
	})

	vertices := make([]*gograph.Vertex[T], len(keys))
	for i, k := range keys {
		vertices[i] = k.vertex
	}

	return vertices
}

// writeEdges writes the edges grouped by source vertex, in the order of
// vertices, and each group in the order the edges were added.
func writeEdges[T comparable](
	w io.Writer,
	g gograph.Graph[T],
	vertices []*gograph.Vertex[T],
	ids map[T]int,
	edgeAttrs func(*gograph.Edge[T]) map[string]string,
) {
	arrow := "->"
	if !g.IsDirected() {
		arrow = "--"
	}

	for i, v := range vertices {
		wroteLoop := false
		for _, neighbor := range v.Neighbors() {
			j := ids[neighbor.Label()]
			if !g.IsDirected() {
				// the other direction of the edge is written from the vertex with the lower ID
				if j < i || (j == i && wroteLoop) {
					continue
				}
				wroteLoop = wroteLoop || j == i
			}

			e := g.GetEdge(v, neighbor)
			var defaults, custom map[string]string
			switch {
			case e.Label() != "":
				defaults = map[string]string{"label": e.Label()}
			case g.IsWeighted():
				defaults = map[string]string{"label": strconv.FormatFloat(e.Weight(), 'g', -1, 64)}
			}
			if edgeAttrs != nil {
				custom = edgeAttrs(e)
			}
			_, _ = fmt.Fprintf(w, "\tn%d %s n%d%s;\n", i, arrow, j, attributes(defaults, custom))
		}
	}
}

// attributes returns the DOT attribute list for the defaults and the
// attributes set by an option, which take precedence, sorted by name. It
// returns an empty string if there are none.
func attributes(defaults, custom map[string]string) string {
	if len(defaults) == 0 && len(custom) == 0 {
		return ""
	}

	keys := slices.Collect(maps.Keys(defaults))
	for key := range custom {
		if _, ok := defaults[key]; !ok {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)

	var b strings.Builder
	b.WriteString(" [")
	for i, key := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(id(key))
		b.WriteByte('=')
		if v, ok := custom[key]; ok {
			b.WriteString(value(v))
		} else {
			b.WriteString(quote(defaults[key]))
		}
	}
	b.WriteByte(']')

	return b.String()
}

// keywords can't be used as an ID without quotes. DOT keywords are case
// independent.
var keywords = map[string]bool{
	"node": true, "edge": true, "graph": true, "digraph": true, "subgraph": true, "strict": true,
}

// id returns s as a DOT ID: as is if it's a name of ASCII letters, digits
// and underscores that doesn't start with a digit and isn't a keyword, and
// quoted otherwise.
func id(s string) string {
	if s == "" || keywords[strings.ToLower(s)] {
		return quote(s)
	}

	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return quote(s)
		}
	}

	return s
}

// value returns an attribute value set by an option: as is if it's an
// HTML-like string, and quoted otherwise.
func value(s string) string {
	if len(s) >= 2 && s[0] == '<' && s[len(s)-1] == '>' {
		return s
	}

	return quote(s)
}

// escaper escapes the characters that end or change a quoted DOT string,
// and writes line breaks as \n, which Graphviz draws as a new line.
var escaper = strings.NewReplacer(
	`\`, `\\`,
	`"`, `\"`,
	"\r\n", `\n`,
	"\n", `\n`,
	"\r", `\n`,
)

func quote(s string) string {
	return `"` + escaper.Replace(s) + `"`
}
