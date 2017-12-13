package main

import (
	"flag"

	"github.com/lizs/gomsg"
)

type handler struct {
}

func (h *handler) OnOpen(s *gomsg.Session) {
	s.SetHandler(h)
}

func (h *handler) OnReq(s *gomsg.Session, data []byte, ch chan *gomsg.Result) {
	ch <- &gomsg.Result{En: gomsg.Success, Data: nil}
}

func (h *handler) OnPush(s *gomsg.Session, data []byte) uint16 {
	return gomsg.Success
}

func main() {
	host := flag.String("h", "localhost:6000", "specify the client/server host address.\n\tUsage: -h localhost:6000")
	runAsServer := flag.Bool("s", false, "whether to run as a tcp server.\n\tUsage : -s true/false")
	flag.Parse()

	if *runAsServer {
		s := gomsg.NewServer(*host, &handler{})
		s.Start()
	} else {
		c := gomsg.NewClient(*host, &handler{}, true)
		c.Start()

		ch := make(chan int)
		<-ch
	}
}
