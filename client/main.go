package client

import (
	"fmt"
	"log"

	"github.com/hoffa2/chord/netutils"
	"github.com/urfave/cli"
)

func Run(c *cli.Context) error {
	port := c.String("port")
	if port == "" {
		port = "8000"
	}

	nameServerAddr := c.String("nameserver")
	log.Printf("Address of nameserver: %s\n", nameServerAddr)
	_ = &Client{nameServer: nameServerAddr}

	ips, err := netutils.GetNodeIPs(nameServerAddr + ":" + "8000")
	if err != nil {
		return err
	}

	fmt.Println(ips)

	return nil
}
