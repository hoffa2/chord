package node

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/urfave/cli"
)

func Run(c *cli.Context) error {
	if c.NArg() < 2 {
		panic(fmt.Sprintf("Usage: node port nameserver\n"))
	}

	port := c.String("port")
	NameServerAddr := c.String("namserver")

	r := mux.NewRouter()

	node := &Node{nameServer: NameServerAddr}

	// Registering the put and get methods
	r.HandleFunc("/{key}", node.GetKey).Methods("GET")
	r.HandleFunc("/{key}", node.PutKey).Methods("PUT")

	return http.ListenAndServe(":"+port, r)
}
