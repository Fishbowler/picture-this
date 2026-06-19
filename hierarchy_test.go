package main

import "testing"

func TestParseBounds(t *testing.T) {
	cases := []struct {
		in                     string
		x1, y1, x2, y2         int
		ok                     bool
	}{
		{"[0,36][606,1244]", 0, 36, 606, 1244, true},
		{" [15,51][51,87] ", 15, 51, 51, 87, true}, // surrounding whitespace tolerated
		{"[480,669][565,709]", 480, 669, 565, 709, true},
		{"", 0, 0, 0, 0, false},
		{"0,0,606,1280", 0, 0, 0, 0, false},
		{"[0,0]", 0, 0, 0, 0, false},
		{"[a,b][c,d]", 0, 0, 0, 0, false},
	}
	for _, c := range cases {
		x1, y1, x2, y2, ok := parseBounds(c.in)
		if ok != c.ok {
			t.Errorf("parseBounds(%q) ok = %v, want %v", c.in, ok, c.ok)
			continue
		}
		if ok && (x1 != c.x1 || y1 != c.y1 || x2 != c.x2 || y2 != c.y2) {
			t.Errorf("parseBounds(%q) = (%d,%d,%d,%d), want (%d,%d,%d,%d)",
				c.in, x1, y1, x2, y2, c.x1, c.y1, c.x2, c.y2)
		}
	}
}

// sample tree: a container with no identity, holding a labelled child and a
// nested grandchild. Used to check that simple mode filters per-node and keeps
// traversing through elided parents.
func sampleTree() Node {
	return Node{
		Attributes: map[string]string{"bounds": "[0,0][100,100]"}, // no identity
		Children: []Node{
			{
				Attributes: map[string]string{"bounds": "[0,0][50,50]", "text": "Hello"},
			},
			{
				Attributes: map[string]string{"bounds": "[50,0][100,50]"}, // no identity (elided in simple)
				Children: []Node{
					{Attributes: map[string]string{"bounds": "[60,10][90,40]", "resource-id": "deep"}},
				},
			},
			{
				Attributes: map[string]string{"bounds": "[0,50][100,100]", "text": "\t\t  "}, // whitespace only -> no identity
			},
		},
	}
}

func TestWalkFull(t *testing.T) {
	got := Walk(sampleTree(), false)
	// root + 3 children + 1 grandchild, all of which have bounds.
	if len(got) != 5 {
		t.Fatalf("full walk returned %d elements, want 5", len(got))
	}
}

func TestWalkSimple(t *testing.T) {
	got := Walk(sampleTree(), true)
	if len(got) != 2 {
		t.Fatalf("simple walk returned %d elements, want 2", len(got))
	}
	// The two kept elements are the "Hello" text and the deeply nested
	// resource-id node — proving traversal continued through the elided parent.
	labels := map[string]bool{}
	for _, e := range got {
		labels[label(e.Attrs)] = true
	}
	if !labels["Hello"] || !labels["deep"] {
		t.Errorf("simple walk kept %v, want Hello and deep", labels)
	}
}

func TestHasIdentity(t *testing.T) {
	if hasIdentity(map[string]string{"text": "  ", "resource-id": ""}) {
		t.Error("whitespace-only text should not count as identity")
	}
	if !hasIdentity(map[string]string{"hintText": "Search"}) {
		t.Error("hintText should count as identity")
	}
}
