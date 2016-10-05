package nameserver

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/hoffa2/chord/util"
	"github.com/urfave/cli"
)

const (
	Ip   = "ip"
	NoIp = "No IpAddress present in query"
)

type NameServer struct {
	IpAdresses []string
	mu         sync.RWMutex
	states     []NodeState
}

type NodeState struct {
	Next string
	ID   string
	Prev string
}

func Run(c *cli.Context) error {
	port := c.String("port")
	if port == "" {
		port = "8030"
	}

	ns := &NameServer{}

	r := mux.NewRouter()
	r.HandleFunc("/", ns.GetNodeList).Methods("GET")
	r.HandleFunc("/unregister", ns.unRegister).Methods("POST")
	r.HandleFunc("/", ns.registerNode).Methods("POST")
	r.HandleFunc("/nodes", ns.getNodeState).Methods("GET")
	return http.ListenAndServe(":"+port, r)
}

func (n *NameServer) registerNode(w http.ResponseWriter, r *http.Request) {
	ip := r.PostFormValue("ip")
	if len(ip) == 0 {
		util.ErrorResponse(w, NoIp)
		return
	}
	n.mu.Lock()
	n.IpAdresses = append(n.IpAdresses, ip)
	n.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (n *NameServer) unRegister(w http.ResponseWriter, r *http.Request) {
	ip := r.PostFormValue("ip")
	if len(ip) == 0 {
		util.ErrorResponse(w, NoIp)
		return
	}
	n.mu.Lock()
	defer n.mu.Lock()

	for i, node := range n.IpAdresses {
		if node == ip {
			n.IpAdresses = append(n.IpAdresses[:i], n.IpAdresses[i+1:]...)
			break
		}
	}
}

func (n *NameServer) GetNodeList(w http.ResponseWriter, r *http.Request) {
	n.mu.RLock()
	util.WriteJson(w, n.IpAdresses)
	n.mu.RUnlock()
}

func (n *NameServer) getNodeState(w http.ResponseWriter, r *http.Request) {
	n.mu.RLock()
	util.WriteJson(w, n.states)
	n.mu.RUnlock()
}
