package cli

import (
	"fmt"
	"os"
)

var docs = `
excaliburl - extract URLs and paths from textual data

Usage:
  excaliburl [options / files]

Description:
  excaliburl heuristically extracts URLs, paths, and related artifacts
  from text files and textual data sources.

  Binary files are ignored, only text-based files are processed.

  Input may be provided through standard input, explicit files via -f,
  recursive directory traversal via -R, or positional file arguments.

  Relative paths may be resolved against a base URL using -b.

Output:
  Results are written as JSON Lines (JSONL) to stdout by default.
  Use -o to write to a file instead.

Options:
  -f <file>       Read input from file (repeatable).

  -R <directory>  Recursively traverse directory and process all
                  discovered files (repeatable).

  -i <pattern>    Include only results matching pattern (repeatable).

  -e <pattern>    Exclude results matching pattern (repeatable).

  -o <file>       Write output to file instead of stdout.
                  Parent directories are created if needed.

  -b <url>        Resolve relative paths against url.

  -h, --help      Show this help and exit.
`

func Usage() {
	fmt.Fprintf(os.Stderr, "%s\n", docs)
}
