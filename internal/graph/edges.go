package graph

import "github.com/hb-chen/opskills/internal/state"

// Edge represents a graph edge
type Edge struct {
	From      string
	To        string
	Condition func(*state.State) bool // Optional condition for conditional routing
}

// ConditionalEdge creates an edge with a condition
func ConditionalEdge(from, to string, condition func(*state.State) bool) Edge {
	return Edge{
		From:      from,
		To:        to,
		Condition: condition,
	}
}

// SimpleEdge creates a simple unconditional edge
func SimpleEdge(from, to string) Edge {
	return Edge{
		From:      from,
		To:        to,
		Condition: nil,
	}
}
