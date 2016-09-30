package node

import (
	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/util"
)

// FindPredecessor RPC call to find a predecessor of Key on node n
func (n *Node) FindPredecessor(args *comm.Args, reply *comm.NodeID) error {
	n.nMu.RLock()
	defer n.nMu.RUnlock()
	key := util.Identifier(args.ID)

	if n.ID.IsEqual(n.prev.ID) {
		reply.ID = n.ID.ToString()
		reply.IP = n.IP
		return nil
	}

	pre, err := n.findPredecessor(key)
	if err != nil {
		return err
	}

	reply.ID = pre.ID.ToString()
	reply.IP = pre.IP
	return err
}

// FindPredecessor RPC call to find a predecessor of Key on node n
func (n *Node) FindSuccessor(args *comm.Args, reply *comm.NodeID) error {
	n.nMu.RLock()
	defer n.nMu.RUnlock()
	key := util.Identifier(args.ID)

	if n.ID.IsEqual(n.fingers[0].node.ID) {
		reply.ID = n.ID.ToString()
		reply.IP = n.IP
		return nil
	}

	succ, err := n.findSuccessor(key)
	if err != nil {
		return err
	}

	reply.ID = succ.ID.ToString()
	reply.IP = succ.IP
	return err
}

// FindSuccessor Finding the successor of n
func (n *Node) GetSuccessor(args *comm.Empty, reply *comm.NodeID) error {
	n.nMu.RLock()
	defer n.nMu.RUnlock()

	reply.IP = n.fingers[0].node.IP
	reply.ID = n.fingers[0].node.ID.ToString()
	return nil
}

func (n *Node) GetPredecessor(args *comm.Empty, reply *comm.NodeID) error {
	n.nMu.RLock()
	defer n.nMu.RUnlock()

	reply.IP = n.prev.IP
	reply.ID = n.prev.ID.ToString()
	return nil
}

// UpdatePredecessor Updates n's predecessor and initializes an RPC connection
func (n *Node) UpdatePredecessor(args *comm.NodeID, reply *comm.Empty) error {
	IP := args.IP
	ID := args.ID

	err := n.setPredecessor(&comm.Rnode{ID: util.StringToID(ID), IP: IP})
	if err != nil {
		return err
	}
	return nil
}

// PutRemote Gets an RPC put request to store a Key/Value pair
func (n *Node) PutRemote(args *comm.KeyValue, reply *comm.Empty) error {
	n.putValue(util.StringToID(args.Key), []byte(args.Value))
	return nil
}

// GetRemote Gets an RPC put request to store a Key/Value pair
func (n *Node) GetRemote(args *comm.KeyValue, reply *comm.KeyValue) error {
	val, err := n.getValue(util.StringToID(args.Key))
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

	err := n.setSuccessor(&comm.Rnode{ID: util.StringToID(ID), IP: IP})
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

func (n *Node) ClosestPreFinger(args *string, reply *comm.NodeID) error {
	rnode := n.closestPreFinger(util.StringToID(*args))
	*reply = comm.NodeID{ID: rnode.ID.ToString(), IP: rnode.IP}
	return nil
}

// UpdateFingerTable Updates n's fingertable's i'th entry
func (n *Node) UpdateFingerTable(args *comm.FingerEntry, reply *comm.Empty) error {
	return n.updateFTable(util.StringToID(args.S.ID), args.S.IP, args.IDX)
}

func (n *Node) GetKeysInInterval(ival *comm.Interval, reply *comm.Keys) error {
	*reply = n.migrateKeys(ival.From, ival.To)
	return nil
}

func (n *Node) Notify(node *comm.Rnode, reply *comm.Empty) {
	n.notify(node)
}
