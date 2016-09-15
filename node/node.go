package node

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/netutils"
	"github.com/hoffa2/chord/util"
)

const (
	Keysize  = 12
	NoValue  = "No value in body"
	NotFound = "No value on key: %s"
	Internal = "Internal error: "
)

var (
	NextToSmall = errors.New("Successor is less than n id")
	PrevToLarge = errors.New("Predecessor is larger than n id")
)

type Neighbor struct {
	ID   util.Identifier
	conn *netutils.NodeComm
	IP   string
}

// Interface struct that represents the state
// of one node
type Node struct {
	// Storing key-value pairs on the respective node
	mu          sync.RWMutex
	objectStore map[string]string
	ID          util.Identifier
	nameServer  string
	conn        http.Client
	finger      []FingerEntry
	next        Neighbor
	prev        Neighbor
	IP          string
}

func readKey(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["key"]
}

func (n Node) findKeySuccessor(k util.Identifier) (Neighbor, error) {
	// I'm the successor
	if k.InKeySpace(n.prev.ID, n.ID) {
		return Neighbor{ID: n.ID}, nil
	}
	s, err := n.next.conn.FindSuccessor(k)
	if err != nil {
		return Neighbor{}, nil
	}
	return Neighbor{
		ID: s.ID,
		IP: s.IP,
	}, nil
}

func (n *Node) putValue(key string, body []byte) {
	n.mu.Lock()
	n.objectStore[key] = string(body)
	n.mu.Unlock()
}

func (n *Node) getValue(key string) (string, error) {
	n.mu.RLock()
	val, ok := n.objectStore[key]
	if !ok {
		n.mu.RUnlock()
		return "", fmt.Errorf(NotFound, key)
	}
	n.mu.RUnlock()
	return val, nil
}

func (n *Node) putKey(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, NoValue, http.StatusBadRequest)
		return
	}

	key := readKey(r)

	kID := util.StringToID(key)

	s, err := n.findKeySuccessor(kID)
	if err != nil {
		http.Error(w, Internal, http.StatusInternalServerError)
		return
	}
	if s.ID.IsEqual(n.ID) {
		n.putValue(key, body)
	} else {
		putValueNeighbor(s)
	}

	w.WriteHeader(http.StatusOK)
}

func (n *Node) getKey(w http.ResponseWriter, r *http.Request) {
	key := readKey(r)

	val, ok := n.objectStore[key]
	if !ok {
		util.ErrorNotFound(w, NotFound, key)
	}

	sendKey(w, val)
}

// Just writes the response - may need to add more
// header later. But, for now ResponseWriter
// handles everything
func sendKey(w http.ResponseWriter, key string) {
	w.Write([]byte(key))
}

func (n Node) registerNode() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", n.nameServer, strings.NewReader(hostname))
	if err != nil {
		return err
	}

	resp, err := n.conn.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Could not register with nameserver: %s", n.nameServer)
	}
	return nil
}

// Findsuccessor Finding the successor of n
func (n Node) FindSuccessor(args *comm.Args, reply *comm.NodeID) error {
	key := args.ID
	var succ comm.NodeID
	var err error
	// If node is the only one in the ring
	if n.ID.IsEqual(n.next.ID) {
		reply.ID = n.ID
		reply.IP = n.IP
		return nil
	}

	if key.InKeySpace(n.prev.ID, n.ID) {
		succ.ID = n.ID
		succ.IP = n.IP
	} else if key.IsLess(n.next.ID) && key.IsLarger(n.ID) {
		succ.ID = n.next.ID
		succ.IP = n.next.IP
	} else if key.IsLarger(n.ID) {
		succ, err = n.next.conn.FindSuccessor(key)
	}
	if err != nil {
		return err
	}

	reply.ID = succ.ID
	reply.IP = succ.IP

	return nil
}

// FindPredecessor RPC call to find a predecessor of Key on node n
func (n Node) FindPredecessor(args *comm.Args, reply *comm.NodeID) error {
	key := args.ID
	var pre comm.NodeID
	var err error

	if n.ID.IsEqual(n.prev.ID) {
		reply.ID = n.ID
		reply.IP = n.IP
	}

	if key.IsBetween(n.ID, n.next.ID) {
		reply.ID = n.ID
		reply.IP = n.IP
	} else if key.IsLarger(n.next.ID) {
		pre, err = n.next.conn.FindPredecessor(key)
		reply.ID = pre.ID
		reply.IP = pre.IP
	}
	if err != nil {
		return err
	}

	return nil
}

// UpdatePredecessor Updates n's predecessor and initializes an RPC connection
func (n *Node) UpdatePredecessor(args *comm.Args, reply *comm.NodeID) error {
	Ip := args.IP
	Id := args.ID

	if Id.IsLess(n.ID) {
		return PrevToLarge
	}

	err := n.setPredecessor(Id, Ip)
	if err != nil {
		return err
	}
	return nil
}

// PutRemote Gets an RPC put request to store a Key/Value pair
func (n *Node) PutRemote(args *comm.KeyValue, reply *comm.NodeID) error {
	n.putValue(args.Key, []byte(args.Value))
	return nil
}

// GetRemote Gets an RPC put request to store a Key/Value pair
func (n *Node) GetRemote(args *comm.KeyValue, reply *comm.KeyValue) error {
	val, err := n.getValue(args.Key)
	if err != nil {
		return err
	}
	reply.Value = val
	return nil
}

// UpdateSuccessor Updates node n's successor and initializes an RPC connection
func (n *Node) UpdateSuccessor(args *comm.Args, reply *comm.NodeID) error {
	Ip := args.IP
	Id := args.ID

	if Id.IsLarger(n.ID) {
		return PrevToLarge
	}

	err := n.setSuccessor(Id, Ip)
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) setPredecessor(id util.Identifier, ip string) error {
	c, err := netutils.ConnectRPC(ip)
	if err != nil {
		return err
	}

	n.prev = Neighbor{
		ID:   id,
		IP:   ip,
		conn: c,
	}

	return nil
}

func (n *Node) setSuccessor(id util.Identifier, ip string) error {
	if id.IsEqual(n.ID) {
		return nil
	}

	c, err := netutils.ConnectRPC(ip)
	if err != nil {
		return err
	}
	n.next = Neighbor{
		ID:   id,
		IP:   ip,
		conn: c,
	}

	return nil
}

// registers itself in the network by asking a random
// node found by issuing a request to the nameserver
func (n *Node) JoinNetwork(id string) error {
	err := n.registerNode()
	if err != nil {
		return err
	}

	nodes, err := netutils.GetNodeIPs(n.nameServer)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		// I'm alone!!!
		n.setSuccessor(util.StringToID(id), n.IP)
		return nil
	}

	// Asking the first node who's my successor
	c, err := netutils.ConnectRPC(nodes[0])
	if err != nil {
		return err
	}

	// setting RPC connection object
	// Todo: find a better way to do this
	n.next.conn = c

	succ, err := c.FindSuccessor(n.ID)
	if err != nil {
		return err
	}

	err = n.setSuccessor(succ.ID, succ.IP)
	if err != nil {
		return err
	}

	return nil
}
