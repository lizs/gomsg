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

	handler IServerHandler
}

// IServerHandler 服务器处理器
type IServerHandler interface {
	OnOpen(s *Session)
}

// NewServer 创建Tcp服务器
func NewServer(host string, h IServerHandler) *Server {
	return &Server{
		addr:     host,
		sessions: make(map[int32]*Session),
		seed:     0,
		handler:  h,
	}
}

// Start 服务器启动
func (s *Server) Start() {
	listener, err := net.Listen("tcp4", s.addr)
	if err != nil {
		log.Panicf("listen failed : %v", err)
	} else {
		log.Printf("server running on [%s]\n", s.addr)
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
	for _, s := range s.sessions {
		s.Close()
	}

	log.Println("server stopped.")
}

func handleConn(s *Server, conn net.Conn) {
	s.seed++
	session := &Session{ID: s.seed, conn: conn}
	s.sessions[s.seed] = session

	log.Printf("conn [%d] established.\n", s.seed)

	//go echo(session)
	if s.handler != nil {
		s.handler.OnOpen(session)
	}

	go session.scan()
}

func echo(s *Session) {
	input := bufio.NewScanner(s.conn)
	for input.Scan() {
		fmt.Println(s.conn, input.Text())
	}
}
