package target

import (
	"slices"
	"strings"
)

type element struct {
	ID      string   `json:"id"`
	Targets []Target `json:"targets,omitempty"`
}

type Router []element

func NewRouter(targets []Target) Router {
	m := make(map[string][]Target)
	for _, target := range targets {
		m[target.GetExecutionID()] = append(m[target.GetExecutionID()], target)
	}
	router := make(Router, 0, len(m))
	for id, targets := range m {
		router = append(router, element{ID: id, Targets: targets})
	}
	slices.SortFunc(router, func(a, b element) int {
		return strings.Compare(a.ID, b.ID)
	})
	return router
}

// Get returns execution targets by exact execution ID match.
func (r Router) Get(executionID string) ([]Target, bool) {
	i, ok := slices.BinarySearchFunc(r, executionID, func(a element, b string) int {
		return strings.Compare(a.ID, b)
	})
	if ok {
		return r[i].Targets, true
	}
	return nil, false
}

// GetEventBestMatch uses an exact match first, then the longest matching event
// prefix. Event wildcard routes use the suffix .*.
func (r Router) GetEventBestMatch(executionID string) ([]Target, bool) {
	targets, ok := r.Get(executionID)
	if ok {
		return targets, true
	}
	var bestMatch element
	for _, entry := range r {
		if strings.HasPrefix(executionID, entry.ID) {
			bestMatch, ok = entry, true
		}

		prefix, wildcard := strings.CutSuffix(entry.ID, ".*")
		if wildcard && strings.HasPrefix(executionID, prefix) {
			bestMatch, ok = entry, true
		}
	}
	return bestMatch.Targets, ok
}

func (r Router) IsZero() bool {
	return len(r) == 0
}
