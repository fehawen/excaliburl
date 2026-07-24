package extractor

import (
	"net/url"
	"strings"
)

func trimURL(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'`)
	s = strings.TrimRight(s, `\`)

	return s
}

func (e *Extractor) normalize(raw string) string {
	if e.base == nil {
		return raw
	}

	ref, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	if ref.IsAbs() {
		return ref.String()
	}

	return e.base.ResolveReference(ref).String()
}
