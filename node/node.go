package node

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/hoffa2/chord/util"
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
	nameServer  string
	conn        http.Client
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

func (n *Node) RegisterWithNameServer() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("PUT", n.nameServer, []byte(hostname))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := n.conn.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Could not register with nameserver: %s", n.nameServer)
	}
}
