package assignment

import (
	"sort"
	"strings"
)

type Rule struct {
	ID       string
	Category string
	UnitID   string
	Priority int
	Active   bool
}

type RuleSet struct{ Rules []Rule }

func (s RuleSet) Resolve(category string) (Rule, bool) {
	candidates := make([]Rule, 0)
	for _, rule := range s.Rules {
		if rule.Active && strings.EqualFold(rule.Category, category) {
			candidates = append(candidates, rule)
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Priority == candidates[j].Priority {
			return candidates[i].ID < candidates[j].ID
		}
		return candidates[i].Priority > candidates[j].Priority
	})
	if len(candidates) == 0 {
		return Rule{}, false
	}
	return candidates[0], true
}
