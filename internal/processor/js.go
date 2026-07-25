package processor

import (
	"regexp"

	"github.com/fehawen/excaliburl/internal/result"
)

type JSProcessor struct {
	re *regexp.Regexp
}

func NewJSProcessor() *JSProcessor {
	return &JSProcessor{
		re: regexp.MustCompile(
			`["']((?:\/|\.\.?\/)[^"'\\\s]+)["']`,
		),
	}
}

func (p *JSProcessor) Process(data []byte, file string, emit Emitter) {
	matches := p.re.FindAllSubmatch(data, -1)

	for _, m := range matches {
		emit(result.Result{
			File:      file,
			Processor: "js",
			Raw:       string(m[1]),
		})
	}
}
