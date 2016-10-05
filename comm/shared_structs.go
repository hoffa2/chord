package comm

import "github.com/hoffa2/chord/util"

// Args arguments to an RPC
type Args struct {
	// Identifier of a node
	ID string
}

// FingerEntry
type FingerEntry struct {
	S   NodeID
	IDX int
}

type Interval struct {
	From string
	To   string
}

type Keys map[string]string

// KeyValue Arguments used in
// RPC calls involving remote get/put operations
type KeyValue struct {
	Key   string
	Value string
}

type Test struct {
}
type Empty struct{}

type NodeID struct {
	ID string
	IP string
}

type Rnodes []Rnode

type Rnode struct {
	ID util.Identifier
	IP string
}

func (slice Rnodes) Len() int {
	return len(slice)
}

func (slice Rnodes) Less(i, j int) bool {
	return slice[i].ID.IsLess(slice[j].ID)
}

func (slice Rnodes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
