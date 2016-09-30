package node

import (
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
	s, err := n.remote.FindSuccessor(*n.fingers[0].node, k)
	if err != nil {
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
		go n.reportState()
		return nil
	}

	n.initFTable(false)
	node := n.getRandomNode(nodes)
	rnode := &comm.Rnode{IP: node}

	err = n.setSuccFingers(rnode)
	if err != nil {
		return err
	}
	err = n.updateNeighbors()
	if err != nil {
		return err
	}

	return n.retrieveKeys()
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

func (n *Node) setSuccFingers(node *comm.Rnode) error {
	// Who's y successor?
	succ, err := n.remote.FindSuccessor(*node, n.fingers[0].start)
	if err != nil {
		return err
	}

	n.setSuccessor(succ)

	temp, _ := n.remote.GetSuccessor(*succ)
	if succ.ID.IsEqual(temp.ID) {
		err := n.remote.UpdateSuccessor(*succ, n.ID, n.IP)
		if err != nil {
			return err
		}
	}

	pre, err := n.remote.GetPredecessor(*succ)
	if err != nil {
		return err
	}

	n.setPredecessor(pre)

	err = n.remote.UpdatePredecessor(*succ, n.ID, n.IP)
	if err != nil {
		return err
	}

	// Update Fingers according to new successor
	for i := 0; i < KeySize-1; i++ {
		if n.fingers[i+1].start.InLowerInclude(n.ID, n.fingers[i].node.ID) ||
			n.fingers[i+1].start.InKeySpace(n.ID, n.fingers[i].node.ID) {
			n.fingers[i+1].node = n.fingers[i].node
		} else {
			succ, err = n.remote.FindSuccessor(*node, n.fingers[i+1].start)
			if err != nil {
				return err
			}
			n.fingers[i+1].node = succ
		}
	}
	go n.reportState()
	return nil
}

func (n *Node) updateFTable(id util.Identifier, ip string, idx int) error {
	if idx >= KeySize {
		return ErrInvalidIndex
	}
	newEntry := new(comm.Rnode)

	n.nMu.Lock()
	defer n.nMu.Unlock()

	if id.InLowerInclude(n.ID, n.fingers[idx].node.ID) {
		newEntry.IP = ip
		newEntry.ID = id
		n.fingers[idx].node = newEntry
		err := n.remote.UpdateFingerTable(*n.prev, id, ip, idx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) state(w http.ResponseWriter, r *http.Request) {
	p := &struct {
		IP   string
		Next string
		Prev string
	}{
		n.IP,
		n.prev.IP,
		n.fingers[0].node.IP,
	}
	util.WriteJson(w, p)
}

func (n *Node) updateNeighbors() error {
	for i := 0; i < KeySize; i++ {
		pre, err := n.findPredecessor(n.ID.PreID(int64(i)))
		if err != nil {
			return err
		}
		log.Println("%d", pre.ID.IsEqual(n.ID))
		err = n.remote.UpdateFingerTable(*pre, n.ID, n.IP, i)
		if err != nil {
			return err
		}
	}
	return nil
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
			if err != nil {
				return nil, err
			}
		}
		if tnode.ID.IsEqual(n.ID) {
			succ = n.fingers[0].node
		} else {
			succ, err = n.remote.GetSuccessor(*tnode)
			fmt.Println(succ)
			if err != nil {
				return nil, err
			}
		}
	}
	return tnode, nil
}

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
func (n *Node) closestPreFinger(id util.Identifier) *comm.Rnode {
	for i := KeySize - 1; i >= 0; i-- {
		if n.fingers[i].node.ID.IsBetween(n.ID, id) {
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

func (n *Node) notify(rn *comm.Rnode) {
	if (n.prev.ID.IsEqual(n.ID)) || rn.ID.IsBetween(n.prev.ID, n.ID) {
		n.setPredecessor(rn)
	}
}

func (n *Node) fixFinger() {
	idx := rand.Intn(KeySize-1) + 1
	newSucc, err := n.findSuccessor(n.fingers[idx].start)
	if err != nil {
		log.Println(err)
	}

	n.fingers[idx].node = newSucc
}

func (n *Node) stabilize() {
	temp, err := n.remote.GetPredecessor(*n.fingers[0].node)
	if err != nil {
		log.Println(err)
	}
	if temp.ID.IsBetween(n.ID, n.fingers[0].node.ID) {
		n.setSuccessor(temp)
	}
	n.remote.Notify(n.fingers[0].node, n.Rnode)
}
