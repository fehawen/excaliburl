package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/fehawen/excaliburl/internal/cli"
	"github.com/fehawen/excaliburl/internal/encoder"
	"github.com/fehawen/excaliburl/internal/extractor"
	"github.com/fehawen/excaliburl/internal/input"
	"github.com/fehawen/excaliburl/internal/processor"
)

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

	flag.Usage = cli.Usage
	flag.Parse()

	var inputs []input.Input

	seenInputs := make(map[string]struct{})

	addInput := func(in input.Input) {
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

	if input.StdinAvailable() {
		addInput(input.Input{
			Path:  "-",
			Stdin: true,
		})
	}

	files = append(files, flag.Args()...)

	for _, f := range files {
		abs, err := filepath.Abs(f)
		if err != nil {
			cli.Log("cannot resolve path \"%s\"", f)
			continue
		}

		addInput(input.Input{
			Path: abs,
		})
	}

	for _, dir := range dirs {
		found, err := input.WalkDir(dir)
		if err != nil {
			cli.Log("directory walk error \"%s\": %v", dir, err)
			continue
		}

		for _, in := range found {
			addInput(in)
		}
	}

	if len(inputs) == 0 {
		cli.Die(2, "no input provided")
	}

	var out io.Writer = os.Stdout

	if output != "" {
		err := os.MkdirAll(filepath.Dir(output), 0755)
		if err != nil {
			cli.Die(1, "failed to create output directory: %v", err)
		}

		f, err := os.Create(output)
		if err != nil {
			cli.Die(1, "failed to create output file: %v", err)
		}

		defer f.Close()

		out = f
	}

	enc := encoder.New(out)

	e := extractor.New(base, incl, excl, enc)
	e.AddProcessor(processor.NewAbsoluteProcessor())
	e.AddProcessor(processor.NewHTMLProcessor())
	e.AddProcessor(processor.NewJSProcessor())
	e.AddProcessor(processor.NewJSTemplateProcessor())
	e.AddProcessor(processor.NewQuotedPathProcessor())
	e.AddProcessor(processor.NewCSSProcessor())
	e.AddProcessor(processor.NewSourceMapProcessor())

	valid := false

	for _, in := range inputs {
		if in.Stdin {
			valid = true

			if err := extractFromReader(e, os.Stdin, "-"); err != nil {
				cli.Log("stdin error: %v", err)
			}

			continue
		}

		f, err := os.Open(in.Path)
		if err != nil {
			cli.Log("cannot read \"%s\"", in.Path)
			continue
		}

		valid = true

		if err := extractFromReader(e, f, in.Path); err != nil {
			cli.Log("read error %q: %v", in.Path, err)
		}

		f.Close()
	}

	if !valid {
		cli.Die(2, "no readable inputs provided")
	}
}

func extractFromReader(e *extractor.Extractor, r io.Reader, path string) error {
	data, err := input.ReadText(r)
	if err != nil {
		return err
	}

	if data == nil {
		return nil
	}

	e.Process(data, path)
	return nil
}
