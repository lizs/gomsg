package gomsg

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

// Server 结构体
type Server struct {
	addr     string
	sessions map[int32]*Session
	seed     int32
}

// NewServer 创建Tcp服务器
func NewServer(host string) *Server {
	return &Server{
		addr:     host,
		sessions: make(map[int32]*Session),
		seed:     0,
	}
}

// Start 服务器启动
func (s *Server) Start() {
	listener, err := net.Listen("tcp4", s.addr)
	if err != nil {
		log.Panicf("listen failed : %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}

		handleConn(s, conn)
	}
}

// Stop 服务器停止
func (s *Server) Stop() {
}

func handleConn(s *Server, conn net.Conn) {
	s.seed++
	session := &Session{ID: s.seed, Conn: conn}
	s.sessions[s.seed] = session

	log.Printf("conn [%d] established.\n", s.seed)

	//go echo(session)
	go dispatch(session)
}

func dispatch(s *Session) {
	input := bufio.NewScanner(s.Conn)
	input.Split(s.Scan)

	for input.Scan() {
		log.Printf("conn : %d=> [%d] bytes\n", s.ID, len(input.Bytes()))
	}

	log.Printf("conn [%d] closed.\n", s.ID)
}

func echo(s *Session) {
	input := bufio.NewScanner(s.Conn)
	for input.Scan() {
		fmt.Println(s.Conn, input.Text())
	}
}
