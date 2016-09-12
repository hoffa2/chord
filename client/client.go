package client

import (
	"encoding/json"
	"log"
	"net/http"
)

type Client struct {
	// IP addresses to chord nodes
	IPs        []string
	// Ip address of nameserver
	nameServer string
	// Connection object - used for all http interaction
	conn       http.Client
}

func (c *Client) getNodeIPs() {
	req, err := http.NewRequest("GET", c.nameServer, nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := c.conn.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatal("could not retrieve ipadresses from nameserver: %s", c.nameServer)
	}

	err = json.NewDecoder(resp.Body).Decode(c.IPs)
	if err != nil {
		log.Fatal(err)
	}
}

func putKey() {

}

func getKey() {

}
