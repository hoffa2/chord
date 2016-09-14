package netutils

import (
	"encoding/json"
	"fmt"
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
func ConnectRPC(host string) (*NodeComm, error) {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}
	nCom := &NodeComm{client: rpc.NewClient(conn)}
	return nCom, nil
}

// SetupRPCServer Instantiates a RPC Server
func SetupRPCServer(port string, api comm.NodeComm) error {
	s := rpc.NewServer()

	registerCommAPI(s, api)
	rpc.HandleHTTP()

	// the start means that we'll listen to
	// all traffic; Not just localhost
	l, err := net.Listen("tcp", "*:"+port)
	if err != nil {
		return err
	}

	go http.Serve(l, nil)
	return nil
}

func (n *NodeComm) FindSuccessor(id string) (string, error) {
	args := &comm.Args{ID: id}
	var reply comm.NodeID
	err := n.client.Call("NodeComm.FindSuccessor", args, &reply)
	if err != nil {
		return "", err
	}
	return reply.ID, nil
}

func (n *NodeComm) FindPredecessor(id string) (string, error) {
	args := &comm.Args{ID: id}
	var reply comm.NodeID
	err := n.client.Call("NodeComm.FindPredecessor", args, &reply)
	if err != nil {
		return "", err
	}
	return reply.ID, nil
}

func GetNodeIPs(address string) ([]string, error) {
	var list []string
	c := http.Client{}

	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not retrieve ipadresses from nameserver: %s", address)
	}

	err = json.NewDecoder(resp.Body).Decode(list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
