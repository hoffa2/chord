package node

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
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
	id   util.Identifier
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
	id util.Identifier
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
	if k.InKeySpace(n.prev.id, n.id) {
		return Neighbor{id: n.id}, nil
	}
	fmt.Printf("Looking for %s on node %s\n", string(k), n.next.IP)
	// TODO: Maybe we should check whether the key is in our successor's keyspace
	s, err := n.next.conn.FindSuccessor(k)
	if err != nil {
		return Neighbor{}, nil
	}
	fmt.Printf("Found %s's successor: %s\n", string(k), s.IP)
	return Neighbor{
		id: util.StringToID(s.ID),
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
	if id.IsEqual(n.next.id) {
		return n.next.conn
	} else if id.IsEqual(n.prev.id) {
		return n.prev.conn
	}
	return nil
}

func (n Node) sendToSuccessor(key, val string, s Neighbor) error {
	var err error
	close := false
	c := n.findExistingConn(s.id)

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

func (n Node) getFromSuccessor(key string, s Neighbor) (string, error) {
	var err error
	close := false
	c := n.findExistingConn(s.id)

	// if we got an unfamiliar node
	if c == nil {
		c, err = netutils.ConnectRPC(s.IP)
		if err != nil {
			return "", err
		}
		close = true
	}

	val, err := c.GetRemote(key)
	if err != nil {
		return "", err
	}

	// if we got an unknown node - close connection
	if close {
		return val, netutils.CloseRPC(c)
	}

	return val, nil
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
	if s.id.IsEqual(n.id) {
		fmt.Printf("Putting key %s on node %s\n", key, n.IP)
		n.putValue(key, body)
	} else {
		fmt.Printf("sending key %s to %s\n", key, s.IP)
		err = n.sendToSuccessor(key, string(body), s)
		if err != nil {
			// TODO: Notify the actual error in some way
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (n *Node) getKey(w http.ResponseWriter, r *http.Request) {
	var val string
	var err error
	key := readKey(r)

	s, err := n.findKeySuccessor(util.StringToID(key))
	if err != nil {
		http.Error(w, Internal, http.StatusInternalServerError)
		return
	}
	fmt.Printf("Successor(%s) is %s\n", key, s.IP)
	if s.id.IsEqual(n.id) {
		val, err = n.getValue(key)
	} else {
		fmt.Printf("Getting key %s from %s\n", key, s.IP)
		val, err = n.getFromSuccessor(key, s)
	}
	if err == ErrNotFound {
		util.ErrorNotFound(w, "Key %s not found", key)
		return
	} else if err != nil {
		log.Println(err.Error())
		http.Error(w, Internal, http.StatusInternalServerError)
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

	resp, err := n.conn.PostForm(fmt.Sprintf("http://%s/", n.nameServer), url.Values{"ip": {hostname}})
	if err != nil {
		log.Println("error when connectiong to nameserver")
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error: %d; Could not register with nameserver: %s", resp.StatusCode, n.nameServer)
	}
	return nil
}

func (n *Node) unlinkNext() error {
	if n.next.id.IsEqual(n.id) {
		return nil
	}
	if n.next.conn != nil {
		return netutils.CloseRPC(n.next.conn)
	}
	return nil
}

func (n *Node) unlinkPrev() error {
	if n.prev.id.IsEqual(n.id) {
		return nil
	}
	if n.prev.conn != nil {
		return netutils.CloseRPC(n.prev.conn)
	}
	return nil
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
		id:   id,
		IP:   ip,
		conn: c,
	}
	fmt.Printf("Node %s updated pre to %s\n", n.IP, ip)
	n.nMu.Unlock()
	return nil
}

func (n *Node) setSuccessor(id util.Identifier, ip string) error {
	n.nMu.Lock()
	if id.IsEqual(n.id) {
		fmt.Println("I'm alone")
		n.nMu.Unlock()
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
		id:   id,
		IP:   ip,
		conn: c,
	}

	fmt.Printf("Node %s updated next to %s\n", n.IP, ip)
	n.nMu.Unlock()
	return nil
}

func (n Node) getRandomNode(nodes []string) string {
	for _, node := range nodes {
		if node != n.IP {
			return node
		}
	}
	return ""
}

// JoinNetwork registers itself in the network by asking a random
// node found by issuing a request to the nameserver
func JoinNetwork(n *Node, id string) error {
	var rnode string
	err := n.registerNode()
	if err != nil {
		return err
	}

	nodes, err := netutils.GetNodeIPs(n.nameServer)
	if err != nil {
		return err
	}

	if len(nodes) == 1 && nodes[0] == n.IP {
		// I'm alone!!!
		n.setSuccessor(util.StringToID(id), n.IP)
		return nil
	}

	rnode = n.getRandomNode(nodes)

	// Asking the first node who's my successor
	c, err := netutils.ConnectRPC(rnode)
	if err != nil {
		return err
	}

	fmt.Printf("%s asks %s\n", n.IP, rnode)
	succ, err := c.FindSuccessor(n.id)
	if err != nil {
		return err
	}

	fmt.Printf("REturned %s\n", succ.IP)
	err = n.setSuccessor(util.StringToID(succ.ID), succ.IP)
	if err != nil {
		return err
	}

	fmt.Printf("%s asks %s\n", n.IP, n.next.IP)
	pre, err := n.next.conn.FindPredecessor(n.id)
	if err != nil {
		return err
	}

	fmt.Printf("RETURNED %s\n", succ.IP)
	err = n.next.conn.UpdatePredecessor(n.id, n.IP)
	if err != nil {
		return err
	}

	err = n.setPredecessor(util.StringToID(pre.ID), pre.IP)
	if err != nil {
		return err
	}

	err = n.prev.conn.UpdateSuccessor(n.id, n.IP)
	if err != nil {
		return err
	}
	fmt.Printf("Node: %s joined network; Pre = %s, Next = %s\n", n.IP, n.prev.IP, n.next.IP)
	return netutils.CloseRPC(c)
}
