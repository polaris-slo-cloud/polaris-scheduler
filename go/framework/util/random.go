package util

import (
	crand "crypto/rand"
	"math/big"
	mrand "math/rand"
)

var (
	_ Random = (*mathRandom)(nil)
	_ Random = (*cryptoRandom)(nil)
)

// A random number generator interface, for easily switching between multiple random number generator implementations.
// Implementations of this interface are generally NOT thread-safe.
type Random interface {
	// Generates a random integer in the interval [0, max)
	Int(max int) int
}

// Creates an instance of the default Random implementation.
func NewDefaultRandom() Random {
	return NewCryptoRandom()
}

// A Random implementation using the math/rand package.
// This is not thread-safe.
// rand.Rand is not thread-safe and using the global rand.Int() function could cause performance loss, because it uses a mutex to sync access to a single Rand.
type mathRandom struct {
	random *mrand.Rand
}

// Creates an instance of a Random implementation using the math/rand package,
// seeded using rand.Int63().
func NewMathRandom() Random {
	seed := mrand.Int63()
	random := &mathRandom{
		random: mrand.New(mrand.NewSource(seed)),
	}
	return random
}

func (mr *mathRandom) Int(max int) int {
	return mr.random.Intn(max)
}

// A Random implementation using the crypto/rand package.
// cryptoRandom is thread-safe.
type cryptoRandom struct {
}

// Creates an instance of a Random implementation using the crypto/rand package.
func NewCryptoRandom() Random {
	return &cryptoRandom{}
}

func (cr *cryptoRandom) Int(max int) int {
	res, err := crand.Int(crand.Reader, big.NewInt(int64(max)))
	if err == nil {
		return int(res.Int64())
	} else {
		panic(err)
	}
}
