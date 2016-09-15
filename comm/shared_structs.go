package comm

import "github.com/hoffa2/chord/util"

// Args arguments to an RPC
type Args struct {
	// Identifier of a node
	ID util.Identifier
	IP string
}

// KeyValue Arguments used in
// RPC calls involving remote get/put operations
type KeyValue struct {
	Key   string
	Value string
}

type Empty struct{}

type NodeID struct {
	ID util.Identifier
	IP string
}
