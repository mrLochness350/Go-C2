package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"os"
	utils "scanner/src/c2/utils"
	"time"

	"github.com/akamensky/argparse"
)

var verbose bool

func GenerateSessionID() string {
	var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	id := make([]byte, 5)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}
	for i := range id {
		id[i] = byte(runes[int(id[i])%len(runes)])
	}
	return string(id)
}

var sessions []utils.Session

func WriteConnections(sessions []utils.Session) {
	file, err := os.Create("sessions.json")
	if err != nil {
		fmt.Println(err)
	}
	for _, session := range sessions {
		ser, err := json.Marshal(session)
		if err != nil {
			fmt.Println(err)
		}
		file.Write(ser)
	}
	defer file.Close()
}

func main() {
	parser := argparse.NewParser("client", "does stuff")
	client := parser.String("c", "client", &argparse.Options{Required: false, Help: "client address (default 127.0.0.1)", Default: "127.0.0.1"})
	port := parser.Int("p", "port", &argparse.Options{Required: false, Help: "port to connect to", Default: 9001})
	version := parser.Flag("V", "version", &argparse.Options{Required: false, Help: "prints the version"})
	verbosity := parser.Flag("v", "verbose", &argparse.Options{Required: false, Help: "adds verbosity"})
	timeout := parser.Int("t", "timeout", &argparse.Options{Required: false, Help: "sets the timeout for disconnection", Default: 3})
	session := parser.String("s", "session", &argparse.Options{Required: false, Help: "session to execute commands on"})
	cmd := parser.String("", "command", &argparse.Options{Required: false, Help: "command to execute on session"})
	list := parser.Flag("l", "list", &argparse.Options{Required: false, Help: "list all active sessions"})
	interactive := parser.Flag("i", "interactive", &argparse.Options{Required: false, Help: "list all active sessions"})

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
	if *interactive {
		CreateSession(client, port, timeout)
	}

	if *session != "" {
		ExecuteFromCli(*cmd, *session)
	}
	if *list {
		ListSessions()
	}
	//defer WriteConnections(sessions)
}

func CreateSession(client *string, port *int, timeout *int) {
	conn := Connect(*client, *port, *timeout)
	logger := utils.Log{}
	sesh := &utils.Session{}
	sesh.Ip_Addr = *client
	sesh.Port = *port
	sesh.Session_UID = GenerateSessionID()
	logger.RelatedSession = *sesh
	logger.CreateLogFile(sesh.Ip_Addr)
	sessions = append(sessions, *sesh)
	ser, err1 := json.Marshal(sesh)
	if err1 != nil {
		fmt.Println(err1)
	}
	err := os.WriteFile("sessions.json", ser, fs.ModeAppend)
	if err != nil {
		fmt.Println(err)
	}
	HandleConnection(conn, &logger)
}
func Connect(client string, port int, timeout int) net.Conn {
	target := fmt.Sprintf("%s:%d", client, port)
	dialer := &net.Dialer{
		Timeout: time.Duration(time.Second * time.Duration(timeout)),
	}
	conn, err := dialer.Dial("tcp", target)
	if err != nil {
		fmt.Println("Host is either down or disconnected")
		if verbose {
			fmt.Printf("Error: %s\n", err.Error())
		}
	}
	return conn
}

func DoesSessionExist(sessionID string) (bool, string, int) {
	data, err := os.ReadFile("sessions.json")
	if err != nil {
		fmt.Println(err)
	}

	ret_session := utils.Session{}
	var s utils.Session
	err = json.Unmarshal(data, &s)
	if err != nil {
		fmt.Println(err)
	}
	if s.Session_UID == sessionID {
		ret_session = s
		return true, ret_session.Ip_Addr, ret_session.Port
	} else {
		return false, "", 0
	}
}

func ListSessions() {
	file, err := os.Open("sessions.json")
	if err != nil {
		fmt.Println(err)
	}
	fstats, err := os.Stat("sessions.json")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if fstats.Size() == 0 {
		fmt.Println("[-] No sessions found")
	} else {
		for scanner.Scan() {
			var session = utils.Session{}
			err := json.Unmarshal(scanner.Bytes(), &session)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("[+] Session ID: %s | IsActive: %v\n", session.Session_UID, session.IsActive)
		}
	}

}

func ExecuteFromCli(cmd string, sessionID string) {
	exists, ip, port := DoesSessionExist(sessionID)
	if exists {
		conn := Connect(ip, port, 3)
		fmt.Println(cmd)
		conn.Write([]byte(cmd))
		buf := make([]byte, (4086 * 2))
		n, err := conn.Read(buf)
		if err != nil {
			if verbose {
				fmt.Println(err)
			}
		}
		fmt.Print(string(buf[:n]))
	}
}

func HandleConnection(conn net.Conn, logger *utils.Log) {
	verbose := true
	if conn != nil {
		shell_pointer := fmt.Sprintf("(%s)> ", conn.RemoteAddr().String())
		defer conn.Close()
		for {
			fmt.Print(shell_pointer)
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				fmt.Print(shell_pointer)
				line := scanner.Text()
				err := logger.WriteLog(line)
				if err != nil {
					fmt.Println(err)
				}
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

}
