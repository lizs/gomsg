package gomsg

import (
	"log"
	"net"
)

// Server struct
type Server struct {
	Node
	sessions map[int32]*Session
	seed     int32
}

// OnOpen ...
func (s *Server) OnOpen(session *Session) {
	s.sessions[s.seed] = session
	log.Printf("conn [%d] established.\n", session.ID)

	s.handler.OnOpen(session)
}

// OnClose ...
func (s *Server) OnClose(session *Session, force bool) {
	delete(s.sessions, session.ID)
	s.handler.OnClose(session, force)
}

// OnReq ...
func (s *Server) OnReq(session *Session, data []byte, cb Callback) {
	s.handler.OnReq(session, data, cb)
}

// OnPush ...
func (s *Server) OnPush(session *Session, data []byte) int16 {
	return s.handler.OnPush(session, data)
}

// NewServer new tcp server
func NewServer(host string, h IHandler) *Server {
	return &Server{
		sessions: make(map[int32]*Session),
		seed:     0,
		Node:     newNode(host, h, 2),
	}
}

// keep alive
func (s *Server) keepAlive() {
	for _, session := range s.sessions {
		if session.elapsedSinceLastResponse() > 4 {
			session.Close(true)
		}
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
	go s.ioCounter()

	// accept incomming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}

		s.handleConn(conn)
	}
}

// Stop server shutdown
func (s *Server) Stop() {
	for _, s := range s.sessions {
		s.Close(true)
	}

	s.sessions = make(map[int32]*Session)
	s.Node.Stop()
	log.Println("server stopped.")
}

func (s *Server) handleConn(conn net.Conn) {
	s.seed++

	// make session
	session := newSession(s.seed, conn, &s.Node)

	// notify
	s.OnOpen(session)

	// io
	go session.scan()
}
