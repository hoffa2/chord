package netutils

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/util"
)

type NodeRPC struct {
	client *rpc.Client
}

var (
	PORT = ":8001"
)

func RegisterNewType(t interface{}) {
	gob.Register(t)
}

func registerCommAPI(server *rpc.Server, comm comm.NodeComm) {
	server.RegisterName("NodeComm", comm)
}

// ConnectRPC Instantiates a RPC connections
func ConnectRPC(host string) (*NodeRPC, error) {
	fmt.Println(host)
	conn, err := net.Dial("tcp", host+PORT)
	if err != nil {
		return nil, err
	}
	client := rpc.NewClient(conn)
	var r comm.NodeID
	err = client.Call("NodeComm.Init", &comm.Args{ID: "init"}, &r)
	if err != nil {
		client.Close()
		return nil, err
	}
	if r.ID != "init" {
		client.Close()
		return nil, fmt.Errorf("Init failed")
	}
	nCom := &NodeRPC{client: client}
	return nCom, nil
}

func CloseRPC(c *NodeRPC) error {
	return c.client.Close()
}

// SetupRPCServer Instantiates a RPC Server
func SetupRPCServer(port string, api comm.NodeComm) error {
	s := rpc.NewServer()

	registerCommAPI(s, api)

	// the start means that we'll listen to
	// all traffic; Not just localhost
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	go s.Accept(l)
	return nil
}

func (n *NodeRPC) FindSuccessor(id util.Identifier) (comm.NodeID, error) {
	//args := &comm.NodeID{ID: []byte(id)}
	var reply comm.NodeID
	err := n.client.Call("NodeComm.FindSuccessor", &comm.Args{ID: string(id)}, &reply)
	if err != nil {
		return comm.NodeID{}, err
	}
	return reply, nil
}

func (n *NodeRPC) FindPredecessor(id util.Identifier) (comm.NodeID, error) {
	args := &comm.NodeID{ID: string(id)}
	var reply comm.NodeID
	err := n.client.Call("NodeComm.FindPredecessor", args, &reply)
	if err != nil {
		return comm.NodeID{}, err
	}
	return reply, nil
}

// PutRemote Stores a value in its respective node
func (n *NodeRPC) PutRemote(key, value string) error {
	args := &comm.KeyValue{Key: key, Value: value}
	err := n.client.Call("NodeComm.PutRemote", args, nil)
	if err != nil {
		return err
	}
	return nil
}

// PutRemote Stores a value in its respective node
func (n *NodeRPC) GetRemote(key string) (string, error) {
	args := &comm.KeyValue{Key: key}
	reply := comm.KeyValue{}
	err := n.client.Call("NodeComm.GetRemote", args, &reply)
	if err != nil {
		return "", err
	}
	return reply.Key, nil
}

func (n *NodeRPC) UpdatePredecessor(id util.Identifier, ip string) error {
	args := &comm.NodeID{ID: string(id), IP: ip}
	err := n.client.Call("NodeComm.UpdatePredecessor", args, nil)
	if err != nil {
		return err
	}
	return nil
}

func (n *NodeRPC) UpdateSuccessor(id util.Identifier, ip string) error {
	args := &comm.NodeID{ID: string(id), IP: ip}
	err := n.client.Call("NodeComm.UpdateSuccessor", args, nil)
	if err != nil {
		return err
	}
	return nil
}

func GetNodeIPs(address string) ([]string, error) {
	var list []string
	c := http.Client{Timeout: time.Duration(time.Second * 2)}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/", address), nil)
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

	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
