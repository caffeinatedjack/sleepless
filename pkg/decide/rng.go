// Package decide implements random choice utilities.
package decide

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	mrand "math/rand"
)

// RNG is the small interface this package needs for "pick one" decisions.
type RNG interface {
	// Intn returns a random integer in [0, n).
	Intn(n int) (int, error)
}

// CryptoRNG draws randomness from crypto/rand.
type CryptoRNG struct{}

// NewCryptoRNG returns a ready-to-use CryptoRNG.
func NewCryptoRNG() *CryptoRNG {
	return &CryptoRNG{}
}

// Intn returns a cryptographically random integer in [0, n).
func (r *CryptoRNG) Intn(n int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("invalid argument to Intn: %d", n)
	}
	max := big.NewInt(int64(n))
	result, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return int(result.Int64()), nil
}

// SeededRNG is deterministic and meant for repeatable runs.
type SeededRNG struct {
	rand *mrand.Rand
}

// NewSeededRNG creates a new SeededRNG with the given seed.
func NewSeededRNG(seed int64) *SeededRNG {
	return &SeededRNG{
		rand: mrand.New(mrand.NewSource(seed)),
	}
}

// Intn returns a deterministic random integer in [0, n).
func (r *SeededRNG) Intn(n int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("invalid argument to Intn: %d", n)
	}
	return r.rand.Intn(n), nil
}

// NewRNG picks between CryptoRNG (no seed) and SeededRNG (seeded).
func NewRNG(seed *int64) RNG {
	if seed == nil {
		return NewCryptoRNG()
	}
	return NewSeededRNG(*seed)
}

// RandomSeed generates a random seed using crypto/rand.
func RandomSeed() (int64, error) {
	var seed int64
	err := binary.Read(rand.Reader, binary.LittleEndian, &seed)
	if err != nil {
		return 0, err
	}
	return seed, nil
}
