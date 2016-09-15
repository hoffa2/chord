package node

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/hoffa2/chord/netutils"
	"github.com/hoffa2/chord/util"
	"github.com/urfave/cli"
)

func Run(c *cli.Context) error {
	port := c.String("port")
	if port == "" {
		port = "8000"
	}
	NameServerAddr := c.String("namserver")

	r := mux.NewRouter()

	n, err := os.Hostname()
	if err != nil {
		return err
	}

	node := &Node{
		nameServer:  NameServerAddr,
		ID:          util.StringToID(n),
		objectStore: make(map[string]string),
	}

	netutils.SetupRPCServer("8001", node)

	err = node.JoinNetwork(n)
	if err != nil {
		return err
	}

	// Registering the put and get methods
	r.HandleFunc("/{key}", node.getKey).Methods("GET")
	r.HandleFunc("/{key}", node.putKey).Methods("PUT")

	return http.ListenAndServe(":"+port, r)
}
