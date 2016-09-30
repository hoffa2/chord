package node

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/netutils"
	"github.com/hoffa2/chord/util"
	"github.com/urfave/cli"
)

type Logger struct {
	err  *log.Logger
	info *log.Logger
}

// Run Runs a chord node
func Run(c *cli.Context) error {
	port := c.String("port")
	if port == "" {
		port = "8030"
	}
	NameServerAddr := c.String("nameserver")

	r := mux.NewRouter()
	n, err := os.Hostname()
	if err != nil {
		return err
	}

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGKILL)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT)

	infolog := log.New(os.Stdout, "\x1b[32m"+n+"\x1b[0m"+" --> ", log.Lshortfile)
	errlog := log.New(os.Stderr, "\x1b[31m"+n+"\x1b[0m"+" --> ", log.Lshortfile)

	client := http.Client{
		Timeout: time.Duration(time.Second * 2),
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
		log:         &Logger{err: errlog, info: infolog},
		remote:      netutils.NewRemote(),
	}
	l, err := netutils.SetupRPCServer("8011", node)
	if err != nil {
		return err
	}
	defer l.Close()
	go func() {
		<-ch
		l.Close()
		fmt.Println("KILLED")
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
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.info.Printf(message, args)
}

func (l *Logger) Error(message string, args ...interface{}) {
	l.err.Printf(message, args)
}
