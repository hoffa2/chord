package launch

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/goware/prefixer"
	"github.com/hoffa2/chord/util"
	"github.com/kvz/logstreamer"
	"github.com/urfave/cli"
)

var (
	RunNodeCmd       = "chord node --port 8030 --nameserver %s:8030"
	RunClientCmd     = []string{"chord", "client", "--port 8000"}
	RunNameServerCmd = "chord nameserver"
	ListHosts        = "rocks_list_hosts.sh"
)

type SSHConn struct {
	inn     bytes.Buffer
	out     bytes.Buffer
	host    string
	process string
	cmd     *exec.Cmd
}

type Connection struct {
	conns []*SSHConn
}

func Run(c *cli.Context) error {
	numhosts := c.String("hosts")
	nameserver := c.String("nameserver")

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	conns := new(Connection)
	// run nameserver
	err = conns.runSSHCommand(nameserver, cwd, RunNameServerCmd)
	if err != nil {
		return err
	}

	err = conns.runNodes(cwd, numhosts, nameserver)
	if err != nil {
		return err
	}

	conns.runCommand("chord", "client", fmt.Sprintf("--nameserver=%s", nameserver))
	if err != nil {
		return err
	}

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT)

	// wait for termination signal
	<-ch
	conns.CloseConnections()

	return nil
}

func (c *Connection) CloseConnections() {
	for _, conn := range c.conns {
		if conn.host != "localhost" {
			_, err := exec.Command("ssh", "-f", conn.host, fmt.Sprintf("pgrep -f \"%s\" | xargs kill -s SIGINT\n", conn.process)).Output()
			if err != nil {
				log.Println(err)
			}
		}
		err := conn.cmd.Process.Kill()
		if err != nil {
			log.Println(err)
		}
	}
}

func (c *Connection) runCommand(command string, args ...string) error {

	cmd := exec.Command(command, args...)

	sshconn := &SSHConn{
		host:    "localhost",
		cmd:     cmd,
		process: command,
	}

	prefixReader := prefixer.New(&sshconn.out, "Localhost")
	cmd.Stdout = &sshconn.out
	prefixReader.WriteTo(os.Stdout)
	cmd.Stdin = &sshconn.inn

	c.conns = append(c.conns, sshconn)
	err := cmd.Start()
	if err != nil {
		return err
	}
	return nil
}

func (c *Connection) runSSHCommand(host, cwd, command string) error {

	log.Printf("Running %s on %s\n", command, host)
	cmd := exec.Command("ssh", "-f", host, fmt.Sprintf("cd %s; %s", cwd+"/..", command))

	logger := log.New(os.Stdout, host+" -->", log.Ldate|log.Ltime)

	logStreamerOut := logstreamer.NewLogstreamer(logger, "stdout", false)
	logStreamerErr := logstreamer.NewLogstreamer(logger, "stderr", false)

	sshconn := &SSHConn{
		host:    host,
		cmd:     cmd,
		process: command,
	}
	fmt.Println(command)
	cmd.Stdout = logStreamerOut
	cmd.Stdin = &sshconn.inn
	cmd.Stderr = logStreamerErr

	c.conns = append(c.conns, sshconn)
	err := cmd.Start()
	if err != nil {
		return err
	}

	return nil
}

func (c *Connection) runNodes(cwd, numhosts, nameserver string) error {
	hosts, err := getNodeList(numhosts)
	if err != nil {
		return err
	}

	nodecmd := fmt.Sprintf(RunNodeCmd, nameserver)
	log.Printf("Running %s nodes\n", numhosts)
	for _, host := range hosts {
		time.Sleep(1 * time.Second)
		err = c.runSSHCommand(host, cwd, nodecmd)
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

	end, err := strconv.Atoi(numhosts)
	if err != nil {
		return nil, err
	}
	hosts := strings.Split(string(output), " ")

	if end >= len(hosts) {
		return nil, fmt.Errorf("Something went wrong when converting numhosts")
	}

	return hosts[:end], nil
}
