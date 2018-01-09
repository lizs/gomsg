package main

import (
	"flag"
	"time"

	"github.com/lizs/gomsg"
)

type handler struct {
}

func (h *handler) OnOpen(s *gomsg.Session) {
}

func (h *handler) OnClose(s *gomsg.Session) {
}

func (h *handler) OnReq(s *gomsg.Session, serial uint16, data []byte) {
	// simulate an async handle
	time.AfterFunc(time.Second*2, func() {
		gomsg.STA().Ret <- gomsg.NewRet(s, serial, &gomsg.Result{En: int16(gomsg.Success), Data: nil})
	})
}

func (h *handler) OnPush(s *gomsg.Session, data []byte) int16 {
	return int16(gomsg.Success)
}

func main() {
	host := flag.String("h", "localhost:6000", "specify the client/server host address.\n\tUsage: -h localhost:6000")
	runAsServer := flag.Bool("s", false, "whether to run as a tcp server.\n\tUsage : -s true/false")
	flag.Parse()

	// start STA service
	go gomsg.STA().Start()

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
