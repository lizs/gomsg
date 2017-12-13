package gomsg

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

// Server struct
type Server struct {
	addr         string
	sessions     map[int32]*Session
	seed         int32
	handler      IServerHandler
	readCounter  chan int
	writeCounter chan int
}

// IServerHandler connection handler
type IServerHandler interface {
	OnOpen(s *Session)
}

// NewServer new tcp server
func NewServer(host string, h IServerHandler) *Server {
	return &Server{
		addr:         host,
		sessions:     make(map[int32]*Session),
		seed:         0,
		handler:      h,
		readCounter:  make(chan int),
		writeCounter: make(chan int),
	}
}

// Start server startup
func (s *Server) Start() {
	listener, err := net.Listen("tcp4", s.addr)
	if err != nil {
		log.Panicf("listen failed : %v", err)
	} else {
		log.Printf("server running on [%s]\n", s.addr)
	}

	defer listener.Close()

	// io counter
	go s.io()

	// accept incomming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}

		handleConn(s, conn)
	}
}

// Stop server shutdown
func (s *Server) Stop() {
	for _, s := range s.sessions {
		s.Close()
	}

	log.Println("server stopped.")
}

func (s *Server) io() {
	reads := 0
	writes := 0
	duration := time.Second * 5
	t := time.NewTimer(duration)
	for {
		select {
		case m := <-s.readCounter:
			reads += m
		case m := <-s.writeCounter:
			writes += m
		case <-t.C:
			t.Reset(duration)
			log.Printf("Read : %d/s\tWrite : %d/s\n", reads/5, writes/5)
			reads, writes = 0, 0
		}
	}
}

func handleConn(s *Server, conn net.Conn) {
	s.seed++
	session := &Session{ID: s.seed, conn: conn, readCounter: s.readCounter, writeCounter: s.writeCounter}
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
