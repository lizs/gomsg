package gomsg

import (
	"log"
	"time"
)

// Callback request callback
type Callback func(*Result)

// internal with keep alive handling
type iinternal interface {
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
	addr            string
	internalHandler iinternal
	signal          chan int
	keepAliveSignal chan int
	ReadCounter     chan int
	WriteCounter    chan int
}

// NewNode make node ptr
func newNode(addr string, h iinternal) Node {
	return Node{
		addr:            addr,
		internalHandler: h,
		ReadCounter:     make(chan int),
		WriteCounter:    make(chan int),
		signal:          make(chan int),
		keepAliveSignal: make(chan int, 1),
	}
}

// Stop stop the IOCounter service
func (n *Node) Stop() {
	n.signal <- 1
	n.keepAliveSignal <- 1
}

// Start
func (n *Node) Start() {
	go n.ioCounter()
	go n.internalHandler.keepAlive()
}

// IOCounter io couter
func (n *Node) ioCounter() {
	defer Recover()

	log.Println("IOCounter running.")
	reads := 0
	writes := 0
	d := time.Second * 5
	t := time.NewTimer(d)

	stop := false
	for !stop {
		select {
		case m := <-n.ReadCounter:
			reads += m

		case m := <-n.WriteCounter:
			writes += m

		case <-t.C:
			t.Reset(d)
			log.Printf("Read : %d/s\tWrite : %d/s\n", reads/5, writes/5)
			reads, writes = 0, 0

		case <-n.signal:
			stop = true
		}
	}

	log.Println("IOCounter stoped.")
}
