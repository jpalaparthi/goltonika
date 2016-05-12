//tcp server application
package main

import (
	"fileops"
	"fmt"
	"log"
	"os"
	"os/exec"
	"server"
	"time"
)

var logger *log.Logger

func main() {
	var addr string

	/*if len(os.Args) > 1 {
		addr = os.Args[1]
	} else {
		fmt.Println("IP address is not provided.\nPlease provide IP address and port(e.g. 10.12.43.345:6000) as commong line argument")
		return
	}*/
	addr, err := getIP()
	addr = addr + ":27000"
	if err != nil {
		panic(err)
	}

	s := &server.Server{}

	go OnMessage(s)

	s.Init(addr)

	s.Listen()

	go s.AddOrRemoveClient(getConLogger("ActiveClients.txt"))

	//Just uncomment the below code to check connection object incoming data
	//go ShowConnections(s)

	s.ListenAndAccept()

}

func OnMessage(s *server.Server) {
	for m := range s.Message {
		fmt.Println(m.MSG)
		switch m.MSGType {
		case 1: //Success
			{
				if m.MSGWith == "Client" {
					logMessage(logger, fileops.GetCombDate(time.Now())+"Client-Success-log.txt", m.Module, m.MSG)
				} else if m.MSGWith == "Server" {
					logMessage(logger, fileops.GetCombDate(time.Now())+"Server-Success-log.txt", m.Module, m.MSG)
				}
			}
		case 2: //Error
			{
				if m.MSGWith == "Client" {
					logMessage(logger, fileops.GetCombDate(time.Now())+"Client-Error-log.txt", m.Module, m.MSG)
				} else if m.MSGWith == "Server" {
					logMessage(logger, fileops.GetCombDate(time.Now())+"Server-Error-log.txt", m.Module, m.MSG)
				}
			}
		default:
			{
				// do nothing at this point of time.
			}

		}

	}
}

func logMessage(logger *log.Logger, filename, subject, Message string) {
	logfile, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	logger = log.New(logfile, subject, log.Lshortfile|log.LstdFlags)
	logger.Println(Message)
}

func getConLogger(filename string) *log.Logger {
	var conlogger *log.Logger
	file, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	file.Truncate(0) // this deletes the contents of the file every time and recreates the active clients data
	conlogger = log.New(file, "Acive Connecions: ", log.Lshortfile|log.LstdFlags)
	return conlogger
}

func ShowConnections(s *server.Server) {
	for {
		time.Sleep(time.Second * 10)
		fmt.Println("Displaying the List of the connected clients")
		for i := 0; i < len(s.Clients); i++ {
			fmt.Println(s.Clients[i].Conn.RemoteAddr().String())
			fmt.Println(s.Clients[i].InData)
			fmt.Println()

		}
	}
}

func getIP() (string, error) {
	cmd := "ip route get 8.8.8.8 | awk '{print $NF; exit}'"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	return string(out), err
}
