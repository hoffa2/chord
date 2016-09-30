package comm

import "github.com/hoffa2/chord/comm"

// NodeComm Chord RPC interface
type NodeComm interface {
	FindPredecessor(args *Args, reply *NodeID) error
	FindSuccessor(args *Args, reply *NodeID) error
	GetPredecessor(args *Empty, reply *NodeID) error
	GetSuccessor(args *Empty, reply *NodeID) error
	PutRemote(args *KeyValue, reply *Empty) error
	GetRemote(args *KeyValue, reply *KeyValue) error
	UpdatePredecessor(args *NodeID, reply *Empty) error
	UpdateSuccessor(args *NodeID, reply *Empty) error
	Init(args *Args, reply *NodeID) error
	UpdateFingerTable(args *FingerEntry, reply *Empty) error
	ClosestPreFinger(id *string, reply *NodeID) error
	GetKeysInInterval(ival *Interval, reply *Keys) error
	Notify(node *comm.Rnode, reply *Empty)
}
