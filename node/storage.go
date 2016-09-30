package node

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/util"
)

func (n *Node) assertPlacement(key string) {
	k := util.StringToID(key)
	if !k.InKeySpace(n.prev.ID, n.ID) {
		log.Printf("%s should not be located on %s\n", key, n.IP)
	}
}

func (n *Node) putValue(key util.Identifier, body []byte) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.objectStore[key.ToString()] = string(body)
}

func (n *Node) getValue(key util.Identifier) (string, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	val, ok := n.objectStore[key.ToString()]
	if !ok {
		return "", ErrNotFound
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

	KID := util.StringToID(key)

	s, err := n.findKeySuccessor(KID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s.ID.IsEqual(n.ID) {
		n.putValue(KID, body)
	} else {
		err = n.sendToSuccessor(KID.ToString(), string(body), s)
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
	KID := util.StringToID(key)

	s, err := n.findKeySuccessor(KID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s.ID.IsEqual(n.ID) {
		val, err = n.getValue(KID)
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
