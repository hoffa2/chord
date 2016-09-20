package node

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/hoffa2/chord/netutils"
	"github.com/hoffa2/chord/util"
	"github.com/urfave/cli"
)

func Run(c *cli.Context) error {
	port := c.String("port")
	if port == "" {
		port = "8030"
	}
	NameServerAddr := c.String("nameserver")

	r := mux.NewRouter()

	n, err := os.Hostname()
	if err != nil {
		return err
	}

	client := http.Client{
		Timeout: time.Duration(time.Second * 2),
	}

	node := &Node{
		nameServer:  NameServerAddr,
		IP:          n,
		id:          util.StringToID(util.HashValue(n)),
		objectStore: make(map[string]string),
		conn:        client,
	}

	go netutils.SetupRPCServer("8001", node)

	err = JoinNetwork(node, n)
	if err != nil {
		return err
	}

	// Registering the put and get methods
	r.HandleFunc("/{key}", node.getKey).Methods("GET")
	r.HandleFunc("/{key}", node.putKey).Methods("PUT")

	return http.ListenAndServe(":"+port, r)
}
