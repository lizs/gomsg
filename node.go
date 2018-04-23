package gomsg

import (
	"log"
	"time"
)

// Callback request callback
type Callback func(*Result)

// internal with keep alive handling
type iInternalHandler interface {
	IHandler
	keepAlive()
}

// IHandler node handler
type IHandler interface {
	OnOpen(*Session)
	OnClose(*Session, bool)
	OnReq(*Session, []byte, Callback)
	OnPush(*Session, []byte) int16
}

// Node struct
type Node struct {
	addr              string
	handler           IHandler
	signal            chan int
	ReadCounter       chan int
	WriteCounter      chan int
	keepAliveInterval int
}

// NewNode make node ptr
func newNode(addr string, h IHandler, keepAliveInterval int) Node {
	return Node{
		addr:              addr,
		handler:           h,
		ReadCounter:       make(chan int),
		WriteCounter:      make(chan int),
		signal:            make(chan int),
		keepAliveInterval: keepAliveInterval,
	}
}

func (n *Node) keepAlive() {}

func (n *Node) OnOpen(*Session) {}

func (n *Node) OnClose(*Session, bool) {}

func (n *Node) OnReq(*Session, []byte, Callback) {}

func (n *Node) OnPush(*Session, []byte) int16 {
	return 0
}

// Stop stop the IOCounter service
func (n *Node) Stop() {
	n.signal <- 1
}

// IOCounter io couter
func (n *Node) ioCounter() {
	defer Recover()

	reads := 0
	writes := 0
	d1 := time.Second * 5
	t1 := time.NewTimer(d1)

	d2 := time.Duration(n.keepAliveInterval) * time.Second
	t2 := time.NewTimer(d2)

	stop := false
	for !stop {
		select {
		case m := <-n.ReadCounter:
			reads += m

		case m := <-n.WriteCounter:
			writes += m

		case <-t1.C:
			t1.Reset(d1)
			log.Printf("Read : %d/s\tWrite : %d/s\n", reads/5, writes/5)
			reads, writes = 0, 0

		case <-t2.C:
			t2.Reset(d2)
			n.keepAlive()

		case <-n.signal:
			stop = true
		}
	}

	log.Println("IOCounter stoped.")
}
