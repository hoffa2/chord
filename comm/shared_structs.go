package comm

// Args arguments to an RPC
type Args struct {
	// Identifier of a node
	ID string
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
