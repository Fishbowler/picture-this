package main

import (
	"regexp"
	"strconv"
	"strings"
)

// Node mirrors the shape of a Maestro view-hierarchy JSON node. Each node has a
// bag of string attributes and a list of children. Other sibling keys that
// Maestro emits (clickable, enabled, ...) are ignored — everything we need
// lives under "attributes".
type Node struct {
	Attributes map[string]string `json:"attributes"`
	Children   []Node            `json:"children"`
}

// Element is a single drawable rectangle extracted from a node that had
// parseable bounds, together with that node's attributes (used for labelling).
type Element struct {
	X1, Y1, X2, Y2 int
	Attrs          map[string]string
}

// Width and Height of the element's bounding box.
func (e Element) Width() int  { return e.X2 - e.X1 }
func (e Element) Height() int { return e.Y2 - e.Y1 }

// boundsRe matches the Maestro bounds format, e.g. "[0,36][606,1244]".
var boundsRe = regexp.MustCompile(`^\[(-?\d+),(-?\d+)\]\[(-?\d+),(-?\d+)\]$`)

// parseBounds extracts the two corners from a bounds string. ok is false when
// the string is empty or does not match the expected format.
func parseBounds(s string) (x1, y1, x2, y2 int, ok bool) {
	m := boundsRe.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, 0, 0, 0, false
	}
	// All four groups are guaranteed numeric by the regex.
	x1, _ = strconv.Atoi(m[1])
	y1, _ = strconv.Atoi(m[2])
	x2, _ = strconv.Atoi(m[3])
	y2, _ = strconv.Atoi(m[4])
	return x1, y1, x2, y2, true
}

// identityKeys are the attributes that give an element a user-facing identity.
// In simple mode, an element is elided when all of these are empty.
var identityKeys = []string{"text", "accessibilityText", "hintText", "resource-id"}

// hasIdentity reports whether any identity attribute is non-empty (after
// trimming). Note: some text is purely whitespace/tabs (e.g. "\t\t  ") — that
// counts as no identity.
func hasIdentity(attrs map[string]string) bool {
	for _, k := range identityKeys {
		if strings.TrimSpace(attrs[k]) != "" {
			return true
		}
	}
	return false
}

// Walk traverses the whole tree and returns every node that has parseable
// bounds, in pre-order (parents before children). When simple is true, nodes
// without an identity are skipped — but their children are still visited, so
// filtering is per-node rather than pruning whole subtrees.
func Walk(root Node, simple bool) []Element {
	var out []Element
	var visit func(n Node)
	visit = func(n Node) {
		if x1, y1, x2, y2, ok := parseBounds(n.Attributes["bounds"]); ok {
			if !simple || hasIdentity(n.Attributes) {
				out = append(out, Element{X1: x1, Y1: y1, X2: x2, Y2: y2, Attrs: n.Attributes})
			}
		}
		for _, c := range n.Children {
			visit(c)
		}
	}
	visit(root)
	return out
}

// label returns the most descriptive label for an element: the first non-empty
// of text, resource-id, accessibilityText, hintText (trimmed). Empty if none.
func label(attrs map[string]string) string {
	for _, k := range []string{"text", "resource-id", "accessibilityText", "hintText"} {
		if v := strings.TrimSpace(attrs[k]); v != "" {
			return v
		}
	}
	return ""
}
