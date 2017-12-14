package gomsg

import (
	"log"
	"net"
)

// Client struct
type Client struct {
	node             *Node
	session          *Session
	autoRetryEnabled bool
	sta              *STAService
}

// NewClient new tcp client
func NewClient(host string, h INodeHandler, autoRetry bool) *Client {
	return &Client{
		autoRetryEnabled: autoRetry,
		session:          nil,
		node: &Node{
			addr:         host,
			handler:      h,
			readCounter:  make(chan int),
			writeCounter: make(chan int),
		},
	}
}

// Start client startup
func (c *Client) Start() {
	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp4", c.node.addr)
		if err != nil {
			log.Printf("connect failed : %v\n", err)

			if c.autoRetryEnabled {
				continue
			}

			return
		}

		break
	}

	// io counter
	go c.node.IOCounter()

	// make session
	c.session = &Session{
		ID:           0,
		conn:         conn,
		readCounter:  c.node.readCounter,
		writeCounter: c.node.writeCounter,
	}

	// notify
	if c.node.handler != nil {
		c.node.handler.OnOpen(c.session)
	}

	// io
	go c.session.scan()

	log.Printf("conn [%d] established.\n", c.session.ID)
}

// Stop client shutdown
func (c *Client) Stop() {
	if c.session != nil {
		c.session.Close()
		c.session = nil
	}

	log.Println("client stopped.")
}
