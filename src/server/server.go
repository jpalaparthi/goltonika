// Teltonika server package consists of all methods to the server.
package server

import (
	 "fmt"
	"log"
	"net"
	"io"
   	"strconv"
)

//Message Type defines Success Failure Information and Fatal message types which is a signal to the routine execution status
type MessageType int

const (
	Success MessageType = iota
	Error
	Data
	Info
	Failure
	Warning
	Fatal
)

//Server type contains of all server elements.
//Server remote address, listner, list of the clients connected.
//ClientInfo channel notifies when a client is connected or disconnected
//Message channel notifies if there is any success or failure message occures.This is very useful particularly to notify a message in unconditionally loops where method cannot yeilds return type
type Server struct {
	tcpAddr *net.TCPAddr
	Clients  []ClientInfo
	listner *net.TCPListener
	Client  chan ClientInfo
	Message chan MessageInfo
}

type ClientInfo struct {
	Conn        net.Conn
	InData      string
	OutData     string
	isNewClient bool
	StepData []string
}

//MessageInfo type is used to define a return message to the subscripber
//Methods can return error or success. Error or Success can be for ClientInfo or Server
//Message Info type gives appropriate messages based on the context.
//Message Type tells whether message is failure or success.
//For successful messages MSG is always Success
type MessageInfo struct {
	MSGType MessageType
	MSGWith string //ClientInfo or Server.
	Module  string // Which function or Method this is called.
	MSG     string
}

//HandleClient method ideally handles each connection as a concurrent go routine
func (s *Server) HandleClient(con net.Conn) {
	ci:=s.GetClientInfo(con)
	buf:= make([]byte, 1024)
	for {
		n, err := con.Read(buf)
		defer con.Close()
		if err != nil || n == 0 || err ==io.EOF {
			con.Close()
			s.removeClient(con)
			s.Client <- ClientInfo{Conn: con, isNewClient: false}
			break
		}
		ci.InData = ci.InData + string(buf[0:n])
		fmt.Println(ci.InData)
		s.Process(ci)
		//WriteTo(con,[]byte(con.RemoteAddr().String()))
		//fmt.Println(ci.InData)
	}
}

//Init method is called to initilize required server elements.Unless Init method called the subscriber of this package cannot move further
//Init method resolves passed address, makes a channel for the message and client notifications
func (s *Server) Init(addr string) MessageInfo {
	s.Message = make(chan MessageInfo)
	s.Client = make(chan ClientInfo)
	var m MessageInfo
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		m = MessageInfo{MSGType: 2, MSGWith: "Server", Module: "Init", MSG: err.Error()}
		s.Message <- m
	}
	if err == nil {
		s.tcpAddr = tcpAddr
		m = MessageInfo{MSGType: 1, MSGWith: "Server", Module: "Init", MSG: "Success"}
		s.Message <- m
	}
	return m
}

//Listen method is called to create listner based on the tcp address provided from the Init method
//Listen method returns MessageInfo on success or failure with details.
func (s *Server) Listen() MessageInfo {
	var err error
	var m MessageInfo
	s.listner, err = net.ListenTCP("tcp", s.tcpAddr)
	if err != nil {
		m = MessageInfo{MSGType: 2, MSGWith: "Server", Module: "Listen", MSG: err.Error()}
		s.Message <- m

	} else if err == nil {
		m = MessageInfo{MSGType: 1, MSGWith: "Server", Module: "Listen", MSG: "Success"}
		s.Message <- m
	}
	return m
}

//Listen&Accept method runs unconditionally, hence it is not apt to return
//On success or failuere MessageInfo channel is assigned to Server.Message channel
func (s *Server) ListenAndAccept() {
	for {
		con, err := s.listner.Accept()
		if err != nil {
			s.Message <- MessageInfo{MSGType: 2, MSGWith: "Server", Module: "ListenAndAccept", MSG: err.Error()}
		}

		if err == nil {
			go s.HandleClient(con)
			s.addClient(con)
			s.Client <- ClientInfo{Conn: con, isNewClient: true}
		}
	}
}

//addClient occures when ever a new connection is accepted by the listner
//This is a private method.Subscriber cannot add a clinet explicitely, hence called internally
//This method adds a client with a new connection to the server.
//All added clients are appended to the Clients element of the server object
func (s *Server) addClient(con net.Conn) {
	var exist bool = false
	var ci ClientInfo
	ci = ClientInfo{Conn: con}

	if len(s.Clients) == 0 {
		_ = append(s.Clients, ci)
	}
	for i := 0; i < len(s.Clients); i++ {
		if s.Clients[i].Conn == con {
			exist = true
			break
		}
	}
	if exist == false {
		s.Clients = append(s.Clients, ci)
	}
}

//removeClient occures when ever existing connect closes due to any reason.
//This is a private method.Subscriber cannot remove a client explicitely,hence called internally
func (s *Server) removeClient(con net.Conn) {
	for i := 0; i < len(s.Clients); i++ {
		if s.Clients[i].Conn == con {
			s.Clients = append(s.Clients[:i], s.Clients[i+1:]...)
		}
	}

}

//AddOrRemoveClient method has to work concurrently.
func (s *Server) AddOrRemoveClient(logger *log.Logger) {
	for c := range s.Client {

		if c.isNewClient == true {
			s.addClient(c.Conn)
		} else if c.isNewClient == false {
			s.removeClient(c.Conn)
		} else {
			//do nothing for the moment
		}
		s.LogActiveConnections(logger);
	}
}

func (s *Server) LogActiveConnections(logger *log.Logger) {
	for _, c := range s.Clients {
		logger.Println(c.Conn.RemoteAddr().String())
	}

}

//GetClientInfo method gives client object based on the connection provided
func (s *Server) GetClientInfo(con net.Conn) *ClientInfo {
	var ci *ClientInfo
	for i := 0; i < len(s.Clients); i++ {
		if s.Clients[i].Conn == con {
			ci = &s.Clients[i]
			break
		}
	}
 	return ci
}

//Process method takes incoming data from the client and processes based on the server protocol
//in case of any errors, returns error object else returns nil
func (s *Server) Process(c *ClientInfo)(error){
	var IMEIlen int
	var Datalen int
	if(len(c.InData)>=4 && IMEIlen==0){
		IMEIlen = hex2int(c.InData[0:4])
	}
	//fmt.Println(IMEIlen)
	if(len(c.InData) >= 4+(IMEIlen*2) && IMEIlen!=0){
		c.Conn.Write([]byte("01"))
		c.StepData =append(c.StepData,c.InData)
		c.InData=""
		return nil
	}
	if(len(c.InData)>=10 && Datalen==0){
		Datalen=hex2int(c.InData[8:16])
		fmt.Println("Data length is ",c.InData[8:16],":",Datalen)
	}
	if(len(c.InData)>= 8 + (Datalen*2) && Datalen!=0){
		c.Conn.Write([]byte(c.InData[18:20]))
		fmt.Println(len(c.InData))
	}
	//fmt.Println(len(c.InData))
	
	return nil
}

type Processor interface{
//	[]Data
	Process([]byte)(error)
}

func WriteTo(con net.Conn,data []byte){
	con.Write(data)
}

 func hex2int(hexStr string) int {
          // base 16 for hexadecimal
          result, _ := strconv.ParseInt(hexStr, 16, 64)
          return int(result)
  }