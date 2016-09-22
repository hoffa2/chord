package client

import (
	"math/rand"
	"net/http"
)

type Client struct {
	// IP addresses to chord nodes
	IPs []string
	// Ip address of nameserver
	nameServer string
	// Connection object - used for all http interaction
	conn http.Client
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func putKey() {

}

func getKey() {

}

func (c *Client) RunTests() {

}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
