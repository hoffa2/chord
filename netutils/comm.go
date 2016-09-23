package netutils

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"

	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/util"
)

type NodeRPC struct {
	client *rpc.Client
	host   string
	sync.Mutex
}

var (
	PORT = ":8010"
)

func RegisterNewType(t interface{}) {
	gob.Register(t)
}

func registerCommAPI(server *rpc.Server, comm comm.NodeComm) {
	server.RegisterName("NodeComm", comm)
}

// ConnectRPC Instantiates a RPC connections
func ConnectRPC(host string) (*NodeRPC, error) {
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
	nCom := &NodeRPC{client: client, host: host + PORT}
	return nCom, nil
}

func (c *NodeRPC) reDial() error {
	err := c.client.Close()
	conn, err := net.Dial("tcp", c.host)
	c.client = rpc.NewClient(conn)
	return err
}

func CloseRPC(c *NodeRPC) error {
	return c.client.Close()
}

// SetupRPCServer Instantiates a RPC Server
func SetupRPCServer(port string, api comm.NodeComm) (net.Listener, error) {
	s := rpc.NewServer()

	registerCommAPI(s, api)

	// the start means that we'll listen to
	// all traffic; Not just localhost
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}

	go s.Accept(l)
	return l, nil
}

func (n *NodeRPC) FindSuccessor(id util.Identifier) (comm.NodeID, error) {
	//args := &comm.NodeID{ID: []byte(id)}
	var reply comm.NodeID
	err := n.Call("NodeComm.FindSuccessor", &comm.Args{ID: string(id)}, &reply)
	if err != nil {
		return comm.NodeID{}, err
	}
	return reply, nil
}

func (n *NodeRPC) FindPredecessor(id util.Identifier) (comm.NodeID, error) {
	args := &comm.NodeID{ID: string(id)}
	var reply comm.NodeID
	err := n.Call("NodeComm.FindPredecessor", args, &reply)
	if err != nil {
		return comm.NodeID{}, err
	}
	return reply, nil
}

// PutRemote Stores a value in its respective node
func (n *NodeRPC) PutRemote(key, value string) error {
	args := &comm.KeyValue{Key: key, Value: value}
	err := n.Call("NodeComm.PutRemote", args, nil)
	if err != nil {
		return err
	}
	return nil
}

// PutRemote Stores a value in its respective node
func (n *NodeRPC) GetRemote(key string) (string, error) {
	args := &comm.KeyValue{Key: key}
	reply := comm.KeyValue{}
	err := n.Call("NodeComm.GetRemote", args, &reply)
	if err != nil {
		return "", err
	}

	return reply.Value, nil
}

func (n *NodeRPC) UpdatePredecessor(id util.Identifier, ip string) error {
	args := &comm.NodeID{ID: string(id), IP: ip}
	err := n.Call("NodeComm.UpdatePredecessor", args, nil)
	if err != nil {
		return err
	}
	return nil
}

func (n *NodeRPC) UpdateSuccessor(id util.Identifier, ip string) error {
	args := &comm.NodeID{ID: string(id), IP: ip}
	err := n.Call("NodeComm.UpdateSuccessor", args, nil)
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

func (c *NodeRPC) Call(method string, args interface{}, reply interface{}) error {
	err := c.client.Call(method, args, reply)
	if err == rpc.ErrShutdown {
		c.Mutex.Lock()
		err = c.reDial()
		c.Mutex.Unlock()
		if err != nil {
			return err
		}
		return c.client.Call(method, args, reply)
	} else if err != nil {
		return err
	}
	return nil
}
