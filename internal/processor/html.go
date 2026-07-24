package processor

import (
	"regexp"

	"github.com/fehawen/excaliburl/internal/result"
)

type HTMLProcessor struct {
	re *regexp.Regexp
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

		emit(result.Result{
			File:      file,
			Processor: "html",
			Raw:       string(m[2]),
		})
	}
}
