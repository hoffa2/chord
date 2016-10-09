package launch

import (
	"bufio"
	"bytes"
	"encoding/json"
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

	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/netutils"
	"github.com/hoffa2/chord/util"
	"github.com/urfave/cli"
)

var (
	RunNodeCmd       = "chord node --port 8030 --nameserver %s:8030"
	RunClientCmd     = []string{"chord", "client", "--port 8000"}
	RunNameServerCmd = "chord nameserver"
	ListHosts        = "rocks_list_hosts.sh"
	KILLED           = "KILLED"
	ALIVE            = "ALIVE"
)

type cmdfunc func([]string) error

type SSHConn struct {
	inn     bytes.Buffer
	out     bytes.Buffer
	host    string
	process string
	cmd     *exec.Cmd
	state   string
}

type NodeState struct {
	IP         string
	Next       string
	Prev       string
	Successors comm.Rnodes
}

type Connection struct {
	conns      []*SSHConn
	freeNodes  []string
	nameserver string
	cwd        string
	graph      int
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
	fmt.Printf("Nodes Running: %d\n", len(c.conns)-1)
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
			fmt.Printf(Blue+"%s  "+Red+"KILLED\n"+White, node.host)
			continue
		}

		err = json.NewDecoder(resp.Body).Decode(&n)
		if err != nil {
			return err
		}
		fmt.Printf("Node: "+Blue+"%s"+White+" ==> ("+Green+"\t%s "+Red+"%s"+White+")\n",
			n.IP, n.Prev, n.Next)
		fmt.Println("Successors:")
		for _, succ := range n.Successors {
			fmt.Printf("%s\n", succ.IP)
		}
	}
	return nil
}

func (c *Console) RunConsole() {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(Blue + "--> " + White)
		input, _ := in.ReadString('\n')
		cmd := strings.Split(strings.TrimSpace(input), " ")
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
	command := nodecmd
	if c.graph != 0 {
		command += fmt.Sprintf(" --graph=%d", c.graph)
	}
	return c.runSSHCommand(node, c.cwd, command)

}

func (c *Connection) killNode(args []string) error {
	idx, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	if idx >= len(c.conns) {
		return fmt.Errorf("Out of range")
	}

	killConnection(c.conns[idx])

	return nil
}

func (c *Connection) RunTests(args []string) error {
	fmt.Printf(Red + "Running Test: " + White)
	if len(args) < 1 {
		return fmt.Errorf("")
	}
	c.runCommand("chord", "client", fmt.Sprintf("--nameserver=%s", c.nameserver),
		fmt.Sprintf("--tests=%s", args[0]))
	return nil
}

func (c *Connection) leaveNode(args []string) error {
	noderpc, err := netutils.ConnectRPC(args[0])
	if err != nil {
		return err
	}
	err = noderpc.Call("NodeComm.Leave", &comm.Empty{}, &comm.Empty{})
	if err != nil {
		return err
	}
	for _, conn := range c.conns {
		if conn.host == args[0] {
			conn.state = KILLED
		}
	}
	return nil
}

func InitConsole(conn *Connection) *Console {
	c := new(Console)
	c.commands = make(map[string]cmdfunc)
	c.commands["ls"] = conn.ListNodes
	c.commands["shutdown"] = conn.CloseConnections
	c.commands["add"] = conn.AddNode
	c.commands["kill"] = conn.killNode
	c.commands["leave"] = conn.leaveNode
	c.commands["test"] = conn.RunTests
	return c
}

func Run(c *cli.Context) error {
	nameserver := c.String("nameserver")
	graph := c.Int("graph")
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

	nodes, err := getNodeList("40")
	if err != nil {
		return err
	}

	conns.freeNodes = nodes
	conns.graph = graph
	cons := InitConsole(conns)
	go cons.RunConsole()

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT)

	defer func() {
		if r := recover(); r != nil {
			conns.CloseConnections([]string{})
		}
	}()

	<-ch
	conns.CloseConnections([]string{})
	return nil
}

func killConnection(s *SSHConn) {
	_, err := exec.Command("ssh", "-f", s.host,
		fmt.Sprintf("pgrep -f \"%s\" | xargs kill -s SIGINT\n", s.process)).Output()
	if err != nil {
		log.Println(err)
	}
	err = s.cmd.Process.Kill()
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("Shutdown -> %s with err: %s\n", s.host, err)
	s.state = KILLED
}

func (c *Connection) CloseConnections(args []string) error {
	for _, conn := range c.conns {
		killConnection(conn)
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
	cmd := exec.Command("ssh", "-f", host, fmt.Sprintf("cd %s; %s; ulimit -n 2000", cwd+"/..", command))
	fmt.Printf("\x1b[32m"+"Started \x1b[0m"+" --> "+" %s\n", host)

	sshconn := &SSHConn{
		host:    host,
		cmd:     cmd,
		process: command,
	}
	cmd.Stdout = c.logfile
	cmd.Stdin = &sshconn.inn
	cmd.Stderr = c.logfile
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
