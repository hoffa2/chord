package comm

import "github.com/hoffa2/chord/util"

// Args arguments to an RPC
type Args struct {
	// Identifier of a node
	ID util.Identifier
}

type NodeID struct {
	ID util.Identifier
}
