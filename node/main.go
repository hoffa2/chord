package node

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func Run(args []string) error {
	if len(os.Args) < 2 {
		panic(fmt.Sprintf("Usage: %s port\n", os.Args[0]))
	}

	port := os.Args[0]
	NameServerAddr := os.Args[1]

	r := mux.NewRouter()

	node := &Node{nameServer: NameServerAddr}

	// Registering the put and get methods
	r.HandleFunc("/{key}", node.GetKey).Methods("GET")
	r.HandleFunc("/{key}", node.PutKey).Methods("PUT")

	return http.ListenAndServe(":"+port, r)
}
