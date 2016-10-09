package node

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/netutils"
	"github.com/hoffa2/chord/util"
	"github.com/urfave/cli"
)

type Logger struct {
	Err  *log.Logger
	Info *log.Logger
}

// Run Runs a chord node
func Run(c *cli.Context) error {
	port := c.String("port")
	if port == "" {
		port = "8030"
	}

	NameServerAddr := c.String("nameserver")
	graph := c.Int("graph")

	r := mux.NewRouter()
	n, err := os.Hostname()
	if err != nil {
		return err
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	http.DefaultTransport.(*http.Transport).IdleConnTimeout = time.Second * 2
	http.DefaultTransport.(*http.Transport).MaxIdleConns = 10000

	// CTRL-C HANDLER
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGKILL)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT)

	infolog := log.New(os.Stdout, "\x1b[32m"+n+"\x1b[0m"+" --> ", log.Lshortfile)
	errlog := log.New(os.Stderr, "\x1b[31m"+n+"\x1b[0m"+" --> ", log.Lshortfile)

	client := http.Client{
		Timeout: time.Duration(time.Second * 3),
	}
	node := &Node{
		nameServer: NameServerAddr,
		Rnode: &comm.Rnode{
			IP: n,
			ID: util.StringToID(util.HashValue(n)),
		},
		objectStore: make(map[string]string),
		conn:        client,
		fingers:     make([]FingerEntry, KeySize),
		log:         &Logger{Err: errlog, Info: infolog},
		exitChan:    make(chan string),
		graphIP:     "129.242.22.74:8080",
		graph:       graph != 0,
	}

	node.remote = netutils.NewRemote(node.failhandler)
	l, err := netutils.SetupRPCServer("8011", node)
	if err != nil {
		return err
	}

	defer l.Close()

	// CTRL-C
	go func() {
		<-ch
		l.Close()
		fmt.Println("KILLED")
		os.Exit(1)
	}()

	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered: ", r)
		}
		l.Close()
		os.Exit(1)
	}()

	err = JoinNetwork(node, n)
	if err != nil {
		log.Println(err)
		return err
	}

	// Registering the put and get methods
	r.HandleFunc("/{key}", node.getKey).Methods("GET")
	r.HandleFunc("/{key}", node.putKey).Methods("PUT")
	r.HandleFunc("/state/get", node.state).Methods("GET")

	// Used in sending errors from httplisten
	errchan := make(chan error)

	srv := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      r,
	}
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			errchan <- err
		}
	}()

	select {
	case err := <-errchan:
		l.Close()
		return err
	// Used to make a node leave the network
	case <-node.exitChan:
		node.leave()
		l.Close()
		node.log.Err.Println("GOT LEAVE MESSAGE!!!")
		node.log.Err.Println("AFTER EXIT!!!")
		os.Exit(0)

	}
	// Assertion
	panic("Reached end")
}
