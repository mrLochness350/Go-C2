package utils

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Log struct {
	RelatedSession Session
	LogName        string
	LogLocation    string
	LogPointer     *os.File
}

func (log *Log) CreateLogFile(host string) error {
	formatted := strings.Replace(host, ".", "_", -1)
	logname := fmt.Sprintf("logs/%s.log", formatted)
	init_log_str := fmt.Sprintf("----------LOG BEGINNING: %s ----------\n", host)
	file, err := os.Create(logname)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("log file %q already exists", logname)
		}
		return err
	}
	//defer file.Close()
	log.LogName = logname
	log.LogPointer = file
	_, err = file.Write([]byte(init_log_str))
	//defer log.LogPointer.Close()
	return err
}

func (log *Log) WriteLog(cmd string) error {
	if log.LogPointer == nil {
		log.CreateLogFile(string(log.RelatedSession.Ip_Addr))
	}
	currTime := time.Now()
	formatted := currTime.Format("01-02-2006 15:04:05")
	cmdAppend := fmt.Sprintf("%s: %s\n", formatted, cmd)
	_, err := log.LogPointer.Write([]byte(cmdAppend))
	return err
}
