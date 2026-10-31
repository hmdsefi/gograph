package gograph

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	// ErrIncompatibleGraphs is returned by Diff when one graph is directed and
	// the other is undirected, since their edges can't be matched.
	ErrIncompatibleGraphs = errors.New("graphs are not both directed or both undirected")

	// ErrDiffMismatch is returned by GraphDiff.Apply when the diff doesn't fit
	// the graph it's applied to.
	ErrDiffMismatch = errors.New("diff does not match the graph")
)

// VertexEntry is a vertex that a GraphDiff adds or removes.
type VertexEntry[T comparable] struct {
	Label  T
	Weight float64
}

// VertexChange is a vertex whose weight differs between the two graphs.
type VertexChange[T comparable] struct {
	Label                T
	OldWeight, NewWeight float64
}

// EdgeEntry is an edge that a GraphDiff adds or removes, with its weight and
// its edge label.
type EdgeEntry[T comparable] struct {
	From, To T
	Weight   float64
	Label    string
}

// EdgeChange is an edge whose weight or edge label differs between the two
// graphs.
type EdgeChange[T comparable] struct {
	From, To             T
	OldWeight, NewWeight float64
	OldLabel, NewLabel   string
}

// PropertiesSummary holds the properties a graph reports through GraphType.
type PropertiesSummary struct {
	Directed, Weighted, Acyclic bool
}

// GraphDiff holds the differences that turn one graph into another, as Diff
// returns them. Its fields are plain exported values, so encoding/json can
// write and read it, as long as no weight is NaN or infinite.
type GraphDiff[T comparable] struct {
	AddedVertices   []VertexEntry[T]  `json:",omitempty"`
	RemovedVertices []VertexEntry[T]  `json:",omitempty"`
	ChangedVertices []VertexChange[T] `json:",omitempty"`

	AddedEdges   []EdgeEntry[T]  `json:",omitempty"`
	RemovedEdges []EdgeEntry[T]  `json:",omitempty"`
	ChangedEdges []EdgeChange[T] `json:",omitempty"`

	// OldProperties and NewProperties are set when the graphs differ in
	// whether they're weighted or acyclic.
	OldProperties, NewProperties *PropertiesSummary `json:",omitempty"`
}

// DiffOption changes how Diff compares two graphs.
type DiffOption func(*diffOptions)

type diffOptions struct {
	tolerance float64
}

// WithWeightTolerance makes Diff treat two weights as equal when they differ
// by at most eps. A negative or NaN eps keeps the exact comparison.
func WithWeightTolerance(eps float64) DiffOption {
	return func(o *diffOptions) {
		o.tolerance = eps
	}
}

// sameWeight reports whether x and y are equal, both NaN, or within the
// tolerance of each other.
func (o diffOptions) sameWeight(x, y float64) bool {
	return x == y || (math.IsNaN(x) && math.IsNaN(y)) || math.Abs(x-y) <= o.tolerance
}

// Diff returns the differences that turn graph a into graph b. It matches
// vertices by label and edges by the labels of their two vertices. It
// compares the weights of vertices and edges, the edge labels, and whether
// the graphs are weighted or acyclic. It leaves out metadata, the order of
// the vertices and edges, and the edge insertion indexes.
//
// Weights are equal when they're == or both NaN, so a graph has no
// differences from itself. WithWeightTolerance also accepts weights that are
// close. A removed vertex removes its edges, and RemovedEdges lists those
// edges too, so the diff is complete on its own.
//
// In an undirected graph, each edge is listed once, in the orientation in
// which it's found first: in a for a removed edge, and in b for an added or
// changed one. A - B in a and B - A in b are the same edge. If one graph is
// directed and the other is undirected, it returns ErrIncompatibleGraphs.
//
// Removed vertices and edges come in the order of a, and added and changed
// ones in the order of b, so the same graphs always give the same diff.
//
// For graphs from New, it takes O(V+E) time for the vertices and edges of
// both graphs, and allocates only the diff it returns. Any other Graph is
// read through Vertices, Edges, GetVertexByID and GetEdge, and an undirected
// one also through GetAllVertices, with a map of O(V) space. It panics if a
// or b is nil.
func Diff[T comparable](a, b Graph[T], opts ...DiffOption) (*GraphDiff[T], error) {
	if isNilGraph(a) || isNilGraph(b) {
		panic("gograph: Diff requires non-nil graphs")
	}
	if a.IsDirected() != b.IsDirected() {
		return nil, fmt.Errorf("%w: a is directed: %v, b is directed: %v",
			ErrIncompatibleGraphs, a.IsDirected(), b.IsDirected())
	}

	var o diffOptions
	for _, opt := range opts {
		opt(&o)
	}

	d := &GraphDiff[T]{}
	if before, after := summarize(a), summarize(b); before != after {
		d.OldProperties, d.NewProperties = &before, &after
	}

	from, to := indexGraph(a), indexGraph(b)
	for v := range Vertices(a) {
		if to.vertex(v.label) == nil {
			d.RemovedVertices = append(d.RemovedVertices, VertexEntry[T]{Label: v.label, Weight: v.Weight()})
		}
	}
	for v := range Vertices(b) {
		old := from.vertex(v.label)
		switch {
		case old == nil:
			d.AddedVertices = append(d.AddedVertices, VertexEntry[T]{Label: v.label, Weight: v.Weight()})
		case !o.sameWeight(old.Weight(), v.Weight()):
			d.ChangedVertices = append(d.ChangedVertices, VertexChange[T]{
				Label: v.label, OldWeight: old.Weight(), NewWeight: v.Weight(),
			})
		}
	}

	for e := range Edges(a) {
		if from.firstFound(e) && to.edge(e.source.label, e.dest.label) == nil {
			d.RemovedEdges = append(d.RemovedEdges, edgeEntry(e))
		}
	}
	for e := range Edges(b) {
		if !to.firstFound(e) {
			continue
		}
		old := from.edge(e.source.label, e.dest.label)
		switch {
		case old == nil:
			d.AddedEdges = append(d.AddedEdges, edgeEntry(e))
		case !o.sameWeight(old.Weight(), e.Weight()) || old.Label() != e.Label():
			d.ChangedEdges = append(d.ChangedEdges, EdgeChange[T]{
				From: e.source.label, To: e.dest.label,
				OldWeight: old.Weight(), NewWeight: e.Weight(),
				OldLabel: old.Label(), NewLabel: e.Label(),
			})
		}
	}

	return d, nil
}

func isNilGraph[T comparable](g Graph[T]) bool {
	bg, ok := g.(*baseGraph[T])
	return g == nil || (ok && bg == nil)
}

func summarize[T comparable](g Graph[T]) PropertiesSummary {
	return PropertiesSummary{Directed: g.IsDirected(), Weighted: g.IsWeighted(), Acyclic: g.IsAcyclic()}
}

func edgeEntry[T comparable](e *Edge[T]) EdgeEntry[T] {
	return EdgeEntry[T]{From: e.source.label, To: e.dest.label, Weight: e.Weight(), Label: e.Label()}
}

// graphIndex looks up the vertices and edges of a graph by label.
type graphIndex[T comparable] struct {
	g    Graph[T]
	base *baseGraph[T]

	// place holds the index of each vertex in GetAllVertices for an
	// undirected graph that New didn't build
	place map[T]int

	// from and dests are the last source looked up and its edges, since
	// Edges yields the edges of each source together
	from   T
	dests  map[T]*Edge[T]
	cached bool
}

func indexGraph[T comparable](g Graph[T]) graphIndex[T] {
	x := graphIndex[T]{g: g}
	x.base, _ = g.(*baseGraph[T])
	if x.base == nil && !g.IsDirected() {
		vertices := g.GetAllVertices()
		x.place = make(map[T]int, len(vertices))
		for i, v := range vertices {
			x.place[v.label] = i
		}
	}

	return x
}

func (x graphIndex[T]) vertex(label T) *Vertex[T] {
	if x.base != nil {
		return x.base.vertices[label]
	}

	return x.g.GetVertexByID(label)
}

func (x *graphIndex[T]) edge(from, to T) *Edge[T] {
	if x.base != nil {
		if !x.cached || x.from != from {
			x.from, x.dests, x.cached = from, x.base.edges[from], true
		}
		return x.dests[to]
	}

	return x.g.GetEdge(x.g.GetVertexByID(from), x.g.GetVertexByID(to))
}

// firstFound reports whether e is the first of the two stored directions of
// its edge in the order of Edges. Every edge of a directed graph is.
func (x graphIndex[T]) firstFound(e *Edge[T]) bool {
	switch {
	case x.g.IsDirected():
		return true
	case x.base != nil:
		return e.source.position <= e.dest.position
	default:
		return x.place[e.source.label] <= x.place[e.dest.label]
	}
}

// IsEmpty reports whether d has no differences. A nil diff is empty.
func (d *GraphDiff[T]) IsEmpty() bool {
	return d == nil || (len(d.AddedVertices) == 0 && len(d.RemovedVertices) == 0 &&
		len(d.ChangedVertices) == 0 && len(d.AddedEdges) == 0 && len(d.RemovedEdges) == 0 &&
		len(d.ChangedEdges) == 0 && d.OldProperties == nil && d.NewProperties == nil)
}

// String returns the differences one per line: "+" for an added vertex or
// edge, "-" for a removed one and "~" for a change, as in
//
//	~ weighted false -> true
//	+ vertex D
//	- vertex B weight 2
//	~ vertex C weight 1 -> 2
//	+ edge A -> D label "uses"
//	- edge A -> B
//	~ edge A -> C weight 3 -> 5
//
// A changed property comes first, then the added, removed and changed
// vertices, then the edges in the same order, each group in the order of the
// diff. An added or removed item shows its weight when it isn't 0, and an
// edge its label when it isn't empty. An undirected edge shows the
// orientation of the diff. An empty diff returns "".
func (d *GraphDiff[T]) String() string {
	if d == nil {
		return ""
	}

	var b strings.Builder
	line := func(format string, args ...any) {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, format, args...)
	}

	if d.OldProperties != nil || d.NewProperties != nil {
		var before, after PropertiesSummary
		if d.OldProperties != nil {
			before = *d.OldProperties
		}
		if d.NewProperties != nil {
			after = *d.NewProperties
		}
		for _, p := range []struct {
			name          string
			before, after bool
		}{
			{"directed", before.Directed, after.Directed},
			{"weighted", before.Weighted, after.Weighted},
			{"acyclic", before.Acyclic, after.Acyclic},
		} {
			if p.before != p.after {
				line("~ %s %v -> %v", p.name, p.before, p.after)
			}
		}
	}

	for _, v := range d.AddedVertices {
		line("+ vertex %v%s", v.Label, weightText(v.Weight))
	}
	for _, v := range d.RemovedVertices {
		line("- vertex %v%s", v.Label, weightText(v.Weight))
	}
	for _, v := range d.ChangedVertices {
		line("~ vertex %v weight %s -> %s", v.Label, formatWeight(v.OldWeight), formatWeight(v.NewWeight))
	}
	for _, e := range d.AddedEdges {
		line("+ edge %v -> %v%s%s", e.From, e.To, weightText(e.Weight), labelText(e.Label))
	}
	for _, e := range d.RemovedEdges {
		line("- edge %v -> %v%s%s", e.From, e.To, weightText(e.Weight), labelText(e.Label))
	}
	for _, e := range d.ChangedEdges {
		var changes []string
		if e.OldWeight != e.NewWeight && !(math.IsNaN(e.OldWeight) && math.IsNaN(e.NewWeight)) {
			changes = append(changes, " weight "+formatWeight(e.OldWeight)+" -> "+formatWeight(e.NewWeight))
		}
		if e.OldLabel != e.NewLabel {
			changes = append(changes, " label "+strconv.Quote(e.OldLabel)+" -> "+strconv.Quote(e.NewLabel))
		}
		line("~ edge %v -> %v%s", e.From, e.To, strings.Join(changes, ","))
	}

	return b.String()
}

func formatWeight(w float64) string {
	return strconv.FormatFloat(w, 'g', -1, 64)
}

func weightText(w float64) string {
	if w == 0 {
		return ""
	}

	return " weight " + formatWeight(w)
}

func labelText(label string) string {
	if label == "" {
		return ""
	}

	return " label " + strconv.Quote(label)
}

// Apply returns a copy of base with the differences of d made to it, so
// Diff(a, b) applied to a gives a graph with no differences from b. The copy
// comes from Clone, so it's a graph from New, and base doesn't change.
//
// It removes the edges, then the vertices, adds the vertices, then the
// edges, and then sets the changed weights, edge labels and properties. A
// removed or changed vertex or edge has to be in base, an added one can't
// be, and an added edge needs both of its vertices. When d changes the
// properties, base has to have the old ones. The first difference that
// doesn't fit returns ErrDiffMismatch, wrapped with what it found. Added
// edges that close a cycle in an acyclic graph return ErrDAGCycle, and
// making a graph with a cycle acyclic returns ErrDAGHasCycle. On an error it
// returns nil. A nil diff returns a copy of base.
//
// Copying base takes O(V+E) time and space. A removed edge or vertex then
// costs what RemoveEdges or RemoveVertices does, and every other difference
// O(1). When the result is acyclic and d adds edges or makes it acyclic, one
// O(V+E) search checks it for a cycle, so a large diff to an acyclic graph
// costs no more than copying it. It panics if base is nil.
func (d *GraphDiff[T]) Apply(base Graph[T]) (Graph[T], error) {
	g, _ := Clone(base).(*baseGraph[T])
	if d == nil {
		return g, nil
	}
	if err := d.applyTo(g); err != nil {
		return nil, err
	}

	return g, nil
}

func (d *GraphDiff[T]) applyTo(g *baseGraph[T]) error {
	after, changed, err := d.checkProperties(g)
	if err != nil {
		return err
	}
	wasAcyclic := g.properties.isAcyclic
	acyclic := wasAcyclic
	if changed {
		acyclic = after.Acyclic
	}

	// One edge at a time, an acyclic graph can reorder its vertices on each
	// edge, so the edges go in unchecked and one search looks for a cycle.
	if wasAcyclic && (!acyclic || len(d.AddedEdges) > 0) {
		g.properties.isAcyclic = false
		g.ord = nil
	}

	if err := d.remove(g); err != nil {
		return err
	}
	if err := d.add(g); err != nil {
		return err
	}
	if err := d.change(g); err != nil {
		return err
	}

	if changed {
		g.properties.isWeighted = after.Weighted
	}
	if acyclic && !g.properties.isAcyclic {
		if g.hasDirectedCycle() {
			if wasAcyclic {
				return ErrDAGCycle
			}
			return ErrDAGHasCycle
		}
		g.properties.isAcyclic = true
		g.ord = nil
	}

	return nil
}

// checkProperties returns the properties d gives the graph, and whether d
// changes them.
func (d *GraphDiff[T]) checkProperties(g *baseGraph[T]) (PropertiesSummary, bool, error) {
	if d.OldProperties == nil && d.NewProperties == nil {
		return PropertiesSummary{}, false, nil
	}
	if d.OldProperties == nil || d.NewProperties == nil {
		return PropertiesSummary{}, false, mismatchf("OldProperties and NewProperties aren't both set")
	}

	before, after := *d.OldProperties, *d.NewProperties
	if before.Directed != after.Directed || (after.Acyclic && !after.Directed) {
		return PropertiesSummary{}, false, fmt.Errorf("%w: the diff gives the properties %+v", ErrIncompatibleGraphs, after)
	}
	if got := summarize[T](g); got != before {
		return PropertiesSummary{}, false, mismatchf("the graph has the properties %+v, not %+v", got, before)
	}

	return after, true, nil
}

func (d *GraphDiff[T]) remove(g *baseGraph[T]) error {
	for _, e := range d.RemovedEdges {
		edge := g.edges[e.From][e.To]
		if edge == nil {
			return mismatchf("edge %v -> %v isn't in the graph", e.From, e.To)
		}
		g.removeAllEdges(edge)
	}
	for _, v := range d.RemovedVertices {
		vertex := g.vertices[v.Label]
		if vertex == nil {
			return mismatchf("vertex %v isn't in the graph", v.Label)
		}
		g.removeVertex(vertex)
	}

	return nil
}

func (d *GraphDiff[T]) add(g *baseGraph[T]) error {
	for _, v := range d.AddedVertices {
		if g.vertices[v.Label] != nil {
			return mismatchf("vertex %v is already in the graph", v.Label)
		}
		g.addVertex(NewVertex(v.Label, WithVertexWeight(v.Weight)))
	}
	for _, e := range d.AddedEdges {
		from, to := g.vertices[e.From], g.vertices[e.To]
		switch {
		case from == nil:
			return mismatchf("vertex %v of edge %v -> %v isn't in the graph", e.From, e.From, e.To)
		case to == nil:
			return mismatchf("vertex %v of edge %v -> %v isn't in the graph", e.To, e.From, e.To)
		case g.edges[e.From][e.To] != nil:
			return mismatchf("edge %v -> %v is already in the graph", e.From, e.To)
		}
		weight, label := WithEdgeWeight(e.Weight), WithEdgeLabel(e.Label)
		if !g.properties.isDirected && from != to {
			g.link(to, from, weight, label)
		}
		g.link(from, to, weight, label)
	}

	return nil
}

func (d *GraphDiff[T]) change(g *baseGraph[T]) error {
	for _, v := range d.ChangedVertices {
		vertex := g.vertices[v.Label]
		if vertex == nil {
			return mismatchf("vertex %v isn't in the graph", v.Label)
		}
		vertex.properties.weight = v.NewWeight
	}
	for _, e := range d.ChangedEdges {
		edge := g.edges[e.From][e.To]
		if edge == nil {
			return mismatchf("edge %v -> %v isn't in the graph", e.From, e.To)
		}
		edge.properties = EdgeProperties{weight: e.NewWeight, label: e.NewLabel}
		if !g.properties.isDirected {
			g.edges[e.To][e.From].properties = edge.properties
		}
	}

	return nil
}

func mismatchf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrDiffMismatch, fmt.Sprintf(format, args...))
}
