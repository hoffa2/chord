package netutils

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/util"
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

func CloseRPC(c *NodeComm) error {
	return c.client.Close()
}

// SetupRPCServer Instantiates a RPC Server
func SetupRPCServer(port string, api comm.NodeComm) error {
	s := rpc.NewServer()

	registerCommAPI(s, api)

	// the start means that we'll listen to
	// all traffic; Not just localhost
	l, err := net.Listen("tcp", "*:"+port)
	if err != nil {
		return err
	}

	go s.Accept(l)
	return nil
}

func (n *NodeComm) FindSuccessor(id util.Identifier) (comm.NodeID, error) {
	args := &comm.NodeID{ID: id}
	var reply comm.NodeID
	err := n.client.Call("NodeComm.FindSuccessor", args, &reply)
	if err != nil {
		return comm.NodeID{}, err
	}
	return reply, nil
}

func (n *NodeComm) FindPredecessor(id util.Identifier) (comm.NodeID, error) {
	args := &comm.NodeID{ID: id}
	var reply comm.NodeID
	err := n.client.Call("NodeComm.FindPredecessor", args, &reply)
	if err != nil {
		return comm.NodeID{}, err
	}
	return reply, nil
}

// PutRemote Stores a value in its respective node
func (n *NodeComm) PutRemote(key, value string) error {
	args := &comm.KeyValue{Key: key, Value: value}
	err := n.client.Call("NodeComm.PutRemote", args, nil)
	if err != nil {
		return err
	}
	return nil
}

// PutRemote Stores a value in its respective node
func (n *NodeComm) GetRemote(key string) error {
	args := &comm.KeyValue{Key: key}
	reply := comm.KeyValue{}
	err := n.client.Call("NodeComm.GetRemote", args, &reply)
	if err != nil {
		return err
	}
	return nil
}

func (n *NodeComm) UpdatePredecessor(id util.Identifier, ip string) error {
	args := &comm.NodeID{ID: id, IP: ip}
	err := n.client.Call("NodeComm.UpdatePredecessor", args, nil)
	if err != nil {
		return err
	}
	return nil
}

func (n *NodeComm) UpdateSuccessor(id util.Identifier, ip string) error {
	args := &comm.NodeID{ID: id, IP: ip}
	err := n.client.Call("NodeComm.UpdateSuccessor", args, nil)
	if err != nil {
		return err
	}
	return nil
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
