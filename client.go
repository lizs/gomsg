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
	externalHandler  IHandler
}

func (c *Client) OnOpen(s *Session) {
	c.externalHandler.OnOpen(s)
}

func (c *Client) OnClose(s *Session, force bool) {
	c.externalHandler.OnClose(s, force)

	// reconnect
	if !force && c.autoRetryEnabled {
		c.Stop()
		c.Start()
	}
}

func (c *Client) OnReq(s *Session, data []byte, cb Callback) {
	c.externalHandler.OnReq(s, data, cb)
}

func (c *Client) OnPush(s *Session, data []byte) int16 {
	return c.externalHandler.OnPush(s, data)
}

// NewClient new tcp client
func NewClient(host string, h IHandler, autoRetry bool) *Client {
	ret := &Client{
		autoRetryEnabled: autoRetry,
		session:          nil,
		externalHandler:  h,
	}

	ret.node = NewNode(host, ret)
	return ret
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
		c.session.Close(true)
		c.session = nil
	}

	c.node.Stop()
	log.Println("client stopped.")
}
