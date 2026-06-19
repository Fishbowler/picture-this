package main

import (
	"fmt"
	"io"
	"math"
	"strings"
)

// renderSVG writes an SVG drawing of the given elements to w. Each element is
// drawn as an unfilled coloured rectangle ("lines where each element is"); when
// labels is true, a small label is drawn at each box's top-left corner.
func renderSVG(w io.Writer, elements []Element, labels bool) error {
	width, height := canvasSize(elements)

	var b strings.Builder
	fmt.Fprintf(&b,
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" font-family="sans-serif">`+"\n",
		width, height, width, height)
	// White background for legibility against the coloured strokes.
	fmt.Fprintf(&b, `  <rect x="0" y="0" width="%d" height="%d" fill="white"/>`+"\n", width, height)

	for i, e := range elements {
		color := distinctColor(i)
		fmt.Fprintf(&b,
			`  <rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1.5"/>`+"\n",
			e.X1, e.Y1, e.Width(), e.Height(), color)

		if labels {
			if text := label(e.Attrs); text != "" {
				// Anchor the label just inside the top-left corner of the box.
				fmt.Fprintf(&b,
					`  <text x="%d" y="%d" font-size="8" fill="%s">%s</text>`+"\n",
					e.X1+2, e.Y1+9, color, escapeXML(truncate(text, 40)))
			}
		}
	}

	b.WriteString("</svg>\n")
	_, err := io.WriteString(w, b.String())
	return err
}

// canvasSize returns the dimensions needed to contain every element. It uses
// the maximum right/bottom edge so off-by-one rows (e.g. ...[606,1245]) still
// fit. Falls back to a 1x1 canvas when there are no elements.
func canvasSize(elements []Element) (int, int) {
	w, h := 0, 0
	for _, e := range elements {
		if e.X2 > w {
			w = e.X2
		}
		if e.Y2 > h {
			h = e.Y2
		}
	}
	if w == 0 {
		w = 1
	}
	if h == 0 {
		h = 1
	}
	return w, h
}

// distinctColor returns a visually distinct colour for index i by rotating the
// hue with the golden angle, so adjacent elements contrast strongly.
func distinctColor(i int) string {
	hue := math.Mod(float64(i)*137.508, 360)
	r, g, b := hslToRGB(hue, 0.70, 0.45)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// hslToRGB converts an HSL colour (h in degrees, s and l in [0,1]) to 8-bit RGB.
func hslToRGB(h, s, l float64) (int, int, int) {
	c := (1 - math.Abs(2*l-1)) * s
	hp := h / 60
	x := c * (1 - math.Abs(math.Mod(hp, 2)-1))
	var r, g, b float64
	switch {
	case hp < 1:
		r, g, b = c, x, 0
	case hp < 2:
		r, g, b = x, c, 0
	case hp < 3:
		r, g, b = 0, c, x
	case hp < 4:
		r, g, b = 0, x, c
	case hp < 5:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	m := l - c/2
	return int(math.Round((r + m) * 255)),
		int(math.Round((g + m) * 255)),
		int(math.Round((b + m) * 255))
}

// truncate shortens s to at most n runes, appending an ellipsis when cut.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// escapeXML escapes the characters that are unsafe in SVG text content.
func escapeXML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
		"\t", " ",
		"\n", " ",
	)
	return replacer.Replace(s)
}
