package main

import "strings"

// semver is a version split into its parts. The numbers are kept as
// strings, since a valid version has no leading zeros and comparing the
// length first, then the digits, orders them without overflowing.
type semver struct {
	major, minor, patch string
	prerelease          string
}

// compareVersions compares two module versions in semantic version order,
// the order minimal version selection uses. Build metadata such as
// "+incompatible" doesn't count, and an invalid version sorts before every
// valid one. It returns -1, 0 or +1.
func compareVersions(a, b string) int {
	va, okA := parseVersion(a)
	vb, okB := parseVersion(b)
	switch {
	case !okA && !okB:
		return 0
	case !okA:
		return -1
	case !okB:
		return 1
	}

	if c := compareNumbers(va.major, vb.major); c != 0 {
		return c
	}
	if c := compareNumbers(va.minor, vb.minor); c != 0 {
		return c
	}
	if c := compareNumbers(va.patch, vb.patch); c != 0 {
		return c
	}
	return comparePrerelease(va.prerelease, vb.prerelease)
}

// parseVersion splits a version such as v1.2.3-pre.1+build. Like the go
// command, it reads v1 and v1.2 as v1.0.0 and v1.2.0, but only when nothing
// follows them.
func parseVersion(v string) (semver, bool) {
	rest, ok := strings.CutPrefix(v, "v")
	if !ok {
		return semver{}, false
	}

	rest, build, hasBuild := strings.Cut(rest, "+")
	if hasBuild && !validIdentifiers(build, false) {
		return semver{}, false
	}
	core, prerelease, hasPrerelease := strings.Cut(rest, "-")
	if hasPrerelease && !validIdentifiers(prerelease, true) {
		return semver{}, false
	}

	parts := strings.Split(core, ".")
	if len(parts) > 3 || len(parts) < 3 && (hasPrerelease || hasBuild) {
		return semver{}, false
	}
	for _, part := range parts {
		if !isNumber(part) || len(part) > 1 && part[0] == '0' {
			return semver{}, false
		}
	}
	for len(parts) < 3 {
		parts = append(parts, "0")
	}

	return semver{major: parts[0], minor: parts[1], patch: parts[2], prerelease: prerelease}, true
}

// validIdentifiers reports whether s is one or more dot-separated
// identifiers of ASCII letters, digits and hyphens. In a prerelease, a
// numeric identifier can't have a leading zero.
func validIdentifiers(s string, prerelease bool) bool {
	for _, id := range strings.Split(s, ".") {
		if id == "" {
			return false
		}
		for _, c := range id {
			if !isIdentifierChar(c) {
				return false
			}
		}
		if prerelease && isNumber(id) && len(id) > 1 && id[0] == '0' {
			return false
		}
	}
	return true
}

func isIdentifierChar(c rune) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '-'
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// compareNumbers compares two decimal numbers without leading zeros.
func compareNumbers(a, b string) int {
	if len(a) != len(b) {
		if len(a) < len(b) {
			return -1
		}
		return 1
	}
	return strings.Compare(a, b)
}

// comparePrerelease compares two prereleases. A version without one sorts
// after the same version with one. Otherwise the identifiers are compared
// in turn: numbers by value, other identifiers by ASCII, and a number before
// any other identifier. If one list of identifiers starts with the other,
// the shorter one sorts first.
func comparePrerelease(a, b string) int {
	switch {
	case a == b:
		return 0
	case a == "":
		return 1
	case b == "":
		return -1
	}

	idsA, idsB := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(idsA) && i < len(idsB); i++ {
		x, y := idsA[i], idsB[i]
		if x == y {
			continue
		}
		numX, numY := isNumber(x), isNumber(y)
		switch {
		case numX && numY:
			return compareNumbers(x, y)
		case numX:
			return -1
		case numY:
			return 1
		default:
			return strings.Compare(x, y)
		}
	}
	switch {
	case len(idsA) < len(idsB):
		return -1
	case len(idsA) > len(idsB):
		return 1
	}
	return 0
}
