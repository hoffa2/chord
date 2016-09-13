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
	RunNodeCmd       = "go run main.go node --port=8000 --nameserver=%s --pre=%s --succ=%s"
	RunClientCmd     = "go run main.go client --port=8000 --nameserver=%s"
	RunNameServerCmd = "go run main.go nameserver --port=8000"
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
	if host == "localhost" {
		cmd = exec.Command(command)
	} else {
		cmd = exec.Command("ssh -f %s 'cd %s; %s'", host, cwd, command)
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
	var pre int
	var succ int

	hosts, err := getNodeList(numhosts)
	if err != nil {
		return err
	}

	ipToHost, h := hashValuesSorted(hosts)

	for i, host := range h {
		pre = (i - 1) % len(hosts)
		succ = (i + 1) % len(hosts)
		nodecmd := fmt.Sprintf(RunNodeCmd, nameserver, ipToHost[h[pre]], ipToHost[h[succ]])
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
	output, err := exec.Command("./rocks_list_hosts.sh %s", numhosts).Output()
	if err != nil {
		return []string{}, err
	}

	hosts := strings.Split(string(output), " ")
	return hosts, nil
}
