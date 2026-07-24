package processor

import (
	"regexp"

	"github.com/fehawen/excaliburl/internal/result"
)

type QuotedPathProcessor struct {
	re *regexp.Regexp
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

		emit(result.Result{
			File:      file,
			Processor: "quoted-path",
			Raw:       string(m[1]),
		})
	}
}
