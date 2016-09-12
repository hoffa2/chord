package node

import "github.com/hoffa2/chord/netutils"

// FingerEntry One entry in the fingertable
type FingerEntry struct {
	// Start of identifier space of node
	start string
	// IpAddress of node
	ipAdress string
	node     *netutils.NodeComm
}
