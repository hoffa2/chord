package node

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/hoffa2/chord/netutils"
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

	id := createNodeID(n)

	node := &Node{
		nameServer: NameServerAddr,
		ID:         id,
	}

	netutils.SetupRPCServer("8001", node)

	// Registering the put and get methods
	r.HandleFunc("/{key}", node.getKey).Methods("GET")
	r.HandleFunc("/{key}", node.putKey).Methods("PUT")

	return http.ListenAndServe("*:"+port, r)
}
