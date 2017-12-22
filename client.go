package gomsg

import (
	"log"
	"net"
	"time"
)

// Client struct
type Client struct {
	node             *Node
	session          *Session
	autoRetryEnabled bool
	sta              *STAService
}

// NewClient new tcp client
func NewClient(host string, h IHandler, autoRetry bool) *Client {
	return &Client{
		autoRetryEnabled: autoRetry,
		session:          nil,
		node:             NewNode(host, h),
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
				select {
				case <-time.After(time.Second * 2):
					log.Printf("reconnecting ...")
				}

				continue
			}

			log.Println("you can set `autoRetryEnabled` true to do auto reconnect stuff.")
			return
		}

		break
	}

	// io counter
	go c.node.IOCounter()

	// make session
	c.session = NewSession(0, conn, c.node)

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

		if c.node.handler != nil {
			c.node.handler.OnClose(c.session)
		}

		c.session = nil
	}

	c.node.Stop()
	log.Println("client stopped.")
}
