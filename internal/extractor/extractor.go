package extractor

import (
	"net/url"
	"regexp"

	"github.com/fehawen/excaliburl/internal/cli"
	"github.com/fehawen/excaliburl/internal/processor"
	"github.com/fehawen/excaliburl/internal/result"
)

type Encoder interface {
	Write(result.Result) error
}

type Extractor struct {
	processors []processor.Processor

	include []*regexp.Regexp
	exclude []*regexp.Regexp

	seen map[string]struct{}

	base *url.URL

	encoder Encoder
}

func New(base string, incl []string, excl []string, enc Encoder) *Extractor {
	e := &Extractor{
		seen:    make(map[string]struct{}),
		include: compilePatterns(incl),
		exclude: compilePatterns(excl),
		encoder: enc,
	}

	if base != "" {
		u, err := url.Parse(base)
		if err != nil {
			cli.Die(2, "invalid base URL: %v", err)
		}

		if u.Scheme == "" || u.Host == "" {
			cli.Die(2, "base URL must include scheme and host")
		}

		e.base = u
	}

	return e
}

func (e *Extractor) AddProcessor(p processor.Processor) {
	e.processors = append(e.processors, p)
}

func (e *Extractor) Process(data []byte, file string) {
	for _, p := range e.processors {
		p.Process(data, file, e.emit)
	}
}

func (e *Extractor) emit(r result.Result) {
	r.Raw = trimURL(r.Raw)

	if r.Raw == "" {
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

	if err := e.encoder.Write(r); err != nil {
		return
	}
}
