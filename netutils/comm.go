package netutils

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"sync"
	"time"

	"github.com/hoffa2/chord/comm"
)

type NodeRPC struct {
	sync.Mutex
	host    string
	c       *rpc.Client
	timeout time.Duration
}

var (
	PORT       = ":8011"
	ErrTimeout = errors.New("RPC call timed out")
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
	return &NodeRPC{c: client, host: host, timeout: time.Duration(time.Second * 2)}, nil
}

func (n *NodeRPC) reDial() error {
	err := n.c.Close()
	conn, err := net.Dial("tcp4", n.host+PORT)
	n.c = rpc.NewClient(conn)
	return err
}

// SetupRPCServer Instantiates a RPC Server
func SetupRPCServer(port string, api comm.NodeComm) (net.Listener, error) {
	s := rpc.NewServer()

	registerCommAPI(s, api)

	// the start means that we'll listen to
	// all traffic; Not just localhost
	l, err := net.Listen("tcp4", ":"+port)
	if err != nil {
		return nil, err
	}

	go s.Accept(l)
	return l, nil
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

func UnRegister(ip, nameserver string) {
	http.PostForm(fmt.Sprintf("http://%s/unregister", nameserver),
		url.Values{"ip": {ip}})
}
func (n *NodeRPC) Call(method string, args interface{}, reply interface{}) error {
	call := n.c.Go(method, args, reply, nil)
	select {
	case <-time.After(n.timeout):
		return ErrTimeout
	case call := <-call.Done:
		if call.Error != nil {
			return call.Error
		}
	}
	return nil
}
