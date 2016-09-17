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
					Name:  "pre",
					Usage: "predecessor of node",
				},
				cli.StringFlag{
					Name:  "succ",
					Usage: "successor of node",
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
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
