package client

import "github.com/urfave/cli"

func Run(c *cli.Context) error {
	port := c.String("port")
	if port == "" {
		port = "8000"
	}

	nameServerAddr := c.String("nameserver")

	client := &Client{nameServer: nameServerAddr}

	client.getNodeIPs()

	return nil
}

