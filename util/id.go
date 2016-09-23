package util

import (
	"bytes"
	"math/big"
)

type Identifier []byte

// IsLrger checks whether id is larger than node
func (id Identifier) IsLarger(b Identifier) bool {
	return bytes.Compare(id, b) == 1
}

// IsLess checks whether id is less than node
func (id Identifier) IsLess(b Identifier) bool {
	return bytes.Compare(id, b) == -1
}

// isEqual returns if id is equal to nodeid
func (id Identifier) IsEqual(b Identifier) bool {
	return bytes.Compare(id, b) == 0
}

// StringToID Converts a string to a sequence of bytes
// represented as NodeID
func StringToID(str string) Identifier {
	var n Identifier
	b := []byte(str)
	for _, val := range b {
		n = append(n, val)
	}
	return n
}

// InKeySpace asserts whether nodeID is in
// the keyspace between one and two.
func (id Identifier) InKeySpace(one, two Identifier) bool {

	// check whether ring wraps around
	if bytes.Compare(one, two) == 1 {
		return bytes.Compare(one, id) == -1 || bytes.Compare(two, id) >= 0
	}
	return bytes.Compare(one, id) == -1 &&
		bytes.Compare(two, id) >= 0
}

func (id Identifier) IsBetween(one, two Identifier) bool {

	// check whether ring wraps around
	if bytes.Compare(one, two) == -1 {
		return bytes.Compare(one, id) == 1 && (bytes.Compare(two, id) <= 0 || bytes.Compare(two, id) >= 0)
	}
	return bytes.Compare(id, two) <= 0 &&
		bytes.Compare(id, one) == 1
}

func (id Identifier) Add(id2 Identifier) Identifier {
	res := make(Identifier, len(id))
	for i, _ := range res {
		res[i] = id[i] + id2[i]
	}
	return res
}

func Mod(id1, id2 Identifier) Identifier {
	var one big.Int
	var two big.Int
	var res big.Int
	one.SetBytes(id1)
	two.SetBytes(id2)

	res.Mod(&one, &two)
	return res.Bytes()
}

func (id Identifier) CalculateStart(k, m int64) Identifier {
	msquared := new(big.Int).Exp(big.NewInt(2), big.NewInt(m), nil)
	ksquared := new(big.Int).Exp(big.NewInt(2), big.NewInt(k-1), nil)
	left := new(big.Int).Add(big.NewInt(0).SetBytes(id), ksquared)
	start := new(big.Int).Mod(left, msquared)
	return start.Bytes()
}
