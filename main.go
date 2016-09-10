package main

import (
	"log"
	"os"
	"github.com/hoffa2/chord/nameserver"
	"github.com/hoffa2/chord/client"
	"github.com/hoffa2/chord/launch"
	"github.com/hoffa2/chord/node"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Key-Value Store"
	app.Usage = "Run one of the components"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "nameserver, ns",
			Usage: "Specify ip of nameserver",
		},
		cli.StringFlag{
			Name: "nodes",
			Usage: "Specify a list of node seperated by space",
		},
		cli.StringFlag{
			Name: "run-tests",
			Usage: "run all tests",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "node",
			Usage: "run node",
			Action: func(c *cli.Context) error {
				return node.Run(c)
			},
		},
		{
			Name:  "client",
			Usage: "run client",
			Action: func(c *cli.Context) error {
				return client.Run(c)
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
			Name: "RunAll",
			Usage: "Run all components together",
			Action: func(c *cli.Context) error {
				return launch.Run(c)
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
