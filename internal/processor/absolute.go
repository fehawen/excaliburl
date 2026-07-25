package processor

import (
	"regexp"

	"github.com/fehawen/excaliburl/internal/result"
)

type AbsoluteProcessor struct {
	re *regexp.Regexp
}

func NewAbsoluteProcessor() *AbsoluteProcessor {
	return &AbsoluteProcessor{
		re: regexp.MustCompile(
			`([a-zA-Z][a-zA-Z0-9+.-]*://|www\.)[^\s"'<>\\\]\)\},]+`,
		),
	}
}

func (p *AbsoluteProcessor) Name() string {
	return "absolute"
}

func (p *AbsoluteProcessor) Process(data []byte, file string, emit Emitter) {
	matches := p.re.FindAll(data, -1)

	for _, m := range matches {
		emit(result.Result{
			File:      file,
			Processor: "absolute",
			Raw:       string(m),
		})
	}
}
