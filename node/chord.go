package node

import (
	"errors"
	"net/http"
	"sync"

	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/netutils"
	"github.com/hoffa2/chord/util"
)

// FingerEntry One entry in the fingertable
type FingerEntry struct {
	// Start of identifier space of node
	start util.Identifier
	node  *comm.Rnode
}

const (
	// NoValue if a put request does not have a body
	NoValue = "No value in body"
	// NotFound If key cannot be found

	// Internal Internal error
	Internal = "Internal error: "
)

var (
	// Keysize size of keyspace
	KeySize         = 160
	ErrInvalidIndex = errors.New("ftable index is invalid")
	// ErrNotFound if key does not exist
	ErrNotFound = errors.New("No value on key")
	// ErrNextToSmall if the successor is too small
	ErrNextToSmall = errors.New("Successor is less than n id")
	// ErrPrevToLarge If the predecessor is too large
	ErrPrevToLarge = errors.New("Predecessor is larger than n id")
)

// Neighbor Describing an adjacent node in the ring

// Node Interface struct that represents the state
// of one node
type Node struct {
	// Storing key-value pairs on the respective node
	mu  sync.RWMutex
	nMu sync.RWMutex
	// Representing the local node
	*comm.Rnode
	// Map in which keys are stored
	objectStore map[string]string
	// IP Address of nameserver
	nameServer string
	conn       http.Client
	// FingerTable
	fingers []FingerEntry
	// Predecessor of node
	prev *comm.Rnode
	// RPC connection wrapper
	remote *netutils.Remote

	successors []*comm.Rnode

	log *Logger
}
