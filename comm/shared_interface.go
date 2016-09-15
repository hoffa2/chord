package comm

// NodeComm Chord RPC interface
type NodeComm interface {
	FindSuccessor(args *Args, reply *NodeID) error
	FindPredecessor(args *Args, reply *NodeID) error
	PutRemote(args *KeyValue, reply *NodeID) error
	GetRemote(args *KeyValue, reply *KeyValue) error
}
