package gomsg

import (
	"log"
	"time"
)

// INodeHandler node handler
type INodeHandler interface {
	OnOpen(s *Session)
}

// Node struct
type Node struct {
	addr         string
	handler      INodeHandler
	readCounter  chan int
	writeCounter chan int
}

// IOCounter io couter
func (n *Node) IOCounter() {
	reads := 0
	writes := 0
	duration := time.Second * 5
	t := time.NewTimer(duration)
	for {
		select {
		case m := <-n.readCounter:
			reads += m
		case m := <-n.writeCounter:
			writes += m
		case <-t.C:
			t.Reset(duration)
			log.Printf("Read : %d/s\tWrite : %d/s\n", reads/5, writes/5)
			reads, writes = 0, 0
		}
	}
}
