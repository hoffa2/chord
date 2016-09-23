package comm

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
