# picture-this

Draw a [Maestro](https://maestro.dev) view-hierarchy JSON as an SVG picture.
Each UI element's `bounds` is drawn as a distinctly coloured box so you can see
the on-screen layout at a glance.

## Install

`picture-this` is a single self-contained binary — no runtime required.

### Download a prebuilt binary

Grab the archive for your platform from the
[latest release](https://github.com/Fishbowler/picture-this/releases/latest),
unpack it, and put the binary on your `PATH`:

```sh
# macOS / Linux example
tar -xzf picture-this_v0.1.0_darwin_arm64.tar.gz
sudo mv picture-this /usr/local/bin/
```

On Windows, unzip and place `picture-this.exe` somewhere on your `PATH`.
Each release also ships a `checksums.txt` for verification.

### With the Go toolchain

```sh
go install github.com/Fishbowler/picture-this@latest
```

This compiles and installs into `$(go env GOPATH)/bin` (usually `~/go/bin`).

### From source

```sh
go build -o picture-this .
```

Cross-compiling is built in, e.g.:

```sh
GOOS=linux   GOARCH=amd64 go build -o picture-this-linux .
GOOS=windows GOARCH=amd64 go build -o picture-this.exe .
```

## Usage

```sh
picture-this [flags] [file]
```

Reads a Maestro hierarchy JSON from `[file]`, or from stdin when no file is
given, and writes an SVG to stdout (or to `--output`).

```sh
# Render the whole hierarchy to a file
picture-this hierarchy.json -o screen.svg

# Pipe from stdin
maestro hierarchy | picture-this > screen.svg

# Simple mode: only elements that have a text/accessibilityText/hintText/resource-id
picture-this --simple hierarchy.json -o screen.svg

# Boxes only, no labels
picture-this --no-labels hierarchy.json -o screen.svg
```

Open the resulting `.svg` in any browser.

### Flags

| Flag | Description |
| --- | --- |
| `-s`, `--simple` | Only draw elements that have a `text`, `accessibilityText`, `hintText`, or `resource-id`. Structural containers (layouts/views with no identity) are elided. |
| `--no-labels` | Don't draw element labels (labels are drawn by default). |
| `-o`, `--output FILE` | Write the SVG to `FILE` instead of stdout. |

## How it works

The tool recursively walks the hierarchy, parses each node's `bounds`
(`[x1,y1][x2,y2]`), and draws an unfilled coloured rectangle for it. Colours are
spread across the hue wheel by the golden angle so neighbouring boxes contrast.
The canvas is sized to the outermost bounds. Labels (when on) use the first
non-empty of `text`, `resource-id`, `accessibilityText`, `hintText`.
