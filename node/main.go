package node

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/urfave/cli"
)

func Run(c *cli.Context) error {
	port := c.String("port")
	if port == "" {
		port = "8000"
	}

	NameServerAddr := c.String("namserver")

	r := mux.NewRouter()

	node := &Node{nameServer: NameServerAddr}

	// Registering the put and get methods
	r.HandleFunc("/{key}", node.GetKey).Methods("GET")
	r.HandleFunc("/{key}", node.PutKey).Methods("PUT")

	return http.ListenAndServe(":"+port, r)
}
