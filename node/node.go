package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/netutils"
	"github.com/hoffa2/chord/util"
)

func readKey(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["key"]
}

func (n Node) findKeySuccessor(k util.Identifier) (*comm.Rnode, error) {
	// I'm the successor
	if k.InKeySpace(n.prev.ID, n.ID) {
		return n.Rnode, nil
	}
	// TODO: Maybe we should check whether the key is in our successor's keyspace
	s, err := n.findSuccessor(k)
	if err != nil {
		n.log.Err.Printf("Unable to locate successor on key: %s", err.Error())
		return nil, err
	}
	return s, nil
}

func (n Node) registerNode() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	resp, err := n.conn.PostForm(fmt.Sprintf("http://%s/", n.nameServer),
		url.Values{"ip": {hostname}})
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error: %d; Could not register with nameserver: %s",
			resp.StatusCode, n.nameServer)
	}
	return nil
}

func (n *Node) setPredecessor(rn *comm.Rnode) error {
	n.nMu.Lock()
	n.prev = rn
	n.nMu.Unlock()
	return nil
}

func (n *Node) setSuccessor(rn *comm.Rnode) error {
	n.nMu.Lock()
	n.fingers[0].node = rn
	if !rn.ID.IsEqual(n.ID) {
		n.successors = append([]comm.Rnode{*rn}, n.successors...)
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
		n.setSuccessor(n.Rnode)
		n.setPredecessor(n.Rnode)
		n.initFTable(true)
		n.addNodeToGraph()
		go n.periodicRun()
		go n.pushState()
		return nil
	}

	n.initFTable(false)
	node := n.getRandomNode(nodes)
	rnode := &comm.Rnode{IP: node}
	n.log.Info.Println("Came here")
	succ, err := n.remote.FindSuccessor(*rnode, n.ID)
	if err != nil {
		return err
	}
	n.setSuccessor(succ)
	n.setPredecessor(n.Rnode)
	go n.periodicRun()
	n.addNodeToGraph()
	go n.pushState()
	return nil
}

func (n *Node) retrieveKeys() error {
	keys, err := n.remote.GetKeysInInterval(*n.prev, n.prev.ID, n.ID)
	if err != nil {
		return err
	}
	n.mu.Lock()
	defer n.mu.Unlock()

	for k, v := range *keys {
		n.objectStore[k] = v
	}

	return nil
}

func (n *Node) initFTable(alone bool) {
	for i := 0; i < KeySize; i++ {
		n.fingers[i].start = n.ID.PowMod(int64(i), int64(KeySize))
		if alone {
			n.fingers[i].node = n.Rnode
		}
	}
}

func (n *Node) reportState() {
	for {
		time.Sleep(time.Second * 5)
		log.Printf("State (%s): (%s:%s)\n", n.IP, n.prev.IP, n.fingers[0].node.IP)
	}
}

func (n *Node) state(w http.ResponseWriter, r *http.Request) {
	p := &struct {
		IP         string
		Next       string
		Prev       string
		Successors []comm.Rnode
	}{
		n.IP,
		n.fingers[0].node.IP,
		n.prev.IP,
		n.successors,
	}
	util.WriteJson(w, p)
}

func (n *Node) findPredecessor(id util.Identifier) (*comm.Rnode, error) {
	var tnode *comm.Rnode
	var succ *comm.Rnode
	var err error

	tnode = n.Rnode
	succ = n.fingers[0].node

	if id.InKeySpace(tnode.ID, succ.ID) {
		return tnode, nil
	}

	// Checking if n is id's predecessor
	for !id.InKeySpace(tnode.ID, succ.ID) {

		if tnode.ID.IsEqual(n.ID) {
			tnode = n.closestPreFinger(id)
		} else {
			tnode, err = n.remote.ClosestPreFinger(*tnode, id)
			if err == netutils.ErrTimeout {
				tnode = n.skipClosestFinger(tnode, id)
			} else if err != nil {
				return nil, err
			}
		}
		if tnode.ID.IsEqual(n.ID) {
			succ = n.fingers[0].node
		} else {
			succ, err = n.remote.GetSuccessor(*tnode)
			if err == netutils.ErrTimeout {
				tnode = n.skipClosestFinger(tnode, id)
			} else if err != nil {
				return nil, err
			}
		}
	}
	return tnode, nil
}

func (n *Node) skipClosestFinger(rn *comm.Rnode, id util.Identifier) *comm.Rnode {
	var cf *comm.Rnode
	for i, succ := range n.successors {
		if succ.ID.IsEqual(rn.ID) {
			if i == len(n.successors)-1 {
				cf = &n.successors[i-1]
				break
			} else if i < len(n.successors)-1 {
				cf = &n.successors[i+1]
				break
			}
		}
	}
	if cf == nil {
		return &n.successors[0]
	}
	return cf
}

// TODO: replay query to closest predeceding node
func (n *Node) findSuccessor(id util.Identifier) (*comm.Rnode, error) {
	pre, err := n.findPredecessor(id)
	if err != nil {
		return nil, err
	}
	succ, err := n.remote.GetSuccessor(*pre)
	if err != nil {
		return nil, err
	}
	return succ, nil
}

// Finding closest predeceeding finger
// TODO: Iterate successor list
func (n *Node) closestPreFinger(id util.Identifier) *comm.Rnode {
	for i := KeySize - 1; i >= 0; i-- {
		if n.fingers[i].node != nil && n.fingers[i].node.ID.IsBetween(n.ID, id) {
			return n.fingers[i].node
		}
	}
	return n.Rnode
}

func (n *Node) migrateKeys(from, to string) map[string]string {
	n.mu.Lock()
	defer n.mu.Unlock()
	mk := make(map[string]string)

	fromID := util.StringToID(from)
	toID := util.StringToID(to)

	for k, v := range n.objectStore {
		if util.StringToID(k).InKeySpace(fromID, toID) {
			mk[k] = v
			delete(n.objectStore, k)
		}
	}
	return mk
}

//
func (n *Node) notify(rn *comm.Rnode) {
	if n.prev.ID.IsEqual(n.ID) || rn.ID.IsBetween(n.prev.ID, n.ID) {
		n.setPredecessor(rn)
	}
	if alive, _ := n.remote.IsAlive(*n.prev); !alive {
		n.setPredecessor(rn)
	}

	if n.fingers[0].node.ID.IsEqual(n.ID) {
		n.setSuccessor(rn)
	}
	if len(n.successors) == 1 && n.successors[0].ID.IsEqual(n.ID) {
		n.setSuccessor(rn)
	}

	//TODO: Handle key replication
}

// fixFinger
func (n *Node) fixFinger() {
	idx := rand.Intn(KeySize-1) + 1
	newSucc, err := n.findSuccessor(n.fingers[idx].start)
	if err != nil {
		log.Println(err)
	}

	n.fingers[idx].node = newSucc
	n.updateSuccessors(newSucc)
}

// stabilize
func (n *Node) stabilize() {
	// check if successor has died
	var temp *comm.Rnode
	var err error

	if n.fingers[0].node.ID.IsEqual(n.ID) {
		return
	}
	successor := n.successors[0]

	// Try to query the first successor - skip the next on failure
	for temp, err = n.remote.GetPredecessor(successor); ; {
		if err != nil {
			n.log.Err.Println(err)
			successor, err = n.skipSuccessor(&successor)
			if err == ErrExhausted {
				n.setSuccessor(n.Rnode)
				return
			}
		} else {
			break
		}
	}
	if temp == nil {
		return
	}

	if temp.ID.IsBetween(n.ID, successor.ID) ||
		n.fingers[0].node.ID.IsEqual(n.ID) {
		n.setSuccessor(temp)

	}
	n.remote.Notify(successor, n.Rnode)

	n.checkSuccessors()

	n.fixFinger()
}

func (n *Node) checkSuccessors() {
	for i := 0; i < len(n.successors)-1; i++ {
		s, err := n.remote.GetSuccessor(n.successors[i])
		if err == netutils.ErrTimeout {
			n.log.Err.Println(err)
		} else if err != nil {
			n.successors = append(n.successors[i:], n.successors[i+1:]...)
		} else {
			n.successors[i+1] = *s
		}
	}
}

// Adds a successor to the successor list
func (n *Node) updateSuccessors(nsucc *comm.Rnode) error {
	n.nMu.Lock()
	defer n.nMu.Unlock()

	if nsucc.ID.IsEqual(n.ID) {
		return nil
	}

	found := false
	for _, succ := range n.successors {
		if succ.ID.IsEqual(nsucc.ID) {
			found = true
			break
		}
	}
	if !found {
		if len(n.successors) == 0 {
			n.successors = append(n.successors, *nsucc)
		} else {
			for i, succ := range n.successors {
				if nsucc.ID.IsBetweenEqual(succ.ID, n.successors[(i+1)%len(n.successors)].ID) {
					n.successors = append(n.successors, comm.Rnode{})
					copy(n.successors[i+1:], n.successors[i:])
					n.successors[i+1] = *nsucc
				}
			}
		}
	}
	return nil
}

// Skips to the next successor if the nearmost successor has failed
func (n *Node) skipSuccessor(s *comm.Rnode) (comm.Rnode, error) {
	if !n.successors[0].ID.IsEqual(s.ID) {
		return comm.Rnode{}, ErrNotFirst
	}
	n.successors = append(n.successors[:0], n.successors[1:]...)
	if len(n.successors) == 0 {
		return comm.Rnode{}, ErrExhausted
	}
	n.log.Info.Printf("Skipping successor to %s\n", n.successors[0].IP)
	n.fingers[0].node = &n.successors[0]
	return n.successors[0], nil
}

// Callback
func (n *Node) failhandler(rn *comm.Rnode) {

}

func (n *Node) periodicRun() {
	for {
		time.Sleep(time.Second * 1)
		n.stabilize()
	}
}

func (n *Node) pushState() {

	for {
		b := n.createState()
		n.sendState("update", b)
		time.Sleep(time.Millisecond * 300)
	}
}

func (n *Node) createState() bytes.Buffer {
	state := &struct {
		Next string
		ID   string
		Prev string
	}{
		n.fingers[0].node.IP,
		n.IP,
		n.prev.IP,
	}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(state)
	if err != nil {
		n.log.Err.Println(err)
	}
	return *b
}

func (n *Node) sendState(method string, b bytes.Buffer) {
	url := fmt.Sprintf("http://%s/%s", n.graphIP, method)
	n.log.Err.Println(b.String())
	_, err := n.conn.Post(url, "application/json", &b)
	if err != nil {
		n.log.Err.Println(err)
	}

}

func (n *Node) leave() {
	b := n.createState()
	n.log.Info.Println(b)
	n.sendState("remove", b)
}

func (n *Node) addNodeToGraph() {
	b := n.createState()
	n.sendState("add", b)
}
