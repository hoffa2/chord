package comm

type NodeComm interface {
	FindSuccessor(args *Args, reply *NodeID) error
	FindPredecessor(args *Args, reply *NodeID) error
}
