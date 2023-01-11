package main

import (
	"fmt"
	"net"
	"os"
	"scanner/src/agent/utils"
	"sync"
	"time"

	"github.com/akamensky/argparse"
)

type Config struct {
	PortsToScan []int
	Delay       int
	Timeout     int    `default:"3"`
	Target      string `default:"127.0.0.1"`
}

type InitConf struct {
	Host       string
	Port       int
	Connection net.Conn
}

type Process struct {
	Name string
	PID  int
	UID  int
}

type SessionIdentifier struct {
	OS         string
	Username   string
	OsVersion  string
	any1       any
	Hostname   string
	MACAddress []string
}

func main() {

	parser := argparse.NewParser("listener", "does stuff")
	client := parser.String("c", "client", &argparse.Options{Required: false, Help: "client address", Default: ""})
	port := parser.Int("p", "port", &argparse.Options{Required: false, Help: "port to connect to", Default: 9001})
	verbosity := parser.Flag("v", "verbose", &argparse.Options{Required: false, Help: "adds verbosity"})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}
	if *verbosity {
		utils.Verbose = true
	}
	Connect(client, port, 3)

}

func Connect(client *string, port *int, timeout int) {
	target := fmt.Sprintf(":%d", *port)
	utils.Conf.Host = *client
	utils.Conf.Port = *port
	dial := net.Dialer{
		Timeout: time.Duration(time.Second * time.Duration(timeout)),
	}
	conn, err := dial.Dial("tcp", target)
	utils.Conf.Connection = conn
	if err != nil {
		fmt.Println(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		utils.HandleConnection(conn)

		defer wg.Done()
	}()
	wg.Wait()

}
