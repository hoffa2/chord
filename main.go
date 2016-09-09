package main

import (
	"log"
	"os"

	"github.com/hoffa2/chord/nameserver"

	"github.com/hoffa2/chord/client"

	"github.com/hoffa2/chord/node"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:  "node",
			Usage: "run node",
			Action: func(c *cli.Context) error {
				return node.Run(c.Args())
			},
		},
		{
			Name:  "client",
			Usage: "run client",
			Action: func(c *cli.Context) error {
				return client.Run(c.Args())
			},
		},
		{
			Name:  "nameserver",
			Usage: "run nameserver",
			Action: func(c *cli.Context) error {
				return nameserver.Run(c.Args())
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
