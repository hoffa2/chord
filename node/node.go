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
	"github.com/hoffa2/chord/netutils"
	"github.com/hoffa2/chord/util"
)

const (
	// NoValue if a put request does not have a body
	NoValue = "No value in body"
	// NotFound If key cannot be found

	// Internal Internal error
	Internal = "Internal error: "
)

var (
	ErrNotFound = errors.New("No value on key")
	// ErrNextToSmall if the successor is too small
	ErrNextToSmall = errors.New("Successor is less than n id")
	// ErrPrevToLarge If the predecessor is too large
	ErrPrevToLarge = errors.New("Predecessor is larger than n id")
)

// Neighbor Describing an adjacent node in the ring
type Neighbor struct {
	ID   util.Identifier
	conn *netutils.NodeRPC
	IP   string
}

// Node Interface struct that represents the state
// of one node
type Node struct {
	// Storing key-value pairs on the respective node
	mu  sync.RWMutex
	nMu sync.RWMutex
	//
	objectStore map[string]string
	// Node Identifier
	ID util.Identifier
	// IP Address of nameserver
	nameServer string
	conn       http.Client
	finger     []FingerEntry
	// Successor of node
	next Neighbor
	// Predecessor of node
	prev Neighbor
	// Node IP
	IP string
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
	// TODO: Maybe we should check whether the key is in our successor's keyspace
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
		return "", ErrNotFound
	}
	n.mu.RUnlock()
	return val, nil
}

func (n Node) findExistingConn(id util.Identifier) *netutils.NodeRPC {
	if id.IsEqual(n.next.ID) {
		return n.next.conn
	} else if id.IsEqual(n.prev.ID) {
		return n.prev.conn
	}
	return nil
}

func (n Node) sendToSuccessor(key, val string, s Neighbor) error {
	var err error
	close := false
	c := n.findExistingConn(s.ID)

	// if we got an unfamiliar node
	if c == nil {
		c, err = netutils.ConnectRPC(s.IP)
		if err != nil {
			return err
		}
		close = true
	}

	err = c.PutRemote(key, val)
	if err != nil {
		return err
	}

	// if we got an unknown node - close connection
	if close {
		return netutils.CloseRPC(c)
	}

	return nil
}

func (n *Node) putKey(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, NoValue, http.StatusBadRequest)
		return
	}

	key := readKey(r)

	KID := util.StringToID(key)

	s, err := n.findKeySuccessor(KID)
	if err != nil {
		http.Error(w, Internal, http.StatusInternalServerError)
		return
	}
	if s.ID.IsEqual(n.ID) {
		n.putValue(key, body)
	} else {
		err = n.sendToSuccessor(key, string(body), s)
		if err != nil {
			// TODO: Notify the actual error in some way
			http.Error(w, Internal, http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (n *Node) getKey(w http.ResponseWriter, r *http.Request) {
	key := readKey(r)

	val, err := n.getValue(key)
	if err == ErrNotFound {
		util.ErrorNotFound(w, "Key %s not found", key)
		return
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

func (n *Node) unlinkNext() error {
	if n.next.ID.IsEqual(n.ID) {
		return nil
	}
	return netutils.CloseRPC(n.next.conn)
}

func (n *Node) unlinkPrev() error {
	if n.prev.ID.IsEqual(n.ID) {
		return nil
	}
	return netutils.CloseRPC(n.prev.conn)
}
func (n *Node) setPredecessor(id util.Identifier, ip string) error {
	n.nMu.Lock()
	c, err := netutils.ConnectRPC(ip)
	if err != nil {
		n.nMu.Unlock()
		return err
	}

	err = n.unlinkPrev()
	if err != nil {
		n.nMu.Unlock()
		return err
	}

	n.prev = Neighbor{
		ID:   id,
		IP:   ip,
		conn: c,
	}
	n.nMu.Unlock()
	return nil
}

func (n *Node) setSuccessor(id util.Identifier, ip string) error {
	n.nMu.Lock()
	if id.IsEqual(n.ID) {
		return nil
	}

	c, err := netutils.ConnectRPC(ip)
	if err != nil {
		n.nMu.Unlock()
		return err
	}
	err = n.unlinkNext()
	if err != nil {
		n.nMu.Unlock()
		return err
	}

	n.next = Neighbor{
		ID:   id,
		IP:   ip,
		conn: c,
	}
	n.nMu.Unlock()
	return nil
}

// JoinNetwork registers itself in the network by asking a random
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
