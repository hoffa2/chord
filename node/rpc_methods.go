package node

import "github.com/hoffa2/chord/comm"

// FindPredecessor RPC call to find a predecessor of Key on node n
func (n Node) FindPredecessor(args *comm.Args, reply *comm.NodeID) error {
	key := args.ID
	var pre comm.NodeID
	var err error

	if n.ID.IsEqual(n.prev.ID) {
		reply.ID = n.ID
		reply.IP = n.IP
	}

	if key.IsBetween(n.ID, n.next.ID) {
		reply.ID = n.ID
		reply.IP = n.IP
	} else if key.IsLarger(n.next.ID) {
		pre, err = n.next.conn.FindPredecessor(key)
		reply.ID = pre.ID
		reply.IP = pre.IP
	}
	if err != nil {
		return err
	}

	return nil
}

// FindSuccessor Finding the successor of n
func (n Node) FindSuccessor(args *comm.Args, reply *comm.NodeID) error {
	key := args.ID
	var succ comm.NodeID
	var err error
	// If node is the only one in the ring
	if n.ID.IsEqual(n.next.ID) {
		reply.ID = n.ID
		reply.IP = n.IP
		return nil
	}

	if key.InKeySpace(n.prev.ID, n.ID) {
		succ.ID = n.ID
		succ.IP = n.IP
	} else if key.IsLess(n.next.ID) && key.IsLarger(n.ID) {
		succ.ID = n.next.ID
		succ.IP = n.next.IP
	} else if key.IsLarger(n.ID) {
		succ, err = n.next.conn.FindSuccessor(key)
	}
	if err != nil {
		return err
	}

	reply.ID = succ.ID
	reply.IP = succ.IP

	return nil
}

// UpdatePredecessor Updates n's predecessor and initializes an RPC connection
func (n *Node) UpdatePredecessor(args *comm.NodeID, reply *comm.Empty) error {
	IP := args.IP
	ID := args.ID

	err := n.setPredecessor(ID, IP)
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

	err := n.setSuccessor(ID, IP)
	if err != nil {
		return err
	}
	return nil
}
