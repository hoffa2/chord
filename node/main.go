package node

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/hoffa2/chord/netutils"
	"github.com/hoffa2/chord/util"
	"github.com/tylerb/graceful"
	"github.com/urfave/cli"
)

// Run Runs a chord node
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
		fingers:     make([]FingerEntry, KeySize),
	}
	l, err := netutils.SetupRPCServer("8010", node)
	if err != nil {
		return err
	}
	defer l.Close()

	err = JoinNetwork(node, n)
	if err != nil {
		return err
	}

	// Registering the put and get methods
	r.HandleFunc("/{key}", node.getKey).Methods("GET")
	r.HandleFunc("/{key}", node.putKey).Methods("PUT")

	graceful.Run(":"+port, time.Microsecond*100, r)
	return nil
}
