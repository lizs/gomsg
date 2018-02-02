package gomsg

import (
	"log"
	"net"
)

// Server struct
type Server struct {
	node     *Node
	sessions map[int32]*Session
	seed     int32
}

// OnOpen ...
func (s *Server) OnOpen(session *Session) {
	s.sessions[s.seed] = session
	log.Printf("conn [%d] established.\n", session.ID)

	s.node.handler.OnOpen(session)
}

// OnClose ...
func (s *Server) OnClose(session *Session, force bool) {
	delete(s.sessions, session.ID)
	s.node.handler.OnClose(session, force)
}

// OnReq ...
func (s *Server) OnReq(session *Session, data []byte, cb Callback) {
	s.node.handler.OnReq(session, data, cb)
}

// OnPush ...
func (s *Server) OnPush(session *Session, data []byte) int16 {
	return s.node.handler.OnPush(session, data)
}

// NewServer new tcp server
func NewServer(host string, h IHandler) *Server {
	return &Server{
		sessions: make(map[int32]*Session),
		seed:     0,
		node:     NewNode(host, h),
	}
}

// Start server startup
func (s *Server) Start() {
	listener, err := net.Listen("tcp4", s.node.addr)
	if err != nil {
		log.Panicf("listen failed : %v", err)
	} else {
		log.Printf("server running on [%s]\n", s.node.addr)
	}

	defer listener.Close()

	// io counter
	go s.node.IOCounter()

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
	s.node.Stop()
	log.Println("server stopped.")
}

func (s *Server) handleConn(conn net.Conn) {
	s.seed++

	// make session
	session := NewSession(s.seed, conn, s.node)

	// notify
	s.OnOpen(session)

	// io
	go session.scan()
}
