package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Emitter func(Result)

type Processor interface {
	Name() string
	Process(data []byte, file string, emit Emitter)
}

type Input struct {
	Path  string
	Stdin bool
}

type Result struct {
	File       string `json:"file,omitempty"`
	Processor  string `json:"processor,omitempty"`
	Raw        string `json:"raw"`
	Normalized string `json:"normalized,omitempty"`
}

type Extractor struct {
	processors []Processor

	include []*regexp.Regexp
	exclude []*regexp.Regexp

	seen map[string]struct{}

	base *url.URL

	writer io.Writer
}

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

func usage() {
	fmt.Fprintf(os.Stderr, "%s\n", docs)
}

func log(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[excaliburl] "+format+"\n", args...)
}

func hint() {
	log("try 'excaliburl -h' for more information")
}

func die(code int, format string, args ...any) {
	log(format, args...)

	if code == 2 {
		hint()
	}

	os.Exit(code)
}

func trimURL(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'`)
	s = strings.TrimRight(s, `\`)

	return s
}

func isText(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	sample := data

	if len(sample) > 8192 {
		sample = sample[:8192]
	}

	var control int

	for _, b := range sample {
		if b == 0 {
			return false
		}

		if b < 32 &&
			b != '\n' &&
			b != '\r' &&
			b != '\t' {
			control++
		}
	}

	ratio := float64(control) / float64(len(sample))

	return ratio < 0.10
}

func compilePatterns(patterns []string) []*regexp.Regexp {
	var out []*regexp.Regexp

	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			die(2, "invalid pattern '%s': %v", p, err)
		}

		out = append(out, re)
	}

	return out
}

func matchesAny(s string, patterns []*regexp.Regexp) bool {
	for _, re := range patterns {
		if re.MatchString(s) {
			return true
		}
	}

	return false
}

func stdinAvailable() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return (stat.Mode() & os.ModeCharDevice) == 0
}

func walkDir(root string) ([]Input, error) {
	var inputs []Input

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			return nil
		}

		inputs = append(inputs, Input{
			Path: abs,
		})

		return nil
	})

	return inputs, err
}

func NewExtractor(base string, incl []string, excl []string, w io.Writer) *Extractor {
	e := &Extractor{
		seen:    make(map[string]struct{}),
		include: compilePatterns(incl),
		exclude: compilePatterns(excl),
		writer:  w,
	}

	if base != "" {
		u, err := url.Parse(base)
		if err != nil {
			die(2, "invalid base URL: %v", err)
		}

		if u.Scheme == "" || u.Host == "" {
			die(2, "base URL must include scheme and host")
		}

		e.base = u
	}

	return e
}

func (e *Extractor) AddProcessor(p Processor) {
	e.processors = append(e.processors, p)
}

func (e *Extractor) Process(data []byte, file string) {
	for _, p := range e.processors {
		p.Process(data, file, e.emit)
	}
}

func (e *Extractor) emit(r Result) {
	r.Raw = trimURL(r.Raw)

	if len(r.Raw) <= 1 {
		return
	}

	normalized := e.normalize(r.Raw)

	value := r.Raw

	if normalized != r.Raw {
		r.Normalized = normalized
		value = normalized
	}

	if len(e.include) > 0 && !matchesAny(value, e.include) {
		return
	}

	if len(e.exclude) > 0 && matchesAny(value, e.exclude) {
		return
	}

	key := r.File + "\x00" + value

	if _, ok := e.seen[key]; ok {
		return
	}

	e.seen[key] = struct{}{}

	out, err := json.Marshal(r)
	if err != nil {
		return
	}

	fmt.Fprintln(e.writer, string(out))
}

func (e *Extractor) normalize(raw string) string {
	if e.base == nil {
		return raw
	}

	ref, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	if ref.IsAbs() {
		return ref.String()
	}

	return e.base.ResolveReference(ref).String()
}

type AbsoluteProcessor struct {
	re *regexp.Regexp
}

func (p *AbsoluteProcessor) Name() string {
	return "absolute"
}

func NewAbsoluteProcessor() *AbsoluteProcessor {
	return &AbsoluteProcessor{
		re: regexp.MustCompile(
			`((([a-zA-Z]{1,10})|s3):\/\/|www\.)[^\s"'<>\\\]\)\},]+`,
		),
	}
}

func (p *AbsoluteProcessor) Process(data []byte, file string, emit Emitter) {
	matches := p.re.FindAll(data, -1)

	for _, m := range matches {
		emit(Result{
			File:      file,
			Processor: p.Name(),
			Raw:       string(m),
		})
	}
}

type HTMLProcessor struct {
	re *regexp.Regexp
}

func (p *HTMLProcessor) Name() string {
	return "html"
}

func NewHTMLProcessor() *HTMLProcessor {
	return &HTMLProcessor{
		re: regexp.MustCompile(
			`(?i)(href|src|action|data|poster)\s*=\s*["'](\/[^"' \t\r\n>]*|\.\.?\/[^"' \t\r\n>]*)["']`,
		),
	}
}

func (p *HTMLProcessor) Process(data []byte, file string, emit Emitter) {
	matches := p.re.FindAllSubmatch(data, -1)

	for _, m := range matches {
		if len(m) < 3 {
			continue
		}

		emit(Result{
			File:      file,
			Processor: p.Name(),
			Raw:       string(m[2]),
		})
	}
}

type JSProcessor struct {
	re *regexp.Regexp
}

func (p *JSProcessor) Name() string {
	return "js"
}

func (p *JSProcessor) Process(data []byte, file string, emit Emitter) {
	matches := p.re.FindAllSubmatch(data, -1)

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}

		emit(Result{
			File:      file,
			Processor: p.Name(),
			Raw:       string(m[1]),
		})
	}
}

func NewJSProcessor() *JSProcessor {
	return &JSProcessor{
		re: regexp.MustCompile(
			`["']((?:\/|\.\.?\/)[^"'\\\s]+)["']`,
		),
	}
}

type JSTemplateProcessor struct {
	staticRe  *regexp.Regexp
	dynamicRe *regexp.Regexp
}

func (p *JSTemplateProcessor) Name() string {
	return "js-template"
}

func NewJSTemplateProcessor() *JSTemplateProcessor {
	return &JSTemplateProcessor{
		staticRe: regexp.MustCompile(
			"`((?:\\/|\\.\\.?/)[^`$]*)`",
		),
		dynamicRe: regexp.MustCompile(
			"`((?:\\/|\\.\\.?/)[^`]*)\\$\\{",
		),
	}
}

func (p *JSTemplateProcessor) Process(data []byte, file string, emit Emitter) {
	staticMatches := p.staticRe.FindAllSubmatch(data, -1)

	for _, m := range staticMatches {
		if len(m) < 2 {
			continue
		}

		emit(Result{
			File:      file,
			Processor: p.Name(),
			Raw:       string(m[1]),
		})
	}

	dynamicMatches := p.dynamicRe.FindAllSubmatch(data, -1)

	for _, m := range dynamicMatches {
		if len(m) < 2 {
			continue
		}

		emit(Result{
			File:      file,
			Processor: p.Name(),
			Raw:       string(m[1]),
		})
	}
}

type QuotedPathProcessor struct {
	re *regexp.Regexp
}

func (p *QuotedPathProcessor) Name() string {
	return "quoted-path"
}

func NewQuotedPathProcessor() *QuotedPathProcessor {
	return &QuotedPathProcessor{
		re: regexp.MustCompile(
			`["']((?:\/|\.\.?\/)[a-zA-Z0-9_./?=&%:#-]+)["']`,
		),
	}
}

func (p *QuotedPathProcessor) Process(data []byte, file string, emit Emitter) {
	matches := p.re.FindAllSubmatch(data, -1)

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}

		emit(Result{
			File:      file,
			Processor: p.Name(),
			Raw:       string(m[1]),
		})
	}
}

type SourceMapProcessor struct {
	re *regexp.Regexp
}

func (p *SourceMapProcessor) Name() string {
	return "sourcemap"
}

func NewSourceMapProcessor() *SourceMapProcessor {
	return &SourceMapProcessor{
		re: regexp.MustCompile(
			`sourceMappingURL=([^\s*]+)`,
		),
	}
}

func (p *SourceMapProcessor) Process(data []byte, file string, emit Emitter) {
	matches := p.re.FindAllSubmatch(data, -1)

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}

		emit(Result{
			File:      file,
			Processor: p.Name(),
			Raw:       string(m[1]),
		})
	}
}

type CSSProcessor struct {
	re *regexp.Regexp
}

func (p *CSSProcessor) Name() string {
	return "css"
}

func NewCSSProcessor() *CSSProcessor {
	return &CSSProcessor{
		re: regexp.MustCompile(
			`url\(\s*["']?((?:\/|\.\.?\/)[^)"'\s]+)["']?\s*\)`,
		),
	}
}

func (p *CSSProcessor) Process(data []byte, file string, emit Emitter) {
	matches := p.re.FindAllSubmatch(data, -1)

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}

		emit(Result{
			File:      file,
			Processor: p.Name(),
			Raw:       string(m[1]),
		})
	}
}

func processReader(e *Extractor, r io.Reader, path string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if !isText(data) {
		return nil
	}

	e.Process(data, path)

	return nil
}

func main() {
	var (
		base   string
		output string
		files  []string
		dirs   []string
		incl   []string
		excl   []string
	)

	flag.StringVar(&base, "b", "", "")
	flag.StringVar(&output, "o", "", "")

	flag.Func("f", "", func(s string) error {
		files = append(files, s)
		return nil
	})

	flag.Func("R", "", func(s string) error {
		dirs = append(dirs, s)
		return nil
	})

	flag.Func("i", "", func(s string) error {
		incl = append(incl, s)
		return nil
	})

	flag.Func("e", "", func(s string) error {
		excl = append(excl, s)
		return nil
	})

	flag.Usage = usage
	flag.Parse()

	var inputs []Input

	seenInputs := make(map[string]struct{})

	addInput := func(in Input) {
		key := in.Path

		if in.Stdin {
			key = "__stdin__"
		}

		if _, ok := seenInputs[key]; ok {
			return
		}

		seenInputs[key] = struct{}{}
		inputs = append(inputs, in)
	}

	if stdinAvailable() {
		addInput(Input{
			Path:  "-",
			Stdin: true,
		})
	}

	files = append(files, flag.Args()...)

	for _, f := range files {
		abs, err := filepath.Abs(f)
		if err != nil {
			log("cannot resolve path \"%s\"", f)
			continue
		}

		addInput(Input{
			Path: abs,
		})
	}

	for _, dir := range dirs {
		found, err := walkDir(dir)
		if err != nil {
			log("directory walk error \"%s\": %v", dir, err)
			continue
		}

		for _, in := range found {
			addInput(in)
		}
	}

	if len(inputs) == 0 {
		die(2, "no input provided")
	}

	var out io.Writer = os.Stdout

	if output != "" {
		err := os.MkdirAll(filepath.Dir(output), 0755)
		if err != nil {
			die(1, "failed to create output directory: %v", err)
		}

		f, err := os.Create(output)
		if err != nil {
			die(1, "failed to create output file: %v", err)
		}

		defer f.Close()

		out = f
	}

	e := NewExtractor(base, incl, excl, out)
	e.AddProcessor(NewAbsoluteProcessor())
	e.AddProcessor(NewHTMLProcessor())
	e.AddProcessor(NewJSProcessor())
	e.AddProcessor(NewJSTemplateProcessor())
	e.AddProcessor(NewQuotedPathProcessor())
	e.AddProcessor(NewCSSProcessor())
	e.AddProcessor(NewSourceMapProcessor())

	valid := false

	for _, in := range inputs {
		if in.Stdin {
			valid = true

			if err := processReader(e, os.Stdin, in.Path); err != nil {
				log("stdin error: %v", err)
			}

			continue
		}

		f, err := os.Open(in.Path)
		if err != nil {
			log("cannot read \"%s\"", in.Path)
			continue
		}

		valid = true

		if err := processReader(e, f, in.Path); err != nil {
			log("read error \"%s\": %v", in.Path, err)
		}

		f.Close()
	}

	if !valid {
		die(2, "no readable inputs provided")
	}
}
