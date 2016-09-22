package util

import "bytes"

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
	if bytes.Compare(one, two) == -1 {
		return bytes.Compare(one, id) == -1 ||
			bytes.Compare(two, id) >= 0
	}
	return bytes.Compare(id, one) >= 0 &&
		bytes.Compare(id, two) == -1
}

func (id Identifier) IsBetween(one, two Identifier) bool {

	// check whether ring wraps around
	if bytes.Compare(one, two) == -1 {
		return bytes.Compare(one, id) == 1 && (bytes.Compare(two, id) <= 0 || bytes.Compare(two, id) >= 0)
	}
	return bytes.Compare(id, two) <= 0 &&
		bytes.Compare(id, one) == 1
}
