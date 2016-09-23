package node

import (
	"fmt"

	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/util"
)

// FindPredecessor RPC call to find a predecessor of Key on node n
func (n *Node) FindPredecessor(args *comm.Args, reply *comm.NodeID) error {
	n.nMu.RLock()
	key := util.Identifier(args.ID)
	var pre comm.NodeID
	var err error

	if n.prev.NodeRPC == nil {
		reply.ID = string(n.id)
		reply.IP = n.IP
		n.nMu.RUnlock()
		return nil
	}

	if n.id.IsEqual(n.prev.id) {
		reply.ID = string(n.id)
		reply.IP = n.IP
		n.nMu.RUnlock()
		return nil
	}

	if key.InKeySpace(n.id, n.next.id) {
		reply.ID = string(n.id)
		reply.IP = n.IP
	} else if key.InKeySpace(n.prev.id, n.id) {
		reply.ID = string(n.prev.id)
		reply.IP = n.prev.IP
	} else {
		pre, err = n.next.FindPredecessor(key)
		reply.ID = pre.ID
		reply.IP = pre.IP
	}
	if err != nil {
		return err
	}

	n.nMu.RUnlock()
	return err
}

// FindSuccessor Finding the successor of n
func (n *Node) FindSuccessor(args *comm.Args, reply *comm.NodeID) error {
	n.nMu.RLock()
	key := util.Identifier(args.ID)
	var succ comm.NodeID
	var err error
	// If node is the only one in the ring
	if n.next.NodeRPC == nil {
		reply.ID = string(n.id)
		reply.IP = n.IP
		n.nMu.RUnlock()
		return nil
	}
	if key.InKeySpace(n.prev.id, n.id) {
		succ.ID = string(n.id)
		succ.IP = n.IP
	} else if key.InKeySpace(n.id, n.next.id) {
		succ.ID = string(n.next.id)
		succ.IP = n.next.IP
	} else if key.IsLarger(n.id) || key.IsLess(n.id) {
		succ, err = n.next.FindSuccessor(key)
	} else {
		// Should not reach this point
		n.nMu.RUnlock()
		return fmt.Errorf("Could not find successor")
	}

	reply.ID = string(succ.ID)
	reply.IP = succ.IP
	n.nMu.RUnlock()
	return err
}

// UpdatePredecessor Updates n's predecessor and initializes an RPC connection
func (n *Node) UpdatePredecessor(args *comm.NodeID, reply *comm.Empty) error {
	IP := args.IP
	ID := args.ID

	err := n.setPredecessor(util.StringToID(ID), IP)
	if err != nil {
		return err
	}
	return nil
}

// PutRemote Gets an RPC put request to store a Key/Value pair
func (n *Node) PutRemote(args *comm.KeyValue, reply *comm.Empty) error {
	n.putValue(args.Key, []byte(args.Value))
	return nil
}

// GetRemote Gets an RPC put request to store a Key/Value pair
func (n *Node) GetRemote(args *comm.KeyValue, reply *comm.KeyValue) error {
	val, err := n.getValue(args.Key)
	if err != nil {
		return err
	}
	reply.Value = val
	return nil
}

// UpdateSuccessor Updates node n's successor and initializes an RPC connection
func (n *Node) UpdateSuccessor(args *comm.NodeID, reply *comm.Empty) error {
	IP := args.IP
	ID := args.ID

	err := n.setSuccessor(util.StringToID(ID), IP)
	if err != nil {
		return err
	}
	return nil
}

// Init convenience function to assert successful RPC init
func (n *Node) Init(args *comm.Args, reply *comm.NodeID) error {
	reply.ID = args.ID
	return nil
}

// UpdateFingerTable Updates n's fingertable's i'th entry
func (n *Node) UpdateFingerTable(args *comm.FingerEntry, reply *comm.Empty) error {
	return n.updateFTable(args.S.ID, args.S.IP, args.IDX)
}
