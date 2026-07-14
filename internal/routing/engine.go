package routing

import (
	"context"
	"fmt"
	"strings"

	netbridge "github.com/netbridge/netbridge"
)

type Engine struct {
	rules   map[string]*netbridge.RouteRule
	counter int
}

func NewEngine() *Engine {
	return &Engine{
		rules: make(map[string]*netbridge.RouteRule),
	}
}

func (e *Engine) AddRule(ctx context.Context, rule netbridge.RouteRule) error {
	if rule.ID == "" {
		rule.ID = e.generateRuleID()
	}
	rule.Enabled = true
	e.rules[rule.ID] = &rule
	return nil
}

func (e *Engine) RemoveRule(ctx context.Context, id string) error {
	if _, ok := e.rules[id]; !ok {
		return fmt.Errorf("rule not found: %s", id)
	}
	delete(e.rules, id)
	return nil
}

func (e *Engine) ListRules(ctx context.Context) ([]*netbridge.RouteRule, error) {
	list := make([]*netbridge.RouteRule, 0, len(e.rules))
	for _, r := range e.rules {
		list = append(list, r)
	}
	return list, nil
}

func (e *Engine) ClearRules(ctx context.Context) error {
	e.rules = make(map[string]*netbridge.RouteRule)
	return nil
}

func (e *Engine) Resolve(destination string) (string, error) {
	destination = strings.ToLower(destination)

	var bestMatch *netbridge.RouteRule
	bestPriority := -1

	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}
		if matchRule(destination, rule) && rule.Priority > bestPriority {
			bestMatch = rule
			bestPriority = rule.Priority
		}
	}

	if bestMatch != nil {
		return bestMatch.ProfileID, nil
	}
	return "", nil
}

func (e *Engine) Apply(ctx context.Context) error {
	return nil
}

func matchRule(dest string, rule *netbridge.RouteRule) bool {
	switch rule.RuleType {
	case "domain":
		return dest == strings.ToLower(rule.Pattern)
	case "domain_suffix":
		return strings.HasSuffix(dest, "."+strings.ToLower(rule.Pattern)) || dest == strings.ToLower(rule.Pattern)
	case "keyword":
		return strings.Contains(dest, strings.ToLower(rule.Pattern))
	case "ip_cidr":
		return false
	case "geoip":
		return false
	default:
		return false
	}
}

func (e *Engine) generateRuleID() string {
	e.counter++
	return fmt.Sprintf("rule-%d", e.counter)
}
