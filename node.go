package gomsg

import (
	"log"
	"time"
)

// IHandler node handler
type IHandler interface {
	OnOpen(s *Session)
	OnClose(s *Session)

	OnReq(s *Session, serial uint16, data []byte)
	OnPush(s *Session, data []byte) int16
}

// Node struct
type Node struct {
	addr         string
	handler      IHandler
	signal       chan int
	ReadCounter  chan int
	WriteCounter chan int
}

// NewNode make node ptr
func NewNode(addr string, h IHandler) *Node {
	return &Node{
		addr:         addr,
		handler:      h,
		ReadCounter:  make(chan int),
		WriteCounter: make(chan int),
		signal:       make(chan int),
	}
}

// Stop stop the IOCounter service
func (n *Node) Stop() {
	n.signal <- 1
}

// IOCounter io couter
func (n *Node) IOCounter() {
	defer Recover()

	reads := 0
	writes := 0
	duration := time.Second * 5
	t := time.NewTimer(duration)

	stop := false
	for !stop {
		select {
		case m := <-n.ReadCounter:
			reads += m
		case m := <-n.WriteCounter:
			writes += m
		case <-t.C:
			t.Reset(duration)
			log.Printf("Read : %d/s\tWrite : %d/s\n", reads/5, writes/5)
			reads, writes = 0, 0
		case <-n.signal:
			stop = true
		}
	}

	log.Println("IOCounter stoped.")
}
