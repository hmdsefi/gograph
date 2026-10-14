// Package mermaid writes graphs as Mermaid flowcharts, which GitHub, GitLab
// and many documentation tools render inside Markdown.
//
// Every vertex gets a generated node ID (n0, n1, ...) and its label is
// written in quotes, so labels with spaces, punctuation or Mermaid keywords
// such as "end" are safe. Characters that Mermaid treats specially are
// written as entity codes, for example #quot; for a double quote, and a
// newline becomes a line break in the node.
//
// Directed graphs use --> for edges and undirected graphs use ---. Each
// undirected edge is written once, even though the graph stores it in both
// directions. Vertices without edges are still declared, so they appear in
// the diagram. Edges show their label by default when one is set, otherwise
// weighted graphs show the edge weight.
//
// The output is stable: a graph built the same way gives the same text on
// every run, so it can be checked into a repository or compared in tests.
//
// # Size limits
//
// mermaid.js refuses to render a diagram above its maxTextSize and maxEdges
// settings, which default to 50,000 characters and 500 edges. GitHub uses
// the defaults, so large graphs show an error instead of a diagram. Draw a
// part of the graph instead, for example one vertex with its neighbors.
package mermaid

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/hmdsefi/gograph"
)

type classDef struct {
	name  string
	style string
}

type options[T comparable] struct {
	direction   string
	vertexLabel func(*gograph.Vertex[T]) string
	edgeLabel   func(*gograph.Edge[T]) string
	vertexClass func(*gograph.Vertex[T]) string
	classDefs   []classDef
}

// Option configures Marshal and Write.
type Option[T comparable] func(*options[T])

// WithDirection sets the flowchart direction: "TD" (the default) or "TB" for
// top to bottom, "BT" for bottom to top, "LR" for left to right, or "RL" for
// right to left. Marshal and Write return an error for any other value.
func WithDirection[T comparable](dir string) Option[T] {
	return func(o *options[T]) {
		o.direction = dir
	}
}

// WithVertexLabel sets the text shown in each node. The default is
// fmt.Sprint of the vertex label.
func WithVertexLabel[T comparable](fn func(*gograph.Vertex[T]) string) Option[T] {
	return func(o *options[T]) {
		o.vertexLabel = fn
	}
}

// WithEdgeLabel sets the text shown on each edge. An empty string means no
// text. By default, an edge's label is shown when set; otherwise its weight is
// shown on weighted graphs.
func WithEdgeLabel[T comparable](fn func(*gograph.Edge[T]) string) Option[T] {
	return func(o *options[T]) {
		o.edgeLabel = fn
	}
}

// WithVertexClass assigns a class to each vertex, for example to highlight a
// path or a cycle. An empty string means no class. Class names can contain
// letters, digits and '_', can't start with a digit, and can't be a word
// that Mermaid reads as a keyword, such as end, class or style. Declare the
// style of each class with WithClassDef.
func WithVertexClass[T comparable](fn func(*gograph.Vertex[T]) string) Option[T] {
	return func(o *options[T]) {
		o.vertexClass = fn
	}
}

// WithClassDef declares a class and its style, for example
// WithClassDef[string]("critical", "fill:#f96,stroke:#333"). The name follows
// the same rules as in WithVertexClass, and the style can't contain a
// newline or a semicolon.
func WithClassDef[T comparable](name, style string) Option[T] {
	return func(o *options[T]) {
		o.classDefs = append(o.classDefs, classDef{name: name, style: style})
	}
}

// Marshal returns the graph as a Mermaid flowchart.
func Marshal[T comparable](g gograph.Graph[T], opts ...Option[T]) ([]byte, error) {
	var buf bytes.Buffer
	if err := Write(&buf, g, opts...); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Write writes the graph to w as a Mermaid flowchart. If an option is
// invalid, it returns an error before writing anything.
func Write[T comparable](w io.Writer, g gograph.Graph[T], opts ...Option[T]) error {
	if g == nil {
		return errors.New("mermaid: graph is nil")
	}

	o := options[T]{
		direction: "TD",
		vertexLabel: func(v *gograph.Vertex[T]) string {
			return fmt.Sprint(v.Label())
		},
		edgeLabel: func(e *gograph.Edge[T]) string {
			if e.Label() != "" {
				return e.Label()
			}
			if g.IsWeighted() {
				return strconv.FormatFloat(e.Weight(), 'g', -1, 64)
			}
			return ""
		},
	}
	for _, opt := range opts {
		opt(&o)
	}
	if err := o.validate(); err != nil {
		return err
	}

	vertices := sortedVertices(g)
	ids := make(map[T]int, len(vertices))
	for i, v := range vertices {
		ids[v.Label()] = i
	}

	classes, err := groupClasses(vertices, o.vertexClass)
	if err != nil {
		return err
	}

	bw := bufio.NewWriter(w)
	_, _ = fmt.Fprintf(bw, "flowchart %s\n", o.direction)
	for i, v := range vertices {
		label := escape(o.vertexLabel(v))
		if label == "" {
			label = " " // Mermaid rejects a node with empty text
		}
		_, _ = fmt.Fprintf(bw, "    n%d[\"%s\"]\n", i, label)
	}

	writeEdges(bw, g, vertices, ids, o.edgeLabel)

	for _, d := range o.classDefs {
		_, _ = fmt.Fprintf(bw, "    classDef %s %s\n", d.name, d.style)
	}
	for _, c := range classes {
		_, _ = fmt.Fprintf(bw, "    class %s %s\n", strings.Join(c.nodes, ","), c.name)
	}

	return bw.Flush()
}

func (o *options[T]) validate() error {
	switch o.direction {
	case "TD", "TB", "BT", "LR", "RL":
	default:
		return fmt.Errorf("mermaid: invalid direction %q, want TD, TB, BT, LR or RL", o.direction)
	}

	for _, d := range o.classDefs {
		if !validClassName(d.name) {
			return fmt.Errorf("mermaid: invalid class name %q", d.name)
		}
		if d.style == "" || strings.ContainsAny(d.style, ";\r\n") {
			return fmt.Errorf("mermaid: invalid style %q for class %q", d.style, d.name)
		}
	}

	return nil
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
	edgeLabel func(*gograph.Edge[T]) string,
) {
	arrow := "-->"
	if !g.IsDirected() {
		arrow = "---"
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

			label := ""
			if edgeLabel != nil {
				label = edgeLabel(g.GetEdge(v, neighbor))
			}
			if label == "" {
				_, _ = fmt.Fprintf(w, "    n%d %s n%d\n", i, arrow, j)
			} else {
				_, _ = fmt.Fprintf(w, "    n%d %s|\"%s\"| n%d\n", i, arrow, escape(label), j)
			}
		}
	}
}

type class struct {
	name  string
	nodes []string
}

// groupClasses returns the node IDs of each class, with the classes in the
// order they first appear.
func groupClasses[T comparable](vertices []*gograph.Vertex[T], vertexClass func(*gograph.Vertex[T]) string) ([]class, error) {
	if vertexClass == nil {
		return nil, nil
	}

	var classes []class
	index := make(map[string]int)
	for i, v := range vertices {
		name := vertexClass(v)
		if name == "" {
			continue
		}
		if !validClassName(name) {
			return nil, fmt.Errorf("mermaid: invalid class name %q for vertex %v", name, v.Label())
		}

		k, ok := index[name]
		if !ok {
			k = len(classes)
			index[name] = k
			classes = append(classes, class{name: name})
		}
		classes[k].nodes = append(classes[k].nodes, "n"+strconv.Itoa(i))
	}

	return classes, nil
}

// reservedClassNames are the words that the Mermaid flowchart parser reads
// as keywords, so a diagram that uses them as class names doesn't parse.
var reservedClassNames = map[string]bool{
	"end": true, "graph": true, "flowchart": true, "subgraph": true,
	"class": true, "classDef": true, "style": true, "linkStyle": true,
	"click": true, "call": true, "href": true, "interpolate": true,
	"_self": true, "_blank": true, "_parent": true, "_top": true,
}

func validClassName(name string) bool {
	if name == "" || reservedClassNames[name] {
		return false
	}

	for i, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}

	return true
}

// escaper writes the characters that Mermaid parses inside quoted text as
// entity codes.
var escaper = strings.NewReplacer(
	"#", "#35;",
	`"`, "#quot;",
	"&", "#amp;",
	"<", "#lt;",
	">", "#gt;",
	"|", "#124;",
	"`", "#96;",
	"\r\n", "<br>",
	"\n", "<br>",
	"\r", "<br>",
)

func escape(s string) string {
	return escaper.Replace(s)
}
