package processor

import (
	"regexp"
	"strings"

	"github.com/fehawen/excaliburl/internal/result"
)

type JSTemplateProcessor struct {
	re *regexp.Regexp
}

func NewJSTemplateProcessor() *JSTemplateProcessor {
	return &JSTemplateProcessor{
		re: regexp.MustCompile(
			"`((?:\\/|\\.\\.?/)[^`]*)`",
		),
	}
}

func (p *JSTemplateProcessor) Process(data []byte, file string, emit Emitter) {
	matches := p.re.FindAllSubmatch(data, -1)

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}

		emit(result.Result{
			File:      file,
			Processor: "js-template",
			Raw:       trimTemplate(string(m[1])),
		})
	}
}

func trimTemplate(raw string) string {
	if i := strings.Index(raw, "${"); i >= 0 {
		if i > 0 {
			raw = raw[:i]
		}
	}

	return raw
}
