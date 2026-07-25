package processor

import (
	"regexp"
	"strings"

	"github.com/fehawen/excaliburl/internal/result"
)

type SourceMapProcessor struct {
	re *regexp.Regexp
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
		emit(result.Result{
			File:      file,
			Processor: "sourcemap",
			Raw:       cleanSourceMap(string(m[1])),
		})
	}
}

func cleanSourceMap(s string) string {
	if i := strings.Index(s, `\n`); i >= 0 {
		s = s[:i]
	}

	if i := strings.Index(s, `\"`); i >= 0 {
		s = s[:i]
	}

	return s
}
