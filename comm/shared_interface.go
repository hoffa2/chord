package comm

// NodeComm Chord RPC interface
type NodeComm interface {
	FindSuccessor(args *Args, reply *NodeID) error
	FindPredecessor(args *Args, reply *NodeID) error
	PutRemote(args *KeyValue, reply *Empty) error
	GetRemote(args *KeyValue, reply *KeyValue) error
	UpdatePredecessor(args *NodeID, reply *Empty) error
	UpdateSuccessor(args *NodeID, reply *Empty) error
	Init(args *Args, reply *NodeID) error
}
