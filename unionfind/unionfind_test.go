package unionfind

import (
	"reflect"
	"testing"
)

func twoGroups() *Sets[string] {
	s := New[string]()
	s.Union("a", "b")
	s.Union("c", "d")
	s.Union("b", "e")
	s.Union("d", "f")
	s.Union("e", "a")
	return s
}

func TestUnionFormsGroups(t *testing.T) {
	s := twoGroups()

	connected := []struct {
		a, b string
		want bool
	}{
		{"a", "b", true},
		{"a", "e", true},
		{"b", "e", true},
		{"c", "d", true},
		{"c", "f", true},
		{"a", "c", false},
		{"e", "f", false},
		{"a", "a", true},
	}
	for _, tc := range connected {
		if got := s.Connected(tc.a, tc.b); got != tc.want {
			t.Errorf("Connected(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}

	want := [][]string{{"a", "b", "e"}, {"c", "d", "f"}}
	if got := s.Groups(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Groups() = %v, want %v", got, want)
	}
}

func TestPathCompressionKeepsConnected(t *testing.T) {
	s := New[int]()
	// A long chain, then union two chains so ranks are equal and paths deepen.
	for i := 1; i < 64; i++ {
		s.Union(i-1, i)
	}
	for i := 101; i < 164; i++ {
		s.Union(i-1, i)
	}

	before := make(map[[2]int]bool)
	pairs := [][2]int{{0, 63}, {5, 40}, {100, 163}, {0, 100}, {63, 163}}
	for _, p := range pairs {
		before[p] = s.Connected(p[0], p[1])
	}
	// Find compresses every path it walks.
	for i := 0; i < 64; i++ {
		if _, ok := s.Find(i); !ok {
			t.Fatalf("Find(%d) not found", i)
		}
	}
	for _, p := range pairs {
		if got := s.Connected(p[0], p[1]); got != before[p] {
			t.Errorf("Connected(%d, %d) = %v after compression, was %v", p[0], p[1], got, before[p])
		}
	}
	for i := 0; i < 64; i++ {
		if s.parent[s.index[i]] != s.root(s.index[i]) {
			t.Fatalf("label %d does not point at its root after Find", i)
		}
	}

	s.Union(0, 100)
	if !s.Connected(63, 163) {
		t.Fatal("Connected(63, 163) = false after merging the chains")
	}
}

func TestFindUnknown(t *testing.T) {
	s := twoGroups()
	if r, ok := s.Find("z"); ok || r != "" {
		t.Fatalf("Find(unknown) = %q, %v, want \"\", false", r, ok)
	}
	if s.Connected("z", "z") {
		t.Fatal("Connected(unknown, unknown) = true")
	}
	if s.Connected("a", "z") || s.Connected("z", "a") {
		t.Fatal("Connected with an unknown label = true")
	}
}

func TestFindRepresentative(t *testing.T) {
	s := twoGroups()
	ra, _ := s.Find("a")
	for _, x := range []string{"b", "e"} {
		if r, ok := s.Find(x); !ok || r != ra {
			t.Errorf("Find(%q) = %q, %v, want %q, true", x, r, ok, ra)
		}
	}
	if rc, _ := s.Find("c"); rc == ra {
		t.Fatalf("groups share representative %q", ra)
	}
}

func TestUnionWithItself(t *testing.T) {
	s := New[string]()
	s.Union("x", "x")
	if r, ok := s.Find("x"); !ok || r != "x" {
		t.Fatalf("Find(x) = %q, %v, want x, true", r, ok)
	}
	if !s.Connected("x", "x") {
		t.Fatal("Connected(x, x) = false")
	}
	if got, want := s.Groups(), [][]string{{"x"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Groups() = %v, want %v", got, want)
	}
}

func TestAdd(t *testing.T) {
	s := New[string]()
	s.Add("a")
	s.Add("b")
	s.Add("a")
	if s.Connected("a", "b") {
		t.Fatal("Add merged two labels")
	}
	s.Union("a", "b")
	s.Add("b")
	if !s.Connected("a", "b") {
		t.Fatal("Add on an existing label split its group")
	}
	if got, want := s.Groups(), [][]string{{"a", "b"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Groups() = %v, want %v", got, want)
	}
}

func TestEmpty(t *testing.T) {
	s := New[string]()
	if got := s.Groups(); got == nil || len(got) != 0 {
		t.Fatalf("Groups() = %#v, want an empty slice", got)
	}
}

func TestDeterministic(t *testing.T) {
	run := func() ([][]string, []string) {
		s := New[string]()
		s.Union("d", "c")
		s.Union("a", "b")
		s.Union("c", "b")
		s.Add("e")
		s.Union("f", "e")
		var reps []string
		for _, x := range []string{"a", "b", "c", "d", "e", "f"} {
			r, _ := s.Find(x)
			reps = append(reps, r)
		}
		return s.Groups(), reps
	}
	g1, r1 := run()
	g2, r2 := run()
	if !reflect.DeepEqual(g1, g2) || !reflect.DeepEqual(r1, r2) {
		t.Fatalf("runs differ: %v %v vs %v %v", g1, r1, g2, r2)
	}
	want := [][]string{{"d", "c", "a", "b"}, {"e", "f"}}
	if !reflect.DeepEqual(g1, want) {
		t.Fatalf("Groups() = %v, want %v", g1, want)
	}
}

func TestUnionByRank(t *testing.T) {
	s := New[string]()
	s.Union("a", "b") // rank(a) = 1
	s.Union("c", "a") // smaller group joins the larger one
	if r, _ := s.Find("c"); r != "a" {
		t.Fatalf("Find(c) = %q, want a", r)
	}
	s.Union("d", "e") // rank(d) = 1
	s.Union("a", "d") // equal ranks, a stays the representative
	if r, _ := s.Find("e"); r != "a" {
		t.Fatalf("Find(e) = %q, want a", r)
	}
}
