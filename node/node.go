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
	"time"

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
	// Keysize size of keyspace
	KeySize         = 160
	ErrInvalidIndex = errors.New("ftable index is invalid")
	// ErrNotFound if key does not exist
	ErrNotFound = errors.New("No value on key")
	// ErrNextToSmall if the successor is too small
	ErrNextToSmall = errors.New("Successor is less than n id")
	// ErrPrevToLarge If the predecessor is too large
	ErrPrevToLarge = errors.New("Predecessor is larger than n id")
)

// Neighbor Describing an adjacent node in the ring
type Neighbor struct {
	id util.Identifier
	*netutils.NodeRPC
	IP string
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
	fingers    []FingerEntry
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
		return Neighbor{id: n.id, IP: n.IP}, nil
	}
	// TODO: Maybe we should check whether the key is in our successor's keyspace
	s, err := n.next.FindSuccessor(k)
	if err != nil {
		return Neighbor{}, err
	}
	return Neighbor{
		id: util.StringToID(s.ID),
		IP: s.IP,
	}, nil
}

func (n *Node) assertPlacement(key string) {
	k := util.StringToID(key)
	if !k.InKeySpace(n.prev.id, n.id) {
		log.Printf("%s should not be located on %s\n", key, n.IP)
	}
}

func (n *Node) putValue(key string, body []byte) {
	n.mu.Lock()
	n.assertPlacement(key)
	n.objectStore[key] = string(body)
	n.mu.Unlock()
}

func (n *Node) getValue(key string) (string, error) {
	n.mu.RLock()
	n.assertPlacement(key)
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
		return n.next.NodeRPC
	} else if id.IsEqual(n.prev.id) {
		return n.prev.NodeRPC
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s.id.IsEqual(n.id) {
		n.putValue(key, body)
	} else {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s.id.IsEqual(n.id) {
		val, err = n.getValue(key)
	} else {
		val, err = n.getFromSuccessor(key, s)
	}
	if err == ErrNotFound {
		util.ErrorNotFound(w, "Key %s not found", key)
		return
	} else if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	if n.next.NodeRPC != nil {
		return netutils.CloseRPC(n.next.NodeRPC)
	}
	return nil
}

func (n *Node) unlinkPrev() error {
	if n.prev.id.IsEqual(n.id) {
		return nil
	}
	if n.prev.NodeRPC != nil {
		return netutils.CloseRPC(n.prev.NodeRPC)
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
		id:      id,
		IP:      ip,
		NodeRPC: c,
	}
	n.nMu.Unlock()
	return nil
}

func (n *Node) setSuccessor(id util.Identifier, ip string) error {
	n.nMu.Lock()
	if id.IsEqual(n.id) {
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
		id:      id,
		IP:      ip,
		NodeRPC: c,
	}

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
		n.setSuccessor(n.id, n.IP)
		//go n.reportState()
		return nil
	}

	rnode = n.getRandomNode(nodes)

	// Asking the first node who's my successor
	c, err := netutils.ConnectRPC(rnode)
	if err != nil {
		return err
	}

	succ, err := c.FindSuccessor(n.id)
	if err != nil {
		return err
	}

	err = n.setSuccessor(util.StringToID(succ.ID), succ.IP)
	if err != nil {
		return err
	}

	pre, err := n.next.FindPredecessor(n.id)
	if err != nil {
		return err
	}

	err = n.next.UpdatePredecessor(n.id, n.IP)
	if err != nil {
		return err
	}

	err = n.setPredecessor(util.StringToID(pre.ID), pre.IP)
	if err != nil {
		return err
	}

	err = n.prev.UpdateSuccessor(n.id, n.IP)
	if err != nil {
		return err
	}
	fmt.Printf("Node joined. (%s <- %s -> %s)\n", n.prev.IP, n.IP, n.next.IP)
	return netutils.CloseRPC(c)
}

func (n *Node) reportState() {
	for {
		time.Sleep(time.Second * 2)
		log.Printf("State (%s): (%s:%s)\n", n.IP, n.prev.IP, n.next.IP)

	}
}

func (n *Node) initFTable() {
	for i := 1; i < KeySize; i++ {
		n.fingers[i-1].start = n.id.CalculateStart(int64(i), int64(KeySize))
	}

}

func (n *Node) updateFTable(id, ip string, idx int) error {
	if idx >= KeySize {
		return ErrInvalidIndex
	}
	n.nMu.Lock()
	n.fingers[idx].ip = ip
	n.fingers[idx].succ = util.StringToID(id)
	n.nMu.Unlock()

	return nil
}
