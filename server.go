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

// NewServer new tcp server
func NewServer(host string, h INodeHandler) *Server {
	return &Server{
		sessions: make(map[int32]*Session),
		seed:     0,
		node: &Node{
			addr:         host,
			handler:      h,
			readCounter:  make(chan int),
			writeCounter: make(chan int),
		},
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

func handleConn(s *Server, conn net.Conn) {
	s.seed++

	// make session
	session := &Session{
		ID:           s.seed,
		conn:         conn,
		readCounter:  s.node.readCounter,
		writeCounter: s.node.writeCounter,
	}
	s.sessions[s.seed] = session

	// notify
	if s.node.handler != nil {
		s.node.handler.OnOpen(session)
	}

	// io
	go session.scan()

	log.Printf("conn [%d] established.\n", session.ID)
}
