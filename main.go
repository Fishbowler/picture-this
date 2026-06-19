// Command picture-this draws a Maestro view-hierarchy JSON as an SVG picture,
// with each element's bounds shown as a distinctly coloured box.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "picture-this:", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("picture-this", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), `picture-this — draw a Maestro view hierarchy as an SVG.

Usage:
  picture-this [flags] [file]

Reads a Maestro hierarchy JSON from [file], or from stdin when no file is given,
and writes an SVG drawing of every element's bounds.

Flags:
`)
		fs.PrintDefaults()
	}

	var simple, noLabels bool
	var output string
	fs.BoolVar(&simple, "simple", false, "only draw elements with a text, accessibilityText, hintText, or resource-id")
	fs.BoolVar(&simple, "s", false, "shorthand for --simple")
	fs.BoolVar(&noLabels, "no-labels", false, "do not draw element labels (labels are drawn by default)")
	fs.StringVar(&output, "output", "", "write SVG to this file instead of stdout")
	fs.StringVar(&output, "o", "", "shorthand for --output")
	// Parse flags, allowing them to be interspersed with the positional
	// argument (Go's flag package otherwise stops at the first non-flag).
	var files []string
	rest := args
	for {
		if err := fs.Parse(rest); err != nil {
			return err
		}
		if fs.NArg() == 0 {
			break
		}
		files = append(files, fs.Arg(0))
		rest = fs.Args()[1:]
	}

	// Choose input: positional file argument, or stdin.
	var data []byte
	var err error
	switch len(files) {
	case 0:
		data, err = io.ReadAll(stdin)
	case 1:
		data, err = os.ReadFile(files[0])
	default:
		return fmt.Errorf("expected at most one input file, got %d", len(files))
	}
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	var root Node
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	elements := Walk(root, simple)
	if len(elements) == 0 {
		return fmt.Errorf("no elements with bounds found in input")
	}

	// Choose output: file, or stdout.
	var out io.Writer = stdout
	if output != "" {
		f, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		out = f
	}

	if err := renderSVG(out, elements, !noLabels); err != nil {
		return fmt.Errorf("rendering SVG: %w", err)
	}
	return nil
}
