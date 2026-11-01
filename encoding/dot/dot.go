// Package dot writes graphs in the DOT language of Graphviz, which dot, the
// other Graphviz layout tools and many viewers and editor plugins read.
//
// Every vertex gets a generated node ID (n0, n1, ...) and its label goes in
// the label attribute, so two labels that print the same don't collide.
// Attribute values are written in quotes, with double quotes and
// backslashes escaped, & written as &amp; so Graphviz doesn't read it as
// the start of an entity, newlines written as \n, and NUL bytes written as
// U+FFFD. A value set by an option is written as is instead when Graphviz
// reads all of it as one HTML-like label: it starts with <, its angle
// brackets balance, the first < closes at the last character, and there is
// more than spaces between them, as in <<b>bold</b>>. Any other value is
// quoted, so <b>bold</b> is drawn as that text.
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
// graph. By default the graph has no name. A DOT ID can't have a backslash
// at the end or before a double quote or a line break, so Marshal and
// Write return an error for such a name or attribute name.
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
// of its label, and fn can replace it by setting label too. Attribute names
// follow the same rule as the name in WithName.
func WithVertexAttributes[T comparable](fn func(*gograph.Vertex[T]) map[string]string) Option[T] {
	return func(o *options[T]) {
		o.vertexAttrs = fn
	}
}

// WithEdgeAttributes sets attributes for each edge, for example its color
// or style. By default an edge's label attribute is its label when one is
// set, or its weight on a weighted graph, and fn can replace it by setting
// label too. For an undirected edge, fn gets the direction that starts at
// the vertex that is written first. Attribute names follow the same rule as
// the name in WithName.
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

// Write writes the graph to w in the DOT language. If the name or a graph
// attribute name is invalid, it returns an error before writing anything.
// If a vertex or edge attribute name is invalid, it returns an error, and
// w may already hold part of the output.
func Write[T comparable](w io.Writer, g gograph.Graph[T], opts ...Option[T]) error {
	if g == nil {
		return errors.New("dot: graph is nil")
	}

	var o options[T]
	for _, opt := range opts {
		opt(&o)
	}
	if !validID(o.name) {
		return fmt.Errorf("dot: graph name %q can't be written as a DOT ID", o.name)
	}
	for key := range o.graphAttrs {
		if !validID(key) {
			return fmt.Errorf("dot: graph attribute %q can't be written as a DOT ID", key)
		}
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
		attrs, err := attributes(map[string]string{"label": fmt.Sprint(v.Label())}, custom)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(bw, "\tn%d%s;\n", i, attrs)
	}

	if err := writeEdges(bw, g, vertices, ids, o.edgeAttrs); err != nil {
		return err
	}

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
) error {
	arrow := "->"
	if !g.IsDirected() {
		arrow = "--"
	}

	for i, v := range vertices {
		wroteLoop := false
		for neighbor := range v.AllNeighbors() {
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
			attrs, err := attributes(defaults, custom)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(w, "\tn%d %s n%d%s;\n", i, arrow, j, attrs)
		}
	}

	return nil
}

// attributes returns the DOT attribute list for the defaults and the
// attributes set by an option, which take precedence, sorted by name. It
// returns an empty string if there are none, and an error if the name of
// an attribute set by an option can't be written as a DOT ID.
func attributes(defaults, custom map[string]string) (string, error) {
	if len(defaults) == 0 && len(custom) == 0 {
		return "", nil
	}

	keys := slices.Collect(maps.Keys(defaults))
	for key := range custom {
		if !validID(key) {
			return "", fmt.Errorf("dot: attribute %q can't be written as a DOT ID", key)
		}
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

	return b.String(), nil
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
		return quoteID(s)
	}

	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return quoteID(s)
		}
	}

	return s
}

// value returns an attribute value set by an option: as is if it's an
// HTML-like string, and quoted otherwise.
func value(s string) string {
	if htmlLike(s) {
		return s
	}

	return quote(s)
}

// htmlLike reports whether Graphviz reads all of s as one HTML-like string:
// it starts with <, its angle brackets balance, and the first < closes at
// the last character. Graphviz ends the string at the > that balances the
// first <, so anything after it would be a syntax error. Older versions of
// Graphviz also reject an HTML-like string with nothing but spaces in it.
func htmlLike(s string) bool {
	if len(s) < 2 || s[0] != '<' {
		return false
	}

	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '<':
			depth++
		case '>':
			depth--
			if depth == 0 {
				return i == len(s)-1 && strings.TrimSpace(s[1:i]) != ""
			}
		case 0:
			return false
		}
	}

	return false
}

// escaper escapes the characters that end or change a quoted DOT string,
// and writes line breaks as \n, which Graphviz draws as a new line.
// Graphviz decodes entities such as &amp; in quoted strings, so & is
// written as &amp; to be drawn as is. It can't read a NUL byte in a quoted
// string, so NUL is written as U+FFFD.
var escaper = strings.NewReplacer(
	`\`, `\\`,
	`"`, `\"`,
	"&", "&amp;",
	"\r\n", `\n`,
	"\n", `\n`,
	"\r", `\n`,
	"\x00", "\uFFFD",
)

func quote(s string) string {
	return `"` + escaper.Replace(s) + `"`
}

// idEscaper escapes the characters that end a quoted DOT ID. Graphviz
// doesn't unescape IDs a second time, as it does labels, so a backslash is
// written as is.
var idEscaper = strings.NewReplacer(
	`"`, `\"`,
	"\x00", "\uFFFD",
)

func quoteID(s string) string {
	return `"` + idEscaper.Replace(s) + `"`
}

// validID reports whether s can be written as a quoted DOT ID. Graphviz
// reads a backslash before a double quote as an escaped quote and drops a
// backslash before a line break, so neither can be written, and a
// backslash at the end would escape the closing quote.
func validID(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			continue
		}
		if i == len(s)-1 {
			return false
		}
		if next := s[i+1]; next == '"' || next == '\n' || next == '\r' {
			return false
		}
	}

	return true
}
