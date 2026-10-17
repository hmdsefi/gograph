// Package unionfind keeps labels in disjoint groups and merges groups as pairs
// arrive, for example two addresses that shared an input or two records that
// describe the same entity. It answers "are these in the same group" in nearly
// constant time, without building a graph first.
//
// The package does not know about graphs. A caller with a graph passes the
// labels of its vertices.
package unionfind

// Sets is a collection of disjoint groups of labels, kept with union by rank
// and path compression. The zero value is not ready to use, call New.
//
// Sets is not safe for concurrent use.
type Sets[T comparable] struct {
	index  map[T]int
	labels []T
	parent []int
	rank   []uint8
}

// New returns an empty collection of groups.
func New[T comparable]() *Sets[T] {
	return &Sets[T]{index: make(map[T]int)}
}

// Add inserts x as a group of its own. Adding a label that is already in a
// group does nothing. It runs in O(1) amortized time.
func (s *Sets[T]) Add(x T) {
	s.add(x)
}

func (s *Sets[T]) add(x T) int {
	if i, ok := s.index[x]; ok {
		return i
	}
	i := len(s.labels)
	s.index[x] = i
	s.labels = append(s.labels, x)
	s.parent = append(s.parent, i)
	s.rank = append(s.rank, 0)
	return i
}

// root returns the index of the representative of i, compressing the path
// on the way.
func (s *Sets[T]) root(i int) int {
	r := i
	for s.parent[r] != r {
		r = s.parent[r]
	}
	for s.parent[i] != r {
		next := s.parent[i]
		s.parent[i] = r
		i = next
	}
	return r
}

// Union merges the groups of a and b. A label that is not in a group yet is
// added first, so Union(x, x) adds x as a group of one.
//
// When both groups have the same rank, the representative of a's group
// stays the representative, so the same sequence of calls always gives the
// same representatives. It runs in nearly O(1) amortized time.
func (s *Sets[T]) Union(a, b T) {
	ra := s.root(s.add(a))
	rb := s.root(s.add(b))
	if ra == rb {
		return
	}
	switch {
	case s.rank[ra] < s.rank[rb]:
		s.parent[ra] = rb
	case s.rank[ra] > s.rank[rb]:
		s.parent[rb] = ra
	default:
		s.parent[rb] = ra
		s.rank[ra]++
	}
}

// Find returns the representative of the group that x is in. The second
// result is false, and the first is the zero value, if x was never added or
// united. It runs in nearly O(1) amortized time.
func (s *Sets[T]) Find(x T) (T, bool) {
	i, ok := s.index[x]
	if !ok {
		var zero T
		return zero, false
	}
	return s.labels[s.root(i)], true
}

// Connected reports whether a and b are in the same group. A label that was
// never seen is in no group, so Connected returns false for it, even when a
// and b are the same label. It runs in nearly O(1) amortized time.
func (s *Sets[T]) Connected(a, b T) bool {
	ia, ok := s.index[a]
	if !ok {
		return false
	}
	ib, ok := s.index[b]
	if !ok {
		return false
	}
	return s.root(ia) == s.root(ib)
}

// Groups returns the members of each group. Members are in the order they
// were first added, and groups are ordered by their first member. An empty
// collection returns an empty slice. It runs in O(n) time for n labels.
func (s *Sets[T]) Groups() [][]T {
	groups := [][]T{}
	position := make(map[int]int)
	for i, label := range s.labels {
		r := s.root(i)
		p, ok := position[r]
		if !ok {
			p = len(groups)
			position[r] = p
			groups = append(groups, nil)
		}
		groups[p] = append(groups[p], label)
	}
	return groups
}
