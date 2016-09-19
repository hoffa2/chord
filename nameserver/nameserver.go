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
}

func Run(c *cli.Context) error {
	port := c.String("port")
	if port == "" {
		port = "8030"
	}

	ns := &NameServer{}

	r := mux.NewRouter()
	r.HandleFunc("/", ns.GetNodeList).Methods("GET")
	r.HandleFunc("/", ns.registerNode).Methods("POST")
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

func (n *NameServer) GetNodeList(w http.ResponseWriter, r *http.Request) {
	n.mu.RLock()
	util.WriteJson(w, n.IpAdresses)
	n.mu.RUnlock()
}
