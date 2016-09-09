package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	if len(os.Args) < 2 {
		panic(fmt.Sprintf("Usage: %s port\n", os.Args[0]))
	}

	port := os.Args[0]

	r := mux.NewRouter()

	node := &Node{}

	// Registering the put and get methods
	r.HandleFunc("/{key}", node.GetKey).Methods("GET")
	r.HandleFunc("/{key}", node.PutKey).Methods("PUT")

	http.ListenAndServe(":"+port, r)
}
