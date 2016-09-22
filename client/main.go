package client

import (
	"fmt"
	"log"
	"time"

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
	client := &Client{nameServer: nameServerAddr}
	time.Sleep(time.Second * 4)
	ips, err := netutils.GetNodeIPs(nameServerAddr + ":" + "8000")
	if err != nil {
		return err
	}
	client.RunTests()
	fmt.Println(ips)

	return nil
}
