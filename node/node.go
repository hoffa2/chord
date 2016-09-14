package node

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/netutils"
	"github.com/hoffa2/chord/util"
)

const (
	Keysize  = 12
	NoValue  = "No value in body"
	NotFound = "No value on key: %s"
)

type Neighbor struct {
	ID   util.Identifier
	conn *netutils.NodeComm
}

// Interface struct that represents the state
// of one node
type Node struct {
	// Storing key-value pairs on the respective node
	objectStore map[string]string
	ID          util.Identifier
	nameServer  string
	conn        http.Client
	finger      []FingerEntry
	next        Neighbor
	prev        Neighbor
}

func readKey(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["key"]
}

func (n Node) findKeySuccessor(k string) string {
	return ""
}

func (n *Node) putKey(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, NoValue, http.StatusBadRequest)
	}

	key := readKey(r)

	kID := util.HashValue(key)

	n.findKeySuccessor(kID)

	n.objectStore[key] = string(body)

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
func (n *Node) FindSuccessor(args *comm.Args, reply *comm.NodeID) error {
	key := args.ID
	var succ util.Identifier
	// If node is the only one in the ring
	if n.ID.IsEqual(n.next.ID) {
		reply.ID = n.ID
		return nil
	}

	if key.InKeySpace(n.prev.ID, n.ID) {
		succ = n.ID
	} else if key.IsLess(n.next.ID) && key.IsLarger(n.ID) {
		succ = n.next.ID
	} else if key.IsLarger(n.ID) {
		succ, err := n.next.conn.FindSuccessor(key)
	}
	reply.ID = succ
	return nil
}

func (n *Node) FindPredecessor(args *comm.Args, reply *comm.NodeID) error {
	return nil
}

func (n *Node) setSuccessor(id util.Identifier) {

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
		n.setSuccessor(id)
		return nil
	}

	// Asking the first node who's my successor
	c, err := netutils.ConnectRPC(nodes[0])
	if err != nil {
		return err
	}

	succ, err := c.FindSuccessor(n.ID)
	if err != nil {
		return err
	}

	n.setSuccessor(succ)

	return nil
}
