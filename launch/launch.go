package launch

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

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

type cmdfunc func([]string) error

type SSHConn struct {
	inn     bytes.Buffer
	out     bytes.Buffer
	host    string
	process string
	cmd     *exec.Cmd
}

type NodeState struct {
	IP   string
	Next string
	Prev string
}

type Connection struct {
	conns      []*SSHConn
	freeNodes  []string
	nameserver string
	cwd        string
	logfile    *os.File
}

var (
	Blue  = "\x1b[34m"
	Red   = "\x1b[31m"
	Green = "\x1b[32m"
	White = "\x1b[0m"
)

type Console struct {
	commands map[string]cmdfunc
}

func (c *Connection) ListNodes(args []string) error {
	client := http.Client{Timeout: time.Duration(time.Second * 2)}
	var n NodeState
	for _, node := range c.conns {
		if node.host == c.nameserver {
			continue
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:8030/state/get", node.host), nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Could not get state from : %s", node.host)
		}

		err = json.NewDecoder(resp.Body).Decode(&n)
		if err != nil {
			return err
		}
		fmt.Printf("Node: "+Blue+"%s"+White+" ==> ("+Green+"\t%s "+Red+"%s"+White+")\n",
			n.IP, n.Prev, n.Next)
	}
	return nil
}

func (c *Console) RunConsole() {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(Blue + "--> " + White)
		input, _ := in.ReadString('\n')
		cmd := strings.Split(input, " \n")

		if len(cmd) == 0 {
			continue
		}
		f, ok := c.commands[cmd[0]]
		if !ok {
			fmt.Println("Command does not exist")
			continue
		}
		err := f(cmd[1:])
		if err != nil {
			log.Println(err)
		}
	}
}

func (c *Connection) AddNode(args []string) error {
	nodecmd := fmt.Sprintf(RunNodeCmd, c.nameserver)
	node := c.freeNodes[len(c.freeNodes)-1]
	c.freeNodes = c.freeNodes[:len(c.freeNodes)-1]
	return c.runSSHCommand(node, c.cwd, nodecmd)

}

func (c *Connection) killNode(args []string) error {
	return errors.New("Not implemented")
}

func (c *Connection) RunTests(args []string) error {
	return c.runCommand("chord", "client", fmt.Sprintf("--nameserver=%s", c.nameserver))
}

func InitConsole(conn *Connection) *Console {
	c := new(Console)
	c.commands = make(map[string]cmdfunc)
	c.commands["ls"] = conn.ListNodes
	c.commands["shutdown"] = conn.CloseConnections
	c.commands["add"] = conn.AddNode
	c.commands["leave"] = conn.killNode
	c.commands["test"] = conn.RunTests
	return c
}

func Run(c *cli.Context) error {
	nameserver := c.String("nameserver")

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	f, err := os.OpenFile("test.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	conns := new(Connection)

	// run nameserver
	err = conns.runSSHCommand(nameserver, cwd, RunNameServerCmd)
	if err != nil {
		return err
	}

	conns.logfile = f
	conns.nameserver = nameserver
	conns.cwd = cwd

	nodes, err := getNodeList("52")
	if err != nil {
		return err
	}

	conns.freeNodes = nodes
	cons := InitConsole(conns)
	go cons.RunConsole()

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT)

	<-ch
	conns.CloseConnections([]string{})
	return nil
}

func (c *Connection) CloseConnections(args []string) error {
	for _, conn := range c.conns {
		_, err := exec.Command("ssh", "-f", conn.host,
			fmt.Sprintf("pgrep -f \"%s\" | xargs kill -s SIGINT\n", conn.process)).Output()
		if err != nil {
			log.Println(err)
		}
		err = conn.cmd.Process.Kill()
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("Shutdown -> %s with err: %s\n", conn.host, err)
	}
	return nil
}

func (c *Connection) runCommand(command string, args ...string) error {

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (c *Connection) runSSHCommand(host, cwd, command string) error {
	cmd := exec.Command("ssh", "-f", host, fmt.Sprintf("cd %s; %s", cwd+"/..", command))
	logger := log.New(c.logfile, "\x1b[32m"+host+"\x1b[0m"+" --> ", 0)

	logStreamerOut := logstreamer.NewLogstreamer(logger, " ", false)
	logStreamerErr := logstreamer.NewLogstreamer(logger, "stderr", false)
	fmt.Printf("\x1b[32m"+"Started \x1b[0m"+" --> "+" %s\n", host)

	sshconn := &SSHConn{
		host:    host,
		cmd:     cmd,
		process: command,
	}
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
