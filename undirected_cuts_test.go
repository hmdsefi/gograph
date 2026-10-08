package gograph

import (
	"errors"
	"reflect"
	"testing"
)

func TestUndirectedCuts(t *testing.T) {
	for _, tt := range []struct {
		name    string
		n       int
		edges   [][2]int
		bridges [][2]int
		points  []int
	}{
		{"empty", 0, nil, [][2]int{}, []int{}},
		{"isolated", 3, nil, [][2]int{}, []int{}},
		{"chain", 4, [][2]int{{0, 1}, {1, 2}, {2, 3}}, [][2]int{{2, 3}, {1, 2}, {0, 1}}, []int{1, 2}},
		{"two cycles sharing a vertex", 5, [][2]int{{0, 1}, {0, 2}, {1, 2}, {1, 3}, {1, 4}, {3, 4}}, [][2]int{}, []int{1}},
		{"root with two children", 3, [][2]int{{0, 1}, {0, 2}}, [][2]int{{0, 1}, {0, 2}}, []int{0}},
		{"self loop", 2, [][2]int{{0, 0}, {0, 1}}, [][2]int{{0, 1}}, []int{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := New[int]()
			for i := range tt.n {
				g.AddVertexByLabel(i)
			}
			for _, e := range tt.edges {
				if _, err := g.AddEdge(g.GetVertexByID(e[0]), g.GetVertexByID(e[1])); err != nil {
					t.Fatal(err)
				}
			}
			bridges, points, err := UndirectedCuts(g)
			if err != nil {
				t.Fatal(err)
			}
			got := make([][2]int, 0, len(bridges))
			for _, e := range bridges {
				got = append(got, [2]int{e.Source().Label(), e.Destination().Label()})
				if e != g.GetEdge(e.Source(), e.Destination()) {
					t.Fatal("copied edge")
				}
			}
			if !reflect.DeepEqual(got, tt.bridges) {
				t.Fatalf("bridges %v, want %v", got, tt.bridges)
			}
			labels := make([]int, 0, len(points))
			for _, v := range points {
				labels = append(labels, v.Label())
				if v != g.GetVertexByID(v.Label()) {
					t.Fatal("copied vertex")
				}
			}
			if !reflect.DeepEqual(labels, tt.points) {
				t.Fatalf("points %v, want %v", labels, tt.points)
			}
		})
	}
}

func TestUndirectedCutsNotFromNew(t *testing.T) {
	for _, g := range []Graph[int]{nil, (*baseGraph[int])(nil)} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			_, _, _ = UndirectedCuts(g)
		}()
	}
}

func TestUndirectedCutsDirected(t *testing.T) {
	g := New[int](Directed())
	if _, _, err := UndirectedCuts(g); !errors.Is(err, ErrNotUndirected) {
		t.Fatalf("error %v", err)
	}
}

func TestUndirectedCutsAfterRemoval(t *testing.T) {
	g := New[int]()
	for i := range 6 {
		g.AddVertexByLabel(i)
	}
	for i := 1; i < 6; i++ {
		if _, err := g.AddEdge(g.GetVertexByID(i-1), g.GetVertexByID(i)); err != nil {
			t.Fatal(err)
		}
	}
	// Leave gaps in the vertex order. Removing two of six does not compact it.
	g.RemoveVertices(g.GetVertexByID(1), g.GetVertexByID(4))

	bridges, points, err := UndirectedCuts(g)
	if err != nil {
		t.Fatal(err)
	}
	got := make([][2]int, 0, len(bridges))
	for _, e := range bridges {
		got = append(got, [2]int{e.Source().Label(), e.Destination().Label()})
	}
	if !reflect.DeepEqual(got, [][2]int{{2, 3}}) {
		t.Fatalf("bridges %v", got)
	}
	if len(points) != 0 {
		t.Fatalf("points %d", len(points))
	}
}
