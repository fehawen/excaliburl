package extractor

import (
	"regexp"

	"github.com/fehawen/excaliburl/internal/cli"
)

func compilePatterns(patterns []string) []*regexp.Regexp {
	var out []*regexp.Regexp

	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			cli.Die(2, "invalid pattern '%s': %v", p, err)
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
