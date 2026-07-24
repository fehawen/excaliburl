package processor

import (
	"regexp"

	"github.com/fehawen/excaliburl/internal/result"
)

type CSSProcessor struct {
	re *regexp.Regexp
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

		emit(result.Result{
			File:      file,
			Processor: "css",
			Raw:       string(m[1]),
		})
	}
}
