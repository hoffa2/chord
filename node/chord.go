package node

import (
	"github.com/hoffa2/chord/netutils"
	"github.com/hoffa2/chord/util"
)

// FingerEntry One entry in the fingertable
type FingerEntry struct {
	// Start of identifier space of node
	start util.Identifier
	// IpAddress of node
	ip   string
	succ util.Identifier
	*netutils.NodeRPC
}
