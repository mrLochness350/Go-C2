package utils

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/takama/daemon"
)

const (
	name        = "Agent"
	description = "AAA"
	port        = ":6056"
)

var stdlog, errlog *log.Logger

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "ERR: ", log.Ldate|log.Ltime)
}

func accept(listener net.Listener, listen chan<- net.Conn) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		listen <- conn
	}
}

func handleConn(conn net.Conn) {
	buf := make([]byte, 8192)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
		}
		// resp, err := bufio.NewReader(conn).ReadString('\n')
		// if err != nil {
		// 	fmt.Println(err)
		// 	break
		// }
		cmd := string(buf)
		// fmt.Println(cmd)
		if cmd == "stop" {
			break
		}
		//go HandleConnection(conn)
	}

}

type Service struct {
	daemon.Daemon
}

func SendLifeNotif() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go func() {
				conn, err := net.Dial("tcp", Conf.Connection.RemoteAddr().String())
				if err != nil {
					fmt.Println(err)
					return
				}
				defer conn.Close()
				fmt.Fprintln(conn, "RDY")
			}()
		}
	}
}

func (service *Service) DaemonManage(cmd string) (string, error) {
	usage := ""
	if len(cmd) > 1 {
		command := cmd
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return "Error binding", err
	}
	listen := make(chan net.Conn, 100)
	go accept(listener, listen)
	for {
		select {
		case conn := <-listen:
			go handleConn(conn)
		case killSignal := <-interrupt:
			listener.Close()
			if killSignal == os.Interrupt {
				return "", err
			}
			return "", err
		}
	}
}

func StartDaemon(cmd string) {
	srv, err := daemon.New(name, description, daemon.SystemDaemon)
	if err != nil {

		os.Exit(0)
	}
	service := &Service{srv}
	status, err := service.DaemonManage(cmd)
	if err != nil {
		errlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}
}
