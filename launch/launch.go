package launch

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/hoffa2/chord/util"
	"github.com/urfave/cli"
)

const (
	RunNodeCmd       = "go run main.go node --port=8000 --nameserver=%s"
	RunClientCmd     = "go run main.go client --port=8000 --nameserver=%s"
	RunNameServerCmd = "go run main.go nameserver --port=8000"
	ListHosts        = "rocks_list_hosts.sh"
)

func Run(c *cli.Context) error {
	numhosts := c.String("hosts")
	nameserver := c.String("nameserver")

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// run nameserver
	err = runCommand(nameserver, cwd, RunNameServerCmd)
	if err != nil {
		return err
	}

	runNodes(cwd, numhosts, nameserver)

	err = runCommand("localhost", cwd, fmt.Sprintf(RunClientCmd, nameserver))
	if err != nil {
		return err
	}

	return nil
}

func runCommand(host, cwd, command string) error {
	cmd := &exec.Cmd{}
	var finalcmd string
	if host == "localhost" {
		cmd = exec.Command(command)
	} else {
		finalcmd = fmt.Sprintf("ssh -f %s 'cd %s; %s'", host, cwd, command)
		cmd = exec.Command(finalcmd)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}
	return nil
}

func runNodes(cwd, numhosts, nameserver string) error {
	hosts, err := getNodeList(numhosts)
	if err != nil {
		return err
	}

	nodecmd := fmt.Sprintf(RunNodeCmd, nameserver)

	for _, host := range hosts {
		err = runCommand(host, cwd, nodecmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func hashValuesSorted(vals []string) (map[string]string, []string) {
	var ips []string
	m := make(map[string]string)

	for _, val := range vals {
		ips = append(ips, util.HashValue(val))
	}

	for i, val := range ips {
		m[val] = vals[i]
	}
	sort.Strings(ips)

	return m, ips
}

func getNodeList(numhosts string) ([]string, error) {
	// Getting a list of uvrocks hosts
	output, err := exec.Command("sh", ListHosts, numhosts).Output()
	if err != nil {
		return nil, err
	}

	hosts := strings.Split(string(output), " ")
	return hosts, nil
}
