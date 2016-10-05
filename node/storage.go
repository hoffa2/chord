package node

import (
	"io/ioutil"
	"net/http"

	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/util"
)

func (n *Node) putValue(key util.Identifier, body []byte) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.objectStore[key.ToString()] = string(body)
	n.log.Info.Printf("Putting key %s on node %s\n", string(body), n.IP)
	if !key.InKeySpace(n.prev.ID, n.ID) {
		n.log.Err.Printf("Key %s is not in %s's keyspace\n", key.ToString(), n.IP)
	}
}

func (n *Node) getValue(key util.Identifier) (string, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	val, ok := n.objectStore[key.ToString()]
	if !ok {
		return "", ErrNotFound
	}

	if !key.InKeySpace(n.prev.ID, n.ID) {
		n.log.Err.Printf("Key %s is not in %s's keyspace\n", key.ToString(), n.IP)
	}
	return val, nil
}

func (n Node) sendToSuccessor(key, val string, s *comm.Rnode) error {
	var err error

	err = n.remote.PutRemote(*s, key, val)
	if err != nil {
		return err
	}
	return nil
}

func (n Node) getFromSuccessor(key string, s *comm.Rnode) (string, error) {
	var err error

	val, err := n.remote.GetRemote(*s, key)
	if err != nil {
		return "", err
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

	KID := util.StringToID(util.HashValue(key))

	s, err := n.findKeySuccessor(KID)
	if err != nil {
		n.log.Err.Printf("Could not find %s's successor\n", key)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s.ID.IsEqual(n.ID) {
		n.putValue(KID, body)
	} else {
		err = n.sendToSuccessor(KID.ToString(), string(body), s)
		if err != nil {
		n.log.Err.Printf("Could not find %s's successor\n", key)
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
	KID := util.StringToID(util.HashValue(key))

	s, err := n.findKeySuccessor(KID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s.ID.IsEqual(n.ID) {
		n.log.Info.Printf("Getting key %s from node %s\n", key, n.IP)
		val, err = n.getValue(KID)
	} else {
		n.log.Info.Printf("Getting key (%s) from %s\n", key, n.fingers[0].node.IP)
		val, err = n.getFromSuccessor(KID.ToString(), s)
	}
	if err == ErrNotFound {
		util.ErrorNotFound(w, "Key %s not found", key)
		return
	} else if err != nil {
		n.log.Err.Println(err.Error())
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
