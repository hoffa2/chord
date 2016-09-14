package client

import "net/http"

type Client struct {
	// IP addresses to chord nodes
	IPs []string
	// Ip address of nameserver
	nameServer string
	// Connection object - used for all http interaction
	conn http.Client
}

func putKey() {

}

func getKey() {

}
