// Package fake implements fake data generation utilities.
package fake

import (
	"gitlab.com/caffeinatedjack/sleepless/pkg/decide"
)

// RNG is the interface for random number generation.
// It re-exports the decide.RNG interface for consistency.
type RNG = decide.RNG

// NewRNG creates a new RNG, either seeded or cryptographic.
func NewRNG(seed *int64) RNG {
	return decide.NewRNG(seed)
}
