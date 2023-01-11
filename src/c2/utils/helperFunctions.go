package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

var Verbose bool

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
			var session = Session{}
			err := json.Unmarshal(scanner.Bytes(), &session)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("[+] Session ID: %s | IsActive: %v\n", session.Session_UID, session.IsActive)
		}
	}

}
func DoesSessionExist(sessionID string) (bool, string, int) {
	data, err := os.ReadFile("sessions.json")
	if err != nil {
		fmt.Println(err)
	}

	ret_session := Session{}
	var s Session
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
func ExecuteFromCli(cmd string, sessionID string) {
	exists, ip, port := DoesSessionExist(sessionID)
	if exists {
		conn := Connect(ip, port, 3)
		fmt.Println(cmd)
		conn.Write([]byte(cmd))
		buf := make([]byte, (4086 * 2))
		n, err := conn.Read(buf)
		if err != nil {
			if Verbose {
				fmt.Println(err)
			}
		}
		fmt.Print(string(buf[:n]))
	}
}
func Connect(client string, port int, timeout int) net.Conn {
	target := fmt.Sprintf("%s:%d", client, port)
	dialer := &net.Dialer{
		Timeout: time.Duration(time.Second * time.Duration(timeout)),
	}
	conn, err := dialer.Dial("tcp", target)
	if err != nil {
		fmt.Println("Host is either down or disconnected")
		if Verbose {
			fmt.Printf("Error: %s\n", err.Error())
		}
	}
	return conn
}
