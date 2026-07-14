package security

import (
	"regexp"
	"strings"
)

var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[=:]\s*\S+`),
	regexp.MustCompile(`(?i)(secret|token|key)\s*[=:]\s*\S+`),
	regexp.MustCompile(`(?i)(private_key)\s*[=:]\s*\S+`),
}

type Masker struct{}

func (m *Masker) MaskString(s string) string {
	for _, pattern := range sensitivePatterns {
		s = pattern.ReplaceAllStringFunc(s, func(match string) string {
			eqIdx := strings.IndexAny(match, "=:")
			if eqIdx < 0 {
				return match
			}
			return match[:eqIdx+1] + " ***MASKED***"
		})
	}
	return s
}

func (m *Masker) MaskProfile(p map[string]any) map[string]any {
	sensitive := []string{"password", "secret", "token", "private_key", "uuid", "aid"}
	for _, key := range sensitive {
		if val, ok := p[key]; ok {
			if _, ok := val.(string); ok {
				p[key] = "***MASKED***"
			}
		}
	}
	return p
}
