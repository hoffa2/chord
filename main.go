package main

import (
	"errors"
	"log"
	"os"

	"github.com/hoffa2/chord/client"
	"github.com/hoffa2/chord/launch"
	"github.com/hoffa2/chord/nameserver"
	"github.com/hoffa2/chord/node"
	"github.com/urfave/cli"
)

var (
	portNotSet = errors.New("Port is not set")
)

func main() {
	app := cli.NewApp()
	app.Name = "Key-Value Store"
	app.Usage = "Run one of the components"

	app.Commands = []cli.Command{
		{
			Name:  "node",
			Usage: "run node",
			Action: func(c *cli.Context) error {
				if !c.IsSet("nameserver") {
					return errors.New("Nameserver flag must be set")
				}
				return node.Run(c)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "port, p",
					Usage: "Specify port",
				},
				cli.StringFlag{
					Name:  "nameserver, ns",
					Usage: "address of nameserver",
				},
				cli.StringFlag{
					Name:  "id",
					Usage: "optional id",
				},
				cli.IntFlag{
					Name:  "graph",
					Usage: "graph on or of (1/0)",
				},
			},
		},
		{
			Name:  "client",
			Usage: "run client",
			Action: func(c *cli.Context) error {
				if !c.IsSet("nameserver") {
					return errors.New("Nameserver flag must be set")
				}
				return client.Run(c)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "nameserver, ns",
					Usage: "address of nameserver",
				},
				cli.IntFlag{
					Name:  "tests",
					Usage: "Number of queries issued",
				},
				cli.IntFlag{
					Name:  "threads",
					Usage: "Number of concurrent requests",
				},
			},
		},
		{
			Name:  "nameserver",
			Usage: "run nameserver",
			Action: func(c *cli.Context) error {
				return nameserver.Run(c)
			},
		},
		{
			Name:  "runall",
			Usage: "Run all components together",
			Action: func(c *cli.Context) error {
				return launch.Run(c)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "nameserver, ns",
					Usage: "address of nameserver",
				},
				cli.StringFlag{
					Name:  "hosts",
					Usage: "number of hosts to run",
				},
				cli.IntFlag{
					Name:  "graph",
					Usage: "0/1",
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
