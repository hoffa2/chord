package client

import (
	"encoding/json"
	"log"
	"net/http"
)

type Client struct {
	IPs        []string
	nameServer string
	conn       http.Client
}

func (c *Client) getIPAddrs() {
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

	err := json.NewDecoder(resb.Body).Decode(c.IPs)
	if err != nil {
		log.Fatal(err)
	}
}

func putKey() {

}

func getKey() {

}
