package routing

import (
	"fmt"
	"net"
	"strings"
)

type Matcher struct{}

func (m *Matcher) MatchDomain(dest, pattern string) bool {
	dest = strings.ToLower(dest)
	pattern = strings.ToLower(pattern)

	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:]
		return strings.HasSuffix(dest, suffix) || dest == strings.TrimPrefix(pattern, "*.")
	}
	return dest == pattern
}

func (m *Matcher) MatchCIDR(dest, cidr string) bool {
	ip := net.ParseIP(dest)
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return network.Contains(ip)
}

func (m *Matcher) MatchKeyword(dest, keyword string) bool {
	return strings.Contains(strings.ToLower(dest), strings.ToLower(keyword))
}

func (m *Matcher) Match(dest string, ruleType, pattern string) bool {
	switch ruleType {
	case "domain":
		return m.MatchDomain(dest, pattern)
	case "domain_suffix":
		return strings.HasSuffix(strings.ToLower(dest), "."+strings.ToLower(pattern))
	case "ip_cidr":
		return m.MatchCIDR(dest, pattern)
	case "keyword":
		return m.MatchKeyword(dest, pattern)
	default:
		return false
	}
}

var _ = fmt.Sprintf
