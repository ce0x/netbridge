package routing

import (
	netbridge "github.com/netbridge/netbridge"
)

type Compiler struct{}

func (c *Compiler) Compile(rules []*netbridge.RouteRule) map[string]any {
	domains := []string{}
	domainSuffixes := []string{}
	ipCIDRs := []string{}
	keywords := []string{}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		switch rule.RuleType {
		case "domain":
			domains = append(domains, rule.Pattern)
		case "domain_suffix":
			domainSuffixes = append(domainSuffixes, rule.Pattern)
		case "ip_cidr":
			ipCIDRs = append(ipCIDRs, rule.Pattern)
		case "keyword":
			keywords = append(keywords, rule.Pattern)
		}
	}

	return map[string]any{
		"domains":       domains,
		"domain_suffix": domainSuffixes,
		"ip_cidr":       ipCIDRs,
		"keywords":      keywords,
	}
}

func (c *Compiler) CompileXray(rules []*netbridge.RouteRule) []map[string]any {
	routingRules := []map[string]any{}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		var xrayRule map[string]any
		switch rule.RuleType {
		case "domain":
			xrayRule = map[string]any{
				"type":        "field",
				"domain":      []string{rule.Pattern},
				"outboundTag": rule.ProfileID,
			}
		case "domain_suffix":
			xrayRule = map[string]any{
				"type":        "field",
				"domain":      []string{"domain:" + rule.Pattern},
				"outboundTag": rule.ProfileID,
			}
		case "ip_cidr":
			xrayRule = map[string]any{
				"type":        "field",
				"ip":          []string{rule.Pattern},
				"outboundTag": rule.ProfileID,
			}
		}
		if xrayRule != nil {
			routingRules = append(routingRules, xrayRule)
		}
	}
	return routingRules
}
