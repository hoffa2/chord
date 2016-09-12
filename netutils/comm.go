package netutils

import (
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/hoffa2/chord/comm"
)

type NodeComm struct {
	client *rpc.Client
}

func registerCommAPI(server *rpc.Server, comm comm.NodeComm) {
	server.RegisterName("NodeComm", comm)
}

// ConnectRPC Instantiates a RPC connections
func ConnectRPC(host string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", host)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// SetupRPCServer Instantiates a RPC Server
func SetupRPCServer(port string, api comm.NodeComm) {
	s := rpc.NewServer()

	registerCommAPI(s, api)
	rpc.HandleHTTP()

	// the start means that we'll listen to
	// all traffic; Not just localhost
	l, err := net.Listen("tcp", "*:"+port)
	if err != nil {
		log.Fatalf("Could not start listening on port %s. Error: %s", port, err)
	}

	go http.Serve(l, nil)
}

func (n *NodeComm) FindSuccessor(id string) string {
	args := &comm.Args{ID: id}
	var reply comm.NodeID
	err := n.client.Call("NodeComm.FindSuccessor", args, &reply)
	if err != nil {
		log.Fatal("Comm error: ", err)
	}
	return reply.ID
}

func (n *NodeComm) FindPredecessor(id string) string {
	args := &comm.Args{ID: id}
	var reply comm.NodeID
	err := n.client.Call("NodeComm.FindSuccessor", args, &reply)
	if err != nil {
		log.Fatal("Comm error: ", err)
	}
	return reply.ID
}
