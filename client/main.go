package client

import (
	"log"
	"time"

	"github.com/hoffa2/chord/netutils"
	"github.com/urfave/cli"
)

var deftests = 1000

func Run(c *cli.Context) error {
	tests := c.Int("tests")
	if tests == 0 {
		tests = deftests
	}

	nameServerAddr := c.String("nameserver")
	log.Printf("Address of nameserver: %s\n", nameServerAddr)

	client := &Client{
		nameServer: nameServerAddr,
		results:    make(chan time.Duration, tests*2),
		nkeys:      tests,
		keyvalues:  make(map[string]string),
	}

	time.Sleep(time.Second * 10)
	ips, err := netutils.GetNodeIPs(nameServerAddr + ":" + "8030")
	if err != nil {
		return err
	}
	client.IPs = ips
	return client.RunTests()
}
