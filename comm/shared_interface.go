package comm

// NodeComm Chord RPC interface
type NodeComm interface {
	// FindPredecessor RPC call to find the predecessor of an identifer
	FindPredecessor(args *Args, reply *NodeID) error
	// FindSuccessor RPC call to find the Successor of an identifer
	FindSuccessor(args *Args, reply *NodeID) error
	// GetPredecessor RPC call to get a nodes predecessor
	GetPredecessor(args *Empty, reply *NodeID) error
	// GetSuccessor RPC call to get a nodes successor
	GetSuccessor(args *Empty, reply *NodeID) error
	// PutRemote put a key-value par on a remote node
	PutRemote(args *KeyValue, reply *Empty) error
	// GetRemote get request to a remote node
	GetRemote(args *KeyValue, reply *KeyValue) error
	// UpdatePredecessor updates a node's predecessor
	UpdatePredecessor(args *NodeID, reply *Empty) error
	// UpdateSUccessor updates a node's successor
	UpdateSuccessor(args *NodeID, reply *Empty) error
	// Init asserts RPC connection
	Init(args *Args, reply *NodeID) error
	UpdateFingerTable(args *FingerEntry, reply *Empty) error
	// ClosesPreFinger find the closeset predecesing finger in a node's fingertable
	ClosestPreFinger(id *string, reply *NodeID) error
	GetKeysInInterval(ival *Interval, reply *Keys) error
	// Notify RPC call to notify function as per Chord
	Notify(node *Rnode, reply *Empty) error
	// Leave called by an organizing entity to make a node leave the network
	Leave(in *Empty, out *Empty) error
}
