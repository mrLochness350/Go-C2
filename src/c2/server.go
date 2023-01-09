package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func handleConnection(conn net.Conn) {
	if conn != nil {
		shell_pointer := fmt.Sprintf("(%s)> ", conn.RemoteAddr().String())
		defer conn.Close()
		for {
			fmt.Print(shell_pointer)
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				fmt.Print(shell_pointer)
				line := scanner.Text()
				if line == "exit" || line == "quit" || line == "q" {
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
}

func StartListener(port int) {
	format := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", format)
	if err != nil {
		fmt.Println(err)
	}
	defer listener.Close()
	fmt.Printf("Starting listener on port %d\n", port)
	conns := make(map[net.Conn]struct{})
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}
		conns[conn] = struct{}{}
		go handleConnection(conn)
	}

}
