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

// Fingertable Table holding information
// about neighboring node
type FingerTable struct {
	// Numentries in table
	entries int
	finger  []FingerEntry
}
