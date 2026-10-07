package gograph

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Dropping a directed edge whose ends are both selected fails this.
func TestInducedSubgraphKeepsInsideEdges(t *testing.T) {
	g := New[string](Directed(), Weighted())
	a := g.AddVertexByLabel("A", WithVertexWeight(3))
	b := g.AddVertexByLabel("B", WithVertexWeight(4))
	c := g.AddVertexByLabel("C", WithVertexWeight(5))
	if _, err := g.AddEdge(a, b, WithEdgeWeight(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddEdge(b, c, WithEdgeWeight(2)); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddEdge(a, c, WithEdgeWeight(4), WithEdgeLabel("skip")); err != nil {
		t.Fatal(err)
	}

	sub, err := InducedSubgraph(g, "A", "C")
	if err != nil {
		t.Fatalf("InducedSubgraph: %v", err)
	}
	if !sub.IsDirected() || !sub.IsWeighted() || sub.IsAcyclic() {
		t.Fatalf("properties: directed %v weighted %v acyclic %v", sub.IsDirected(), sub.IsWeighted(), sub.IsAcyclic())
	}
	if got := vertexLabels(sub.GetAllVertices()); fmt.Sprint(got) != "[A C]" {
		t.Fatalf("vertices: got %v, want [A C]", got)
	}
	sa, sc := sub.GetVertexByID("A"), sub.GetVertexByID("C")
	if sa.Weight() != 3 || sc.Weight() != 5 {
		t.Fatalf("weights: A %v C %v", sa.Weight(), sc.Weight())
	}
	edge := sub.GetEdge(sa, sc)
	if edge == nil || edge.Weight() != 4 || edge.Label() != "skip" {
		t.Fatalf("edge A to C: %+v", edge)
	}
	if sub.GetEdge(sc, sa) != nil || sub.GetVertexByID("B") != nil {
		t.Fatal("kept a vertex or an edge outside the set")
	}
	if sub.Size() != 1 {
		t.Fatalf("size: got %d, want 1", sub.Size())
	}
}

// Copying an undirected edge twice fails this. AddEdge already stores both directions.
func TestInducedSubgraphUndirectedCopiesOnce(t *testing.T) {
	g := New[string](Weighted())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	if _, err := g.AddEdge(a, b, WithEdgeWeight(1), WithEdgeLabel("ab")); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddEdge(b, c, WithEdgeWeight(2)); err != nil {
		t.Fatal(err)
	}
	ac, err := g.AddEdge(a, c, WithEdgeWeight(3), WithEdgeLabel("ac"))
	if err != nil {
		t.Fatal(err)
	}
	ac.metadata = "forward"
	g.GetEdge(c, a).metadata = "reverse"

	sub, err := InducedSubgraph(g, "C", "A")
	if err != nil {
		t.Fatalf("InducedSubgraph: %v", err)
	}
	if sub.IsDirected() || !sub.IsWeighted() {
		t.Fatalf("properties: directed %v weighted %v", sub.IsDirected(), sub.IsWeighted())
	}
	if got := vertexLabels(sub.GetAllVertices()); fmt.Sprint(got) != "[C A]" {
		t.Fatalf("vertices: got %v, want [C A]", got)
	}
	sc, sa := sub.GetVertexByID("C"), sub.GetVertexByID("A")
	forward := sub.GetEdge(sa, sc)
	reverse := sub.GetEdge(sc, sa)
	if forward == nil || reverse == nil || forward.Weight() != 3 || reverse.Weight() != 3 {
		t.Fatalf("edge A-C: forward %v reverse %v", forward, reverse)
	}
	if forward.Label() != "ac" || reverse.Label() != "ac" {
		t.Fatalf("labels: %q %q", forward.Label(), reverse.Label())
	}
	if forward.Metadata() != "forward" || reverse.Metadata() != "reverse" {
		t.Fatalf("metadata: %v %v", forward.Metadata(), reverse.Metadata())
	}
	if sub.GetEdge(sa, sub.GetVertexByID("B")) != nil || sub.Size() != 2 {
		t.Fatalf("size %d, want the two directions of A-C", sub.Size())
	}
}

// Losing a self-loop fails this.
func TestInducedSubgraphKeepsSelfLoop(t *testing.T) {
	g := New[string](Directed())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	if _, err := g.AddEdge(a, a); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddEdge(a, b); err != nil {
		t.Fatal(err)
	}

	sub, err := InducedSubgraph(g, "A")
	if err != nil {
		t.Fatalf("InducedSubgraph: %v", err)
	}
	sa := sub.GetVertexByID("A")
	if sub.GetEdge(sa, sa) == nil || sub.Size() != 1 || sa.OutDegree() != 1 || sa.InDegree() != 1 {
		t.Fatalf("self-loop: size %d out %d in %d", sub.Size(), sa.OutDegree(), sa.InDegree())
	}

	a.neighbors = append(a.neighbors, a)
	once, err := InducedSubgraph(g, "A")
	if err != nil {
		t.Fatalf("repeated self-loop: %v", err)
	}
	if once.Size() != 1 || once.GetVertexByID("A").OutDegree() != 1 {
		t.Fatalf("repeated self-loop: size %d degree %d", once.Size(), once.GetVertexByID("A").OutDegree())
	}

	u := New[string]()
	ua := u.AddVertexByLabel("A")
	ub := u.AddVertexByLabel("B")
	if _, err := u.AddEdge(ua, ua); err != nil {
		t.Fatal(err)
	}
	if _, err := u.AddEdge(ua, ub); err != nil {
		t.Fatal(err)
	}
	usub, err := InducedSubgraph(u, "A")
	if err != nil {
		t.Fatalf("undirected: %v", err)
	}
	usa := usub.GetVertexByID("A")
	if usub.GetEdge(usa, usa) == nil || usub.Size() != 1 {
		t.Fatalf("undirected self-loop: size %d", usub.Size())
	}
}

// An induced subgraph of a DAG that reports a cycle fails this.
func TestInducedSubgraphKeepsAcyclic(t *testing.T) {
	g := New[string](Weighted(), Acyclic())
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	if _, err := g.AddEdge(a, b, WithEdgeWeight(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddEdge(b, c, WithEdgeWeight(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddEdge(a, c, WithEdgeWeight(1)); err != nil {
		t.Fatal(err)
	}

	sub, err := InducedSubgraph(g, "B", "C", "A")
	if err != nil {
		t.Fatalf("InducedSubgraph: %v", err)
	}
	if !sub.IsAcyclic() || !sub.IsDirected() || !sub.IsWeighted() {
		t.Fatalf("properties: directed %v weighted %v acyclic %v", sub.IsDirected(), sub.IsWeighted(), sub.IsAcyclic())
	}
	if got := vertexLabels(sub.GetAllVertices()); fmt.Sprint(got) != "[B C A]" {
		t.Fatalf("vertices: got %v, want [B C A]", got)
	}
	sb, sc, sa := sub.GetVertexByID("B"), sub.GetVertexByID("C"), sub.GetVertexByID("A")
	if sub.GetEdge(sa, sb) == nil || sub.GetEdge(sb, sc) == nil || sub.GetEdge(sa, sc) == nil {
		t.Fatal("dropped an edge inside the DAG")
	}
	if sub.Size() != 3 {
		t.Fatalf("size: got %d, want 3", sub.Size())
	}
}

// Accepting a missing label, or keeping a repeated label, fails this.
func TestInducedSubgraphLabels(t *testing.T) {
	g := New[string](Directed())
	g.AddVertexByLabel("A")
	g.AddVertexByLabel("B")

	sub, err := InducedSubgraph(g, "A", "missing")
	if sub != nil || !errors.Is(err, ErrVertexDoesNotExist) || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("missing: graph %v err %v, want nil and %v", sub, err, ErrVertexDoesNotExist)
	}

	sub, err = InducedSubgraph(g, "B", "A", "B")
	if err != nil {
		t.Fatalf("duplicates: %v", err)
	}
	if got := vertexLabels(sub.GetAllVertices()); fmt.Sprint(got) != "[B A]" {
		t.Fatalf("vertices: got %v, want [B A]", got)
	}

	empty, err := InducedSubgraph(g)
	if err != nil {
		t.Fatalf("no labels: %v", err)
	}
	if empty.Order() != 0 || empty.Size() != 0 || !empty.IsDirected() || empty.IsWeighted() || empty.IsAcyclic() {
		t.Fatalf("empty: order %d size %d directed %v weighted %v acyclic %v", empty.Order(), empty.Size(), empty.IsDirected(), empty.IsWeighted(), empty.IsAcyclic())
	}

	flags, err := InducedSubgraph(New[int](Weighted(), Acyclic()))
	if err != nil {
		t.Fatal(err)
	}
	if !flags.IsWeighted() || !flags.IsAcyclic() || !flags.IsDirected() {
		t.Fatalf("empty flags: weighted %v acyclic %v directed %v", flags.IsWeighted(), flags.IsAcyclic(), flags.IsDirected())
	}
}

// A result that uses the input's vertex or edge pointers fails this.
func TestInducedSubgraphDoesNotShare(t *testing.T) {
	g := New[string](Directed(), Weighted())
	a := g.AddVertexByLabel("A", WithVertexWeight(3))
	a.metadata = "vertex"
	b := g.AddVertexByLabel("B")
	edge, err := g.AddEdge(a, b, WithEdgeWeight(1), WithEdgeLabel("e"))
	if err != nil {
		t.Fatal(err)
	}
	edge.metadata = "edge"

	sub, err := InducedSubgraph(g, "A", "B")
	if err != nil {
		t.Fatalf("InducedSubgraph: %v", err)
	}
	sa, sb := sub.GetVertexByID("A"), sub.GetVertexByID("B")
	if sa == a || sb == b {
		t.Fatal("result vertex is the input vertex")
	}
	se := sub.GetEdge(sa, sb)
	if se == nil || se == edge || se.Metadata() != "edge" || sa.Metadata() != "vertex" {
		t.Fatalf("copied edge: %v metadata %v", se, sa.Metadata())
	}
	sa.metadata = "changed"
	se.metadata = "changed-edge"
	if a.Metadata() != "vertex" || edge.Metadata() != "edge" {
		t.Fatal("changing the result changed the input")
	}
	sub.RemoveEdges(se)
	if g.GetEdge(a, b) != edge || g.Size() != 1 {
		t.Fatal("removing the result edge removed the input edge")
	}
	if sub.AddVertexByLabel("C") == nil || g.GetVertexByID("C") != nil {
		t.Fatal("adding a vertex to the result changed the input")
	}
	if g.Order() != 2 || g.Size() != 1 {
		t.Fatalf("input order %d size %d", g.Order(), g.Size())
	}
}

// Reordering neighbors to follow the label list, instead of the order they
// were added, fails this.
func TestInducedSubgraphKeepsNeighborOrder(t *testing.T) {
	g := New[string]()
	a := g.AddVertexByLabel("A")
	b := g.AddVertexByLabel("B")
	c := g.AddVertexByLabel("C")
	if _, err := g.AddEdge(a, b); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddEdge(a, c); err != nil {
		t.Fatal(err)
	}

	sub, err := InducedSubgraph(g, "C", "A", "B")
	if err != nil {
		t.Fatalf("InducedSubgraph: %v", err)
	}
	got := vertexLabels(sub.GetVertexByID("A").Neighbors())
	if fmt.Sprint(got) != "[B C]" {
		t.Fatalf("neighbors of A: got %v, want [B C]", got)
	}
	if fmt.Sprint(vertexLabels(sub.GetVertexByID("C").Neighbors())) != "[A]" {
		t.Fatalf("neighbors of C: got %v, want [A]", vertexLabels(sub.GetVertexByID("C").Neighbors()))
	}
}

func TestInducedSubgraphNilPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil graph did not panic")
		}
	}()
	_, _ = InducedSubgraph[string](nil)
}

func TestInducedSubgraphNilBasePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil base graph did not panic")
		}
	}()
	var g *baseGraph[string]
	_, _ = InducedSubgraph[string](g)
}

func ExampleInducedSubgraph() {
	g := New[string](Directed(), Weighted())
	a := g.AddVertexByLabel("api")
	b := g.AddVertexByLabel("auth")
	c := g.AddVertexByLabel("billing")
	_, _ = g.AddEdge(a, b, WithEdgeWeight(1))
	_, _ = g.AddEdge(b, c, WithEdgeWeight(2))
	_, _ = g.AddEdge(a, c, WithEdgeWeight(4))

	sub, err := InducedSubgraph(g, "api", "billing")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, v := range sub.GetAllVertices() {
		fmt.Println(v.Label())
	}
	edge := sub.GetEdge(sub.GetVertexByID("api"), sub.GetVertexByID("billing"))
	fmt.Println(edge.Weight())
	// Output:
	// api
	// billing
	// 4
}

func BenchmarkInducedSubgraph(b *testing.B) {
	const n = 4000
	g := New[int](Directed(), Weighted())
	verts := make([]*Vertex[int], n)
	for i := 0; i < n; i++ {
		verts[i] = g.AddVertexByLabel(i)
	}
	for i := 0; i < n-1; i++ {
		if _, err := g.AddEdge(verts[i], verts[i+1], WithEdgeWeight(1)); err != nil {
			b.Fatal(err)
		}
	}
	labels := make([]int, 256)
	for i := range labels {
		labels[i] = 1000 + i
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sub, err := InducedSubgraph(g, labels...)
		if err != nil {
			b.Fatal(err)
		}
		if sub.Order() != 256 || sub.Size() != 255 {
			b.Fatalf("order %d size %d", sub.Order(), sub.Size())
		}
	}
}
