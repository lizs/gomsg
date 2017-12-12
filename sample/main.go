package main

import (
	"github.com/lizs/gomsg"
)

type handler struct {
}

func (h *handler) OnOpen(s *gomsg.Session) {
	s.SetHandler(h)
}

func (h *handler) OnReq(s *gomsg.Session, data []byte) *gomsg.Result {
	return &gomsg.Result{En: gomsg.Success, Data: nil}
}

func (h *handler) OnPush(s *gomsg.Session, data []byte) uint16 {
	return gomsg.Success
}

func main() {
	s := gomsg.NewServer(":6000", &handler{})
	s.Start()
}
