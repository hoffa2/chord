package node

import (
	"io/ioutil"
	"net/http"

	"github.com/hoffa2/chord/util"

	"github.com/gorilla/mux"
)

const (
	NoValue  = "No value in body"
	NotFound = "No value on key: %s"
)

// Interface struct that represents the state
// of one node
type Node struct {
	// Storing key-value pairs on the respective node
	objectStore map[string]string
}

func readKey(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["key"]
}

func (n *Node) PutKey(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, NoValue, http.StatusBadRequest)
	}

	key := readKey(r)

	n.objectStore[key] = string(body)

	w.WriteHeader(http.StatusOK)
}

func (n *Node) GetKey(w http.ResponseWriter, r *http.Request) {
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
