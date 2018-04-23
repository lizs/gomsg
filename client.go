package gomsg

import (
	"log"
	"net"
	"time"
)

// Client struct
type Client struct {
	Node
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

// keep alive
func (c *Client) keepAlive() {
	c.session.Ping()
}

// NewClient new tcp client
func NewClient(host string, h IHandler, autoRetry bool) *Client {
	ret := &Client{
		autoRetryEnabled: autoRetry,
		session:          nil,
		externalHandler:  h,
	}
	ret.Node = newNode(host, &ret.Node, 1)

	return ret
}

// Start client startup
func (c *Client) Start() {
	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp4", c.Node.addr)
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
	go c.ioCounter()

	// make session
	c.session = newSession(0, conn, &c.Node)

	// notify
	c.Node.OnOpen(c.session)

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

	c.Node.Stop()
	log.Println("client stopped.")
}
