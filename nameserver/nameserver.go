package nameserver

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/urfave/cli"
	"github.com/hoffa2/chord/util"
)

const (
	Ip   = "ip"
	NoIp = "No IpAddress present in query"
)

type NameServer struct {
	IpAdresses []string
}

func Run(c *cli.Context) error {
	port := c.String("port")
	if port == "" {
		port = "8000"
	}

	ns := &NameServer{}

	r := mux.NewRouter()
	r.HandleFunc("/", ns.GetNodeList).Methods("GET")
	r.HandleFunc("/", ns.registerNode).Methods("PUT")

	return http.ListenAndServe(":"+port, r)
}

func (n *NameServer) registerNode(w http.ResponseWriter, r *http.Request) {
	ip := r.PostForm.Get(Ip)
	if len(Ip) == 0 {
		util.ErrorResponse(w, NoIp)
	}

	n.IpAdresses = append(n.IpAdresses, ip)

	w.WriteHeader(http.StatusOK)
}

func (n *NameServer) GetNodeList(w http.ResponseWriter, r *http.Request) {
	util.WriteJson(w, n.IpAdresses)
}
