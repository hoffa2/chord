package launch

import (
	"github.com/urfave/cli"
	"os/exec"
	"fmt"
	"os"
	"strings"
)

const (
	RunNodeCmd = "go run main.go node --port=8000 --nameserver=%s"
	RunClientCmd = "go run main.go client --port=8000 --nameserver=%s"
	RunNameServerCmd = "go run main.go nameserver --port=8000"
)


func Run(c *cli.Context) error {
	numhosts := c.String("hosts")
	nameserver := c.String("nameserver")

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// run nameserver
	runCommand(nameserver, cwd, RunNameServerCmd)

	runNodes(cwd, numhosts, nameserver)

	runCommand("localhost", cwd, fmt.Sprintf(RunClientCmd, nameserver))

	return nil
}

func runCommand(host, cwd, command string) {
	cmd := &exec.Cmd{}
	if  host == "localhost" {
		cmd = exec.Command(command)
	} else {
		cmd = exec.Command("ssh -f %s 'cd %s; %s'", host, cwd, command)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		panic(err)
	}
}

func runNodes(cwd, numhosts, nameserver string) {

	// Getting a list of uvrocks hosts
	output, err := exec.Command("./rocks_list_hosts %s", numhosts).Output()
	if err != nil {
		panic(err)
	}

	nodecmd := fmt.Sprintf(RunNodeCmd, nameserver)

	hosts := strings.Split(string(output), " ")

	for _, host := range hosts {
		runCommand(host, cwd, nodecmd)
	}
}
