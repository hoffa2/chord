package client

import (
	"log"
	"net/http"
	"time"

	"github.com/hoffa2/chord/netutils"
	"github.com/urfave/cli"
)

var deftests = 10000

func Run(c *cli.Context) error {
	tests := c.Int("tests")
	if tests == 0 {
		tests = deftests
	}
	workers := c.Int("threads")

	nameServerAddr := c.String("nameserver")
	log.Printf("Address of nameserver: %s\n", nameServerAddr)
	client := &Client{
		nameServer: nameServerAddr,
		results:    make(chan time.Duration, tests*2),
		nkeys:      tests,
		workers:    workers,
		keyvalues:  make(map[string]string),
		errors:     make(chan error, tests*2),
		conn: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 10000,
			},
		},
	}
	for i := 0; i < client.nkeys; i++ {
		key := RandStringBytes(30)
		client.keyvalues[key] = key
	}
	ips, err := netutils.GetNodeIPs(nameServerAddr + ":" + "8030")
	if err != nil {
		return err
	}
	client.IPs = ips
	for i := 100; i <= 1000; i += 100 {
		client.workers = i
		client.RunTests(i)
	}
	return nil
}
