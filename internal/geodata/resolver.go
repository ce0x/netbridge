package geodata

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Resolver struct {
	domains map[string][]string
	cidrs   map[string][]string
}

func NewResolver() *Resolver {
	return &Resolver{
		domains: make(map[string][]string),
		cidrs:   make(map[string][]string),
	}
}

func (r *Resolver) LoadDomains(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ",", 2)
		if len(parts) == 2 {
			category := strings.TrimSpace(parts[0])
			domain := strings.TrimSpace(parts[1])
			r.domains[category] = append(r.domains[category], domain)
		}
	}

	return scanner.Err()
}

func (r *Resolver) LoadCIDRs(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ",", 2)
		if len(parts) == 2 {
			category := strings.TrimSpace(parts[0])
			cidr := strings.TrimSpace(parts[1])
			r.cidrs[category] = append(r.cidrs[category], cidr)
		}
	}

	return scanner.Err()
}

func (r *Resolver) LookupDomains(category string) ([]string, error) {
	domains, ok := r.domains[category]
	if !ok {
		return nil, fmt.Errorf("category %q not found", category)
	}
	return domains, nil
}

func (r *Resolver) LookupCIDRs(category string) ([]string, error) {
	cidrs, ok := r.cidrs[category]
	if !ok {
		return nil, fmt.Errorf("category %q not found", category)
	}
	return cidrs, nil
}

func (r *Resolver) Categories() []string {
	seen := make(map[string]bool)
	var cats []string

	for cat := range r.domains {
		if !seen[cat] {
			seen[cat] = true
			cats = append(cats, cat)
		}
	}
	for cat := range r.cidrs {
		if !seen[cat] {
			seen[cat] = true
			cats = append(cats, cat)
		}
	}

	return cats
}