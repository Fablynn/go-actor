package random

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/big"
	"math/rand"
	"time"
)

var (
	randObj = rand.New(rand.NewSource(time.Now().UnixNano()))
	//rng     = NewSecureRNG()
)

// [0,n)
func Intn(n int) int {
	return randObj.Intn(n)
}

// [0,n)
func Int32n(n int32) int32 {
	return int32(Intn(int(n)))
}

// SecureRNG 密码学安全 性能较差
type SecureRNG struct{}

func NewSecureRNG() *SecureRNG {
	return &SecureRNG{}
}

// Uint64 made [0, 2^64-1) rand number
func (r *SecureRNG) Uint64() (uint64, error) {
	buf := make([]byte, 8)
	if _, err := crand.Read(buf); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buf), nil
}

// Range made [min, max) rand number
func (r *SecureRNG) Range(max int64) (int64, error) {
	bigMax := big.NewInt(max)
	n, err := crand.Int(crand.Reader, bigMax)
	if err != nil {
		return 0, err
	}
	return n.Int64(), nil
}
