package gomsg

import (
	"log"
	"net"
	"time"
)

// Server struct
type Server struct {
	Node
	sessions map[int32]*Session
	seed     int32
	handler  IHandler
}

// Count return sessions'count
func (s *Server) Count() int {
	if s.sessions != nil {
		return len(s.sessions)
	}

	return 0
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
	defer Recover()

	s.handler.OnReq(session, data, cb)
}

// OnPush ...
func (s *Server) OnPush(session *Session, data []byte) int16 {
	return s.handler.OnPush(session, data)
}

// NewServer new tcp server
func NewServer(host string, h IHandler) *Server {
	ret := &Server{
		sessions: make(map[int32]*Session),
		seed:     0,
		handler:  h,
	}

	ret.Node = newNode(host, ret)
	return ret
}

// keep alive
func (s *Server) keepAlive() {
	defer Recover()

	d := time.Second * 5
	t := time.NewTimer(d)

	stop := false
	for !stop {
		select {
		case <-t.C:
			t.Reset(d)
			for _, session := range s.sessions {
				if session.elapsedSinceLastResponse() > 40 {
					session.Ping()
				} else if session.elapsedSinceLastResponse() > 60 {
					session.Close(true)
				}
			}

		case <-s.keepAliveSignal:
			stop = true
		}
	}

	log.Println("Keep alive stopped.")
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

	// base.start
	s.Node.Start()

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
	s.keepAliveSignal <- 1
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
